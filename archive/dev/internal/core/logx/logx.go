package logx

import (
	"fmt"
	"log"
	"os"
)

type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
)

type Logger struct {
	l     *log.Logger
	level Level
}

func New(level Level) *Logger {
	return &Logger{
		l:     log.New(os.Stderr, "", log.LstdFlags),
		level: level,
	}
}

func (lg *Logger) Printf(level Level, format string, args ...any) {
	// Simple level gate (debug is most verbose)
	if lg.level != Debug && level == Debug {
		return
	}
	lg.l.Printf("[%s] %s", level, fmt.Sprintf(format, args...))
}
