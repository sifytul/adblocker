package logger

import (
    "fmt"
    "log"
    "os"
    "time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
    DEBUG LogLevel = iota
    INFO
    WARN
    ERROR
)

// String returns the string representation of log level
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
    default:
        return "UNKNOWN"
    }
}

// Logger is a structured logger
type Logger struct {
    level      LogLevel
    logger     *log.Logger
    timeFormat string
}

// NewLogger creates a new logger
func NewLogger(levelStr string, outputFile string) (*Logger, error) {
    // Parse level
    level := parseLevel(levelStr)
    
    // Set output
    var output *os.File
    if outputFile == "" {
        output = os.Stdout
    } else {
        var err error
        output, err = os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            return nil, fmt.Errorf("failed to open log file: %w", err)
        }
    }
    
    return &Logger{
        level:      level,
        logger:     log.New(output, "", 0), // We'll format ourselves
        timeFormat: "2006-01-02 15:04:05",
    }, nil
}

// parseLevel converts string to LogLevel
func parseLevel(levelStr string) LogLevel {
    switch levelStr {
    case "debug":
        return DEBUG
    case "info":
        return INFO
    case "warn":
        return WARN
    case "error":
        return ERROR
    default:
        return INFO
    }
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
    // Skip if below minimum level
    if level < l.level {
        return
    }
    
    // Format message
    timestamp := time.Now().Format(l.timeFormat)
    levelStr := level.String()
    message := fmt.Sprintf(format, args...)
    
    // Output: [2024-02-13 15:04:05] INFO: Message here
    l.logger.Printf("[%s] %s: %s", timestamp, levelStr, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
    l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
    l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
    l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
    l.log(ERROR, format, args...)
}

// Query logs a DNS query
func (l *Logger) Query(domain string, blocked bool, clientIP string) {
    if blocked {
        l.Info("BLOCKED: %s (from %s)", domain, clientIP)
    } else {
        l.Debug("ALLOWED: %s (from %s)", domain, clientIP)
    }
}
