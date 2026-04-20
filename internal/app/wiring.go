package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wa_chat_service/config"
	handler_grpc "wa_chat_service/internal/handler/grpc"
	grpc_v1 "wa_chat_service/internal/handler/grpc/v1"
	handler_http "wa_chat_service/internal/handler/http"
	http_v1 "wa_chat_service/internal/handler/http/v1"
	"wa_chat_service/internal/repository"
	repository_firestore "wa_chat_service/internal/repository/firestore"
	repository_meili "wa_chat_service/internal/repository/meili"
	"wa_chat_service/internal/service"
	access_token_service "wa_chat_service/internal/service/access_token"
	encrypt_service "wa_chat_service/internal/service/encrypt"
	google_service "wa_chat_service/internal/service/google"
	jose_service "wa_chat_service/internal/service/jose"
	whatsapp_service "wa_chat_service/internal/service/whatsapp"
	"wa_chat_service/internal/usecase"
	activity_log_usecase "wa_chat_service/internal/usecase/activity_log"
	auth_usecase "wa_chat_service/internal/usecase/auth"
	broadcast_usecase "wa_chat_service/internal/usecase/broadcast"
	chat_usecase "wa_chat_service/internal/usecase/chat"
	message_usecase "wa_chat_service/internal/usecase/message"
	storage_media_usecase "wa_chat_service/internal/usecase/storage_media"
	template_usecase "wa_chat_service/internal/usecase/template"
	tenant_usecase "wa_chat_service/internal/usecase/tenant"
	server_grpc "wa_chat_service/pkg/server/grpc"
	grpc_middleware "wa_chat_service/pkg/server/grpc/middleware"
	server_http "wa_chat_service/pkg/server/http"
	http_middleware "wa_chat_service/pkg/server/http/middleware"
	"wa_chat_service/pkg/utils"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"
	"google.golang.org/api/cloudtasks/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type servers struct {
	grpc *server_grpc.Server
	http *server_http.Server
}

type clients struct {
	// DbPostgres *gorm.DB
	firebaseClient   *firebase.App
	firestoreClient  *firestore.Client
	gcpStorageClient *storage.Client
	txManager        *utils.TxManager
	gcpTaskClient    *cloudtasks.Service
	meiliClient      meilisearch.ServiceManager
}

type services struct {
	AccessToken   service.AccessToken
	Encrypt       service.Encrypt
	GoogleStorage service.GoogleStorage
	// GoogleFirebase *service.GoogleFirebase
	WhatsappBusiness service.WhatsappBusiness
	GoogleTask       service.GoogleTask
	JWT              service.JWT
}

type repositories struct {
	ActivityLog   repository.ActivityLog
	Chat          repository.Chat
	Message       repository.Message
	StorageMedia  repository.StorageMedia
	Tenant        repository.Tenant
	Template      repository.Template
	Broadcast     repository.Broadcast
	MeiliTemplate repository.MeiliTemplate
}

type usecases struct {
	ActivityLog  usecase.ActivityLog
	Chat         usecase.Chat
	Message      usecase.Message
	StorageMedia usecase.StorageMedia
	Tenant       usecase.Tenant
	Template     usecase.Template
	Broadcast    usecase.Broadcast
	Auth         usecase.Auth
}

func NewDefaultWiring(zslog *zap.SugaredLogger, config *config.Config) servers {
	clients := newDefaultClients(config)
	services := newDefaultServices(config, zslog, clients)
	repositories := newDefaultRepositories(clients, services)
	usecases := newDefaultUsecases(zslog, clients, repositories, services)
	servers := newDefaultServers(config, zslog, services, usecases)
	return servers
}

