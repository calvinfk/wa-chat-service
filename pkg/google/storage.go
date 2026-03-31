package google

import (
	"context"

	"cloud.google.com/go/storage"
)

func OpenGCPStorageConnection() *storage.Client {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic("Failed to create Google Storage client: " + err.Error())
	}
	return client
}
