package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := Config{
		Port: 8080,
	}

	assert.Equal(t, 8080, cfg.Port, "Port should be set correctly")
}
