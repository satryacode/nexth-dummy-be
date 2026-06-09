package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_defaults(t *testing.T) {
	cfg := Load()
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "secret", cfg.JWTSecret)
	assert.Equal(t, "8080", cfg.Port)
}

func TestLoad_envOverride(t *testing.T) {
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")
	cfg := Load()
	assert.Equal(t, "9090", cfg.Port)
}