func newDefaultClients(config *config.Config) clients {
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
	txManager := utils.NewTxManager(nil, firestoreClient)
	gcpTaskClient, err := cloudtasks.NewService(context.Background())
	if err != nil {
		panic("Failed to create Cloud Tasks client: " + err.Error())
	}
	meiliClient := meilisearch.New(config.Meili.URL, meilisearch.WithAPIKey(config.Meili.APIKey))
	return clients{
		// DbPostgres: dbPostgres,
		firebaseClient:   firebaseClient,
		firestoreClient:  firestoreClient,
		gcpStorageClient: gcpStorageClient,
		txManager:        txManager,
		gcpTaskClient:    gcpTaskClient,
		meiliClient:      meiliClient,
	}
}

func newDefaultServices(config *config.Config, zslog *zap.SugaredLogger, clients clients) services {
	accessTokenService := access_token_service.NewAccessTokenService(config, zslog)
	encryptService := encrypt_service.NewEncryptService(&config.Encrypt, zslog)
	googleStorageService := google_service.NewGoogleStorageService(clients.gcpStorageClient, &config.GCP, zslog)
	// googleFirebaseService := google_service.NewGoogleFirebaseService(&config.GCP, clients.firebaseMessagingClient, clients.firebaseStorageClient)
	whatsappService := whatsapp_service.NewWhatsappService(zslog)
	jwtService := jose_service.NewJWTService(&config.JOSE, zslog)
	googleTaskService := google_service.NewGoogleTaskService(clients.gcpTaskClient, &config.GCP, jwtService, encryptService, zslog)
	return services{
		AccessToken:   accessTokenService,
		Encrypt:       encryptService,
		GoogleStorage: googleStorageService,
		// GoogleFirebase: googleFirebaseService,
		WhatsappBusiness: whatsappService,
		GoogleTask:       googleTaskService,
		JWT:              jwtService,
	}
}

func newDefaultRepositories(clients clients, services services) repositories {
	activityLogRepository := repository_firestore.NewActivityLogRepository(clients.firestoreClient)
	messageRepository := repository_firestore.NewMessageRepository(clients.firestoreClient, services.GoogleStorage)
	chatRepository := repository_firestore.NewChatRepository(clients.firestoreClient)
	storageMediaRepository := repository_firestore.NewStorageMediaRepository(clients.firestoreClient, services.GoogleStorage)
	tenantRepository := repository_firestore.NewTenantRepository(clients.firestoreClient)
	templateRepository := repository_firestore.NewTemplateRepository(clients.firestoreClient)
	broadcastRepository := repository_firestore.NewBroadcastRepository(clients.firestoreClient)
	meiliTemplateRepository := repository_meili.NewMeiliTemplateRepository(clients.meiliClient)
	return repositories{
		ActivityLog:   activityLogRepository,
		Chat:          chatRepository,
		Message:       messageRepository,
		StorageMedia:  storageMediaRepository,
		Template:      templateRepository,
		Tenant:        tenantRepository,
		Broadcast:     broadcastRepository,
		MeiliTemplate: meiliTemplateRepository,
	}
}

func newDefaultUsecases(zslog *zap.SugaredLogger, clients clients, repositories repositories, services services) usecases {
	activityLogUsecase := activity_log_usecase.NewActivityLogUsecase(repositories.ActivityLog, zslog)
	tenantUsecase := tenant_usecase.NewTenantUsecase(repositories.Tenant, services.Encrypt, zslog)
	templateUsecase := template_usecase.NewTemplateUsecase(repositories.Template, repositories.MeiliTemplate, tenantUsecase, services.WhatsappBusiness, clients.txManager, zslog)
	storageMediaUsecase := storage_media_usecase.NewStorageMediaUsecase(repositories.StorageMedia, tenantUsecase, services.GoogleStorage, services.WhatsappBusiness, zslog)
	messageUsecase := message_usecase.NewMessageUsecase(repositories.Message, repositories.Chat, repositories.StorageMedia, storageMediaUsecase, tenantUsecase, services.WhatsappBusiness, services.GoogleStorage, zslog)
	chatUsecase := chat_usecase.NewChatUsecase(repositories.Chat, zslog)
	broadcastUsecase := broadcast_usecase.NewBroadcastUsecase(repositories.Template, repositories.Broadcast, repositories.Tenant, messageUsecase, tenantUsecase, services.GoogleTask, services.WhatsappBusiness, clients.txManager, zslog)
	authUsecase := auth_usecase.NewAuthUsecase(repositories.Tenant, services.AccessToken, services.Encrypt, zslog)
	return usecases{
		ActivityLog:  activityLogUsecase,
		Template:     templateUsecase,
		Message:      messageUsecase,
		StorageMedia: storageMediaUsecase,
		Chat:         chatUsecase,
		Tenant:       tenantUsecase,
		Broadcast:    broadcastUsecase,
		Auth:         authUsecase,
	}
}

