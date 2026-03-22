//go:build integration

package postgres

import (
	"fmt"
	"os"
	"testing"

	"github.com/byteroom/backend/internal/config"
	"github.com/stretchr/testify/require"
)

// setupIntegrationDB connects to the test database defined in docker-compose.test.yml
// and runs migrations before each test.
func setupIntegrationDB(t *testing.T) *DB {
	t.Helper()

	port := 5433
	if p := os.Getenv("TEST_DB_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	cfg := &config.DatabaseConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     port,
		User:     "test",
		Password: "test",
		Name:     "byteroom_test",
		SSLMode:  "disable",
	}

	db, err := Connect(cfg)
	require.NoError(t, err, "failed to connect to test database — is docker-compose.test.yml running?")

	t.Cleanup(func() { db.Close() })

	return db
}
