package google

import (
	"context"

	firebase "firebase.google.com/go/v4"
)

func OpenFirebaseConnection(projectID string) *firebase.App {
	// Use the application default credentials
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: projectID, StorageBucket: projectID + ".firebasestorage.app"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		panic("Failed to create Firebase app: " + err.Error())
	}
	return app
}
