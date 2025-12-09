package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jesus/FCCUR/internal/api"
	"github.com/jesus/FCCUR/internal/auth"
	"github.com/jesus/FCCUR/internal/storage"
)

func main() {
	// Command-line flags with environment variable fallbacks
	addr := flag.String("addr", getEnv("FCCUR_ADDR", ":8080"), "HTTP server address")
	dbPath := flag.String("db", getEnv("FCCUR_DB", "./data/fccur.db"), "Database connection string (SQLite path or PostgreSQL URL)")
	packagesDir := flag.String("packages", getEnv("FCCUR_PACKAGES_DIR", "./packages"), "Packages directory")
	webDir := flag.String("web", getEnv("FCCUR_WEB_DIR", "./web"), "Web files directory")
	migrationsDir := flag.String("migrations", getEnv("FCCUR_MIGRATIONS_DIR", "./migrations"), "Migrations directory")
	certFile := flag.String("cert", getEnv("FCCUR_CERT_FILE", ""), "TLS certificate file (enables HTTPS)")
	keyFile := flag.String("key", getEnv("FCCUR_KEY_FILE", ""), "TLS private key file (enables HTTPS)")
	authUser := flag.String("auth-user", getEnv("FCCUR_AUTH_USER", ""), "Upload authentication username (optional)")
	authPass := flag.String("auth-pass", getEnv("FCCUR_AUTH_PASS", ""), "Upload authentication password (optional)")
	rateLimit := flag.Int("rate-limit", getEnvAsInt("FCCUR_RATE_LIMIT", 10), "Upload rate limit per IP (uploads per hour, 0 to disable)")
	jwtSecret := flag.String("jwt-secret", getEnv("FCCUR_JWT_SECRET", ""), "JWT secret key (auto-generated if not provided)")
	oauth2ClientID := flag.String("oauth2-client-id", getEnv("FCCUR_OAUTH2_CLIENT_ID", ""), "OAuth2 client ID (Microsoft/Azure AD)")
	oauth2ClientSecret := flag.String("oauth2-client-secret", getEnv("FCCUR_OAUTH2_CLIENT_SECRET", ""), "OAuth2 client secret")
	oauth2RedirectURL := flag.String("oauth2-redirect", getEnv("FCCUR_OAUTH2_REDIRECT_URL", "http://localhost:8080/api/oauth2/callback"), "OAuth2 redirect URL")
	oauth2Tenant := flag.String("oauth2-tenant", getEnv("FCCUR_OAUTH2_TENANT", "common"), "Microsoft tenant ID (common for multi-tenant)")
	flag.Parse()

	// Ensure directories exist
	if err := os.MkdirAll(*packagesDir, 0755); err != nil {
		log.Fatalf("Failed to create packages directory: %v", err)
	}

	// Create data directory for SQLite (if using SQLite)
	dbType := storage.DetectDatabaseType(*dbPath)
	if dbType == storage.DatabaseSQLite {
		dbDir := "./data"
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
	}

	// Initialize database (auto-detects SQLite or PostgreSQL)
	db, err := storage.NewDatabase(*dbPath)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Log database type
	log.Printf("Database: %s", storage.GetDatabaseInfo(db))

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// Generate JWT secret if not provided
	secret := *jwtSecret
	if secret == "" {
		// Generate a random 32-byte secret
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("Error generating JWT secret: %v", err)
		}
		secret = base64.StdEncoding.EncodeToString(b)
		log.Printf("WARNING: Auto-generated JWT secret. Sessions will be invalid after restart.")
		log.Printf("Use -jwt-secret flag to set a persistent secret.")
	}

	// Store migrations path for database operations
	storage.SetMigrationsPath(*migrationsDir)

	// Create API server
	server := api.NewServer(db, *packagesDir, *webDir, secret)

	// Configure authentication if provided
	if *authUser != "" && *authPass != "" {
		server.SetAuth(*authUser, *authPass)
		log.Printf("Upload authentication enabled for user: %s", *authUser)
	} else if *authUser != "" || *authPass != "" {
		log.Printf("Warning: Both -auth-user and -auth-pass must be provided for authentication")
	}

	// Configure rate limiting
	if *rateLimit > 0 {
		server.SetRateLimit(*rateLimit)
		log.Printf("Upload rate limiting enabled: %d uploads per hour per IP", *rateLimit)
	} else {
		log.Printf("Upload rate limiting disabled")
	}

	// Configure OAuth2
	if *oauth2ClientID != "" && *oauth2ClientSecret != "" {
		oauth2Config := auth.NewMicrosoftOAuth2Config(
			*oauth2ClientID,
			*oauth2ClientSecret,
			*oauth2RedirectURL,
			*oauth2Tenant,
		)
		server.SetOAuth2(oauth2Config)
		log.Printf("OAuth2 (Microsoft/Azure AD) enabled")
		log.Printf("OAuth2 Redirect URL: %s", *oauth2RedirectURL)
		log.Printf("OAuth2 Tenant: %s", *oauth2Tenant)
	} else {
		log.Printf("OAuth2 disabled (provide -oauth2-client-id and -oauth2-client-secret to enable)")
	}

	// Start server
	log.Printf("FCCUR server starting on %s", *addr)
	log.Printf("Packages directory: %s", *packagesDir)
	log.Printf("Database: %s", *dbPath)

	// Check if TLS is enabled
	if *certFile != "" && *keyFile != "" {
		// HTTPS mode
		log.Printf("TLS enabled with cert: %s", *certFile)
		log.Printf("Access at: https://localhost%s", *addr)
		if err := http.ListenAndServeTLS(*addr, *certFile, *keyFile, server); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		// HTTP mode
		if *certFile != "" || *keyFile != "" {
			log.Printf("Warning: Both -cert and -key must be provided for HTTPS")
		}
		log.Printf("Access at: http://localhost%s", *addr)
		if err := http.ListenAndServe(*addr, server); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
