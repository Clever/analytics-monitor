package config

import (
	"os"

	"github.com/Clever/analytics-pipeline-monitor/logger"
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

// requiredEnv tries to find a value in the environment variables. If a value is not
// found the program will panic.
func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Critical("required-env", logger.M{"name": key})
		os.Exit(1)
	}
	return value
}
