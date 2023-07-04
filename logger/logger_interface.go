package logger

/* --------- enum LoggerLevel start --------- */

type LoggerLevel uint

const (
	LoggerLevelDebug LoggerLevel = 10
	LoggerLevelInfo  LoggerLevel = 20
	LoggerLevelWarn  LoggerLevel = 30
	LoggerLevelError LoggerLevel = 40
)

/* --------- enum LoggerLevel end --------- */

/* --------- interface LoggerInterface start --------- */

type LoggerInterface interface {
	Log(level LoggerLevel, v ...interface{})
	Logf(level LoggerLevel, format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}

/* --------- interface LoggerInterface end --------- */