func newDefaultServers(config *config.Config, zslog *zap.SugaredLogger, services services, usecases usecases) servers {
	grpcServer := server_grpc.New(
		zslog.Desugar().Named("gRPC"),
		server_grpc.Port(fmt.Sprintf("%d", config.GRPC.Port)),
		server_grpc.ServerOptions(
			grpc.ChainUnaryInterceptor(
				grpc_middleware.UnaryRequestLogger(),
				grpc_middleware.TimingServerInterceptor(30*time.Second),
				grpc_middleware.HMACServerInterceptor(config.GRPC.Secret),
			),
		),
	)
	handlerGRPCV1 := grpc_v1.HandlerGRPCV1{
		StorageMedia: usecases.StorageMedia,
		ZSLog:        zslog,
	}
	handler_grpc.NewRouter(grpcServer.App, handlerGRPCV1)
	if config.App.Environment != "production" {
		reflection.Register(grpcServer.App)
	}

	httpServer := server_http.New(
		zslog.Desugar().Named("HTTP"),
		server_http.StructValidator(utils.NewStructValidator()),
		server_http.Port(fmt.Sprintf("%d", config.App.Port)),
		server_http.Middleware(
			http_middleware.Recover(zslog),
			http_middleware.OptionsRoute(),
			http_middleware.FileSizeLimit(16*1024*1024), // 16MB
		),
	)
	// Router Handler
	handlerHTTPV1 := http_v1.HandlerHTTPV1{
		ActivityLogUsecase:  usecases.ActivityLog,
		MessageUsecase:      usecases.Message,
		StorageMediaUsecase: usecases.StorageMedia,
		ChatUsecase:         usecases.Chat,
		TemplateUsecase:     usecases.Template,
		BroadcastUsecase:    usecases.Broadcast,
		TenantUsecase:       usecases.Tenant,
		AuthUsecase:         usecases.Auth,
		EncryptService:      services.Encrypt,
		JWTService:          services.JWT,
		AccessTokenService:  services.AccessToken,
		ZSLog:               zslog.Named("HTTP"),
	}
	handler_http.NewRouter(httpServer.App, config, handlerHTTPV1)
	return servers{
		grpc: grpcServer,
		http: httpServer,
	}
}

func (s *servers) startServers() {
	s.grpc.Start()
	s.http.Start()
}

func (s *servers) waitForShutdown(zlog *zap.Logger) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error

	select {
	case sig := <-interrupt:
		zlog.Info("app - Run - signal:" + sig.String())
	case err = <-s.http.Notify():
		zlog.Error("app - Run - httpServer.Notify", zap.Error(err))
	case err = <-s.grpc.Notify():
		zlog.Error("app - Run - grpcServer.Notify", zap.Error(err))
	}

	s.shutdownServers(zlog)
}

func (s *servers) shutdownServers(zlog *zap.Logger) {
	if err := s.http.Shutdown(); err != nil {
		zlog.Error("app - Run - httpServer.Shutdown", zap.Error(err))
	}

	if err := s.grpc.Shutdown(); err != nil {
		zlog.Error("app - Run - grpcServer.Shutdown", zap.Error(err))
	}
}
