package logger

import (
	l "log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/Clever/kayvee-go.v6/logger"
)

func init() {
	err := logger.SetGlobalRouting("../kvconfig.yml")
	if err != nil {
		l.Fatal(err)
	}
}

// TestJobFinished verifies that JobFinishedEvent
// log routes to the 'job-finished' rule
func TestJobFinished(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		rule       string
		payload    string
		didSucceed bool
	}{
		{
			rule:       "job-finished",
			payload:    "",
			didSucceed: true,
		},
		{
			rule:       "job-finished",
			payload:    "",
			didSucceed: false,
		},
	}

	for _, test := range tests {
		t.Logf("Routing rule %s", test.rule)

		mocklog := logger.NewMockCountLogger("analytics-pipeline-monitor")
		log = mocklog // Overrides package level logger

		JobFinishedEvent(test.payload, test.didSucceed)
		counts := mocklog.RuleCounts()

		assert.Equal(counts[test.rule], 1)
	}
}

// TestCheckLatency verifies that CheckLatencyEvent
// log routes to the 'check-latency' rule that
// ultimately sends the log to SignalFx
func TestCheckLatency(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		rule             string
		errValue         int
		tableName        string
		latency          string
		latencyThreshold string
	}{
		{
			rule:             "check-latency",
			errValue:         0,
			tableName:        "mongo.districts",
			latency:          "1",
			latencyThreshold: "3h",
		},
		{
			rule:             "check-latency",
			errValue:         1,
			tableName:        "mongo.districts",
			latency:          "1",
			latencyThreshold: "3h",
		},
	}

	for _, test := range tests {
		t.Logf("Routing rule %s", test.rule)

		mocklog := logger.NewMockCountLogger("analytics-pipeline-monitor")
		log = mocklog // Overrides package level logger

		CheckLatencyEvent(test.errValue, test.tableName, test.latency, test.latencyThreshold)
		counts := mocklog.RuleCounts()

		assert.Equal(counts[test.rule], 1)
	}
}
