package app

import (
	"context"
	"log"
	"wa_chat_service/config"
	"wa_chat_service/internal/repository"
	repository_firestore "wa_chat_service/internal/repository/firestore"
	"wa_chat_service/internal/service"
	access_token_service "wa_chat_service/internal/service/access_token"
	encrypt_service "wa_chat_service/internal/service/encrypt"
	google_service "wa_chat_service/internal/service/google"
	jose_service "wa_chat_service/internal/service/jose"
	whatsapp_service "wa_chat_service/internal/service/whatsapp"
	"wa_chat_service/internal/usecase"
	activity_log_usecase "wa_chat_service/internal/usecase/activity_log"
	broadcast_usecase "wa_chat_service/internal/usecase/broadcast"
	chat_usecase "wa_chat_service/internal/usecase/chat"
	message_usecase "wa_chat_service/internal/usecase/message"
	storage_media_usecase "wa_chat_service/internal/usecase/storage_media"
	template_usecase "wa_chat_service/internal/usecase/template"
	tenant_usecase "wa_chat_service/internal/usecase/tenant"
	"wa_chat_service/pkg/transaction"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/cloudtasks/v2"
)

type Clients struct {
	// DbPostgres *gorm.DB
	firebaseClient   *firebase.App
	firestoreClient  *firestore.Client
	gcpStorageClient *storage.Client
	txManager        *transaction.TxManager
	gcpTaskClient    *cloudtasks.Service
}

type Services struct {
	AccessToken   service.AccessToken
	Encrypt       service.Encrypt
	GoogleStorage service.GoogleStorage
	// GoogleFirebase *service.GoogleFirebase
	WhatsappBusiness service.WhatsappBusiness
	GoogleTask       service.GoogleTask
	JWT              service.JWT
}

type Repositories struct {
	ActivityLog  repository.ActivityLog
	Chat         repository.Chat
	Message      repository.Message
	StorageMedia repository.StorageMedia
	Tenant       repository.Tenant
	Template     repository.Template
}

type Usecases struct {
	ActivityLog  usecase.ActivityLog
	Chat         usecase.Chat
	Message      usecase.Message
	StorageMedia usecase.StorageMedia
	Tenant       usecase.Tenant
	Template     usecase.Template
	Broadcast    usecase.Broadcast
}

func NewDefaultWiring(config *config.Config) (Clients, Services, Repositories, Usecases) {
	clients := newDefaultClients(config)
	services := newDefaultServices(config, clients)
	repositories := newDefaultRepositories(clients, services)
	usecases := newDefaultUsecases(clients, repositories, services)
	return clients, services, repositories, usecases
}

func newDefaultClients(config *config.Config) Clients {
	ctx := context.Background()
	// dbPostgres, err := gorm.Open(postgres.Open(config.Database.URL), &gorm.Config{})
	// if err != nil {
	// 	panic("failed to open database: " + err.Error())
	// }
	conf := &firebase.Config{ProjectID: config.GCP.ProjectID, StorageBucket: config.GCP.ProjectID + ".firebasestorage.app"}
	firebaseClient, err := firebase.NewApp(ctx, conf)
	if err != nil {
		panic("Failed to create Firebase app: " + err.Error())
	}
	gcpStorageClient, err := storage.NewClient(ctx)
	if err != nil {
		panic("Failed to create Google Storage client: " + err.Error())
	}
	firestoreClient, err := firebaseClient.Firestore(context.Background())
	if err != nil {
		log.Fatalf("[ERROR][internal/app/app.go][Run] Failed to create Firestore client: %v", err)
	}
	// firebaseMessagingClient, err := firebaseClient.Messaging(context.Background())
	// if err != nil {
	// 	log.Fatalf("[ERROR][internal/app/app.go][Run] Failed to create Firebase Messaging client: %v", err)
	// }
	// firebaseStorageClient, err := firebaseClient.Storage(context.Background())
	// if err != nil {
	// 	log.Fatalf("[ERROR][internal/app/app.go][Run] Failed to create Firebase Storage client: %v", err)
	// }
	txManager := transaction.NewTxManager(nil, firestoreClient)
	gcpTaskClient, err := cloudtasks.NewService(context.Background())
	if err != nil {
		panic("Failed to create Cloud Tasks client: " + err.Error())
	}
	return Clients{
		// DbPostgres: dbPostgres,
		firebaseClient:   firebaseClient,
		firestoreClient:  firestoreClient,
		gcpStorageClient: gcpStorageClient,
		txManager:        txManager,
		gcpTaskClient:    gcpTaskClient,
	}
}

