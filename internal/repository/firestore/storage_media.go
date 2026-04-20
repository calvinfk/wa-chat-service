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

func (r *StorageMediaRepository) Insert(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) (model.StorageMedia, error) {
	var err error
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.db.Collection(r.storageMedia.TableName()).Doc(data.DocumentID)
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
	docRef := r.db.Collection(r.storageMedia.TableName()).Doc(documentID)
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

func (r *StorageMediaRepository) GetByURL(ctx context.Context, url string) (model.StorageMedia, error) {
	docs, err := r.db.Collection(r.storageMedia.TableName()).Where("url", "==", url).Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return r.storageMedia, err
	}
	if len(docs) == 0 {
		return r.storageMedia, errs.ErrGenericNotFound
	}
	var media model.StorageMedia
	for _, doc := range docs {
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &media)
		if err != nil {
			return r.storageMedia, err
		}
	}
	return media, nil
}

func (r *StorageMediaRepository) GetByAccessURL(ctx context.Context, accessURL string) (model.StorageMedia, error) {
	docs, err := r.db.Collection(r.storageMedia.TableName()).Where("access_url", "==", accessURL).Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return r.storageMedia, err
	}
	if len(docs) == 0 {
		return r.storageMedia, errs.ErrGenericNotFound
	}
	var media model.StorageMedia
	for _, doc := range docs {
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &media)
		if err != nil {
			return r.storageMedia, err
		}
	}
	return media, nil
}

func (r *StorageMediaRepository) GetByMediaID(ctx context.Context, mediaID string) (model.StorageMedia, error) {
	docs, err := r.db.Collection(r.storageMedia.TableName()).Where("media_id", "==", mediaID).Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return r.storageMedia, err
	}
	if len(docs) == 0 {
		return r.storageMedia, errs.ErrGenericNotFound
	}
	var media model.StorageMedia
	for _, doc := range docs {
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &media)
		if err != nil {
			return r.storageMedia, err
		}
	}
	return media, nil
}

func (r *StorageMediaRepository) Delete(ctx context.Context, tx *firestore.Transaction, documentID string) error {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.db.Collection(r.storageMedia.TableName()).Doc(documentID)
		return tx.Delete(docRef)
	}
	if tx != nil {
		return execDB(ctx, tx)
	}
	return r.db.RunTransaction(ctx, execDB)
}

func (r *StorageMediaRepository) UpdateAccessURL(ctx context.Context, tx *firestore.Transaction, documentID string, accessURL string) error {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		updates := make([]firestore.Update, 1)
		updates[0] = firestore.Update{Path: "access_url", Value: accessURL}
		docRef := r.db.Collection(r.storageMedia.TableName()).Doc(documentID)
		return tx.Update(docRef, updates)
	}
	if tx != nil {
		return execDB(ctx, tx)
	}
	return r.db.RunTransaction(ctx, execDB)
}

func (r *StorageMediaRepository) Update(ctx context.Context, tx *firestore.Transaction, data model.StorageMedia) error {
	execDB := func(ctx context.Context, tx *firestore.Transaction) error {
		docRef := r.db.Collection(r.storageMedia.TableName()).Doc(data.DocumentID)
		return tx.Set(docRef, data)
	}
	if tx != nil {
		return execDB(ctx, tx)
	}
	return r.db.RunTransaction(ctx, execDB)
}

func (r *StorageMediaRepository) GetFiltered(ctx context.Context, inputData filter_request.FilterRequest[dto.StorageMediaGetListRequest]) (filter_request.FilterResponse[dto.StorageMediaResponse], error) {
	var response filter_request.FilterResponse[dto.StorageMediaResponse]
	filters, sort, paginate, err := filter_request.InitializeFilter(inputData, r.storageMedia.AllowedFilterFields(), r.storageMedia.AllowedSortFields())
	if err != nil {
		return response, err
	}
	collection := r.db.Collection(r.storageMedia.TableName())
	query := collection.Query
	docs, totalData, err := filter_request.ApplyFilterFirestore(ctx, query, filters, sort, paginate)
	if err != nil {
		return response, err
	}
	var result []dto.StorageMediaResponse
	for _, doc := range docs {
		var media model.StorageMedia
		docData := doc.Data()
		docData["id"] = doc.Ref.ID
		err = utils.MapToStruct(docData, &media)
		if err != nil {
			return response, err
		}
		var accessURL *string
		if media.IsURLFromStorage && media.URL != nil {
			signedUrl, err := r.googleStorageService.GenerateV4GetObjectSignedURL(*media.URL, 0)
			if err != nil {
				return response, err
			}
			accessURL = &signedUrl
		} else if media.URL != nil {
			accessURL = media.URL
		}
		result = append(result, dto.StorageMediaResponse{}.FromModel(media, accessURL))
	}
	response = filter_request.NewFilterResponse(result, paginate, totalData)
	return response, nil
}

func (r *StorageMediaRepository) GetByDocumentIDs(ctx context.Context, documentIDs []string) (map[string]model.StorageMedia, error) {
	docs, err := r.db.Collection(r.storageMedia.TableName()).Where(firestore.DocumentID, "in", documentIDs).Documents(ctx).GetAll() // to ensure the documentIDs are valid and not exceeding Firestore limits
	if err != nil {
		return nil, err
	}
	mediaMap := make(map[string]model.StorageMedia)
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
	return mediaMap, nil
}
