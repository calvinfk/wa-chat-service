package google_service

import (
	"context"
	"fmt"
	"log"
	"wa_chat_service/config"

	"firebase.google.com/go/v4/messaging"
	"firebase.google.com/go/v4/storage"
)

type GoogleFirebaseService struct {
	config                  *config.GCP
	firebaseMessagingClient *messaging.Client
	firebaseStorageClient   *storage.Client
}

func NewGoogleFirebaseService(config *config.GCP, firebaseMessagingClient *messaging.Client, firebaseStorageClient *storage.Client) *GoogleFirebaseService {
	return &GoogleFirebaseService{
		config:                  config,
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
