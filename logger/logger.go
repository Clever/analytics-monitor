package logger

import (
	kvLogger "gopkg.in/Clever/kayvee-go.v6/logger"
)

// Logger is exposed as an interface to allow for mocking
// out the specific log functions we use here
type Logger interface {
	JobFinishedEvent(payload string, didSucceed bool)
	CheckLatencyEvent(latencyErrValue int, fullTableName, reportedLatency, threshold string)
}

// M is an alias for map[string]interface{} to make log lines less painful to write.
type M kvLogger.M

// logger is the default implementation of Logger
type logger struct {
	log kvLogger.KayveeLogger
}

const (
	// jobFinished refers to job completions
	jobFinished = "job-finished"

	// checkLatency refers to latency check results
	checkLatency = "check-latency"
)

var defaultLog logger

func init() {
	defaultLog = logger{log: kvLogger.New("analytics-pipeline-monitor")}
}

// SetGlobalRouting installs a new log router with the input config
func SetGlobalRouting(kvconfigPath string) error {
	return kvLogger.SetGlobalRouting(kvconfigPath)
}

// GetLogger returns a reference to the logger singleton
func GetLogger() Logger {
	return &defaultLog
}

// GetKVLogger returns a reference to the underlying Kayvee logger
func GetKVLogger() kvLogger.KayveeLogger {
	return defaultLog.log
}

// JobFinishedEvent logs when analytics-pipeline-monitor has completed
// along with payload and success/failure
func (l *logger) JobFinishedEvent(payload string, didSucceed bool) {
	value := 0
	if didSucceed {
		value = 1
	}
	l.log.GaugeIntD(jobFinished, value, M{
		"payload": payload,
		"success": didSucceed,
	})
}

// CheckLatencyEvent logs the results of a latency check
// to be log routed to SignalFx
func (l *logger) CheckLatencyEvent(latencyErrValue int, fullTableName, reportedLatency, threshold string) {
	l.log.GaugeIntD(checkLatency, latencyErrValue, M{
		"table":             fullTableName,
		"latency":           reportedLatency,
		"latency_threshold": threshold,
	})
}
