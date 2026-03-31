package activity_log_usecase

import (
	"context"
	"log"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"
	"wa_chat_service/pkg/filter_request"
)

// Struct that implements the activity log usecase interface
type ActivityLogUsecase struct {
	activityLogRepository repository.ActivityLog
}

// Creates a new instance of ActivityLogUsecase. This function is typically called when setting up the application's dependencies, allowing the use case to be injected into services and handlers that require access to activity log-related operations in the application.
func NewActivityLogUsecase(activityLogRepository repository.ActivityLog) *ActivityLogUsecase {
	return &ActivityLogUsecase{
		activityLogRepository: activityLogRepository,
	}
}

func (uc *ActivityLogUsecase) Insert(ctx context.Context, inputData dto.ActivityLogCreateRequest) (model.ActivityLog, bool, error) {
	var err error
	activityLog := model.ActivityLog{
		UserID:      inputData.UserID,
		Type:        inputData.Type,
		Description: inputData.Description,
	}
	activityLog, err = uc.activityLogRepository.Insert(ctx, nil, activityLog)
	if err != nil {
		log.Println("[ERROR][internal/usecase/activity_log/activity_log.go][Insert] failed to insert activity log:", err)
		return model.ActivityLog{}, true, err
	}
	return activityLog, false, nil
}

func (uc *ActivityLogUsecase) GetFiltered(ctx context.Context, user model.User, filter filter_request.FilterRequest[dto.ActivityLogFilterRequest]) (filter_request.FilterResponse[dto.ActivityLogResponse], bool, error) {
	var err error
	var response filter_request.FilterResponse[dto.ActivityLogResponse]
	response, err = uc.activityLogRepository.GetFiltered(ctx, filter)
	if err != nil {
		log.Println("[ERROR][internal/usecase/activity_log/activity_log.go][GetFiltered] Error getting activity log ", err.Error())
		return response, true, err
	}
	return response, false, nil
}
