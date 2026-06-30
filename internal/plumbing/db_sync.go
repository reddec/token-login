package plumbing

import (
	"context"
	"log/slog"
	"time"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/web"
)

// SyncStats synchronizes hits (stats) to database.
func SyncStats(ctx context.Context, store dbo.Store, statsCh <-chan web.Hit, aggregate time.Duration) {
	ticker := time.NewTicker(aggregate)
	defer ticker.Stop()

	var done bool

	var stats = make(map[int64]dbo.StatsEntry)
	for !done {
		select {
		case <-ctx.Done():
			return
		case hit, ok := <-statsCh:
			if !ok {
				done = true
				break
			}
			old := stats[int64(hit.ID)]
			if hit.Time.After(old.Last) {
				old.Last = hit.Time
			}
			old.Hits++
			stats[int64(hit.ID)] = old
		case <-ticker.C:
		}

		if len(stats) == 0 {
			slog.Debug("no stats to sync")
			continue
		}
		if err := store.UpdateStats(ctx, stats); err != nil {
			slog.Error("failed dump stats to database", "error", err)
		} else {
			stats = make(map[int64]dbo.StatsEntry)
		}
	}
}
