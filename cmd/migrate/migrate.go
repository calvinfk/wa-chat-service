package main

import (
	"io"
	"log"
	"os"
	"strings"
	"wa_chat_service/config"
	"wa_chat_service/internal/model"
	"wa_chat_service/pkg/database"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("APP_ENVIRONMENT") == "" || os.Getenv("APP_ENVIRONMENT") == "development" {
			log.Fatalf("Error loading .env file: %v, APP_ENVIRONMENT: %v", err, os.Getenv("APP_ENVIRONMENT"))
		}
	}
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()
	// MultiWriter sends logs to both stdout and file
	multi := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multi)

	config, err := config.New()
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}
	db := database.OpenPostgresConnection(config.Database.URL)
	models := []any{
		&model.ActivityLog{},
		&model.Tenant{},
		&model.WhatsappBusinessAccount{},
		&model.PhoneNumber{},
	}

	args := os.Args
	if len(args) == 1 {
		log.Printf("[INFO][cmd/migrate/migrate.go][main] No command provided. Use 'migrate', 'drop', or 'seed'.")
		return
	}
	for arg := range strings.SplitSeq(args[1], ",") {
		switch arg {
		case "migrate":
			autoMigrate(db, models)
		case "drop":
			dropTables(db, models)
		case "seed":
			seedData(db)
		default:
			log.Printf("[INFO][cmd/migrate/migrate.go][main] command '%s' not recognized", args[1])
		}
	}
}

func autoMigrate(db *gorm.DB, models []any) {
	log.Printf("[INFO][cmd/migrate/migrate.go][autoMigrate] Auto-migrating database...")
	err := db.AutoMigrate(models...)
	if err != nil {
		log.Printf("[ERROR][cmd/migrate/migrate.go][autoMigrate] Failed to auto-migrate database: %v", err)
	} else {
		log.Printf("[INFO][cmd/migrate/migrate.go][autoMigrate] Database migration completed successfully")
	}
}

func dropTables(db *gorm.DB, models []any) {
	log.Printf("[INFO][cmd/migrate/migrate.go][dropTables] Dropping tables...")
	err := db.Migrator().DropTable(models...)
	if err != nil {
		log.Printf("[ERROR][cmd/migrate/migrate.go][dropTables] Failed to drop tables: %v", err)
	} else {
		log.Printf("[INFO][cmd/migrate/migrate.go][dropTables] Tables dropped successfully")
	}
}

func seedData(db *gorm.DB) {
	log.Printf("[INFO][cmd/migrate/migrate.go][seedData] Seeding role data...")
	_ = db
	log.Printf("[INFO][cmd/migrate/migrate.go][seedData] Data seeding completed successfully")
}

func generateBcryptHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR][cmd/migrate/migrate.go][generateBcryptHash] Failed to generate bcrypt hash: %v", err)
		return ""
	}
	return string(hash)
}
