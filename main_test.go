package main

import (
	"errors"
	"io/ioutil"
	"log"
	"path"
	"testing"

	"github.com/kardianos/osext"
	"github.com/stretchr/testify/assert"

	"github.com/Clever/analytics-pipeline-monitor/config"
)

// Copy kvconfig.yml to exec dir to simulate main.init()
// loading in a production Docker environment.
// This syntax is used to run the setup before main.init()
var _ = func() (_ struct{}) {
	configContent, err := ioutil.ReadFile("kvconfig.yml")
	if err != nil {
		log.Fatal(err)
	}
	execDir, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(path.Join(execDir, "kvconfig.yml"), configContent, 0777)
	if err != nil {
		log.Fatal(err)
	}
	return
}()

type mockRedshiftClient struct {
	latencyHrs int64
	hasRows    bool
	queryErr   error
}

func (c *mockRedshiftClient) QueryLatency(timestampColumn, schemaName, tableName string) (int64, bool, error) {
	return c.latencyHrs, c.hasRows, c.queryErr
}

type mockLogger struct {
	assertions            *assert.Assertions
	expectedLogValue      int
	expectedLatencyReport string
}

func (l *mockLogger) JobFinishedEvent(payload string, didSucceed bool) {
	// Dummy mocked to satisfy the Logger interface
	return
}

func (l *mockLogger) CheckLatencyEvent(latencyErrValue int, fullTableName, reportedLatency, threshold string) {
	l.assertions.Equal(latencyErrValue, l.expectedLogValue, "Incorrect latency log value")
	l.assertions.Equal(reportedLatency, l.expectedLatencyReport, "Mismatched latency report string")
}

// TestPerformLatencyChecks tests the performLatencyChecks
// function, mocking out latency results and verifying
// that the correct results are being logged
func TestPerformLatencyChecks(t *testing.T) {
	assertions := assert.New(t)

	tests := []struct {
		title string

		// Mocks out the results of QueryLatency
		latencyHrs int64
		hasRows    bool
		queryErr   error

		// Mocks out the config latency threshold
		threshold string

		// Specifies what we expect to log (or error)
		expectedLogValue      int
		expectedLatencyReport string
		expectedPanic         bool
	}{
		{
			title:                 "logs a success value (0) when latencyHrs <= threshold",
			latencyHrs:            1,
			hasRows:               true,
			queryErr:              nil,
			threshold:             "2h",
			expectedLogValue:      0,
			expectedLatencyReport: "1h",
		},
		{
			title:                 "logs a failure value (1) when latencyHrs > threshold",
			latencyHrs:            3,
			hasRows:               true,
			queryErr:              nil,
			threshold:             "2h",
			expectedLogValue:      1,
			expectedLatencyReport: "3h",
		},
		{
			title:                 "logs a failure value (1) when no rows exist",
			latencyHrs:            0,
			hasRows:               false,
			queryErr:              nil,
			threshold:             "2h",
			expectedLogValue:      1,
			expectedLatencyReport: "N/A - no rows",
		},
		{
			title:         "panics when threshold is malformatted",
			latencyHrs:    0,
			hasRows:       false,
			queryErr:      nil,
			threshold:     "2j",
			expectedPanic: true,
		},
		{
			title:         "panics when latency query errors out",
			latencyHrs:    0,
			hasRows:       false,
			queryErr:      errors.New("Data Warehouse out of space - s/Redshift/Blueshift"),
			threshold:     "2h",
			expectedPanic: true,
		},
	}

	for _, test := range tests {
		t.Logf("Testing that performLatencyChecks %s", test.title)

		mockRsClient := &mockRedshiftClient{
			latencyHrs: test.latencyHrs,
			hasRows:    test.hasRows,
			queryErr:   test.queryErr,
		}
		mockLog := &mockLogger{
			assertions:            assertions,
			expectedLogValue:      test.expectedLogValue,
			expectedLatencyReport: test.expectedLatencyReport,
		}
		logger = mockLog // Overrides package level logger
		mockChecks := []config.SchemaChecks{
			config.SchemaChecks{
				SchemaName: "mockSchemaName",
				Checks: []config.TableChecks{
					config.TableChecks{
						TableName: "mockTableName",
						Latency: config.LatencyInfo{
							TimestampColumn: "mockColumn",
							Threshold:       test.threshold,
						},
					},
				},
			},
		}

		if test.expectedPanic {
			assert.Panics(t, func() {
				performLatencyChecks(mockRsClient, mockChecks, "mockClusterName")
			}, "Doesn't error when expected")
		} else {
			performLatencyChecks(mockRsClient, mockChecks, "mockClusterName")
		}
	}
}
