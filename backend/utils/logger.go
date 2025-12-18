package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
)

// Log levels
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
	LogLevelFatal = "FATAL"
)

// Logger struct
type Logger struct {
	prefix string
	file   *os.File
}

var (
	// Default logger instance
	defaultLogger *Logger
	// Enable/disable file logging
	enableFileLogging = true
	// Log file path
	logFilePath = "logs/app.log"
)

// init initializes the default logger
func init() {
	defaultLogger = &Logger{
		prefix: "[TraderCoin]",
	}

	// Create logs directory if not exists
	if enableFileLogging {
		if err := os.MkdirAll("logs", 0755); err != nil {
			log.Printf("Failed to create logs directory: %v", err)
			enableFileLogging = false
		} else {
			// Open log file in append mode
			file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Failed to open log file: %v", err)
				enableFileLogging = false
			} else {
				defaultLogger.file = file
			}
		}
	}
}

// formatLog formats the log message with timestamp, level, and caller info
func formatLog(level, message string, withColor bool) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Get caller info
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		// Get only filename without full path
		parts := strings.Split(file, "/")
		caller = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
	}

	// Format without color (for file logging)
	logMsg := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, caller, message)

	// Add color for console output
	if withColor {
		var color string
		switch level {
		case LogLevelDebug:
			color = ColorGray
		case LogLevelInfo:
			color = ColorGreen
		case LogLevelWarn:
			color = ColorYellow
		case LogLevelError:
			color = ColorRed
		case LogLevelFatal:
			color = ColorPurple
		default:
			color = ColorWhite
		}
		return fmt.Sprintf("%s%s%s", color, logMsg, ColorReset)
	}

	return logMsg
}

// writeLog writes log to console and file
func writeLog(level, message string) {
	// Console output with color
	consoleMsg := formatLog(level, message, true)
	fmt.Println(consoleMsg)

	// File output without color
	if enableFileLogging && defaultLogger.file != nil {
		fileMsg := formatLog(level, message, false)
		defaultLogger.file.WriteString(fileMsg + "\n")
	}
}

// LogDebug logs debug messages
func LogDebug(message string) {
	writeLog(LogLevelDebug, message)
}

// LogInfo logs info messages
func LogInfo(message string) {
	writeLog(LogLevelInfo, message)
}

// LogWarn logs warning messages
func LogWarn(message string) {
	writeLog(LogLevelWarn, message)
}

// LogError logs error messages
func LogError(message string) {
	writeLog(LogLevelError, message)
}

// LogFatal logs fatal messages and exits
func LogFatal(message string) {
	writeLog(LogLevelFatal, message)
	if defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
	os.Exit(1)
}

// LogWithFields logs with structured fields
func LogWithFields(level string, message string, fields map[string]interface{}) {
	var fieldStrings []string
	for key, value := range fields {
		fieldStrings = append(fieldStrings, fmt.Sprintf("%s=%v", key, value))
	}

	fullMessage := message
	if len(fieldStrings) > 0 {
		fullMessage = fmt.Sprintf("%s | %s", message, strings.Join(fieldStrings, ", "))
	}

	writeLog(level, fullMessage)
}

// Close closes the log file
func Close() {
	if defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}

// SetLogFile sets custom log file path
func SetLogFile(path string) error {
	if defaultLogger.file != nil {
		defaultLogger.file.Close()
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defaultLogger.file = file
	logFilePath = path
	enableFileLogging = true

	return nil
}

// DisableFileLogging disables file logging
func DisableFileLogging() {
	enableFileLogging = false
	if defaultLogger.file != nil {
		defaultLogger.file.Close()
		defaultLogger.file = nil
	}
}

// EnableFileLogging enables file logging
func EnableFileLogging() error {
	if !enableFileLogging {
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defaultLogger.file = file
		enableFileLogging = true
	}
	return nil
}
