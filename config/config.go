package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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
		GRPC     GRPC
		Meili    Meili
	}

	// Application metadata and server configuration, such as name, version, environment, port, and URL.
	App struct {
		Name         string `env:"APP_NAME,required"`
		Version      string `env:"APP_VERSION,required"`
		Environment  string `env:"APP_ENVIRONMENT,required,oneof=development production staging"`
		Port         int    `env:"APP_PORT,required"`
		URL          string `env:"APP_URL,required"`
		PublicURL    string `env:"APP_PUBLIC_URL,default=APP_URL"`
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
		ProjectID           string `env:"GCP_PROJECT_ID,required"`
		BroadcastTaskParent string `env:"GCP_TASK_BROADCAST_PARENT,required"`
		DefaultBucket       string // default bucket for general use
	}

	JOSE struct {
		RSAPrivateKey     *rsa.PrivateKey
		AccessTokenExpiry time.Duration
	}

	GRPC struct {
		Port   int    `env:"GRPC_PORT,required"`
		Secret string `env:"GRPC_SECRET,required"`
	}

	Meili struct {
		URL    string `env:"MEILI_URL,required"`
		APIKey string `env:"MEILI_API_KEY,required"`
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

	grpc, err := grpcENV()
	if err != nil {
		return nil, err
	}
	config.GRPC = grpc

	meili, err := meiliEnv()
	if err != nil {
		return nil, err
	}
	config.Meili = meili
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
		PublicURL:    os.Getenv("APP_PUBLIC_URL"),
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
	if config.PublicURL == "" {
		config.PublicURL = config.URL
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
		ProjectID:           os.Getenv("GCP_PROJECT_ID"),
		BroadcastTaskParent: os.Getenv("GCP_TASK_BROADCAST_PARENT"),
	}
	errors := []string{}
	if config.ProjectID == "" {
		errors = append(errors, "GCP_PROJECT_ID is required")
	}
	if config.BroadcastTaskParent == "" {
		errors = append(errors, "GCP_TASK_BROADCAST_PARENT is required")
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
	if accessTokenExpiryStr := os.Getenv("JOSE_ACCESS_TOKEN_EXPIRY"); accessTokenExpiryStr != "" {
		accessTokenExpiry, err := time.ParseDuration(accessTokenExpiryStr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("JOSE_ACCESS_TOKEN_EXPIRY must be a valid duration string: %v", err))
		} else {
			cfg.AccessTokenExpiry = accessTokenExpiry
		}
	} else {
		errors = append(errors, "JOSE_ACCESS_TOKEN_EXPIRY is required")
	}
	if len(errors) > 0 {
		return JOSE{}, fmt.Errorf("missing required JOSE environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}

func loadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	// 600 permissions mean read/write for owner only
	// ensuring they are not accessible by other users on the system.
	file, err := os.OpenFile(keyPath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file: %w", err)
	}
	defer file.Close()

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	var privateKey *rsa.PrivateKey
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	switch block.Type {
	case "PRIVATE KEY":
		// Public Key Cryptography Standards (PKCS) #8 defines a standard syntax for storing private key information
		// including a private key for any public key algorithm and a set of attributes.
		// It is more flexible and can accommodate various types of private keys, including RSA, DSA, and EC keys.
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

func grpcENV() (GRPC, error) {
	var errors []string
	cfg := GRPC{
		Port:   0,
		Secret: os.Getenv("GRPC_SECRET"),
	}
	if portStr := os.Getenv("GRPC_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		} else {
			errors = append(errors, "GRPC_PORT must be a valid integer")
		}
	} else {
		errors = append(errors, "GRPC_PORT is required")
	}
	if cfg.Secret == "" {
		errors = append(errors, "GRPC_SECRET is required")
	}
	if len(errors) > 0 {
		return cfg, fmt.Errorf("missing required GRPC environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}

func meiliEnv() (Meili, error) {
	var errors []string
	cfg := Meili{
		URL:    os.Getenv("MEILI_URL"),
		APIKey: os.Getenv("MEILI_API_KEY"),
	}
	if cfg.URL == "" {
		errors = append(errors, "MEILI_URL is required")
	}
	if cfg.APIKey == "" {
		errors = append(errors, "MEILI_API_KEY is required")
	}
	if len(errors) > 0 {
		return cfg, fmt.Errorf("missing required Meili environment variables: %s", strings.Join(errors, ", "))
	}
	return cfg, nil
}
