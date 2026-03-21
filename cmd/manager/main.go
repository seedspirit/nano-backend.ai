package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/seedspirit/nano-backend.ai/internal/manager"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	addr := ":8090"
	router := manager.NewRouter()

	slog.Info("manager starting", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
