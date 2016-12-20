package main

import (
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
	latencyConfigPath string
	logger            l.Logger
)

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
}

func main() {
	flag.Parse()
	config.Parse()

	defer logger.JobFinishedEvent(strings.Join(os.Args[1:], " "), true)

	fastConnection, err := db.NewRedshiftFastClient()
	fatalIfErr(err, "redshift-fast-failed-init")

	prodConnection, err := db.NewRedshiftProdClient()
	fatalIfErr(err, "redshift-prod-failed-init")

	latencyChecks := config.ParseChecks(latencyConfigPath)

	performLoadErrorsCheck(fastConnection)
	performLoadErrorsCheck(prodConnection)

	performLatencyChecks(fastConnection, latencyChecks.FastProdChecks, "fast-prod")
	performLatencyChecks(prodConnection, latencyChecks.ProdChecks, "prod")
}

// fatalIfErr logs a critical error. Assumes logger is initialized
func fatalIfErr(err error, title string) {
	if err != nil {
		logger.JobFinishedEvent(strings.Join(os.Args[1:], " "), false)
		l.GetKVLogger().CriticalD(title, l.M{"error": err.Error()})
		panic("Encountered fatal error")
	}
}

// TODO (IP-1204): Perform STL_LOAD_ERRORS Latency Check
// Doesn't need to return anything since Kayvee logging should be sufficient
func performLoadErrorsCheck(redshiftClient db.RedshiftClient) {

}

func performLatencyChecks(redshiftClient db.RedshiftClient, clusterConfig []config.SchemaChecks, clusterName string) {
	for _, schemaConfig := range clusterConfig {
		schemaName := schemaConfig.SchemaName
		for _, check := range schemaConfig.Checks {
			threshold, err := time.ParseDuration(check.Latency.Threshold)
			fatalIfErr(err, "parse-duration-error")

			latencyHrs, hasRows, err := redshiftClient.QueryLatency(check.Latency.TimestampColumn,
				schemaName, check.TableName)
			fatalIfErr(err, "query-latency-error")

			latencyErrValue := 0
			if !hasRows || float64(latencyHrs) > threshold.Hours() {
				latencyErrValue = 1
			}

			reportedLatency := strconv.FormatInt(latencyHrs, 10)
			if !hasRows {
				reportedLatency = "N/A - no rows"
			}

			fullTableName := fmt.Sprintf("%s.%s.%s", clusterName, schemaName, check.TableName)
			logger.CheckLatencyEvent(latencyErrValue, fullTableName, reportedLatency, check.Latency.Threshold)
		}
	}
}
