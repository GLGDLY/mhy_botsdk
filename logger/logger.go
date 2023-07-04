package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type DefaultLogger struct {
	bot_id         string
	last_datetime  string
	level_console  LoggerLevel
	level_file     LoggerLevel
	logger_console *log.Logger
	logger_file    *log.Logger
}

func (l *DefaultLogger) ensureFile() {
	datetime := time.Now().Format("2006_01_02")
	if l.last_datetime == datetime {
		return
	}
	l.last_datetime = datetime
	var err error
	err = os.MkdirAll(filepath.Join(".", "log"), os.ModePerm)
	if err != nil {
		l.Error("create log dir error : " + err.Error())
	}
	err = os.MkdirAll(filepath.Join(".", "log", l.bot_id), os.ModePerm)
	if err != nil {
		l.Error("create log dir error : " + err.Error())
	}
	log_file, err := os.OpenFile(filepath.Join(".", "log", l.bot_id, datetime+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		l.Error("open log file error : " + err.Error())
	}
	l.logger_file = log.New(log_file, "", log.LstdFlags)
}

// 设置日志输出console的最低级别，默认为LoggerLevelInfo
func (l *DefaultLogger) SetConsoleLevel(level LoggerLevel) {
	l.level_file = level
}

// 设置日志输出file的最低级别，默认为LoggerLevelDebug
func (l *DefaultLogger) SetFileLevel(level LoggerLevel) {
	l.level_file = level
}

func (l *DefaultLogger) console_formatter(level LoggerLevel, text string) string {
	switch level {
	case LoggerLevelDebug:
		return "[DEBUG] " + text
	case LoggerLevelInfo:
		return "\033[1;32m[INFO]\033[0m " + text
	case LoggerLevelWarn:
		return "\033[1;33m[WARN]\033[0m " + text
	case LoggerLevelError:
		return "\033[1;31m[ERROR]\033[0m " + text
	default:
		return "[UNKNOWN] " + text
	}
}

func (l *DefaultLogger) file_formatter(level LoggerLevel, text string) string {
	switch level {
	case LoggerLevelDebug:
		return "[DEBUG] " + text
	case LoggerLevelInfo:
		return "[INFO] " + text
	case LoggerLevelWarn:
		return "[WARN] " + text
	case LoggerLevelError:
		return "[ERROR] " + text
	default:
		return "[UNKNOWN] " + text
	}
}

func (l *DefaultLogger) log(level LoggerLevel, text string) {
	if level >= l.level_console {
		l.logger_console.Print(l.console_formatter(level, text))
	}
	if level >= l.level_file {
		l.ensureFile()
		l.logger_file.Print(l.file_formatter(level, text))
	}
}

func (l *DefaultLogger) Log(level LoggerLevel, v ...interface{}) {
	l.log(level, fmt.Sprintln(v...))
}

func (l *DefaultLogger) Logf(level LoggerLevel, format string, v ...interface{}) {
	l.log(level, fmt.Sprintf(format, v...))
}

func (l *DefaultLogger) Debug(v ...interface{}) {
	l.Log(LoggerLevelDebug, v...)
}

func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	l.Logf(LoggerLevelDebug, format, v...)
}

func (l *DefaultLogger) Info(v ...interface{}) {
	l.Log(LoggerLevelInfo, v...)
}

func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	l.Logf(LoggerLevelInfo, format, v...)
}

func (l *DefaultLogger) Warn(v ...interface{}) {
	l.Log(LoggerLevelWarn, v...)
}

func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	l.Logf(LoggerLevelWarn, format, v...)
}

func (l *DefaultLogger) Error(v ...interface{}) {
	l.Log(LoggerLevelError, v...)
}

func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	l.Logf(LoggerLevelError, format, v...)
}

func NewDefaultLogger(_bot_id string) *DefaultLogger {
	l := &DefaultLogger{
		bot_id:         _bot_id,
		level_console:  LoggerLevelInfo,
		level_file:     LoggerLevelDebug,
		logger_console: log.New(os.Stdout, "", log.LstdFlags),
	}
	l.ensureFile()
	return l
}
