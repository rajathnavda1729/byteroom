//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/byteroom/backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDBConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     5433,
		User:     "test",
		Password: "test",
		Name:     "byteroom_test",
		SSLMode:  "disable",
	}
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestDB_Connect_ValidConfig_ReturnsPool(t *testing.T) {
	db, err := Connect(testDBConfig())

	require.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}

func TestDB_Ping_HealthyConnection_ReturnsNil(t *testing.T) {
	db, err := Connect(testDBConfig())
	require.NoError(t, err)
	defer db.Close()

	err = db.PingContext(context.Background())

	assert.NoError(t, err)
}

func TestDB_Connect_InvalidHost_ReturnsError(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:    "invalid-host-that-does-not-exist",
		Port:    9999,
		User:    "test",
		Password: "test",
		Name:    "test",
		SSLMode: "disable",
	}

	_, err := Connect(cfg)

	assert.Error(t, err)
}
