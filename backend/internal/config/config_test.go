package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithRequiredEnvVars_ReturnsConfig(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-key")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "test-secret-key", cfg.JWT.Secret)
}

func TestLoad_MissingJWTSecret_ReturnsError(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestLoad_CustomPort_AppliesOverride(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-key")
	t.Setenv("PORT", "9090")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Server.Port)
}

func TestLoad_InvalidDBPort_ReturnsError(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-key")
	t.Setenv("DB_PORT", "not-a-number")
	defer os.Unsetenv("DB_PORT")

	_, err := Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB_PORT")
}

func TestDatabaseConfig_DSN_FormatsCorrectly(t *testing.T) {
	d := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "byteroom",
		Password: "secret",
		Name:     "byteroom",
		SSLMode:  "disable",
	}

	dsn := d.DSN()

	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=byteroom")
	assert.Contains(t, dsn, "dbname=byteroom")
	assert.Contains(t, dsn, "sslmode=disable")
}
