package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zfusx/mbti/internal/config"
	"github.com/zfusx/mbti/internal/scoring"
	"github.com/zfusx/mbti/internal/server"
	"github.com/zfusx/mbti/internal/storage"
	"github.com/zfusx/mbti/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	store, err := storage.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("storage init error: %v", err)
	}
	defer store.Close()

	scorer, err := scoring.NewService(cfg.QuestionCount)
	if err != nil {
		log.Fatalf("scoring init error: %v", err)
	}

	var spaFS http.FileSystem
	if sub, err := fs.Sub(web.Static, "static"); err == nil {
		spaFS = http.FS(sub)
	} else {
		log.Printf("warning: static assets unavailable: %v", err)
	}

	srv := server.New(store, scorer, cfg.AllowedOrigin, cfg.QuestionCount, spaFS)

	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		log.Printf("MBTI API listening on :%s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	log.Println("Shutting down http server...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
