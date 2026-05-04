package repository_firestore

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRepository struct {
	user model.User
	db   *firestore.Client
}

func NewUserRepository(db *firestore.Client) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	docRef, err := r.db.Collection(r.user.TableName()).Where("email", "==", email).Limit(1).Documents(ctx).Next()
	if err != nil {
		if err == iterator.Done {
			return model.User{}, errs.ErrGenericNotFound
		}
		return model.User{}, err
	}
	var user model.User
	docData := docRef.Data()
	docData["id"] = docRef.Ref.ID
	err = utils.MapToStruct(docData, &user)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (model.User, error) {
	doc, err := r.db.Collection(r.user.TableName()).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return model.User{}, errs.ErrGenericNotFound
		}
		return model.User{}, err
	}
	var user model.User
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &user)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByTenantIDFiltered(ctx context.Context, tenantID string, filters filter_request.FilterRequest[dto.UserListRequest]) (filter_request.FilterResponse[dto.UserResponse], error) {
	var response filter_request.FilterResponse[dto.UserResponse]
	filter, sort, paginate, err := filter_request.InitializeFilter(filters, r.user.AllowedFilterFields(), r.user.AllowedSortFields())
	if err != nil {
		return response, err
	}
	query := r.db.Collection(r.user.TableName()).Where("tenant_id", "==", tenantID)
	docs, totalItems, err := filter_request.ApplyFilterFirestore(ctx, query, filter, sort, paginate)
	if err != nil {
		return response, err
	}
	var results []dto.UserResponse
	for _, doc := range docs {
		var user model.User
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &user)
		if err != nil {
			return response, err
		}
		results = append(results, dto.UserResponse{}.FromModel(user))
	}
	response = filter_request.NewFilterResponse(results, paginate, totalItems)
	return response, nil
}

func (r *UserRepository) Upsert(ctx context.Context, tx *firestore.Transaction, user model.User) (model.User, error) {
	docRef := r.db.Collection(r.user.TableName()).Doc(user.DocumentID)
	if tx != nil {
		return user, tx.Set(docRef, user)
	} else {
		_, err := docRef.Set(ctx, user)
		if err != nil {
			return model.User{}, err
		}
		return user, nil
	}
}

func (r *UserRepository) GetBySupervisorID(ctx context.Context, supervisorID string) ([]model.User, error) {
	var users []model.User
	docs, err := r.db.Collection(r.user.TableName()).Where("supervisor_id", "==", supervisorID).Documents(ctx).GetAll()
	if err != nil {
		return users, err
	}
	for _, doc := range docs {
		var user model.User
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &user)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}
