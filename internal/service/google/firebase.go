package google_service

import (
	"context"
	"fmt"
	"log"
	"wa_chat_service/config"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type GoogleFirebaseService struct {
	cfg            *config.GCP
	firebaseClient *firebase.App
}

func NewGoogleFirebaseService(cfg *config.GCP, firebaseClient *firebase.App) *GoogleFirebaseService {
	return &GoogleFirebaseService{
		cfg:            cfg,
		firebaseClient: firebaseClient,
	}
}

func (s *GoogleFirebaseService) SendNotification(ctx context.Context, title string, body string, tokens []string) error {
	// 1. Obtain a Messaging Client
	client, err := s.firebaseClient.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("error getting Messaging client: %v", err)
	}

	// 2. Send the message
	response, err := client.SendEachForMulticast(ctx, &messaging.MulticastMessage{
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
	// 1. Obtain a Messaging Client
	client, err := s.firebaseClient.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("error getting Messaging client: %v", err)
	}
	// 2. Subscribe to topics
	for _, t := range topics {
		response, err := client.SubscribeToTopic(ctx, tokens, t)
		if err != nil {
			return fmt.Errorf("error subscribing to topic: %v", err)
		}
		log.Printf("Successfully subscribed %d tokens to topic %s\n", response.SuccessCount, t)
	}
	return nil
}

func (s *GoogleFirebaseService) UnsubscribeFromTopic(ctx context.Context, topics []string, tokens []string) error {
	// 1. Obtain a Messaging Client
	client, err := s.firebaseClient.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("error getting Messaging client: %v", err)
	}
	// 2. Unsubscribe from topics
	for _, t := range topics {
		response, err := client.UnsubscribeFromTopic(ctx, tokens, t)
		if err != nil {
			return fmt.Errorf("error unsubscribing from topic: %v", err)
		}
		log.Printf("Successfully unsubscribed %d tokens from topic %s\n", response.SuccessCount, t)
	}
	return nil
}

func (s *GoogleFirebaseService) SendNotificationToTopic(ctx context.Context, title string, body string, topic string) error {
	// 1. Obtain a Messaging Client
	client, err := s.firebaseClient.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("error getting Messaging client: %v", err)
	}
	// 2. Send the message
	response, err := client.Send(ctx, &messaging.Message{
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
