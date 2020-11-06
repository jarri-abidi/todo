package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/matryer/is"

	"github.com/jarri-abidi/todolist/config"
)

func TestLoad(t *testing.T) {
	t.Run("Loads from file", func(t *testing.T) {
		is := is.New(t)

		conf, err := config.Load("testdata/", "test.env")
		is.NoErr(err) // could not load config

		is.Equal(conf.ServerAddress, "0.0.0.0:443")
		is.Equal(conf.ServerReadTimeout, 30*time.Second)
		is.Equal(conf.ServerWriteTimeout, 30*time.Second)
		is.Equal(conf.ServerIdleTimeout, 1*time.Minute)
		is.Equal(conf.GracefulShutdownTimeout, 45*time.Second)
	})

	t.Run("Overrides from env", func(t *testing.T) {
		is := is.New(t)

		reset := setEnv("SERVER_ADDRESS", "127.0.0.1:80", is)
		defer reset()
		reset = setEnv("SERVER_WRITE_TIMEOUT", "15s", is)
		defer reset()
		reset = setEnv("SERVER_READ_TIMEOUT", "15s", is)
		defer reset()
		reset = setEnv("SERVER_IDLE_TIMEOUT", "2m", is)
		defer reset()
		reset = setEnv("GRACEFUL_SHUTDOWN_TIMEOUT", "30s", is)
		defer reset()

		conf, err := config.Load("testdata/", "test.env")
		is.NoErr(err) // could not load config

		is.Equal(conf.ServerAddress, "127.0.0.1:80")
		is.Equal(conf.ServerReadTimeout, 15*time.Second)
		is.Equal(conf.ServerWriteTimeout, 15*time.Second)
		is.Equal(conf.ServerIdleTimeout, 2*time.Minute)
		is.Equal(conf.GracefulShutdownTimeout, 30*time.Second)
	})

	t.Run("Returns err when no file exists", func(t *testing.T) {
		is := is.New(t)

		if _, err := config.Load("testdata/", "doesnotexist.env"); err == nil {
			is.Fail() // should return err when file does not exist
		}
	})
}

func setEnv(key string, val string, is *is.I) (reset func()) {
	original := os.Getenv(key)
	is.NoErr(os.Setenv(key, val)) // could not set env
	return func() {
		os.Setenv(key, original)
	}
}
