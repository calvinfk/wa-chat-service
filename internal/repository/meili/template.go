package repository_meili

import (
	"context"
	"wa_chat_service/internal/model"

	"github.com/meilisearch/meilisearch-go"
)

type MeiliTemplateRepository struct {
	template model.Template
	client   meilisearch.ServiceManager
}

func NewMeiliTemplateRepository(client meilisearch.ServiceManager) *MeiliTemplateRepository {
	return &MeiliTemplateRepository{
		client: client,
	}
}

func (r *MeiliTemplateRepository) CreateIndex(ctx context.Context) (*meilisearch.TaskInfo, error) {
	taskInfo, err := r.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        r.template.TableName(),
		PrimaryKey: "id",
	})
	if err != nil {
		return nil, err
	}
	return taskInfo, nil
}

func (r *MeiliTemplateRepository) AddDocuments(ctx context.Context, documents []model.Template) (*meilisearch.TaskInfo, error) {
	primaryKey := r.template.PrimaryKey()
	taskInfo, err := r.client.Index(r.template.TableName()).AddDocuments(documents, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return nil, err
	}
	return taskInfo, nil
}

func (r *MeiliTemplateRepository) DeleteDocuments(ctx context.Context, documentIDs []string) (*meilisearch.TaskInfo, error) {
	primaryKey := r.template.PrimaryKey()
	taskInfo, err := r.client.Index(r.template.TableName()).DeleteDocuments(documentIDs, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return nil, err
	}
	return taskInfo, nil
}
