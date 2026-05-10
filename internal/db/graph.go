package db

import (
	"context"
	"fmt"

	"github.com/FelixWinchester/CodeCartographer/internal/analyzer"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GraphRepository struct {
	pool *pgxpool.Pool
}

func NewGraphRepository(pool *pgxpool.Pool) *GraphRepository {
	return &GraphRepository{pool: pool}
}

// SaveGraph сохраняет весь граф в Postgres.
// Сначала удаляет старые данные по repoPath, потом вставляет новые.
func (r *GraphRepository) SaveGraph(ctx context.Context, graph *analyzer.Graph) error {
	// Начинаем транзакцию — либо всё сохраняется, либо ничего
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if len(graph.Nodes) == 0 {
		return nil
	}

	repoPath := graph.Nodes[0].RepoPath

	// Удаляем старые данные по этому репозиторию
	_, err = tx.Exec(ctx, `DELETE FROM edges WHERE repo_path = $1`, repoPath)
	if err != nil {
		return fmt.Errorf("failed to delete old edges: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM nodes WHERE repo_path = $1`, repoPath)
	if err != nil {
		return fmt.Errorf("failed to delete old nodes: %w", err)
	}

	// Маппинг: локальный ID → реальный ID в Postgres
	// Нужен чтобы правильно записать рёбра после вставки узлов
	localToDBID := map[int64]int64{}

	// Вставляем узлы
	for _, node := range graph.Nodes {
		var dbID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO nodes (repo_path, kind, name, file_path, start_line, end_line)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, node.RepoPath, string(node.Kind), node.Name, node.FilePath, node.StartLine, node.EndLine).Scan(&dbID)
		if err != nil {
			return fmt.Errorf("failed to insert node %s: %w", node.Name, err)
		}
		localToDBID[node.ID] = dbID
	}

	// Вставляем рёбра используя реальные ID из Postgres
	for _, edge := range graph.Edges {
		fromDBID, ok1 := localToDBID[edge.FromID]
		toDBID, ok2 := localToDBID[edge.ToID]
		if !ok1 || !ok2 {
			// Пропускаем рёбра с неизвестными узлами
			continue
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO edges (repo_path, from_id, to_id, kind)
			VALUES ($1, $2, $3, $4)
		`, edge.RepoPath, fromDBID, toDBID, string(edge.Kind))
		if err != nil {
			return fmt.Errorf("failed to insert edge: %w", err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetGraph возвращает все узлы и рёбра для репозитория
func (r *GraphRepository) GetGraph(ctx context.Context, repoPath string) (*analyzer.Graph, error) {
	graph := &analyzer.Graph{}

	// Получаем узлы
	rows, err := r.pool.Query(ctx, `
		SELECT id, repo_path, kind, name, file_path, start_line, end_line
		FROM nodes
		WHERE repo_path = $1
	`, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		node := &analyzer.Node{}
		var kind string
		err := rows.Scan(&node.ID, &node.RepoPath, &kind, &node.Name, &node.FilePath, &node.StartLine, &node.EndLine)
		if err != nil {
			return nil, err
		}
		node.Kind = analyzer.NodeKind(kind)
		graph.Nodes = append(graph.Nodes, node)
	}

	// Получаем рёбра
	edgeRows, err := r.pool.Query(ctx, `
		SELECT id, repo_path, from_id, to_id, kind
		FROM edges
		WHERE repo_path = $1
	`, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}
	defer edgeRows.Close()

	for edgeRows.Next() {
		edge := &analyzer.Edge{}
		var kind string
		err := edgeRows.Scan(&edge.ID, &edge.RepoPath, &edge.FromID, &edge.ToID, &kind)
		if err != nil {
			return nil, err
		}
		edge.Kind = analyzer.EdgeKind(kind)
		graph.Edges = append(graph.Edges, edge)
	}

	return graph, nil
}

// GetNode возвращает один узел по ID
func (r *GraphRepository) GetNode(ctx context.Context, id int64) (*analyzer.Node, error) {
	node := &analyzer.Node{}
	var kind string

	err := r.pool.QueryRow(ctx, `
		SELECT id, repo_path, kind, name, file_path, start_line, end_line
		FROM nodes
		WHERE id = $1
	`, id).Scan(&node.ID, &node.RepoPath, &kind, &node.Name, &node.FilePath, &node.StartLine, &node.EndLine)
	if err != nil {
		return nil, fmt.Errorf("node %d not found: %w", id, err)
	}

	node.Kind = analyzer.NodeKind(kind)
	return node, nil
}
