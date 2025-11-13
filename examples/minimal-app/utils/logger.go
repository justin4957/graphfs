/*
# Module: utils/logger.go
Logging utility for structured logging.

Provides logging functionality with different levels (Info, Debug, Warn, Error).
Used throughout the application for observability and debugging.

## Linked Modules
None (utility module with no dependencies)

## Tags
utility, logging, observability

## Exports
Logger, NewLogger, LogLevel

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .

<#logger.go> a code:Module ;
    code:name "utils/logger.go" ;
    code:description "Logging utility for structured logging" ;
    code:language "go" ;
    code:layer "utility" ;
    code:exports <#Logger>, <#NewLogger>, <#LogLevel> ;
    code:tags "utility", "logging", "observability" ;
    code:isLeaf true .

<#Logger> a code:Type ;
    code:name "Logger" ;
    code:kind "struct" ;
    code:description "Structured logger instance" ;
    code:hasMethod <#Logger.Info>, <#Logger.Debug>, <#Logger.Warn>, <#Logger.Error> .

<#Logger.Info> a code:Method ;
    code:name "Info" ;
    code:description "Logs info-level message" .

<#Logger.Debug> a code:Method ;
    code:name "Debug" ;
    code:description "Logs debug-level message" .

<#Logger.Warn> a code:Method ;
    code:name "Warn" ;
    code:description "Logs warning message" .

<#Logger.Error> a code:Method ;
    code:name "Error" ;
    code:description "Logs error message" .

<#LogLevel> a code:Type ;
    code:name "LogLevel" ;
    code:kind "enum" ;
    code:description "Logging level enumeration" ;
    code:hasValue "INFO", "DEBUG", "WARN", "ERROR" .

<#NewLogger> a code:Function ;
    code:name "NewLogger" ;
    code:description "Creates new logger instance" ;
    code:returns <#Logger> .
<!-- End LinkedDoc RDF -->
*/

package utils

import (
	"fmt"
	"time"
)

// LogLevel represents logging levels
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	DEBUG LogLevel = "DEBUG"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// Logger provides structured logging
type Logger struct {
	level LogLevel
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		level: INFO,
	}
}

// Info logs an info-level message
func (l *Logger) Info(message string) {
	l.log(INFO, message)
}

// Debug logs a debug-level message
func (l *Logger) Debug(message string) {
	l.log(DEBUG, message)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log(WARN, message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.log(ERROR, message)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s: %s\n", timestamp, level, message)
}
