package logger

import (
	"gopkg.in/Clever/kayvee-go.v5/logger"
)

var log *logger.Logger

func init() {
	log = logger.New("analytics-pipeline-monitor")
}

// M is an alias for map[string]interface{} to make log lines less painful to write.
type M logger.M

// Info logs at the info level
func Info(title string, data M) {
	log.InfoD(title, data)
}

// Debug logs at the debug level
func Debug(title string, data M) {
	log.DebugD(title, data)
}

// Warning logs at the warning level
func Warning(title string, data M) {
	log.WarnD(title, data)
}

// Error logs an error at the error level
func Error(title string, data M) {
	log.ErrorD(title, data)
}

// Critical logs at the critical level
func Critical(title string, data M) {
	log.CriticalD(title, data)
}

// Counter logs a counter with a value of 1
func Counter(title string) {
	log.Counter(title)
}

// CounterD logs a counter with a value of <count>
func CounterD(title string, count int, data M) {
	log.CounterD(title, count, data)
}

// GaugeFloatD logs a gauge with a value of <value>
func GaugeFloatD(title string, value float64, data M) {
	log.GaugeFloatD(title, value, data)
}
