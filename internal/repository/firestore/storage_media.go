package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"

	"cloud.google.com/go/firestore"
)

type StorageMediaRepository struct {
	storageMedia model.StorageMedia
	db           *firestore.Client
}

func NewStorageMediaRepository(db *firestore.Client) *StorageMediaRepository {
	return &StorageMediaRepository{
		db: db,
	}
}

func (r *StorageMediaRepository) Insert(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) (model.StorageMedia, error) {
	var err error
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.db.Collection("storage_medias").Doc(data.DocumentID)
		return tx.Set(docRef, data)
	}
	if tx != nil {
		err = execDB(ctx, tx)
	} else {
		err = r.db.RunTransaction(ctx, execDB)
	}
	return data, err
}

func (r *StorageMediaRepository) GetByDocumentID(ctx context.Context, documentID string) (model.StorageMedia, error) {
	docRef := r.db.Collection("storage_medias").Doc(documentID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return model.StorageMedia{}, err
	}
	var media model.StorageMedia
	if err := docSnap.DataTo(&media); err != nil {
		return model.StorageMedia{}, err
	}
	return media, nil
}
