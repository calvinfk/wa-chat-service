package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type (
	// Main configuration struct that holds all application settings, including app metadata, CORS settings, database connection info, and JOSE/JWT configuration.
	Config struct {
		App      App
		CORS     CORS
		Database Database
		Encrypt  Encrypt
		GCP      GCP
	}

	// Application metadata and server configuration, such as name, version, environment, port, and URL.
	App struct {
		Name         string `env:"APP_NAME,required"`
		Version      string `env:"APP_VERSION,required"`
		Environment  string `env:"APP_ENVIRONMENT,required,oneof=development production staging"`
		Port         int    `env:"APP_PORT,required"`
		URL          string `env:"APP_URL,required"`
		SecureCookie bool   `env:"APP_SECURE_COOKIE,default=false"`
	}

	// CORS configuration, including whether CORS is enabled, allowed origins, methods, headers, and credentials settings.
	CORS struct {
		CorsEnabled          bool     `env:"CORS_ENABLED,default=true"`
		CorsAllowOrigins     []string `env:"CORS_ALLOW_ORIGINS,default=*"`
		CorsAllowMethods     []string `env:"CORS_ALLOW_METHODS,default=GET,POST,PUT,OPTIONS"`
		CorsAllowHeaders     []string `env:"CORS_ALLOW_HEADERS,default=Origin,Content-Type,Authorization,X-Real-IP,X-Forwarded-For,X-Forwarded-Proto,X-Target-Host,X-Original-Host"`
		CorsExposeHeaders    []string `env:"CORS_EXPOSE_HEADERS,default=Content-Length"`
		CorsAllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS,default=true"`
	}

	// Database configuration, currently only includes the URL for connecting to the database.
	Database struct {
		URL string `env:"DATABASE_URL,required"`
	}

	// Encryption configuration for AES-GCM encryption. A key is required. IV is kept for backward compatibility with older deployments but is not required.
	Encrypt struct {
		Key []byte `env:"AES_ENCRYPTION_KEY,required"`
	}

	GCP struct {
		ProjectID            string `env:"GCP_PROJECT_ID,required"`
		TaskAPIKey           string `env:"GCP_TASK_API_KEY,required"`
		AttachmentLinkExpiry int    `env:"GCP_ATTACHMENT_LINK_EXPIRY,default=3600"` // in seconds
		AttachmentBucket     string `env:"GCP_ATTACHMENT_BUCKET,required"`
		AttachmentMaxSize    int64  `env:"GCP_ATTACHMENT_MAX_SIZE,default=10485760"` // in bytes (default 10 MB)
	}
)

// Reads environment variables and constructs a Config struct, validating required fields and returning an error if any are missing or invalid.
func New() (*Config, error) {
	cfg := &Config{}

	app, err := appEnv()
	if err != nil {
		return nil, err
	}
	cfg.App = app

	cors, err := corsEnv()
	if err != nil {
		return nil, err
	}
	cfg.CORS = cors

	database, err := databaseEnv()
	if err != nil {
		return nil, err
	}
	cfg.Database = database

	encrypt, err := encryptEnv()
	if err != nil {
		return nil, err
	}
	cfg.Encrypt = encrypt

	gcp, err := gcpEnv()
	if err != nil {
		return nil, err
	}
	cfg.GCP = gcp

	return cfg, nil
}

