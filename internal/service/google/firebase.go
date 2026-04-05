package google_service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wa_chat_service/config"

	gs "cloud.google.com/go/storage"
	"firebase.google.com/go/v4/messaging"
	"firebase.google.com/go/v4/storage"
)

type GoogleFirebaseService struct {
	cfg                     *config.GCP
	firebaseMessagingClient *messaging.Client
	firebaseStorageClient   *storage.Client
}

func NewGoogleFirebaseService(cfg *config.GCP, firebaseMessagingClient *messaging.Client, firebaseStorageClient *storage.Client) *GoogleFirebaseService {
	return &GoogleFirebaseService{
		cfg:                     cfg,
		firebaseMessagingClient: firebaseMessagingClient,
		firebaseStorageClient:   firebaseStorageClient,
	}
}

func (s *GoogleFirebaseService) SendNotification(ctx context.Context, title string, body string, tokens []string) error {
	response, err := s.firebaseMessagingClient.SendEachForMulticast(ctx, &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Tokens: tokens,
	})
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}
	for idx, resp := range response.Responses {
		if !resp.Success {
			log.Printf("Failed to send message to token: %s, error: %v\n", tokens[idx], resp.Error)
		}
	}
	log.Printf("Successfully sent message to %d tokens\n", response.SuccessCount)
	return nil
}

func (s *GoogleFirebaseService) SubscribeToTopic(ctx context.Context, topics []string, tokens []string) error {
	for _, t := range topics {
		response, err := s.firebaseMessagingClient.SubscribeToTopic(ctx, tokens, t)
		if err != nil {
			return fmt.Errorf("error subscribing to topic: %v", err)
		}
		log.Printf("Successfully subscribed %d tokens to topic %s\n", response.SuccessCount, t)
	}
	return nil
}

func (s *GoogleFirebaseService) UnsubscribeFromTopic(ctx context.Context, topics []string, tokens []string) error {
	// 2. Unsubscribe from topics
	for _, t := range topics {
		response, err := s.firebaseMessagingClient.UnsubscribeFromTopic(ctx, tokens, t)
		if err != nil {
			return fmt.Errorf("error unsubscribing from topic: %v", err)
		}
		log.Printf("Successfully unsubscribed %d tokens from topic %s\n", response.SuccessCount, t)
	}
	return nil
}

func (s *GoogleFirebaseService) SendNotificationToTopic(ctx context.Context, title string, body string, topic string) error {
	response, err := s.firebaseMessagingClient.Send(ctx, &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Topic: topic,
	})
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}
	log.Printf("Successfully sent message to topic %s: %s\n", topic, response)
	return nil
}

func (s *GoogleFirebaseService) UploadFile(ctx context.Context, filePath string, file []byte) (*gs.ObjectAttrs, error) {
	bucket, err := s.firebaseStorageClient.DefaultBucket()
	if err != nil {
		return nil, fmt.Errorf("error getting bucket: %v", err)
	}
	object := bucket.Object(filePath)
	writer := object.NewWriter(ctx)
	if _, err := writer.Write(file); err != nil {
		return nil, fmt.Errorf("error writing file: %v", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing writer: %v", err)
	}
	attrs, err := object.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting object attributes: %v", err)
	}
	return attrs, nil
}

func (s *GoogleFirebaseService) DeleteFile(ctx context.Context, filePath string) error {
	filePath = strings.TrimPrefix(filePath, "gs://")
	filePath = strings.TrimPrefix(filePath, s.cfg.ProjectID+".firebasestorage.app/")
	bucket, err := s.firebaseStorageClient.DefaultBucket()
	if err != nil {
		return fmt.Errorf("error getting bucket: %v", err)
	}
	object := bucket.Object(filePath)
	if err := object.Delete(ctx); err != nil {
		return fmt.Errorf("error deleting file: %v", err)
	}
	return nil
}
