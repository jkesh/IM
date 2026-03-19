package main

import (
	"IM/internal/config"
	"IM/internal/controller"
	"IM/internal/logic"
	"IM/internal/service"
	"IM/internal/storage/cache"
	"IM/internal/storage/db"
	"errors"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	logic.SetJWTSecret(cfg.Auth.JWTSecret)
	logic.SetTokenTTL(cfg.Auth.TokenTTL)

	if err := db.InitDB(cfg.Database); err != nil {
		log.Fatal("init db failed: ", err)
	}
	if err := cache.InitRedis(cfg.Redis); err != nil {
		log.Printf("init redis failed, offline message is disabled: %v", err)
	}

	hub := service.NewHub(cfg.Redis.OfflineTTL)
	go hub.Run()

	mux := http.NewServeMux()
	controller.RegisterRoutes(mux, hub)

	server := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           mux,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	log.Printf("IM server started on %s", cfg.Server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("listen and serve failed: ", err)
	}
}
