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

// Config configures latency checks by cluster
type Config struct {
	RedshiftProdChecks []SchemaConfig `json:"redshift-prod"`
	RedshiftFastChecks []SchemaConfig `json:"redshift-fast"`
}

// SchemaConfig configures latency checks by schema
type SchemaConfig struct {
	SchemaName             string       `json:"schema"`
	DefaultThreshold       string       `json:"default_threshold"`
	DefaultTimestampColumn string       `json:"default_timestamp_column"`
	TablesToOmit           []string     `json:"omit_tables"`
	Checks                 []TableCheck `json:"checks"`
}

// TableCheck configures a single latency check for a table
type TableCheck struct {
	TableName string      `json:"table"`
	Latency   LatencyInfo `json:"latency"`
}

// LatencyInfo stores information for a latency check
// `threshold` expects a string formatted Golang duration
type LatencyInfo struct {
	TimestampColumn string `json:"timestamp_column"`
	Threshold       string `json:"threshold"`
}

// Parse reads environment variables and initializes the config.
func Parse() {
	RedshiftProdHost = requiredEnv("REDSHIFT_PROD_HOST")
	RedshiftProdPort = requiredEnv("REDSHIFT_PROD_PORT")
	RedshiftProdDatabase = requiredEnv("REDSHIFT_PROD_DATABASE")
	RedshiftProdUsername = requiredEnv("REDSHIFT_PROD_USER")
	RedshiftProdPassword = requiredEnv("REDSHIFT_PROD_PASSWORD")

	RedshiftFastHost = requiredEnv("REDSHIFT_FAST_HOST")
	RedshiftFastPort = requiredEnv("REDSHIFT_FAST_PORT")
	RedshiftFastDatabase = requiredEnv("REDSHIFT_FAST_DATABASE")
	RedshiftFastUsername = requiredEnv("REDSHIFT_FAST_USER")
	RedshiftFastPassword = requiredEnv("REDSHIFT_FAST_PASSWORD")
}

// ParseChecks reads in the latency check definitions
func ParseChecks(latencyConfigPath string) Config {
	latencyJSON, err := ioutil.ReadFile(latencyConfigPath)
	if err != nil {
		l.GetKVLogger().CriticalD("read-latency-config-error", l.M{"error": err.Error()})
		panic("Unable to read latency config")
	}

	var checks Config
	err = json.Unmarshal(latencyJSON, &checks)
	if err != nil {
		l.GetKVLogger().CriticalD("parse-latency-checks-error", l.M{"error": err.Error()})
		panic("Unable to parse latency checks")
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
