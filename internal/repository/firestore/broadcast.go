package repository_firestore

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (r *BroadcastRepository) Insert(ctx context.Context, tx *firestore.Transaction, broadcast model.Broadcast) error {
	docRef := r.db.Collection(r.broadcast.TableName()).Doc(broadcast.DocumentID)
	if tx != nil {
		return tx.Set(docRef, broadcast)
	} else {
		_, err := docRef.Set(ctx, broadcast)
		return err
	}
}

func (r *BroadcastRepository) GetByID(ctx context.Context, broadcastID string) (model.Broadcast, error) {
	docRef, err := r.db.Collection(r.broadcast.TableName()).Doc(broadcastID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return r.broadcast, errs.ErrGenericNotFound
		}
		return r.broadcast, err
	}
	var broadcast model.Broadcast
	docData := docRef.Data()
	docData["id"] = docRef.Ref.ID
	err = utils.MapToStruct(docData, &broadcast)
	if err != nil {
		return r.broadcast, err
	}
	return broadcast, nil
}

func (r *BroadcastRepository) Update(ctx context.Context, tx *firestore.Transaction, broadcast model.Broadcast) error {
	docRef := r.db.Collection(r.broadcast.TableName()).Doc(broadcast.DocumentID)
	if tx != nil {
		return tx.Set(docRef, broadcast)
	} else {
		_, err := docRef.Set(ctx, broadcast)
		return err
	}
}

func (r *BroadcastRepository) Delete(ctx context.Context, tx *firestore.Transaction, broadcastID string) error {
	docRef := r.db.Collection(r.broadcast.TableName()).Doc(broadcastID)
	if tx != nil {
		return tx.Delete(docRef)
	} else {
		_, err := docRef.Delete(ctx)
		return err
	}
}

func (r *BroadcastRepository) InsertRecipient(ctx context.Context, tx *firestore.Transaction, broadcastRecipient model.BroadcastRecipient) error {
	docRef := r.db.Collection(r.broadcast.TableName()).Doc(broadcastRecipient.BroadcastID).Collection(r.broadcastRecipient.TableName()).Doc(broadcastRecipient.DocumentID)
	if tx != nil {
		return tx.Set(docRef, broadcastRecipient)
	} else {
		_, err := docRef.Set(ctx, broadcastRecipient)
		return err
	}
}

func (r *BroadcastRepository) GetRecipientsByBroadcastID(ctx context.Context, broadcastID string) ([]model.BroadcastRecipient, error) {
	docs, err := r.db.
		Collection(r.broadcast.TableName()).Doc(broadcastID).
		Collection(r.broadcastRecipient.TableName()).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	var recipients []model.BroadcastRecipient
	for _, docRef := range docs {
		var recipient model.BroadcastRecipient
		docData := docRef.Data()
		docData["id"] = docRef.Ref.ID
		docData["broadcast_id"] = broadcastID
		err = utils.MapToStruct(docData, &recipient)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}
	return recipients, nil
}

func (r *BroadcastRepository) UpdateRecipientStatus(ctx context.Context, tx *firestore.Transaction, data model.BroadcastRecipient) error {
	docRef := r.db.Collection(r.broadcast.TableName()).Doc(data.BroadcastID).Collection(r.broadcastRecipient.TableName()).Doc(data.DocumentID)
	updates := []firestore.Update{
		{Path: "status", Value: data.Status},
		{Path: "updated_at", Value: data.UpdatedAt},
		{Path: "errors", Value: data.Errors},
	}
	if tx != nil {
		return tx.Update(docRef, updates)
	} else {
		_, err := docRef.Update(ctx, updates)
		return err
	}
}

func (r *BroadcastRepository) GetFilteredByTenantID(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.BroadcastGetFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastResponse], error) {
	var emptyResponse filter_request.FilterResponse[dto.BroadcastResponse]
	query := r.db.Collection(r.broadcast.TableName()).Query.Where("tenant_id", "==", tenantID)
	filters, sort, paginate, err := filter_request.InitializeFilter(inputData, r.broadcast.AllowedFilterFields(), r.broadcast.AllowedSortFields())
	if err != nil {
		return emptyResponse, err
	}
	docRef, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, sort, paginate)
	if err != nil {
		return emptyResponse, err
	}
	var results []dto.BroadcastResponse
	for _, doc := range docRef {
		var broadcast model.Broadcast
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &broadcast)
		if err != nil {
			return emptyResponse, err
		}
		results = append(results, dto.BroadcastResponse{}.FromModel(broadcast))
	}
	response := filter_request.NewFilterResponse(results, paginate, totalData)
	return response, nil

}

func (r *BroadcastRepository) GetRecipientsFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.BroadcastGetRecipientsFilteredRequest]) (filter_request.FilterResponse[dto.BroadcastRecipientResponse], error) {
	var emptyResponse filter_request.FilterResponse[dto.BroadcastRecipientResponse]
	query := r.db.Collection(r.broadcast.TableName()).Doc(inputData.SpecificFilter.BroadcastID).Collection(r.broadcastRecipient.TableName()).Query
	filters, sort, paginate, err := filter_request.InitializeFilter(inputData, r.broadcastRecipient.AllowedFilterFields(), r.broadcastRecipient.AllowedSortFields())
	if err != nil {
		return emptyResponse, err
	}
	docRef, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, sort, paginate)
	if err != nil {
		return emptyResponse, err
	}
	var results []dto.BroadcastRecipientResponse
	for _, doc := range docRef {
		var broadcast model.BroadcastRecipient
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		docData["broadcast_id"] = inputData.SpecificFilter.BroadcastID
		err = utils.MapToStruct(docData, &broadcast)
		if err != nil {
			return emptyResponse, err
		}
		results = append(results, dto.BroadcastRecipientResponse{}.FromModel(broadcast))
	}
	response := filter_request.NewFilterResponse(results, paginate, totalData)
	return response, nil

}
