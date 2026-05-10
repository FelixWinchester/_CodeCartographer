package main

import (
	"context"
	"log"
	"os"

	"github.com/FelixWinchester/CodeCartographer/internal/analyzer/ast"
	"github.com/FelixWinchester/CodeCartographer/internal/config"
	"github.com/FelixWinchester/CodeCartographer/internal/db"
)

func main() {
	ctx := context.Background()

	// Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Путь к репозиторию берём из аргумента командной строки
	// Пример: go run ./cmd/analyzer /path/to/some/go/project
	if len(os.Args) < 2 {
		log.Fatalf("usage: analyzer <repo-path>")
	}
	repoPath := os.Args[1]

	// Подключаемся к Postgres
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres error: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Postgres connected")

	// Парсим репозиторий
	log.Printf("🔍 Analyzing repository: %s", repoPath)
	parser := ast.NewParser(repoPath)
	graph, err := parser.Parse()
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}
	log.Printf("✅ Found %d nodes, %d edges", len(graph.Nodes), len(graph.Edges))

	// Сохраняем граф в Postgres
	repo := db.NewGraphRepository(pool)
	if err := repo.SaveGraph(ctx, graph); err != nil {
		log.Fatalf("save error: %v", err)
	}
	log.Println("✅ Graph saved to Postgres")
}
