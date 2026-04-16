package activity_log_usecase

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/repository"

	"go.uber.org/zap"
)

// Struct that implements the activity log usecase interface
type ActivityLogUsecase struct {
	activityLogRepository repository.ActivityLog
	zslog                 *zap.SugaredLogger
}

// Creates a new instance of ActivityLogUsecase. This function is typically called when setting up the application's dependencies, allowing the use case to be injected into services and handlers that require access to activity log-related operations in the application.
func NewActivityLogUsecase(activityLogRepository repository.ActivityLog, zslog *zap.SugaredLogger) *ActivityLogUsecase {
	return &ActivityLogUsecase{
		activityLogRepository: activityLogRepository,
		zslog:                 zslog,
	}
}

func (uc *ActivityLogUsecase) Insert(ctx context.Context, inputData dto.ActivityLogCreateRequest) (model.ActivityLog, bool, error) {
	var err error
	var userID *string
	if inputData.UserID != nil {
		userIDStr := inputData.UserID.String()
		userID = &userIDStr
	}
	activityLog := model.ActivityLog{
		UserID:      userID,
		Type:        inputData.Type,
		Description: inputData.Description,
	}
	activityLog, err = uc.activityLogRepository.Insert(ctx, nil, activityLog)
	if err != nil {
		uc.zslog.Errorf("[Insert] failed to insert activity log: %v", err)
		return model.ActivityLog{}, true, err
	}
	return activityLog, false, nil
}
