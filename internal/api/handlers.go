package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nstoddard1002/traceless/internal/db"
)

type rateLimiter struct {
	counts map[string]int
	mu     sync.Mutex
}

var limiter = &rateLimiter{
	counts: make(map[string]int),
}

// Reset counts every 10 seconds
func init() {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			limiter.mu.Lock()
			limiter.counts = make(map[string]int)
			limiter.mu.Unlock()
		}
	}()
}

type Server struct {
	db *db.DB
}

func NewServer(database *db.DB) *Server {
	return &Server{db: database}
}

type CreateSecretRequest struct {
	Ciphertext      string `json:"ciphertext"`
	ExpirationMin   int    `json:"expiration_minutes"`
	ViewLimit       int    `json:"view_limit"`
	PasscodeEnabled bool   `json:"passcode_enabled"`
	Salt            string `json:"salt"`
}

type CreateSecretResponse struct {
	ID string `json:"id"`
}

func (s *Server) CreateSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	var req CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body or size too large", http.StatusBadRequest)
		return
	}

	if req.Ciphertext == "" {
		http.Error(w, "Ciphertext is required", http.StatusBadRequest)
		return
	}
	if req.ExpirationMin <= 0 {
		req.ExpirationMin = 1440
	}
	if req.ViewLimit <= 0 {
		req.ViewLimit = 1
	}

	id, err := GenerateID()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	secret := &db.Secret{
		ID:              id,
		Ciphertext:      req.Ciphertext,
		ExpiresAt:       time.Now().Add(time.Duration(req.ExpirationMin) * time.Minute),
		RemainingViews:  req.ViewLimit,
		PasscodeEnabled: req.PasscodeEnabled,
		Salt:            req.Salt,
	}

	if err := s.db.CreateSecret(r.Context(), secret); err != nil {
		http.Error(w, "Failed to store secret", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateSecretResponse{ID: id})
}

func (s *Server) GetSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[4]

	secret, err := s.db.GetSecret(r.Context(), id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if secret == nil {
		http.Error(w, "Secret not found or already destroyed", http.StatusNotFound)
		return
	}

	if time.Now().After(secret.ExpiresAt) {
		s.db.DeleteSecret(r.Context(), id)
		http.Error(w, "Secret not found or already destroyed", http.StatusNotFound)
		return
	}

	remaining, err := s.db.DecrementRemainingViews(r.Context(), id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if remaining <= 0 {
		if err := s.db.DeleteSecret(r.Context(), id); err != nil {
			// Log error but continue
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

func (s *Server) GetSecretMetaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[4]

	secret, err := s.db.GetSecretMeta(r.Context(), id)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if secret == nil {
		http.Error(w, "Secret not found or already destroyed", http.StatusNotFound)
		return
	}

	if time.Now().After(secret.ExpiresAt) {
		s.db.DeleteSecret(r.Context(), id)
		http.Error(w, "Secret not found or already destroyed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security Headers
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-eval' 'wasm-unsafe-eval' https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline'; connect-src 'self' https://cdnjs.cloudflare.com; img-src 'self' data:; frame-ancestors 'none';")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Rate Limiting - Only for API requests
		if strings.HasPrefix(r.URL.Path, "/api/") {
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = strings.Split(r.RemoteAddr, ":")[0]
			} else {
				ip = strings.TrimSpace(strings.Split(ip, ",")[0])
			}
			
			limiter.mu.Lock()
			limiter.counts[ip]++
			if limiter.counts[ip] > 10 {
				limiter.mu.Unlock()
				http.Error(w, "Too many requests. Please wait a few seconds.", http.StatusTooManyRequests)
				return
			}
			limiter.mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/secrets", s.CreateSecretHandler)
	mux.HandleFunc("/api/v1/secrets/", s.GetSecretHandler)
	mux.HandleFunc("/api/v1/meta/", s.GetSecretMetaHandler)
}

func (s *Server) GetHandler(mux *http.ServeMux) http.Handler {
	return s.securityMiddleware(mux)
}
