package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"wa_chat_service/config"
	http_internal "wa_chat_service/internal/handler/http"
	http_v1 "wa_chat_service/internal/handler/http/v1"
	"wa_chat_service/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

func Run(config *config.Config) {
	log.Printf("[INFO][internal/app/app/Run] Starting %s, version %s, in %s mode", config.App.Name, config.App.Version, config.App.Environment)

	clients, services, _, usecases := NewDefaultWiring(config)

	// Router Handler
	routerHandlerV1 := http_v1.RouterHandlerV1{
		ActivityLogUsecase:  usecases.ActivityLog,
		MessageUsecase:      usecases.Message,
		StorageMediaUsecase: usecases.StorageMedia,
		ChatUsecase:         usecases.Chat,
		TemplateUsecase:     usecases.Template,
		BroadcastUsecase:    usecases.Broadcast,
		EncryptService:      services.Encrypt,
		JWTService:          services.JWT,
	}

	app := fiber.New(fiber.Config{
		AppName:         config.App.Name,
		BodyLimit:       16 * 1024 * 1024, // 16MB
		StructValidator: utils.NewStructValidator(),
	})

	http_internal.NewRouter(app, config, routerHandlerV1)
	go func() {
		if err := app.Listen(":" + fmt.Sprintf("%d", config.App.Port)); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("[ERROR][internal/app/app/Run] Failed to start server: %v", err)
		}
	}()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	log.Printf("[INFO][internal/app/app/Run] Server is running on %s", fmt.Sprintf(":%d", config.App.Port))
	<-interrupt
	log.Println("[INFO][internal/app/app/Run] Shutting down server...")
	// Shutdown gracefully
	if err := app.Shutdown(); err != nil {
		log.Printf("[ERROR][internal/app/app/Run] Server forced to shutdown: %v", err)
	}
	// Shutdown clients
	if err := clients.Shutdown(); err != nil {
		log.Printf("[ERROR][internal/app/app/Run] Failed to close Firestore client: %v", err)
	}
	log.Println("[INFO][internal/app/app/Run] Server exiting")
}
