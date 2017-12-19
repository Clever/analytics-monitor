package db

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var loc, _ = time.LoadLocation("America/Los_Angeles")

func setup(t *testing.T) *postgresClient {
	conf := PostgresCredentials{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: "",
		Password: "",
		Database: "postgres",
	}
	postgres, err := newPostgresClient(conf, "testCluster")
	db := postgres.(*postgresClient)

	assert.NoError(t, err)

	_, err = db.session.Exec("CREATE SCHEMA IF NOT EXISTS test")
	assert.NoError(t, err)

	_, err = db.session.Exec("DROP TABLE IF EXISTS test.latency")
	assert.NoError(t, err)
	_, err = db.session.Exec(`CREATE TABLE test.latency (time timestamp without time zone)`)
	assert.NoError(t, err)

	return db
}

func TestQueryLatency(t *testing.T) {
	n := time.Now()
	past := time.Date(n.Year(), n.Month(), n.Day(), n.Hour()-100, 0, 0, 0, loc)

	db := setup(t)

	latency, valid, err := db.QueryLatency("time", "test", "latency")
	assert.NoError(t, err)
	assert.False(t, valid)

	_, err = db.session.Exec(fmt.Sprintf("INSERT INTO test.latency(time) VALUES ('%s')",
		past.In(time.UTC).Format(time.RFC3339)))
	require.NoError(t, err)

	latency, valid, err = db.QueryLatency("time", "test", "latency")
	assert.NoError(t, err)
	assert.True(t, valid)
	// Give a little leeway for timing
	assert.True(t, latency >= 99 && latency <= 101)
}
