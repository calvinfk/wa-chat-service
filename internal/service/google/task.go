package google_service

import (
	"net/http"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/internal/service"

	"google.golang.org/api/cloudtasks/v2"
)

type GoogleTaskService struct {
	client         *cloudtasks.Service
	cfg            *config.GCP
	jwtService     service.JWT
	encryptService service.Encrypt
}

func NewGoogleTaskService(client *cloudtasks.Service, cfg *config.GCP, jwtService service.JWT, encryptService service.Encrypt) *GoogleTaskService {
	return &GoogleTaskService{
		client:         client,
		cfg:            cfg,
		jwtService:     jwtService,
		encryptService: encryptService,
	}
}

// func (s *GoogleTaskService) CreateGoogleTask(ctx context.Context) error {
// 	// Implementation for creating a Google Task
// 	req := &cloudtasks.CreateTaskRequest{
// 		Task: &cloudtasks.Task{
// 			HttpRequest: &cloudtasks.HttpRequest{
// 				Url: s.cfg.AppBaseURL + "/ping",
// 			},
// 			ScheduleTime: time.Now().Add(1 * time.Minute).Format(time.RFC3339), // Schedule task to run after 1 minute
// 		},
// 	}
// 	// err := s.client.Projects.Locations.Queues.Tasks.Create("projects/"+s.cfg.ProjectID+"/locations/asia-southeast2"+"/queues/"+s.cfg.QueueID, req).Context(ctx).Do()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (s *GoogleTaskService) CreatePingTask() error {
	token, err := s.jwtService.GenerateJWT("ping-task", time.Now().Add(time.Second*5).Unix())
	if err != nil {
		return err
	}
	encryptedToken, err := s.encryptService.Encrypt(token)
	if err != nil {
		return err
	}
	req := &cloudtasks.CreateTaskRequest{
		Task: &cloudtasks.Task{
			HttpRequest: &cloudtasks.HttpRequest{
				Url:        s.cfg.AppBaseURL + "/api/v1/broadcast/send",
				HttpMethod: http.MethodPost,
				Headers: map[string]string{
					"Authorization": "Bearer " + encryptedToken,
				},
			},
			ScheduleTime: time.Now().Add(15 * time.Second).Format(time.RFC3339), // Schedule task to run after 15 seconds
		},
	}
	_, err = s.client.Projects.Locations.Queues.Tasks.Create("projects/"+s.cfg.ProjectID+"/locations/asia-southeast2/queues/broadcast-message", req).Do()
	if err != nil {
		return err
	}
	return nil
}
