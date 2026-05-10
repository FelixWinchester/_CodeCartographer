package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/FelixWinchester/CodeCartographer/internal/api/handlers"
	"github.com/FelixWinchester/CodeCartographer/internal/cache"
	"github.com/FelixWinchester/CodeCartographer/internal/config"
	"github.com/FelixWinchester/CodeCartographer/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres error: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Postgres connected")

	redisClient, err := cache.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis error: %v", err)
	}
	defer redisClient.Close()
	log.Println("✅ Redis connected")

	// Репозитории
	graphRepo := db.NewGraphRepository(pool)
	metricsRepo := db.NewMetricsRepository(pool)

	// Handlers
	graphHandler := handlers.NewGraphHandler(graphRepo)
	nodeHandler := handlers.NewNodeHandler(graphRepo, metricsRepo)

	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS — нужен чтобы фронтенд мог обращаться к API
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// API роуты
	r.Route("/api", func(r chi.Router) {
		r.Get("/graph", graphHandler.GetGraph)
		r.Get("/node/{id}", nodeHandler.GetNode)
	})

	log.Printf("🚀 API server starting on :%s", cfg.APIPort)
	if err := http.ListenAndServe(":"+cfg.APIPort, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
