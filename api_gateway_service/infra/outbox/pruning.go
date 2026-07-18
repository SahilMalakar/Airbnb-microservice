package outbox

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// PruneRunner periodically deletes old processed outbox rows.
type PruneRunner struct {
	db            *sql.DB
	retentionDays int
	interval      time.Duration
	cancel        context.CancelFunc
	done          chan struct{}
}

func NewPruneRunner(db *sql.DB) *PruneRunner {
	days := 30
	if v := os.Getenv("PRUNE_RETENTION_DAYS"); v != "" {
		if d, err := strconv.Atoi(v); err == nil && d > 0 {
			days = d
		}
	}

	hours := 24
	if v := os.Getenv("PRUNE_INTERVAL_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			hours = h
		}
	}

	return &PruneRunner{
		db:            db,
		retentionDays: days,
		interval:      time.Duration(hours) * time.Hour,
		done:          make(chan struct{}),
	}
}

func (p *PruneRunner) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go func() {
		defer close(p.done)
		slog.Info("outbox pruning started", "retentionDays", p.retentionDays, "intervalHours", p.interval.Hours())

		// Run immediately, then on interval
		p.prune(ctx)

		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.prune(ctx)
			}
		}
	}()
}

func (p *PruneRunner) Stop() {
	if p.cancel != nil {
		p.cancel()
		<-p.done
		slog.Info("outbox pruning stopped")
	}
}

func (p *PruneRunner) prune(ctx context.Context) {
	cutoff := time.Now().Add(-time.Duration(p.retentionDays) * 24 * time.Hour)
	query := `DELETE FROM outbox WHERE processed_at IS NOT NULL AND processed_at < $1`

	result, err := p.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		slog.Error("outbox pruning failed", "error", err)
		return
	}

	rows, _ := result.RowsAffected()
	slog.Info("outbox pruning complete", "rowsDeleted", rows)
}
