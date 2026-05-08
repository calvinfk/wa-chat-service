package google_service

import (
	"net/http"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/internal/service"

	"go.uber.org/zap"
	"google.golang.org/api/cloudtasks/v2"
)

type googleTaskService struct {
	client         *cloudtasks.Service
	cfg            *config.GCP
	jwtService     service.JWT
	encryptService service.Encrypt
	zsLog          *zap.SugaredLogger
	baseURL        string
}

// NewGoogleTaskService creates a new instance of googleTaskService with the provided dependencies.
func NewGoogleTaskService(client *cloudtasks.Service, cfg *config.GCP, jwtService service.JWT, encryptService service.Encrypt, zsLog *zap.SugaredLogger, baseURL string) *googleTaskService {
	return &googleTaskService{
		client:         client,
		cfg:            cfg,
		jwtService:     jwtService,
		encryptService: encryptService,
		zsLog:          zsLog,
		baseURL:        baseURL,
	}
}

func (s *googleTaskService) CreateBroadcastTask(broadcastID string, scheduleTime time.Time) error {
	// Generate a JWT token for the broadcast task with a unique identifier and an expiration time slightly after the scheduled time to ensure the task can execute successfully.
	token, err := s.jwtService.GenerateJWT("broadcast_"+broadcastID, scheduleTime.Add(time.Second*20).Unix())
	if err != nil {
		return err
	}
	encryptedToken, err := s.encryptService.Encrypt(token)
	if err != nil {
		return err
	}
	req := &cloudtasks.CreateTaskRequest{
		Task: &cloudtasks.Task{
			Name:         s.cfg.BroadcastTaskParent + "/tasks/" + broadcastID,
			ScheduleTime: scheduleTime.Format(time.RFC3339), // Schedule task to run at specified time
			HttpRequest: &cloudtasks.HttpRequest{
				Url:        s.baseURL + "/api/v1/broadcast/send",
				HttpMethod: http.MethodPost,
				Headers: map[string]string{
					"Authorization": "Bearer " + encryptedToken,
				},
			},
		},
	}
	_, err = s.client.Projects.Locations.Queues.Tasks.Create(s.cfg.BroadcastTaskParent, req).Do()
	if err != nil {
		s.zsLog.Errorf("[CreateBroadcastTask] error creating broadcast task: %v", err)
		return err
	}
	return nil
}

func (s *googleTaskService) DeleteBroadcastTask(broadcastID string) error {
	_, err := s.client.Projects.Locations.Queues.Tasks.Delete(s.cfg.BroadcastTaskParent + "/tasks/" + broadcastID).Do()
	if err != nil {
		s.zsLog.Errorf("[DeleteBroadcastTask] error deleting broadcast task: %v", err)
		return err
	}
	return nil
}

func (s *googleTaskService) CreateReminderSLATask(scheduleTime time.Time) error {
	token, err := s.jwtService.GenerateJWT("reminder-sla", scheduleTime.Add(time.Second*20).Unix())
	if err != nil {
		return err
	}
	encryptedToken, err := s.encryptService.Encrypt(token)
	if err != nil {
		return err
	}
	req := &cloudtasks.CreateTaskRequest{
		Task: &cloudtasks.Task{
			Name:         s.cfg.ScheduleTaskParent + "/tasks/" + "reminder-sla",
			ScheduleTime: scheduleTime.Format(time.RFC3339), // Schedule task to run at specified time
			HttpRequest: &cloudtasks.HttpRequest{
				Url:        s.baseURL + "/api/v1/chat/reminder-sla",
				HttpMethod: http.MethodPost,
				Headers: map[string]string{
					"Authorization": "Bearer " + encryptedToken,
				},
			},
		},
	}
	_, err = s.client.Projects.Locations.Queues.Tasks.Create(s.cfg.ScheduleTaskParent, req).Do()
	if err != nil {
		s.zsLog.Errorf("[CreateReminderSLATask] error creating reminder SLA task: %v", err)
		return err
	}
	return nil
}
