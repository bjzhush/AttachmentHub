package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"atthub/internal/config"
	"atthub/internal/db"
)

func main() {
	var yes bool
	flag.BoolVar(&yes, "yes", false, "confirm deleting all local sqlite data and stored attachments")
	flag.Parse()

	if !yes {
		fmt.Fprintln(os.Stderr, "Refusing to run without --yes")
		os.Exit(1)
	}

	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := os.RemoveAll(cfg.StorageDir); err != nil {
		logger.Error("failed to clear storage directory", "path", cfg.StorageDir, "error", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(cfg.StorageDir, 0o755); err != nil {
		logger.Error("failed to recreate storage directory", "path", cfg.StorageDir, "error", err)
		os.Exit(1)
	}

	for _, path := range []string{
		cfg.DBPath,
		cfg.DBPath + "-wal",
		cfg.DBPath + "-shm",
	} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			logger.Error("failed to remove sqlite file", "path", path, "error", err)
			os.Exit(1)
		}
	}

	sqliteDB, err := db.OpenSQLite(cfg.DBPath)
	if err != nil {
		logger.Error("failed to reinitialize sqlite schema", "error", err)
		os.Exit(1)
	}
	_ = sqliteDB.Close()

	logger.Info("development data reset completed",
		"db_path", cfg.DBPath,
		"storage_dir", cfg.StorageDir,
	)
}
