package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	BatchDuration      time.Duration
	BatchItems         int
	ServerPort         int
	SharedVolumePath   string
	SidecarPort        int
	BackendURL         string
	BufferSize         int
	FlushInterval      time.Duration
	GracePeriodSeconds int
	MetricsPort        int
	JaegerEndpoint     string
}

func Load() (*Config, error) {
	return &Config{
		BatchDuration:      getDuration("BATCH_DURATION", 30*time.Second),
		BatchItems:         getInt("BATCH_ITEMS", 1000),
		ServerPort:         getInt("BATCH_PORT", 8080),
		SharedVolumePath:   getString("SHARED_VOLUME_PATH", "/shared"),
		SidecarPort:        getInt("SIDECAR_PORT", 8081),
		BackendURL:         getString("BACKEND_URL", "http://mock-telemetry-backend:8080"),
		BufferSize:         getInt("BUFFER_SIZE", 10000),
		FlushInterval:      getDuration("FLUSH_INTERVAL", 5*time.Second),
		GracePeriodSeconds: getInt("GRACE_PERIOD_SECONDS", 120),
		MetricsPort:        getInt("METRICS_PORT", 9090),
		JaegerEndpoint:     getString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}, nil
}

func getString(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
