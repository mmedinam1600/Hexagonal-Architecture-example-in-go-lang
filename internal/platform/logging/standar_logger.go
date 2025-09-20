package logging

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Info(message string, kv ...any)
	Warn(message string, kv ...any)
	Error(message string, kv ...any)
}

type StandardLogger struct{ *log.Logger }

func NewStd() Logger {
	return &StandardLogger{Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)}
}

func (l *StandardLogger) Info(message string, kv ...any) {
	l.Printf("INFO  %s %s", message, formatKeyValues(kv...))
}
func (l *StandardLogger) Warn(message string, kv ...any) {
	l.Printf("WARN  %s %s", message, formatKeyValues(kv...))
}
func (l *StandardLogger) Error(message string, kv ...any) {
	l.Printf("ERROR %s %s", message, formatKeyValues(kv...))
}

func formatKeyValues(kv ...any) string {
	if len(kv) == 0 {
		return ""
	}
	result := ""
	for i := 0; i < len(kv); i += 2 {
		key := fmt.Sprint(kv[i])
		var value any = "<missing>"
		if i+1 < len(kv) {
			value = kv[i+1]
		}
		result += fmt.Sprintf("%s=%v ", key, value)
	}
	return result
}
