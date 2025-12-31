package filequeue

import (
	"os"
	"strconv"
	"strings"
)

const (
	defaultWorkers = 4
	defaultSize    = 100

	envWorkers = "FILEQUEUE_WORKERS"
	envSize    = "FILEQUEUE_SIZE"
)

type Config struct {
	Workers int
	Size    int
}

func LoadConfigFromEnv() Config {
	cfg := Config{
		Workers: defaultWorkers,
		Size:    defaultSize,
	}

	if val := strings.TrimSpace(os.Getenv(envWorkers)); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			cfg.Workers = parsed
		}
	}
	if val := strings.TrimSpace(os.Getenv(envSize)); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			cfg.Size = parsed
		}
	}
	return cfg
}
