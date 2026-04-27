package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
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
	collection := r.db.Collection("users")
	docRef, err := collection.Where("email", "==", email).Limit(1).Documents(ctx).Next()
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
