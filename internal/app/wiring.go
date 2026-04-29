package app

import (
	"context"
	"fmt"
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
	auth_usecase "wa_chat_service/internal/usecase/auth"
	broadcast_usecase "wa_chat_service/internal/usecase/broadcast"
	chat_usecase "wa_chat_service/internal/usecase/chat"
	message_usecase "wa_chat_service/internal/usecase/message"
	storage_media_usecase "wa_chat_service/internal/usecase/storage_media"
	template_usecase "wa_chat_service/internal/usecase/template"
	tenant_usecase "wa_chat_service/internal/usecase/tenant"
	wa_business_account_usecase "wa_chat_service/internal/usecase/wa_business_account"
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
	firebaseClient   *firebase.App
	firestoreClient  *firestore.Client
	gcpStorageClient *storage.Client
	txManager        *utils.TxManager
	gcpTaskClient    *cloudtasks.Service
	meiliClient      meilisearch.ServiceManager
}

type services struct {
	AccessToken      service.AccessToken
	Encrypt          service.Encrypt
	GoogleStorage    service.GoogleStorage
	WhatsappBusiness service.WhatsappBusiness
	GoogleTask       service.GoogleTask
	JWT              service.JWT
}

type repositories struct {
	Chat              repository.Chat
	Message           repository.Message
	StorageMedia      repository.StorageMedia
	Tenant            repository.Tenant
	Template          repository.Template
	Broadcast         repository.Broadcast
	SearchTemplate    repository.SearchTemplate
	SearchMessage     repository.SearchMessage
	User              repository.User
	WaBusinessAccount repository.WaBusinessAccount
	WaPhone           repository.WaPhone
}

type usecases struct {
	Chat              usecase.Chat
	Message           usecase.Message
	StorageMedia      usecase.StorageMedia
	Tenant            usecase.Tenant
	Template          usecase.Template
	Broadcast         usecase.Broadcast
	Auth              usecase.Auth
	WaBusinessAccount usecase.WaBusinessAccount
}

func NewDefaultWiring(zsLog *zap.SugaredLogger, cfg *config.Config) servers {
	clients := newDefaultClients(cfg, zsLog)
	services := newDefaultServices(cfg, zsLog, clients)
	repositories := newDefaultRepositories(clients, services, zsLog)
	usecases := newDefaultUsecases(cfg, zsLog, clients, repositories, services)
	servers := newDefaultServers(cfg, zsLog, services, usecases)
	return servers
}

func newDefaultClients(cfg *config.Config, zsLog *zap.SugaredLogger) clients {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	gcpStorageClient, err := storage.NewClient(ctx)
	if err != nil {
		zsLog.Fatalf("Failed to create Google Storage client: " + err.Error())
	}
	conf := &firebase.Config{ProjectID: cfg.GCP.ProjectID, StorageBucket: cfg.GCP.ProjectID + ".firebasestorage.app"}
	firebaseClient, err := firebase.NewApp(ctx, conf)
	if err != nil {
		zsLog.Fatalf("Failed to create Firebase app: " + err.Error())
	}
	firestoreClient, err := firebaseClient.Firestore(ctx)
	if err != nil {
		zsLog.Fatalf("Failed to create Firestore client: %v", err)
	}
	txManager := utils.NewTxManager(nil, firestoreClient)
	gcpTaskClient, err := cloudtasks.NewService(ctx)
	if err != nil {
		zsLog.Fatalf("Failed to create Cloud Tasks client: " + err.Error())
	}
	meiliClient := meilisearch.New(cfg.Meili.URL, meilisearch.WithAPIKey(cfg.Meili.APIKey))
	return clients{
		firebaseClient:   firebaseClient,
		firestoreClient:  firestoreClient,
		gcpStorageClient: gcpStorageClient,
		txManager:        txManager,
		gcpTaskClient:    gcpTaskClient,
		meiliClient:      meiliClient,
	}
}

func newDefaultServices(cfg *config.Config, zsLog *zap.SugaredLogger, clients clients) services {
	accessTokenService := access_token_service.NewAccessTokenService(cfg, zsLog)
	encryptService := encrypt_service.NewEncryptService(&cfg.Encrypt, zsLog)
	googleStorageService := google_service.NewGoogleStorageService(clients.gcpStorageClient, &cfg.GCP, zsLog)
	whatsappService := whatsapp_service.NewWhatsappService(zsLog)
	jwtService := jose_service.NewJWTService(&cfg.JOSE, zsLog)
	googleTaskService := google_service.NewGoogleTaskService(clients.gcpTaskClient, &cfg.GCP, jwtService, encryptService, zsLog, cfg.App.PublicURL)
	return services{
		AccessToken:      accessTokenService,
		Encrypt:          encryptService,
		GoogleStorage:    googleStorageService,
		WhatsappBusiness: whatsappService,
		GoogleTask:       googleTaskService,
		JWT:              jwtService,
	}
}

