package config

import (
	"os"
	"strconv"
)

const (
	defaultPort        = "10001"
	defaultDBPath      = "./data/attachmenthub.db"
	defaultStorageDir  = "./attachments"
	defaultMaxUploadMB = 100
	maxUploadEnvVar    = "ATTHUB_MAX_UPLOAD_MB"
	portEnvVar         = "ATTHUB_PORT"
	dbPathEnvVar       = "ATTHUB_DB_PATH"
	storageDirEnvVar   = "ATTHUB_STORAGE_DIR"
)

type Config struct {
	Port           string
	DBPath         string
	StorageDir     string
	MaxUploadBytes int64
}

func Load() Config {
	maxUploadMB := getenvInt(maxUploadEnvVar, defaultMaxUploadMB)

	return Config{
		Port:           getenv(portEnvVar, defaultPort),
		DBPath:         getenv(dbPathEnvVar, defaultDBPath),
		StorageDir:     getenv(storageDirEnvVar, defaultStorageDir),
		MaxUploadBytes: int64(maxUploadMB) * 1024 * 1024,
	}
}

func getenv(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func getenvInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
