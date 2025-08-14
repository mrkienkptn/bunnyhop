package bunnyhop

import (
	"log"
)

// DefaultLogger implementation mặc định
type DefaultLogger struct {
	debugEnabled bool
}

func NewDefaultLogger(debugEnabled bool) *DefaultLogger {
	return &DefaultLogger{debugEnabled: debugEnabled}
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	if l.debugEnabled {
		log.Printf("[DEBUG] "+msg, args...)
	}
}
