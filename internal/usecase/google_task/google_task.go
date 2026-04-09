package google_task_usecase

import (
	"log"
	"wa_chat_service/internal/service"
)

type GoogleTaskUsecase struct {
	googleTaskService service.GoogleTask
}

func NewGoogleTaskUsecase(googleTaskService service.GoogleTask) *GoogleTaskUsecase {
	return &GoogleTaskUsecase{
		googleTaskService: googleTaskService,
	}
}

func (u *GoogleTaskUsecase) CreatePingTask() (bool, error) {
	// Implementation for creating a Google Task
	err := u.googleTaskService.CreatePingTask()
	if err != nil {
		log.Println("[ERROR][internal/usecase/google_task/google_task.go][CreatePingTask] error: ", err)
		return true, err
	}
	return false, nil
}
