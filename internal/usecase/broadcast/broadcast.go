package broadcast_usecase

import (
	"fmt"
	"log"
	"wa_chat_service/internal/service"
)

type BroadcastUsecase struct {
	googleTaskService service.GoogleTask
}

func NewBroadcastUsecase(googleTaskService service.GoogleTask) *BroadcastUsecase {
	return &BroadcastUsecase{
		googleTaskService: googleTaskService,
	}
}

func (u *BroadcastUsecase) ScheduleBroadcast() (bool, error) {
	// Implementation for creating a Google Task
	err := u.googleTaskService.CreatePingTask()
	if err != nil {
		log.Println("[ERROR][internal/usecase/broadcast/broadcast.go][ScheduleBroadcast] error: ", err)
		return true, err
	}
	return false, nil
}

func (u *BroadcastUsecase) SendBroadcast() (bool, error) {
	// Implementation for sending a broadcast
	return true, fmt.Errorf("SendBroadcast is not implemented yet")
}
