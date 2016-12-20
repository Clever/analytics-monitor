package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	l "github.com/Clever/analytics-pipeline-monitor/logger"
)

var (
	// We have two redshift databases, one that holds all the data and views (prod)
	// and one that holds a small number of faster, materialized views. (fast-prod)
	// RedshiftProd* is for the former
	RedshiftProdHost     string
	RedshiftProdPort     string
	RedshiftProdDatabase string
	RedshiftProdUsername string
	RedshiftProdPassword string

	// RedshiftFast* is the latter
	RedshiftFastHost     string
	RedshiftFastPort     string
	RedshiftFastDatabase string
	RedshiftFastUsername string
	RedshiftFastPassword string
)

// ClusterChecks stores latency checks by cluster
type ClusterChecks struct {
	ProdChecks     []SchemaChecks `json:"prod"`
	FastProdChecks []SchemaChecks `json:"fast-prod"`
}

// SchemaChecks stores an array of checks by schema
type SchemaChecks struct {
	SchemaName string        `json:"schema"`
	Checks     []TableChecks `json:"checks"`
}

// TableChecks stores checks per view or table
type TableChecks struct {
	TableName string      `json:"table"`
	Latency   LatencyInfo `json:"latency"`
}

// LatencyInfo stores information for a latency check
type LatencyInfo struct {
	TimestampColumn string `json:"timestamp_column"`
	Threshold       string `json:"threshold"`
}

// Parse reads environment variables and initializes the config.
func Parse() {
	RedshiftProdHost = requiredEnv("PG_HOST")
	RedshiftProdPort = requiredEnv("PG_PORT")
	RedshiftProdDatabase = requiredEnv("PG_DATABASE")
	RedshiftProdUsername = requiredEnv("PG_USER")
	RedshiftProdPassword = requiredEnv("PG_PASSWORD")

	RedshiftFastHost = requiredEnv("FAST_PG_HOST")
	RedshiftFastPort = requiredEnv("FAST_PG_PORT")
	RedshiftFastDatabase = requiredEnv("FAST_PG_DATABASE")
	RedshiftFastUsername = requiredEnv("FAST_PG_USER")
	RedshiftFastPassword = requiredEnv("FAST_PG_PASSWORD")
}

// ParseChecks reads in the latency check definitions
func ParseChecks(latencyConfigPath string) ClusterChecks {
	latencyJSON, err := ioutil.ReadFile(latencyConfigPath)
	if err != nil {
		l.GetKVLogger().CriticalD("read-latency-config-error", l.M{"error": err.Error()})
		os.Exit(1)
	}

	var checks ClusterChecks
	err = json.Unmarshal(latencyJSON, &checks)
	if err != nil {
		l.GetKVLogger().CriticalD("parse-latency-checks-error", l.M{"error": err.Error()})
		os.Exit(1)
	}

	return checks
}

// requiredEnv tries to find a value in the environment variables. If a value is not
// found the program will panic.
func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		l.GetKVLogger().CriticalD("required-env", l.M{"name": key})
		os.Exit(1)
	}
	return value
}
