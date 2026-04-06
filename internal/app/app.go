package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"wa_chat_service/config"
	http_internal "wa_chat_service/internal/handler/http"
	http_v1 "wa_chat_service/internal/handler/http/v1"
	repository_firestore "wa_chat_service/internal/repository/firestore"
	access_token_service "wa_chat_service/internal/service/access_token"
	encrypt_service "wa_chat_service/internal/service/encrypt"
	google_service "wa_chat_service/internal/service/google"
	whatsapp_service "wa_chat_service/internal/service/whatsapp"
	activity_log_usecase "wa_chat_service/internal/usecase/activity_log"
	chat_usecase "wa_chat_service/internal/usecase/chat"
	message_usecase "wa_chat_service/internal/usecase/message"
	storage_media_usecase "wa_chat_service/internal/usecase/storage_media"
	"wa_chat_service/pkg/formatter"
	"wa_chat_service/pkg/google"

	"github.com/gofiber/fiber/v3"
)

func Run(config *config.Config) {
	log.Printf("[INFO][internal/app/app/Run] Starting %s, version %s, in %s mode", config.App.Name, config.App.Version, config.App.Environment)

	// dbPostgres := database.OpenPostgresConnection(config.Database.URL)
	firebaseClient := google.OpenFirebaseConnection(config.GCP.ProjectID)
	gcpStorageClient := google.OpenGCPStorageConnection()
	firestoreClient, err := firebaseClient.Firestore(context.Background())
	if err != nil {
		log.Fatalf("[ERROR][internal/app/app.go][Run] Failed to create Firestore client: %v", err)
	}
	firebaseMessagingClient, err := firebaseClient.Messaging(context.Background())
	if err != nil {
		log.Fatalf("[ERROR][internal/app/app.go][Run] Failed to create Firebase Messaging client: %v", err)
	}
	firebaseStorageClient, err := firebaseClient.Storage(context.Background())
	if err != nil {
		log.Fatalf("[ERROR][internal/app/app.go][Run] Failed to create Firebase Storage client: %v", err)
	}
	// _ = transaction.NewTxManager(dbPostgres, firestoreClient)

	// Service
	accessTokenService := access_token_service.NewAccessTokenService(config)
	encryptService := encrypt_service.NewEncryptService(&config.Encrypt)
	googleStorageService := google_service.NewGoogleStorageService(gcpStorageClient, &config.GCP)
	googleFirebaseService := google_service.NewGoogleFirebaseService(&config.GCP, firebaseMessagingClient, firebaseStorageClient)
	whatsappService := whatsapp_service.NewWhatsappService()

	// Repository
	activityLogRepository := repository_firestore.NewActivityLogRepository(firestoreClient)
	messageRepository := repository_firestore.NewMessageRepository(firestoreClient, googleStorageService)
	chatRepository := repository_firestore.NewChatRepository(firestoreClient)
	storageMediaRepository := repository_firestore.NewStorageMediaRepository(firestoreClient)
	phoneNumberRepository := repository_firestore.NewPhoneNumberRepository(firestoreClient)

	// Usecase
	activityLogUsecase := activity_log_usecase.NewActivityLogUsecase(activityLogRepository)
	storageMediaUsecase := storage_media_usecase.NewStorageMediaUsecase(storageMediaRepository, phoneNumberRepository, googleFirebaseService, googleStorageService, encryptService, whatsappService)
	messageUsecase := message_usecase.NewMessageUsecase(messageRepository, chatRepository, phoneNumberRepository, storageMediaRepository, storageMediaUsecase, whatsappService, encryptService, googleFirebaseService)
	chatUsecase := chat_usecase.NewChatUsecase(chatRepository)

	// Router Handler
	routerHandlerV1 := http_v1.RouterHandlerV1{
		ActivityLogUsecase:    activityLogUsecase,
		MessageUsecase:        messageUsecase,
		StorageMediaUsecase:   storageMediaUsecase,
		ChatUsecase:           chatUsecase,
		AccessTokenService:    accessTokenService,
		EncryptService:        encryptService,
		GoogleStorageService:  googleStorageService,
		GoogleFirebaseService: googleFirebaseService,
	}

	app := fiber.New(fiber.Config{
		AppName:         config.App.Name,
		BodyLimit:       16 * 1024 * 1024, // 16MB
		StructValidator: formatter.Validator(),
	})

	http_internal.NewRouter(app, config, routerHandlerV1)
	go func() {
		if err := app.Listen(":" + fmt.Sprintf("%d", config.App.Port)); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("[ERROR][internal/app/app/Run] Failed to start server: %v", err)
		}
	}()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	log.Printf("[INFO][internal/app/app/Run] Server is running on %s", fmt.Sprintf(":%d", config.App.Port))
	<-interrupt
	log.Println("[INFO][internal/app/app/Run] Shutting down server...")
	// Shutdown gracefully
	if err := app.Shutdown(); err != nil {
		log.Printf("[ERROR][internal/app/app/Run] Server forced to shutdown: %v", err)
	}
	log.Println("[INFO][internal/app/app/Run] Closing database connection...")
	// if sqlDB, err := dbPostgres.DB(); err != nil {
	// 	log.Printf("[ERROR][internal/app/app.go][Run] Error getting database connection: %v", err)
	// } else {
	// 	if err := sqlDB.Close(); err != nil {
	// 		log.Printf("[ERROR][internal/app/app.go][Run] Error closing database connection: %v", err)
	// 	}
	// }

	log.Println("[INFO][internal/app/app/Run] Server exiting")
}
