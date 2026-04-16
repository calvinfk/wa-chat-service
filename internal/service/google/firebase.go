package google_service

import (
	"context"
	"wa_chat_service/config"

	"firebase.google.com/go/v4/messaging"
	"firebase.google.com/go/v4/storage"
	"go.uber.org/zap"
)

type GoogleFirebaseService struct {
	config                  *config.GCP
	firebaseMessagingClient *messaging.Client
	firebaseStorageClient   *storage.Client
	zslog                   *zap.SugaredLogger
}

func NewGoogleFirebaseService(config *config.GCP, firebaseMessagingClient *messaging.Client, firebaseStorageClient *storage.Client, zslog *zap.SugaredLogger) *GoogleFirebaseService {
	return &GoogleFirebaseService{
		config:                  config,
		firebaseMessagingClient: firebaseMessagingClient,
		firebaseStorageClient:   firebaseStorageClient,
		zslog:                   zslog,
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
		s.zslog.Errorf("[SendNotification] error sending message: %v", err)
		return err
	}
	for idx, resp := range response.Responses {
		if !resp.Success {
			s.zslog.Errorf("[SendNotification] Failed to send message to token: %s, error: %v", tokens[idx], resp.Error)
		}
	}
	return nil
}

func (s *GoogleFirebaseService) SubscribeToTopic(ctx context.Context, topics []string, tokens []string) error {
	for _, t := range topics {
		response, err := s.firebaseMessagingClient.SubscribeToTopic(ctx, tokens, t)
		if err != nil {
			s.zslog.Errorf("[SubscribeToTopic] error subscribing to topic: %v", err)
			return err
		}
		if response.Errors != nil {
			for _, subErr := range response.Errors {
				s.zslog.Errorf("[SubscribeToTopic] Failed to subscribe token: %s to topic: %s, error: %v", subErr.Index, t, subErr.Reason)
			}
		}
	}
	return nil
}

func (s *GoogleFirebaseService) UnsubscribeFromTopic(ctx context.Context, topics []string, tokens []string) error {
	// 2. Unsubscribe from topics
	for _, t := range topics {
		response, err := s.firebaseMessagingClient.UnsubscribeFromTopic(ctx, tokens, t)
		if err != nil {
			s.zslog.Errorf("[UnsubscribeFromTopic] error unsubscribing from topic: %v", err)
			return err
		}
		if response.Errors != nil {
			for _, subErr := range response.Errors {
				s.zslog.Errorf("[UnsubscribeFromTopic] Failed to unsubscribe token: %s from topic: %s, error: %v", subErr.Index, t, subErr.Reason)
			}
		}
	}
	return nil
}

func (s *GoogleFirebaseService) SendNotificationToTopic(ctx context.Context, title string, body string, topic string) error {
	_, err := s.firebaseMessagingClient.Send(ctx, &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Topic: topic,
	})
	if err != nil {
		s.zslog.Errorf("[SendNotificationToTopic] error sending message to topic: %v", err)
		return err
	}
	return nil
}
