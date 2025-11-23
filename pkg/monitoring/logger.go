package monitoring

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel represents logging severity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging
type Logger struct {
	level  LogLevel
	mu     sync.Mutex
	output *log.Logger
	fields map[string]interface{}
}

// NewLogger creates a new logger
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		output: log.New(os.Stdout, "", 0),
		fields: make(map[string]interface{}),
	}
}

// WithField adds a field to the logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		level:  l.level,
		output: l.output,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

// WithFields adds multiple fields to the logger context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		level:  l.level,
		output: l.output,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// log writes a log message
func (l *Logger) log(level LogLevel, msg string) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Build log message
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logMsg := fmt.Sprintf("[%s] %s: %s", timestamp, level, msg)

	// Add fields
	if len(l.fields) > 0 {
		logMsg += " |"
		for k, v := range l.fields {
			logMsg += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	l.output.Println(logMsg)

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(INFO, msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(WARN, msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(ERROR, msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.log(FATAL, msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, fmt.Sprintf(format, args...))
}

// Global logger instance
var globalLogger = NewLogger(INFO)

// SetGlobalLevel sets the global log level
func SetGlobalLevel(level LogLevel) {
	globalLogger.level = level
}

// Global logging functions
func Debug(msg string) {
	globalLogger.Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

func Info(msg string) {
	globalLogger.Info(msg)
}

func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

func Warn(msg string) {
	globalLogger.Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

func Error(msg string) {
	globalLogger.Error(msg)
}

func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

func Fatal(msg string) {
	globalLogger.Fatal(msg)
}

func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}