func newDefaultServices(config *config.Config, clients Clients) Services {
	accessTokenService := access_token_service.NewAccessTokenService(config)
	encryptService := encrypt_service.NewEncryptService(&config.Encrypt)
	googleStorageService := google_service.NewGoogleStorageService(clients.gcpStorageClient, &config.GCP)
	// googleFirebaseService := google_service.NewGoogleFirebaseService(&config.GCP, clients.firebaseMessagingClient, clients.firebaseStorageClient)
	whatsappService := whatsapp_service.NewWhatsappService()
	jwtService := jose_service.NewJWTService(&config.JOSE)
	googleTaskService := google_service.NewGoogleTaskService(clients.gcpTaskClient, &config.GCP, jwtService, encryptService)
	return Services{
		AccessToken:   accessTokenService,
		Encrypt:       encryptService,
		GoogleStorage: googleStorageService,
		// GoogleFirebase: googleFirebaseService,
		WhatsappBusiness: whatsappService,
		GoogleTask:       googleTaskService,
		JWT:              jwtService,
	}
}

func newDefaultRepositories(clients Clients, services Services) Repositories {
	activityLogRepository := repository_firestore.NewActivityLogRepository(clients.firestoreClient)
	messageRepository := repository_firestore.NewMessageRepository(clients.firestoreClient, services.GoogleStorage)
	chatRepository := repository_firestore.NewChatRepository(clients.firestoreClient)
	storageMediaRepository := repository_firestore.NewStorageMediaRepository(clients.firestoreClient, services.GoogleStorage)
	tenantRepository := repository_firestore.NewTenantRepository(clients.firestoreClient)
	templateRepository := repository_firestore.NewTemplateRepository(clients.firestoreClient)
	return Repositories{
		ActivityLog:  activityLogRepository,
		Chat:         chatRepository,
		Message:      messageRepository,
		StorageMedia: storageMediaRepository,
		Template:     templateRepository,
		Tenant:       tenantRepository,
	}
}

func newDefaultUsecases(clients Clients, repositories Repositories, services Services) Usecases {
	activityLogUsecase := activity_log_usecase.NewActivityLogUsecase(repositories.ActivityLog)
	tenantUsecase := tenant_usecase.NewTenantUsecase(repositories.Tenant, services.Encrypt)
	templateUsecase := template_usecase.NewTemplateUsecase(repositories.Template, tenantUsecase, services.WhatsappBusiness, clients.txManager)
	storageMediaUsecase := storage_media_usecase.NewStorageMediaUsecase(repositories.StorageMedia, repositories.Tenant, tenantUsecase, services.GoogleStorage, services.WhatsappBusiness)
	messageUsecase := message_usecase.NewMessageUsecase(repositories.Message, repositories.Chat, repositories.StorageMedia, storageMediaUsecase, tenantUsecase, services.WhatsappBusiness, services.GoogleStorage)
	chatUsecase := chat_usecase.NewChatUsecase(repositories.Chat)
	broadcastUsecase := broadcast_usecase.NewBroadcastUsecase(services.GoogleTask)
	return Usecases{
		ActivityLog:  activityLogUsecase,
		Template:     templateUsecase,
		Message:      messageUsecase,
		StorageMedia: storageMediaUsecase,
		Chat:         chatUsecase,
		Tenant:       tenantUsecase,
		Broadcast:    broadcastUsecase,
	}
}

func (c Clients) Shutdown() error {
	// if c.DbPostgres != nil {
	// 	log.Println("[INFO][internal/app/app.go][Clients.Shutdown] Closing database connection...")
	// 	if sqlDB, err := c.DbPostgres.DB(); err != nil {
	// 		log.Printf("[ERROR][internal/app/app.go][Clients.Shutdown] Error getting database connection: %v", err)
	// 	} else {
	// 		if err := sqlDB.Close(); err != nil {
	// 			log.Printf("[ERROR][internal/app/app.go][Clients.Shutdown] Error closing database connection: %v", err)
	// 		}
	// 	}
	// }
	return nil
}
