package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
		JOSE     JOSE
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
		ProjectID     string `env:"GCP_PROJECT_ID,required"`
		AppBaseURL    string `env:"GCP_APP_BASE_URL,required"`
		DefaultBucket string // default bucket for general use
	}

	JOSE struct {
		RSAPrivateKey *rsa.PrivateKey
	}
)

// Reads environment variables and constructs a Config struct, validating required fields and returning an error if any are missing or invalid.
func New() (*Config, error) {
	config := &Config{}

	app, err := appEnv()
	if err != nil {
		return nil, err
	}
	config.App = app

	cors, err := corsEnv()
	if err != nil {
		return nil, err
	}
	config.CORS = cors

	database, err := databaseEnv()
	if err != nil {
		return nil, err
	}
	config.Database = database

	encrypt, err := encryptEnv()
	if err != nil {
		return nil, err
	}
	config.Encrypt = encrypt

	gcp, err := gcpEnv()
	if err != nil {
		return nil, err
	}
	config.GCP = gcp

	jose, err := joseEnv()
	if err != nil {
		return nil, err
	}
	config.JOSE = jose

	return config, nil
}

// Reads and validates application-related environment variables, returning an App struct or an error if any required variables are missing or invalid.
func appEnv() (App, error) {
	config := App{
		Name:         os.Getenv("APP_NAME"),
		Version:      os.Getenv("APP_VERSION"),
		Environment:  os.Getenv("APP_ENVIRONMENT"),
		Port:         0,
		URL:          os.Getenv("APP_URL"),
		SecureCookie: false,
	}
	errors := []string{}
	if config.Name == "" {
		errors = append(errors, "APP_NAME is required")
	}
	if config.Version == "" {
		errors = append(errors, "APP_VERSION is required")
	}
	if config.Environment == "" {
		errors = append(errors, "APP_ENVIRONMENT is required")
	}
	if config.Environment != "development" && config.Environment != "production" && config.Environment != "staging" {
		errors = append(errors, "APP_ENVIRONMENT must be either 'development', 'production', or 'staging'")
	}
	if appPortString := os.Getenv("APP_PORT"); appPortString != "" {
		appPort, err := strconv.Atoi(appPortString)
		if err != nil || appPort <= 0 {
			errors = append(errors, "APP_PORT must be a positive integer")
		} else {
			config.Port = appPort
		}
	} else {
		errors = append(errors, "APP_PORT is required")
	}
	if config.URL == "" {
		errors = append(errors, "APP_URL is required")
	}
	if secureCookieString := os.Getenv("APP_SECURE_COOKIE"); secureCookieString != "" {
		secureCookie, err := strconv.ParseBool(secureCookieString)
		if err != nil {
			errors = append(errors, "APP_SECURE_COOKIE must be a boolean value")
		} else {
			config.SecureCookie = secureCookie
		}
	}

	if len(errors) > 0 {
		return App{}, fmt.Errorf("missing required app environment variables: %s", strings.Join(errors, ", "))
	}
	return config, nil
}

// Reads and validates CORS-related environment variables, returning a CORS struct or an error if any required variables are missing or invalid.
func corsEnv() (CORS, error) {
	errors := []string{}
	if os.Getenv("CORS_ENABLED") == "" {
		errors = append(errors, "CORS_ENABLED is required")
	}
	config := CORS{
		CorsEnabled: os.Getenv("CORS_ENABLED") == "true",
	}
	if !config.CorsEnabled {
		config.CorsAllowOrigins = []string{}
		config.CorsAllowMethods = []string{}
		config.CorsAllowHeaders = []string{}
		config.CorsExposeHeaders = []string{}
	} else {
		config.CorsAllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
		if len(config.CorsAllowOrigins) == 0 {
			config.CorsAllowOrigins = []string{"*"}
		}
		config.CorsAllowMethods = strings.Split(os.Getenv("CORS_ALLOW_METHODS"), ",")
		if len(config.CorsAllowMethods) == 0 {
			config.CorsAllowMethods = []string{"GET", "POST", "PUT", "OPTIONS"}
		}
		config.CorsAllowHeaders = strings.Split(os.Getenv("CORS_ALLOW_HEADERS"), ",")
		if len(config.CorsAllowHeaders) == 0 {
			config.CorsAllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Real-IP", "X-Forwarded-For", "X-Forwarded-Proto", "X-Target-Host", "X-Original-Host", "Access-Control-Allow-Origin"}
		}
		config.CorsExposeHeaders = strings.Split(os.Getenv("CORS_EXPOSE_HEADERS"), ",")
		if len(config.CorsExposeHeaders) == 0 {
			config.CorsExposeHeaders = []string{"Content-Length"}
		}
		config.CorsAllowCredentials = os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"
	}
	if len(errors) > 0 {
		return CORS{}, fmt.Errorf("missing required CORS environment variables: %s", strings.Join(errors, ", "))
	}
	return config, nil
}

