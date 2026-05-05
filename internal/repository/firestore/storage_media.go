package repository_firestore

import (
	"context"
	"wa_chat_service/internal/dto"
	"wa_chat_service/internal/model"
	"wa_chat_service/internal/service"
	"wa_chat_service/pkg/errs"
	"wa_chat_service/pkg/filter_request"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StorageMediaRepository struct {
	storageMedia         model.StorageMedia
	db                   *firestore.Client
	googleStorageService service.GoogleStorage
}

func NewStorageMediaRepository(db *firestore.Client, googleStorageService service.GoogleStorage) *StorageMediaRepository {
	return &StorageMediaRepository{
		db:                   db,
		googleStorageService: googleStorageService,
	}
}

func (r *StorageMediaRepository) Upsert(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) (model.StorageMedia, error) {
	var err error
	docRef := r.db.Collection(r.storageMedia.TableName()).Doc(data.DocumentID)
	if tx != nil {
		err = tx.Set(docRef, data)
	} else {
		_, err = docRef.Set(ctx, data)
	}
	return data, err
}

func (r *StorageMediaRepository) GetByID(ctx context.Context, ID string) (model.StorageMedia, error) {
	docRef := r.db.Collection(r.storageMedia.TableName()).Doc(ID)
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return r.storageMedia, errs.ErrGenericNotFound
		}
		return r.storageMedia, err
	}
	var media model.StorageMedia
	docData := doc.Data()
	docData["id"] = doc.Ref.ID
	err = utils.MapToStruct(docData, &media)
	if err != nil {
		return r.storageMedia, err
	}
	return media, nil
}

func (r *StorageMediaRepository) Delete(ctx context.Context, tx *firestore.Transaction, ID string) error {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.db.Collection(r.storageMedia.TableName()).Doc(ID)
		return tx.Delete(docRef)
	}
	if tx != nil {
		return execDB(ctx, tx)
	}
	return r.db.RunTransaction(ctx, execDB)
}

func (r *StorageMediaRepository) GetFilteredByTenantID(ctx context.Context, tenantID string, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) ([]model.StorageMedia, filter_request.Paginate, int64, error) {
	var data []model.StorageMedia
	filters, sort, paginate, err := filter_request.InitializeFilter(inputData, r.storageMedia.AllowedFilterFields(), r.storageMedia.AllowedSortFields())
	if err != nil {
		return data, paginate, 0, err
	}
	collection := r.db.Collection(r.storageMedia.TableName())
	query := collection.Where("tenant_id", "==", tenantID)
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, sort, paginate)
	if err != nil {
		return data, paginate, 0, err
	}
	for _, doc := range docs {
		var media model.StorageMedia
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &media)
		if err != nil {
			return data, paginate, totalData, err
		}
		data = append(data, media)
	}
	return data, paginate, totalData, nil
}

func (r *StorageMediaRepository) GetByIDs(ctx context.Context, IDs []string) (map[string]model.StorageMedia, error) {
	mediaMap := make(map[string]model.StorageMedia)
	if len(IDs) == 0 {
		return mediaMap, nil
	}

	// avoid firestore in query limit
	// https://firebase.google.com/docs/firestore/query-data/queries#in_not-in_and_array-contains-any
	// Use the in operator to combine up to 30 equality (==) clauses on the same field with a logical OR.
	const maxInValues = 30
	collection := r.db.Collection(r.storageMedia.TableName())
	for i := 0; i < len(IDs); i += maxInValues {
		end := min(i+maxInValues, len(IDs))

		docRefs := make([]*firestore.DocumentRef, 0, end-i)
		for _, id := range IDs[i:end] {
			docRefs = append(docRefs, collection.Doc(id))
		}

		docs, err := collection.Where(firestore.DocumentID, "in", docRefs).Documents(ctx).GetAll()
		if err != nil {
			return nil, err
		}

		for _, doc := range docs {
			var media model.StorageMedia
			docData := doc.Data()
			docData["id"] = doc.Ref.ID
			err = utils.MapToStruct(docData, &media)
			if err != nil {
				return nil, err
			}
			mediaMap[doc.Ref.ID] = media
		}
	}
	return mediaMap, nil
}
