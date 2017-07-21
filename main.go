package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/kardianos/osext"

	"github.com/Clever/analytics-pipeline-monitor/config"
	"github.com/Clever/analytics-pipeline-monitor/db"
	l "github.com/Clever/analytics-pipeline-monitor/logger"
)

var (
	latencyConfigPath    string
	logger               l.Logger
	globalDefaultLatency string
)

// Checks stores table checks in a nested map,
// indexed by schema name and then table name
type Checks map[string]map[string]config.TableCheck

func init() {
	logger = l.GetLogger()

	// kvconfig.yml must live in the same directory as
	// the executable file in order for log routing to work
	dir, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}
	err = l.SetGlobalRouting(path.Join(dir, "kvconfig.yml"))
	if err != nil {
		log.Fatal(err)
	}

	latencyConfigPath = path.Join(dir, "config/latency_config.json")
	globalDefaultLatency = "24h"
}

func main() {
	flag.Parse()
	config.Parse()

	defer logger.JobFinishedEvent(strings.Join(os.Args[1:], " "), true)

	prodConnection, err := db.NewRedshiftProdClient()
	fatalIfErr(err, "redshift-prod-failed-init")

	fastConnection, err := db.NewRedshiftFastClient()
	fatalIfErr(err, "redshift-fast-failed-init")

	configChecks := config.ParseChecks(latencyConfigPath)

	prodChecks := buildLatencyChecks(configChecks.ProdChecks, prodConnection)
	fastProdChecks := buildLatencyChecks(configChecks.FastProdChecks, fastConnection)

	prodErrors := performLatencyChecks(prodConnection, prodChecks)
	fastProdErrors := performLatencyChecks(fastConnection, fastProdChecks)

	performLoadErrorsCheck(prodConnection)
	performLoadErrorsCheck(fastConnection)

	queryLatencyErrors := append(prodErrors, fastProdErrors...)
	if len(queryLatencyErrors) > 0 {
		var errStrs []string
		for _, latencyErr := range queryLatencyErrors {
			l.GetKVLogger().CriticalD("query-latency-error", l.M{"errors": latencyErr.Error()})
			errStrs = append(errStrs, latencyErr.Error())
		}

		logger.JobFinishedEvent(strings.Join(os.Args[1:], " "), false)
		log.Fatalf("Encountered fatal error querying for latency: %s", strings.Join(errStrs, ","))
	}
}

// fatalIfErr logs a critical error. Assumes logger is initialized
func fatalIfErr(err error, title string) {
	if err != nil {
		logger.JobFinishedEvent(strings.Join(os.Args[1:], " "), false)
		l.GetKVLogger().CriticalD(title, l.M{"error": err.Error()})
		panic(fmt.Sprintf("Encountered fatal error: %s", err.Error()))
	}
}

// buildLatencyChecks constructs the latency checks for a given Redshift instance
// Each check can either be declared explicitly (by specifying table latency
// in schemaConfigs), or implicitly (by falling back on the default latency
// values specified at the schema level).
//
// Returns: a map of checks for each cluster.
// Each map of checks is indexed by cluster name, then table name.
// Each check (see: config.TableCheck) contains:
// A.) Latency threshold as a duration string
// B.) Name of the timestamp column
func buildLatencyChecks(schemaConfigs []config.SchemaConfig, redshiftClient db.RedshiftClient) Checks {
	checks := make(Checks)

	for _, schemaConfig := range schemaConfigs {
		schemaName := schemaConfig.SchemaName
		checks[schemaName] = make(map[string]config.TableCheck)

		tableMetadata, err := redshiftClient.QueryTableMetadata(schemaName)
		if err != nil {
			l.GetKVLogger().CriticalD("query-table-metadata-error", l.M{"error": err.Error()})
			panic("Unable to query table metadata")
		}

		for tableName, metadata := range tableMetadata {
			// Use inferred timestamp column if not specified in schema default
			timestampColumn := schemaConfig.DefaultTimestampColumn
			if timestampColumn == "" {
				timestampColumn = metadata.TimestampColumn
			}

			// Use global default latency if not specified in schema default
			defaultThreshold := schemaConfig.DefaultThreshold
			if defaultThreshold == "" {
				defaultThreshold = globalDefaultLatency
			}

			checks[schemaName][tableName] = config.TableCheck{
				TableName: tableName,
				Latency: config.LatencyInfo{
					TimestampColumn: timestampColumn,
					Threshold:       defaultThreshold,
				},
			}
		}

		// Override per-schema thresholds if specified in config
		for _, configCheck := range schemaConfig.Checks {
			tableName := configCheck.TableName
			if _, ok := checks[schemaName][configCheck.TableName]; ok {
				checks[schemaName][tableName] = config.TableCheck{
					TableName: tableName,
					Latency: config.LatencyInfo{
						TimestampColumn: configCheck.Latency.TimestampColumn,
						Threshold:       configCheck.Latency.Threshold,
					},
				}
			} else {
				l.GetKVLogger().WarnD("missing-table-in-db", l.M{
					"message": fmt.Sprintf("Can't check latency for %s.%s", schemaName, tableName),
				})
			}
		}

		// Finally, omit latency checks for specified tables
		for _, tableToOmit := range schemaConfig.TablesToOmit {
			if _, ok := checks[schemaName][tableToOmit]; ok {
				fmt.Printf("Omitting latency check for %s.%s", schemaName, tableToOmit)
				delete(checks[schemaName], tableToOmit)
			} else {
				l.GetKVLogger().WarnD("missing-table-in-db", l.M{
					"message": fmt.Sprintf("Omit latency for %s.%s is a no-op",
						schemaName, tableToOmit),
				})
			}
		}
	}

	return checks
}

func performLoadErrorsCheck(redshiftClient db.RedshiftClient) {
	loadErrors, err := redshiftClient.QuerySTLLoadErrors()
	if err != nil {
		fmt.Printf("Error with client performing load error check: %v.\n", err)
	} else {
		if loadErrors != nil && len(loadErrors) > 0 {
			loadErrorsJSON, err := json.Marshal(loadErrors)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			logger.CheckLoadErrorEvent(1, string(loadErrorsJSON))
		} else {
			// No load errors in past hour
			logger.CheckLoadErrorEvent(0, "")
		}
	}
}

func performLatencyChecks(redshiftClient db.RedshiftClient, checks Checks) []error {
	var queryLatencyErrors []error
	clusterName := redshiftClient.GetClusterName()

	for schemaName, tableChecks := range checks {
		for tableName, check := range tableChecks {
			threshold, err := time.ParseDuration(check.Latency.Threshold)
			fatalIfErr(err, "parse-duration-error")

			latencyHrs, hasRows, err := redshiftClient.QueryLatency(check.Latency.TimestampColumn,
				schemaName, tableName)
			if err != nil {
				queryLatencyErrors = append(queryLatencyErrors, err)
				continue
			}

			latencyErrValue := 0
			if !hasRows || float64(latencyHrs) > threshold.Hours() {
				latencyErrValue = 1
			}

			reportedLatency := fmt.Sprintf("%sh", strconv.FormatInt(latencyHrs, 10))
			if !hasRows {
				reportedLatency = "N/A - no rows"
			}

			fullTableName := fmt.Sprintf("%s.%s.%s", clusterName, schemaName, tableName)
			logger.CheckLatencyEvent(latencyErrValue, fullTableName, reportedLatency, check.Latency.Threshold)
		}
	}

	return queryLatencyErrors
}
