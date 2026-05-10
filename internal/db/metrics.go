package db

import (
	"context"
	"fmt"

	"github.com/FelixWinchester/CodeCartographer/internal/analyzer/git"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricsRepository struct {
	pool *pgxpool.Pool
}

func NewMetricsRepository(pool *pgxpool.Pool) *MetricsRepository {
	return &MetricsRepository{pool: pool}
}

// SaveFileMetrics сохраняет git метрики по файлам в Postgres
func (r *MetricsRepository) SaveFileMetrics(ctx context.Context, repoPath string, metrics map[string]*git.FileMetrics) error {
	for _, m := range metrics {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO file_metrics (repo_path, file_path, churn_rate, owner)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (repo_path, file_path)
			DO UPDATE SET
				churn_rate = EXCLUDED.churn_rate,
				owner      = EXCLUDED.owner,
				updated_at = NOW()
		`, repoPath, m.FilePath, m.ChurnRate, m.Owner)
		if err != nil {
			return fmt.Errorf("failed to save metrics for %s: %w", m.FilePath, err)
		}
	}
	return nil
}

// SaveCoupling сохраняет change coupling пары в Postgres
func (r *MetricsRepository) SaveCoupling(ctx context.Context, repoPath string, pairs []git.CouplingPair) error {
	for _, p := range pairs {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO change_coupling (repo_path, file_a, file_b, coupling_score)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (repo_path, file_a, file_b)
			DO UPDATE SET
				coupling_score = EXCLUDED.coupling_score,
				updated_at     = NOW()
		`, repoPath, p.FileA, p.FileB, p.CouplingScore)
		if err != nil {
			return fmt.Errorf("failed to save coupling %s<->%s: %w", p.FileA, p.FileB, err)
		}
	}
	return nil
}