// Reads and validates database-related environment variables, returning a Database struct or an error if any required variables are missing or invalid.
func databaseEnv() (Database, error) {
	config := Database{
		URL: os.Getenv("DATABASE_URL"),
	}
	errors := []string{}
	if config.URL == "" {
		errors = append(errors, "DATABASE_URL is required")
	}
	if len(errors) > 0 {
		return Database{}, fmt.Errorf("missing required Database environment variables: %s", strings.Join(errors, ", "))
	}
	return config, nil
}

// Reads and validates AES encryption-related environment variables, returning an Encrypt struct or an error if any required variables are missing or invalid. For AES-GCM only AES_ENCRYPTION_KEY is required.
func encryptEnv() (Encrypt, error) {
	config := Encrypt{
		Key: []byte(os.Getenv("AES_ENCRYPTION_KEY")),
	}
	errors := []string{}
	if len(config.Key) == 0 {
		errors = append(errors, "AES_ENCRYPTION_KEY is required")
	}
	if len(errors) > 0 {
		return Encrypt{}, fmt.Errorf("missing required Encrypt environment variables: %s", strings.Join(errors, ", "))
	}
	return config, nil
}

func gcpEnv() (GCP, error) {
	config := GCP{
		ProjectID:  os.Getenv("GCP_PROJECT_ID"),
		AppBaseURL: os.Getenv("GCP_APP_BASE_URL"),
	}
	errors := []string{}
	if config.ProjectID == "" {
		errors = append(errors, "GCP_PROJECT_ID is required")
	}
	if config.AppBaseURL == "" {
		errors = append(errors, "GCP_APP_BASE_URL is required")
	}
	config.DefaultBucket = config.ProjectID + ".firebasestorage.app"
	if len(errors) > 0 {
		return GCP{}, fmt.Errorf("missing required GCP environment variables: %s", strings.Join(errors, ", "))
	}
	return config, nil
}
func joseEnv() (JOSE, error) {
	var errors []string
	cfg := JOSE{}
	if encodedKey := os.Getenv("JOSE_RSA_PRIVATE_KEY"); encodedKey != "" {
		var err error
		cfg.RSAPrivateKey, err = loadPrivateKey(encodedKey)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to load RSA private key: %v", err))
		}
	} else {
		errors = append(errors, "JOSE_RSA_PRIVATE_KEY is required")
	}
	if len(errors) > 0 {
		return JOSE{}, fmt.Errorf("missing required JOSE environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}

func loadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	file, err := os.OpenFile(keyPath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file: %w", err)
	}
	defer file.Close()
	data := make([]byte, 2048) // Read up to 2KB, adjust if needed
	_, err = file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	var privateKey *rsa.PrivateKey
	block, _ := pem.Decode(data)
	switch block.Type {
	case "RSA PRIVATE KEY":
		// PKCS#1
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)
		}
	case "PRIVATE KEY":
		// PKCS#8
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not RSA private key")
		}
	default:
		return nil, fmt.Errorf("unsupported key type %q", block.Type)
	}
	return privateKey, nil
}
