package plumbing

import (
	"context"
	"log/slog"
	"time"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/web"
)

func SyncStats(ctx context.Context, store dbo.Store, statsCh <-chan web.Hit, aggregate time.Duration) {
	ticker := time.NewTicker(aggregate)
	defer ticker.Stop()

	stats := make(map[int64]dbo.StatsEntry)
	for {
		select {
		case <-ctx.Done():
			return
		case hit, ok := <-statsCh:
			if !ok {
				// Channel closed: flush any remaining stats.
				if len(stats) > 0 {
					if err := store.UpdateStats(ctx, stats); err != nil {
						slog.Error("failed dump stats to database", "error", err)
					}
				}
				return
			}
			old := stats[hit.ID]
			if hit.Time.After(old.Last) {
				old.Last = hit.Time
			}
			old.Hits++
			stats[hit.ID] = old
		case <-ticker.C:
			if len(stats) == 0 {
				continue
			}
			if err := store.UpdateStats(ctx, stats); err != nil {
				slog.Error("failed dump stats to database", "error", err)
			} else {
				stats = make(map[int64]dbo.StatsEntry)
			}
		}
	}
}
