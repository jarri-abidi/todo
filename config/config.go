package config

import "time"

// Config stores all configurations of the application.
// The values are read from a file or environment variables.
type Config struct {
	ServerAddress           string
	ServerWriteTimeout      time.Duration
	ServerReadTimeout       time.Duration
	ServerIdleTimeout       time.Duration
	GracefulShutdownTimeout time.Duration
}

// Load reads configuration from file or environment variables.
func Load(path, name string) (Config, error) {
	return Config{
		ServerAddress:           "0.0.0.0:8085",
		ServerWriteTimeout:      15 * time.Second,
		ServerReadTimeout:       15 * time.Second,
		ServerIdleTimeout:       60 * time.Second,
		GracefulShutdownTimeout: 30 * time.Second,
	}, nil
}
