package api

import (
	"net/http"
	"time"

	"github.com/jesus/FCCUR/internal/auth"
	"github.com/jesus/FCCUR/internal/storage"
)

type Server struct {
	db           storage.Database
	packagesDir  string
	webDir       string
	mux          *http.ServeMux
	startTime    time.Time
	authConfig   AuthConfig
	rateLimiter  *RateLimiter
	cache        *PackageCache
	jwtManager   *auth.JWTManager
	oauth2Config *auth.OAuth2Config
}

// NewServer creates a new API server
func NewServer(db storage.Database, packagesDir, webDir, jwtSecret string) *Server {
	// Default JWT expiration: 24 hours for access, 30 days for refresh
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour, 30*24*time.Hour)

	s := &Server{
		db:          db,
		packagesDir: packagesDir,
		webDir:      webDir,
		mux:         http.NewServeMux(),
		startTime:   time.Now(),
		authConfig:  AuthConfig{Enabled: false}, // Disabled by default
		cache:       NewPackageCache(),
		jwtManager:  jwtManager,
	}

	s.setupRoutes()
	return s
}

// SetAuth configures authentication for upload endpoint
func (s *Server) SetAuth(username, password string) {
	if username != "" && password != "" {
		s.authConfig = AuthConfig{
			Username: username,
			Password: password,
			Enabled:  true,
		}
	}
}

// SetRateLimit configures rate limiting for uploads
// limit: max uploads per hour per IP
func (s *Server) SetRateLimit(limit int) {
	if limit > 0 {
		s.rateLimiter = NewRateLimiter(limit, time.Hour)
	}
}

// SetOAuth2 configures OAuth2 authentication
func (s *Server) SetOAuth2(config *auth.OAuth2Config) {
	s.oauth2Config = config
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Authentication routes
	s.mux.HandleFunc("/api/auth/register", s.withCORS(s.withLogging(s.Register)))
	s.mux.HandleFunc("/api/auth/login", s.withCORS(s.withLogging(s.Login)))
	s.mux.HandleFunc("/api/auth/logout", s.withCORS(s.withLogging(s.Logout)))
	s.mux.HandleFunc("/api/auth/logout-all", s.withCORS(s.withLogging(s.LogoutAll)))
	s.mux.HandleFunc("/api/auth/refresh", s.withCORS(s.withLogging(s.RefreshToken)))
	s.mux.HandleFunc("/api/auth/me", s.withCORS(s.withLogging(s.GetCurrentUser)))
	s.mux.HandleFunc("/api/auth/change-password", s.withCORS(s.withLogging(s.ChangePassword)))
	s.mux.HandleFunc("/api/auth/request-reset", s.withCORS(s.withLogging(s.RequestPasswordReset)))
	s.mux.HandleFunc("/api/auth/reset-password", s.withCORS(s.withLogging(s.ResetPassword)))

	// OAuth2 routes
	s.mux.HandleFunc("/api/oauth2/config", s.withCORS(s.withLogging(s.OAuth2Config)))
	s.mux.HandleFunc("/api/oauth2/login", s.withCORS(s.withLogging(s.OAuth2Login)))
	s.mux.HandleFunc("/api/oauth2/callback", s.withCORS(s.withLogging(s.OAuth2Callback)))

	// API routes with gzip compression
	s.mux.HandleFunc("/api/packages", s.withCORS(s.withLogging(s.withGzip(s.GetPackages))))
	s.mux.HandleFunc("/api/packages/", s.withCORS(s.withLogging(s.withGzip(s.GetPackage))))
	// Upload endpoint with rate limiting and RBAC (admin or professor only)
	s.mux.HandleFunc("/api/upload", s.withCORS(s.withLogging(s.withRateLimit(s.withCanUpload(s.UploadPackage)))))
	// Delete endpoint requires admin role
	s.mux.HandleFunc("/api/delete", s.withCORS(s.withLogging(s.withCanDelete(s.DeletePackage))))
	// Duplicate check endpoint
	s.mux.HandleFunc("/api/check-duplicate", s.withCORS(s.withLogging(s.withGzip(s.CheckDuplicate))))
	// Checksum download endpoints
	s.mux.HandleFunc("/api/checksum", s.withCORS(s.withLogging(s.DownloadChecksum)))
	s.mux.HandleFunc("/api/checksums/all", s.withCORS(s.withLogging(s.DownloadAllChecksums)))
	// Archive preview endpoint
	s.mux.HandleFunc("/api/archive/contents", s.withCORS(s.withLogging(s.withGzip(s.GetArchiveContents))))
	// Thumbnail endpoint
	s.mux.HandleFunc("/api/thumbnail", s.withCORS(s.withLogging(s.ServeThumbnail)))
	s.mux.HandleFunc("/download/", s.withCORS(s.withLogging(s.DownloadPackage)))
	s.mux.HandleFunc("/api/stats", s.withCORS(s.withLogging(s.withGzip(s.GetStats))))
	s.mux.HandleFunc("/health", s.withGzip(s.Health))

	// Static files
	s.mux.Handle("/", http.FileServer(http.Dir(s.webDir)))
}

// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply privacy headers to all requests
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-DNS-Prefetch-Control", "off")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), interest-cohort=()")
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data:; "+
			"font-src 'self'; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'; "+
			"form-action 'self'")

	s.mux.ServeHTTP(w, r)
}
