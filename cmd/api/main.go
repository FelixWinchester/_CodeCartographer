package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/FelixWinchester/CodeCartographer/internal/cache"
	"github.com/FelixWinchester/CodeCartographer/internal/config"
	"github.com/FelixWinchester/CodeCartographer/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()

	// Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Подключаемся к Postgres
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres error: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Postgres connected")

	// Подключаемся к Redis
	redisClient, err := cache.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis error: %v", err)
	}
	defer redisClient.Close()
	log.Println("✅ Redis connected")

	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check — первый эндпоинт, проверяем что сервер живой
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	log.Printf("🚀 API server starting on :%s", cfg.APIPort)
	if err := http.ListenAndServe(":"+cfg.APIPort, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