// Reads and validates application-related environment variables, returning an App struct or an error if any required variables are missing or invalid.
func appEnv() (App, error) {
	cfg := App{
		Name:         os.Getenv("APP_NAME"),
		Version:      os.Getenv("APP_VERSION"),
		Environment:  os.Getenv("APP_ENVIRONMENT"),
		Port:         0,
		URL:          os.Getenv("APP_URL"),
		SecureCookie: false,
	}
	errors := []string{}
	if cfg.Name == "" {
		errors = append(errors, "APP_NAME is required")
	}
	if cfg.Version == "" {
		errors = append(errors, "APP_VERSION is required")
	}
	if cfg.Environment == "" {
		errors = append(errors, "APP_ENVIRONMENT is required")
	}
	if cfg.Environment != "development" && cfg.Environment != "production" && cfg.Environment != "staging" {
		errors = append(errors, "APP_ENVIRONMENT must be either 'development', 'production', or 'staging'")
	}
	if appPortString := os.Getenv("APP_PORT"); appPortString != "" {
		appPort, err := strconv.Atoi(appPortString)
		if err != nil || appPort <= 0 {
			errors = append(errors, "APP_PORT must be a positive integer")
		} else {
			cfg.Port = appPort
		}
	} else {
		errors = append(errors, "APP_PORT is required")
	}
	if cfg.URL == "" {
		errors = append(errors, "APP_URL is required")
	}
	if secureCookieString := os.Getenv("APP_SECURE_COOKIE"); secureCookieString != "" {
		secureCookie, err := strconv.ParseBool(secureCookieString)
		if err != nil {
			errors = append(errors, "APP_SECURE_COOKIE must be a boolean value")
		} else {
			cfg.SecureCookie = secureCookie
		}
	}

	if len(errors) > 0 {
		return App{}, fmt.Errorf("missing required app environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}

// Reads and validates CORS-related environment variables, returning a CORS struct or an error if any required variables are missing or invalid.
func corsEnv() (CORS, error) {
	errors := []string{}
	if os.Getenv("CORS_ENABLED") == "" {
		errors = append(errors, "CORS_ENABLED is required")
	}
	cfg := CORS{
		CorsEnabled: os.Getenv("CORS_ENABLED") == "true",
	}
	if !cfg.CorsEnabled {
		cfg.CorsAllowOrigins = []string{}
		cfg.CorsAllowMethods = []string{}
		cfg.CorsAllowHeaders = []string{}
		cfg.CorsExposeHeaders = []string{}
	} else {
		cfg.CorsAllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
		if len(cfg.CorsAllowOrigins) == 0 {
			cfg.CorsAllowOrigins = []string{"*"}
		}
		cfg.CorsAllowMethods = strings.Split(os.Getenv("CORS_ALLOW_METHODS"), ",")
		if len(cfg.CorsAllowMethods) == 0 {
			cfg.CorsAllowMethods = []string{"GET", "POST", "PUT", "OPTIONS"}
		}
		cfg.CorsAllowHeaders = strings.Split(os.Getenv("CORS_ALLOW_HEADERS"), ",")
		if len(cfg.CorsAllowHeaders) == 0 {
			cfg.CorsAllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Real-IP", "X-Forwarded-For", "X-Forwarded-Proto", "X-Target-Host", "X-Original-Host", "Access-Control-Allow-Origin"}
		}
		cfg.CorsExposeHeaders = strings.Split(os.Getenv("CORS_EXPOSE_HEADERS"), ",")
		if len(cfg.CorsExposeHeaders) == 0 {
			cfg.CorsExposeHeaders = []string{"Content-Length"}
		}
		cfg.CorsAllowCredentials = os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"
	}
	if len(errors) > 0 {
		return CORS{}, fmt.Errorf("missing required CORS environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}

// Reads and validates database-related environment variables, returning a Database struct or an error if any required variables are missing or invalid.
func databaseEnv() (Database, error) {
	cfg := Database{
		URL: os.Getenv("DATABASE_URL"),
	}
	errors := []string{}
	if cfg.URL == "" {
		errors = append(errors, "DATABASE_URL is required")
	}
	if len(errors) > 0 {
		return Database{}, fmt.Errorf("missing required Database environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}

// Reads and validates AES encryption-related environment variables, returning an Encrypt struct or an error if any required variables are missing or invalid. For AES-GCM only AES_ENCRYPTION_KEY is required.
func encryptEnv() (Encrypt, error) {
	cfg := Encrypt{
		Key: []byte(os.Getenv("AES_ENCRYPTION_KEY")),
	}
	errors := []string{}
	if len(cfg.Key) == 0 {
		errors = append(errors, "AES_ENCRYPTION_KEY is required")
	}
	if len(errors) > 0 {
		return Encrypt{}, fmt.Errorf("missing required Encrypt environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}

func gcpEnv() (GCP, error) {
	cfg := GCP{
		ProjectID:            os.Getenv("GCP_PROJECT_ID"),
		TaskAPIKey:           os.Getenv("GCP_TASK_API_KEY"),
		AttachmentLinkExpiry: 3600, // default to 1 hour
		AttachmentBucket:     os.Getenv("GCP_ATTACHMENT_BUCKET"),
		AttachmentMaxSize:    10485760, // default to 10 MB
	}
	errors := []string{}
	if cfg.ProjectID == "" {
		errors = append(errors, "GCP_PROJECT_ID is required")
	}
	if cfg.TaskAPIKey == "" {
		errors = append(errors, "GCP_TASK_API_KEY is required")
	}
	if expiry := os.Getenv("GCP_ATTACHMENT_LINK_EXPIRY"); expiry != "" {
		if val, err := strconv.Atoi(expiry); err == nil && val >= -1 {
			cfg.AttachmentLinkExpiry = val
		} else {
			errors = append(errors, "GCP_ATTACHMENT_LINK_EXPIRY must be a positive integer or -1 to disable")
		}
	}
	if maxSize := os.Getenv("GCP_ATTACHMENT_MAX_SIZE"); maxSize != "" {
		if val, err := strconv.ParseInt(maxSize, 10, 64); err == nil && val >= 0 {
			cfg.AttachmentMaxSize = val
		} else {
			errors = append(errors, "GCP_ATTACHMENT_MAX_FILE must be a positive integer")
		}
	}
	if cfg.AttachmentBucket == "" {
		errors = append(errors, "GCP_ATTACHMENT_BUCKET is required")
	}
	if len(errors) > 0 {
		return GCP{}, fmt.Errorf("missing required GCP environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}
