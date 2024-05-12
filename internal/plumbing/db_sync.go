package plumbing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/reddec/token-login/internal/ent"
	"github.com/reddec/token-login/internal/ent/token"
	"github.com/reddec/token-login/web"
)

// SyncStats synchronizes hits (stats) to database.
func SyncStats(ctx context.Context, client *ent.Client, statsCh <-chan web.Hit, aggregate time.Duration) {
	ticker := time.NewTicker(aggregate)
	defer ticker.Stop()

	var done bool

	var stats = make(map[int]aggregation)
	for !done {
		select {
		case <-ctx.Done():
			return
		case hit, ok := <-statsCh:
			if !ok {
				// probably finished - dump and quit
				done = true
				break
			}
			old := stats[hit.ID]
			if hit.Time.After(old.Last) {
				old.Last = hit.Time
			}
			old.Hits++
			stats[hit.ID] = old
		case <-ticker.C:
		}

		// dump
		if len(stats) == 0 {
			slog.Debug("no stats to sync")
			continue
		}
		if err := dumpStats(ctx, client, stats); err != nil {
			slog.Error("failed dump stats to database", "error", err)
			// keep
		} else {
			stats = make(map[int]aggregation) // clear
		}
	}
}

func dumpStats(ctx context.Context, client *ent.Client, stats map[int]aggregation) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	for id, stat := range stats {
		if err := tx.Token.Update().Where(token.ID(id)).AddRequests(stat.Hits).SetLastAccessAt(stat.Last).Exec(ctx); err != nil {
			return errors.Join(fmt.Errorf("update stat %v: %w", id, err), tx.Rollback())
		}
	}

	return tx.Commit()
}

type aggregation struct {
	Hits int64
	Last time.Time
}
