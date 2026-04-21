package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		println("Please provide the command to run")
		return
	}
	command := args[1]
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("APP_ENVIRONMENT") == "" || os.Getenv("APP_ENVIRONMENT") == "development" {
			log.Fatalf("Error loading .env file: %v, APP_ENVIRONMENT: %v", err, os.Getenv("APP_ENVIRONMENT"))
		}
	}

	config, err := config.New()
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	zlog, err := utils.NewZapLogger(config.App.Environment, nil)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	zsLog := zlog.Sugar()
	zsLog.Infof("Running command: %s", command)
	conf := &firebase.Config{ProjectID: config.GCP.ProjectID, StorageBucket: config.GCP.ProjectID + ".firebasestorage.app"}
	firebaseClient, err := firebase.NewApp(context.Background(), conf)
	if err != nil {
		zsLog.Fatalf("Failed to create Firebase app: " + err.Error())
	}
	firestoreClient, err := firebaseClient.Firestore(context.Background())
	if err != nil {
		zsLog.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()
	meiliClient := meilisearch.New(config.Meili.URL, meilisearch.WithAPIKey(config.Meili.APIKey))
	switch command {
	case "message":
		syncMessage(zsLog, firestoreClient, meiliClient)
	default:
		println("Unknown command: ", command)
	}
}

func syncMessage(zsLog *zap.SugaredLogger, firestoreClient *firestore.Client, meiliClient meilisearch.ServiceManager) {
	pkName := "id"
	task, err := meiliClient.Index("messages").DeleteAllDocuments(nil)
	if err != nil {
		zsLog.Fatalf("Failed to delete all documents from meili index: %v", err)
	}
	// check if task is completed
	for {
		task, err := meiliClient.Index("messages").GetTask(task.TaskUID)
		if err != nil {
			zsLog.Fatalf("Failed to get task status: %v", err)
		}
		if task.Status == "succeeded" {
			break
		} else if task.Status == "failed" {
			zsLog.Fatalf("Failed to delete all documents from meili index: %v", task.Error)
		} else {
			zsLog.Infof("Waiting for task to complete")
		}
		time.Sleep(1 * time.Second)
	}
	ctx := context.Background()
	docChat := firestoreClient.Collection("chats").Documents(ctx)
	for {
		doc, err := docChat.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			zsLog.Fatalf("Failed to iterate chat documents: %v", err)
		}
		chatID := doc.Ref.ID
		docMessage := doc.Ref.Collection("messages").Documents(ctx)
		var messages []map[string]any
		for {
			docMsg, err := docMessage.Next()
			if err != nil {
				if err == iterator.Done {
					break
				}
				zsLog.Fatalf("Failed to iterate message documents for chat %s: %v", chatID, err)
			}
			messageID := docMsg.Ref.ID
			fmt.Printf("Syncing message %s of chat %s\n", messageID, chatID)
			// Here you can add the logic to sync the message to your search engine or database
			messageData := docMsg.Data()
			messageData["id"] = messageID
			messageData["chat_id"] = chatID
			messages = append(messages, messageData)
		}
		if len(messages) > 0 {
			task, err := meiliClient.Index("messages").AddDocuments(messages, &meilisearch.DocumentOptions{
				PrimaryKey: &pkName,
			})
			if err != nil {
				zsLog.Fatalf("Failed to add documents to meili index for chat %s: %v", chatID, err)
			}
			// check if task is completed
			for {
				task, err := meiliClient.Index("messages").GetTask(task.TaskUID)
				if err != nil {
					zsLog.Fatalf("Failed to get task status: %v", err)
				}
				if task.Status == "succeeded" {
					break
				} else if task.Status == "failed" {
					zsLog.Fatalf("Failed to add documents to meili index for chat %s: %v", chatID, task.Error)
				} else {
					zsLog.Infof("Waiting for task to complete")
				}
				time.Sleep(1 * time.Second)
			}
		}
	}
}
