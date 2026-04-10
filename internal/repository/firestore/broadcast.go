package repository_firestore

import (
	"context"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
)

type BroadcastRepository struct {
	broadcast          model.Broadcast
	broadcastRecipient model.BroadcastRecipient
	db                 *firestore.Client
}

func NewBroadcastRepository(db *firestore.Client) *BroadcastRepository {
	return &BroadcastRepository{
		db: db,
	}
}

func (r *BroadcastRepository) Insert(ctx context.Context, broadcast model.Broadcast) error {
	_, err := r.db.Collection(r.broadcast.TableName()).Doc(broadcast.DocumentID).Set(ctx, broadcast)
	return err
}

func (r *BroadcastRepository) GetByID(ctx context.Context, broadcastID string) (model.Broadcast, error) {
	docRef, err := r.db.Collection(r.broadcast.TableName()).Doc(broadcastID).Get(ctx)
	if err != nil {
		return model.Broadcast{}, err
	}
	var broadcast model.Broadcast
	doc := docRef.Data()
	doc[firestore.DocumentID] = docRef.Ref.ID
	err = utils.MapToStruct(doc, &broadcast)
	if err != nil {
		return model.Broadcast{}, err
	}
	return broadcast, nil
}

func (r *BroadcastRepository) Update(ctx context.Context, broadcast model.Broadcast) error {
	_, err := r.db.Collection(r.broadcast.TableName()).Doc(broadcast.DocumentID).Set(ctx, broadcast)
	return err
}

func (r *BroadcastRepository) Delete(ctx context.Context, broadcastID string) error {
	_, err := r.db.Collection(r.broadcast.TableName()).Doc(broadcastID).Delete(ctx)
	return err
}

func (r *BroadcastRepository) InsertRecipient(ctx context.Context, broadcastRecipient model.BroadcastRecipient) error {
	_, err := r.db.
		Collection(r.broadcast.TableName()).Doc(broadcastRecipient.BroadcastID).
		Collection(r.broadcastRecipient.TableName()).Doc(broadcastRecipient.DocumentID).
		Set(ctx, broadcastRecipient)
	return err
}

func (r *BroadcastRepository) GetRecipietsByBroadcastID(ctx context.Context, broadcastID string) ([]model.BroadcastRecipient, error) {
	query := r.db.Collection(r.broadcast.TableName()).Doc(broadcastID).Collection(r.broadcastRecipient.TableName())
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	var recipients []model.BroadcastRecipient
	for _, docRef := range docs {
		var recipient model.BroadcastRecipient
		doc := docRef.Data()
		doc[firestore.DocumentID] = docRef.Ref.ID
		doc["broadcast_id"] = broadcastID
		err = utils.MapToStruct(doc, &recipient)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}
	return recipients, nil
}

func (r *BroadcastRepository) UpdateRecipientStatus(ctx context.Context, data model.BroadcastRecipient) error {
	_, err := r.db.
		Collection(r.broadcast.TableName()).Doc(data.BroadcastID).
		Collection(r.broadcastRecipient.TableName()).Doc(data.DocumentID).
		Update(ctx, []firestore.Update{
			{Path: "status", Value: data.Status},
			{Path: "updated_at", Value: data.UpdatedAt},
			{Path: "errors", Value: data.Errors},
		})
	return err
}