func newDefaultRepositories(clients clients, services services, zsLog *zap.SugaredLogger) repositories {
	messageRepository := repository_firestore.NewMessageRepository(clients.firestoreClient, services.GoogleStorage)
	chatRepository := repository_firestore.NewChatRepository(clients.firestoreClient)
	storageMediaRepository := repository_firestore.NewStorageMediaRepository(clients.firestoreClient, services.GoogleStorage)
	tenantRepository := repository_firestore.NewTenantRepository(clients.firestoreClient)
	templateRepository := repository_firestore.NewTemplateRepository(clients.firestoreClient)
	broadcastRepository := repository_firestore.NewBroadcastRepository(clients.firestoreClient)
	meiliTemplateRepository := repository_meili.NewMeiliTemplateRepository(clients.meiliClient, zsLog)
	meiliMessageRepository := repository_meili.NewMeiliMessageRepository(clients.meiliClient, zsLog)
	userRepository := repository_firestore.NewUserRepository(clients.firestoreClient)
	whatsappBusinessAccountRepository := repository_firestore.NewWhatsappBusinessAccountRepository(clients.firestoreClient)
	waPhoneRepository := repository_firestore.NewWaPhoneRepository(clients.firestoreClient)
	return repositories{
		Chat:              chatRepository,
		Message:           messageRepository,
		StorageMedia:      storageMediaRepository,
		Template:          templateRepository,
		Tenant:            tenantRepository,
		Broadcast:         broadcastRepository,
		SearchTemplate:    meiliTemplateRepository,
		SearchMessage:     meiliMessageRepository,
		User:              userRepository,
		WaBusinessAccount: whatsappBusinessAccountRepository,
		WaPhone:           waPhoneRepository,
	}
}

func newDefaultUsecases(cfg *config.Config, zsLog *zap.SugaredLogger, clients clients, repositories repositories, services services) usecases {
	waBusinessAccountUsecase := wa_business_account_usecase.NewWaBusinessAccountUsecase(repositories.WaBusinessAccount, services.Encrypt, repositories.WaPhone, zsLog)
	tenantUsecase := tenant_usecase.NewTenantUsecase(repositories.Tenant, repositories.WaBusinessAccount, repositories.WaPhone, services.Encrypt, zsLog)
	templateUsecase := template_usecase.NewTemplateUsecase(repositories.Template, repositories.SearchTemplate, repositories.WaBusinessAccount, waBusinessAccountUsecase, services.WhatsappBusiness, clients.txManager, zsLog)
	storageMediaUsecase := storage_media_usecase.NewStorageMediaUsecase(repositories.StorageMedia, waBusinessAccountUsecase, services.GoogleStorage, services.WhatsappBusiness, services.Encrypt, zsLog, cfg.App.PublicURL)
	messageUsecase := message_usecase.NewMessageUsecase(repositories.Message, repositories.Chat, repositories.StorageMedia, repositories.SearchMessage, repositories.Tenant, storageMediaUsecase, waBusinessAccountUsecase, services.WhatsappBusiness, services.GoogleStorage, clients.txManager, zsLog)
	chatUsecase := chat_usecase.NewChatUsecase(repositories.Chat, repositories.WaPhone, repositories.WaBusinessAccount, repositories.User, repositories.Message, clients.txManager, zsLog)
	broadcastUsecase := broadcast_usecase.NewBroadcastUsecase(repositories.Template, repositories.Broadcast, repositories.Tenant, messageUsecase, waBusinessAccountUsecase, services.GoogleTask, services.WhatsappBusiness, clients.txManager, zsLog)
	authUsecase := auth_usecase.NewAuthUsecase(repositories.User, repositories.Tenant, services.AccessToken, services.Encrypt, zsLog)
	return usecases{
		Template:          templateUsecase,
		Message:           messageUsecase,
		StorageMedia:      storageMediaUsecase,
		Chat:              chatUsecase,
		Tenant:            tenantUsecase,
		Broadcast:         broadcastUsecase,
		Auth:              authUsecase,
		WaBusinessAccount: waBusinessAccountUsecase,
	}
}

func newDefaultServers(cfg *config.Config, zsLog *zap.SugaredLogger, services services, usecases usecases) servers {
	grpcServer := server_grpc.New(
		zsLog.Desugar().Named("gRPC"),
		server_grpc.Port(fmt.Sprintf("%d", cfg.GRPC.Port)),
		server_grpc.ServerOptions(
			grpc.ChainUnaryInterceptor(
				grpc_middleware.UnaryRequestLogger(),
				grpc_middleware.TimingServerInterceptor(30*time.Second),
				grpc_middleware.HMACServerInterceptor(cfg.GRPC.Secret),
			),
		),
	)
	handlerGRPCV1 := grpc_v1.HandlerGRPCV1{
		App:               grpcServer.App,
		StorageMedia:      usecases.StorageMedia,
		Message:           usecases.Message,
		WaBusinessAccount: usecases.WaBusinessAccount,
		Chat:              usecases.Chat,
		ZSLog:             zsLog,
	}
	handler_grpc.NewRouter(handlerGRPCV1)
	if cfg.App.Environment != "production" {
		reflection.Register(grpcServer.App)
	}

	httpServer := server_http.New(
		zsLog.Desugar().Named("HTTP"),
		server_http.StructValidator(utils.NewStructValidator()),
		server_http.Port(fmt.Sprintf("%d", cfg.App.Port)),
		server_http.Middleware(
			http_middleware.Recover(zsLog),
			http_middleware.OptionsRoute(),
			http_middleware.FileSizeLimit(16*1024*1024), // 16MB
		),
	)
	// Router Handler
	handlerHTTPV1 := http_v1.HandlerHTTPV1{
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
		ZSLog:               zsLog.Named("HTTP"),
	}
	handler_http.NewRouter(httpServer.App, cfg, handlerHTTPV1)
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
