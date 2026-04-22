package google_service

import (
	"net/http"
	"time"
	"wa_chat_service/config"
	"wa_chat_service/internal/service"

	"go.uber.org/zap"
	"google.golang.org/api/cloudtasks/v2"
)

type GoogleTaskService struct {
	client         *cloudtasks.Service
	cfg            *config.GCP
	jwtService     service.JWT
	encryptService service.Encrypt
	zslog          *zap.SugaredLogger
	baseURL        string
}

func NewGoogleTaskService(client *cloudtasks.Service, cfg *config.GCP, jwtService service.JWT, encryptService service.Encrypt, zslog *zap.SugaredLogger, baseURL string) *GoogleTaskService {
	return &GoogleTaskService{
		client:         client,
		cfg:            cfg,
		jwtService:     jwtService,
		encryptService: encryptService,
		zslog:          zslog,
		baseURL:        baseURL,
	}
}

func (s *GoogleTaskService) CreateBroadcastTask(broadcastID string, scheduleTime time.Time) error {
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
		s.zslog.Errorf("[CreateBroadcastTask] error creating broadcast task: %v", err)
		return err
	}
	return nil
}

func (s *GoogleTaskService) DeleteBroadcastTask(broadcastID string) error {
	_, err := s.client.Projects.Locations.Queues.Tasks.Delete(s.cfg.BroadcastTaskParent + "/tasks/" + broadcastID).Do()
	if err != nil {
		s.zslog.Errorf("[DeleteBroadcastTask] error deleting broadcast task: %v", err)
		return err
	}
	return nil
}
