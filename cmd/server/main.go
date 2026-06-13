package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nstoddard1002/traceless/internal/api"
	"github.com/nstoddard1002/traceless/internal/db"
	"github.com/nstoddard1002/traceless/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Database
	database, err := db.NewDB(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize API Server
	apiServer := api.NewServer(database)
	mux := http.NewServeMux()

	// Register API Routes
	apiServer.RegisterRoutes(mux)

	// Serve Static Files & SPA Routing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If requesting a file that exists, serve it
		path := "web" + r.URL.Path
		if _, err := os.Stat(path); err == nil && r.URL.Path != "/" {
			http.ServeFile(w, r, path)
			return
		}
		// Otherwise, serve index.html for SPA routing
		http.ServeFile(w, r, "web/index.html")
	})

	// Wrap mux with middleware
	handler := apiServer.GetHandler(mux)

	// Start Cleanup Worker
	go worker.StartCleanupWorker(ctx, database, 5*time.Minute)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Start Server
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for termination signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
