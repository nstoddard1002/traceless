package worker

import (
	"context"
	"log"
	"time"

	"github.com/nstoddard1002/traceless/internal/db"
)

// StartCleanupWorker starts a background goroutine that deletes expired secrets.
func StartCleanupWorker(ctx context.Context, database *db.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting cleanup worker with interval: %v", interval)

	for {
		select {
		case <-ticker.C:
			deleted, err := database.DeleteExpiredSecrets(ctx)
			if err != nil {
				log.Printf("Cleanup worker error: %v", err)
			} else if deleted > 0 {
				log.Printf("Cleanup worker deleted %d expired secrets", deleted)
			}
		case <-ctx.Done():
			log.Println("Stopping cleanup worker")
			return
		}
	}
}
