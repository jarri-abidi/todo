package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jarri-abidi/todolist/config"
)

func TestLoad(t *testing.T) {
	t.Run("Loads from file", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		conf, err := config.Load("testdata/", "test.env")
		require.NoError(err) // could not load config

		assert.Equal(conf.ServerAddress, "0.0.0.0:443")
		assert.Equal(conf.ServerReadTimeout, 30*time.Second)
		assert.Equal(conf.ServerWriteTimeout, 30*time.Second)
		assert.Equal(conf.ServerIdleTimeout, 1*time.Minute)
		assert.Equal(conf.GracefulShutdownTimeout, 45*time.Second)
	})

	t.Run("Overrides from env", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		reset := setEnv("SERVER_ADDRESS", "127.0.0.1:80", assert)
		defer reset()
		reset = setEnv("SERVER_WRITE_TIMEOUT", "15s", assert)
		defer reset()
		reset = setEnv("SERVER_READ_TIMEOUT", "15s", assert)
		defer reset()
		reset = setEnv("SERVER_IDLE_TIMEOUT", "2m", assert)
		defer reset()
		reset = setEnv("GRACEFUL_SHUTDOWN_TIMEOUT", "30s", assert)
		defer reset()

		conf, err := config.Load("testdata/", "test.env")
		require.NoError(err) // could not load config

		assert.Equal(conf.ServerAddress, "127.0.0.1:80")
		assert.Equal(conf.ServerReadTimeout, 15*time.Second)
		assert.Equal(conf.ServerWriteTimeout, 15*time.Second)
		assert.Equal(conf.ServerIdleTimeout, 2*time.Minute)
		assert.Equal(conf.GracefulShutdownTimeout, 30*time.Second)
	})

	t.Run("Returns err when no file exists", func(t *testing.T) {
		if _, err := config.Load("testdata/", "doesnotexist.env"); err == nil {
			t.Fail() // should return err when file does not exist
		}
	})
}

func setEnv(key string, val string, assert *assert.Assertions) (reset func()) {
	original := os.Getenv(key)
	assert.NoError(os.Setenv(key, val)) // could not set env
	return func() {
		os.Setenv(key, original)
	}
}
