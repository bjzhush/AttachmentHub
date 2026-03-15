package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"atthub/internal/api"
	"atthub/internal/attachment"
	"atthub/internal/config"
	"atthub/internal/db"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	sqliteDB, err := db.OpenSQLite(cfg.DBPath)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer sqliteDB.Close()

	repo := attachment.NewRepository(sqliteDB)
	storage := attachment.NewLocalStorage(cfg.StorageDir)
	service := attachment.NewService(repo, storage)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(service, cfg, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("AttachmentHub API started",
		"port", cfg.Port,
		"db_path", cfg.DBPath,
		"storage_dir", cfg.StorageDir,
	)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
