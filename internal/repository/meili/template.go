package repository_meili

import (
	"context"
	"wa_chat_service/internal/model"

	"github.com/meilisearch/meilisearch-go"
)

type MeiliTemplateRepository struct {
	template model.Template
	db       meilisearch.ServiceManager
}

func NewMeiliTemplateRepository(db meilisearch.ServiceManager) *MeiliTemplateRepository {
	return &MeiliTemplateRepository{
		db: db,
	}
}

func (r *MeiliTemplateRepository) AddDocuments(ctx context.Context, documents []model.Template) error {
	primaryKey := r.template.PKName()
	_, err := r.db.Index(r.template.TableName()).AddDocuments(documents, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *MeiliTemplateRepository) DeleteDocuments(ctx context.Context, documentIDs []string) error {
	primaryKey := r.template.PKName()
	_, err := r.db.Index(r.template.TableName()).DeleteDocuments(documentIDs, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return err
	}
	return nil
}
