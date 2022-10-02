package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// Config stores all configurations of the application.
// The values are read from a file or environment variables.
type Config struct {
	ServerAddress              string        `envconfig:"SERVER_ADDRESS"`
	ServerWriteTimeout         time.Duration `envconfig:"SERVER_WRITE_TIMEOUT"`
	ServerReadTimeout          time.Duration `envconfig:"SERVER_READ_TIMEOUT"`
	ServerIdleTimeout          time.Duration `envconfig:"SERVER_IDLE_TIMEOUT"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	DBSource                   string        `envconfig:"DB_SOURCE"`
	DBConnectTimeout           time.Duration `envconfig:"DB_CONNECT_TIMEOUT"`
	OTELExporterJaegerEndpoint string        `envconfig:"OTEL_EXPORTER_JAEGER_ENDPOINT"`
}

// Load returns configuration from file or environment variables.
// Values from file take precedence over environment variables.
func Load(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, errors.Wrapf(err, "could not load .env file %s", path)
	}

	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return nil, errors.Wrap(err, "could not load env vars")
	}

	return &config, nil
}
