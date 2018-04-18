package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseChecks verifies that the latency_config
// JSON can successfully be unmarshalled.
// If this test fails, it is likely that there is a
// formatting issue with the latency_config
func TestParseChecks(t *testing.T) {
	assert.NotPanics(t, func() {
		ParseChecks("example_config.json")
	}, "Unable to parse latency checks")
}
