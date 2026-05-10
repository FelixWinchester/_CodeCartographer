package main

import (
	"context"
	"log"
	"os"

	"github.com/FelixWinchester/CodeCartographer/internal/analyzer/ast"
	gitanalyzer "github.com/FelixWinchester/CodeCartographer/internal/analyzer/git"
	"github.com/FelixWinchester/CodeCartographer/internal/config"
	"github.com/FelixWinchester/CodeCartographer/internal/db"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatalf("usage: analyzer <repo-path>")
	}
	repoPath := os.Args[1]

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres error: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Postgres connected")

	// AST анализ
	log.Printf("🔍 Analyzing repository: %s", repoPath)
	parser := ast.NewParser(repoPath)
	graph, err := parser.Parse()
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}
	log.Printf("✅ Found %d nodes, %d edges", len(graph.Nodes), len(graph.Edges))

	graphRepo := db.NewGraphRepository(pool)
	if err := graphRepo.SaveGraph(ctx, graph); err != nil {
		log.Fatalf("save graph error: %v", err)
	}
	log.Println("✅ Graph saved to Postgres")

	// Git аналитика
	log.Println("📊 Analyzing git history...")
	gitAnalyzer := gitanalyzer.NewAnalyzer(repoPath)

	fileMetrics, err := gitAnalyzer.AnalyzeMetrics()
	if err != nil {
		log.Printf("⚠️  git metrics error (skipping): %v", err)
	} else {
		metricsRepo := db.NewMetricsRepository(pool)
		if err := metricsRepo.SaveFileMetrics(ctx, repoPath, fileMetrics); err != nil {
			log.Fatalf("save metrics error: %v", err)
		}
		log.Printf("✅ File metrics saved (%d files)", len(fileMetrics))
	}

	coupling, err := gitAnalyzer.AnalyzeCoupling()
	if err != nil {
		log.Printf("⚠️  coupling analysis error (skipping): %v", err)
	} else {
		metricsRepo := db.NewMetricsRepository(pool)
		if err := metricsRepo.SaveCoupling(ctx, repoPath, coupling); err != nil {
			log.Fatalf("save coupling error: %v", err)
		}
		log.Printf("✅ Coupling pairs saved (%d pairs)", len(coupling))
	}

	log.Println("🎉 Analysis complete")
}
