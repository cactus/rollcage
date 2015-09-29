// Package slog implements a very simple levelled logger
package slog

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"
)

var timeFormat = "2006-01-02T15:04:05.000000"
var Logger = New(INFO, timeFormat, "")

type severity int32 // sync/atomic int32

const (
	DEBUG severity = iota
	INFO
	WARN
	ERROR
	FATAL
)

var severityName = []string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

type LeveledLogger struct {
	timeformat string
	prefix     string
	severity   severity
	mx         sync.Mutex
}

func (l *LeveledLogger) header(s severity, t *time.Time) *bytes.Buffer {
	b := new(bytes.Buffer)
	if l.timeformat != "" {
		fmt.Fprintf(b, "%s ", t.Format(l.timeformat))
	}
	fmt.Fprintf(b, "%-5.5s %s", severityName[s], l.prefix)
	return b
}

func (l *LeveledLogger) logln(s severity, v ...interface{}) {
	if s >= l.severity {
		t := time.Now()
		buf := l.header(s, &t)
		fmt.Fprintln(buf, v...)
		buf.WriteTo(os.Stderr)
	}
}

func (l *LeveledLogger) logf(s severity, format string, v ...interface{}) {
	if s >= l.severity {
		t := time.Now()
		buf := l.header(s, &t)
		fmt.Fprintf(buf, format, v...)
		if buf.Bytes()[buf.Len()-1] != '\n' {
			buf.WriteByte('\n')
		}
		buf.WriteTo(os.Stderr)
	}
}

func (l *LeveledLogger) Write(p []byte) (n int, err error) {
	t := time.Now()
	buf := l.header(INFO, &t)
	buf.Write(p)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	written, err := buf.WriteTo(os.Stderr)
	if err != nil {
		return int(written), err
	}
	return int(written), nil
}

func (l *LeveledLogger) GetLevel() severity {
	return l.severity
}

func (l *LeveledLogger) SetLevel(s severity) {
	l.severity = s
}

func (l *LeveledLogger) IsDebug() bool {
	if l.severity == DEBUG {
		return true
	}
	return false
}

func (l *LeveledLogger) Debugf(format string, v ...interface{}) {
	l.logf(DEBUG, format, v...)
}

func (l *LeveledLogger) Debugln(v ...interface{}) {
	l.logln(DEBUG, v...)
}

func (l *LeveledLogger) Infof(format string, v ...interface{}) {
	l.logf(INFO, format, v...)
}

func (l *LeveledLogger) Infoln(v ...interface{}) {
	l.logln(INFO, v...)
}

func (l *LeveledLogger) Warnf(format string, v ...interface{}) {
	l.logf(WARN, format, v...)
}

func (l *LeveledLogger) Warnln(v ...interface{}) {
	l.logln(WARN, v...)
}

func (l *LeveledLogger) Errorf(format string, v ...interface{}) {
	l.logf(ERROR, format, v...)
}

func (l *LeveledLogger) Errorln(v ...interface{}) {
	l.logln(ERROR, v...)
}

func (l *LeveledLogger) Fatalf(format string, v ...interface{}) {
	l.logf(FATAL, format, v...)
	os.Exit(1)
}

func (l *LeveledLogger) Fatalln(v ...interface{}) {
	l.logln(FATAL, v...)
	os.Exit(1)
}

func (l *LeveledLogger) Panicf(format string, v ...interface{}) {
	l.logf(FATAL, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (l *LeveledLogger) Panicln(v ...interface{}) {
	l.logln(FATAL, v...)
	panic(fmt.Sprintln(v...))
}

func New(level severity, timeformat string, prefix string) *LeveledLogger {
	return &LeveledLogger{
		timeformat,
		prefix,
		level,
		sync.Mutex{},
	}
}

/*
// isatty returns true if f is a TTY, false otherwise.
func isatty(f *os.File) bool {
	switch runtime.GOOS {
	case "darwin":
	case "linux":
	default:
		return false
	}
	var t [2]byte
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		f.Fd(), syscall.TIOCGPGRP,
		uintptr(unsafe.Pointer(&t)))
	return errno == 0
}
*/

func GetLevel() severity {
	return Logger.GetLevel()
}

func SetLevel(s severity) {
	Logger.SetLevel(s)
}

func IsDebug() bool {
	return Logger.IsDebug()
}

func Debugf(format string, v ...interface{}) {
	Logger.Debugf(format, v...)
}

func Debugln(v ...interface{}) {
	Logger.Debugln(v...)
}

func Infof(format string, v ...interface{}) {
	Logger.Infof(format, v...)
}

func Infoln(v ...interface{}) {
	Logger.Infoln(v...)
}

func Warnf(format string, v ...interface{}) {
	Logger.Warnf(format, v...)
}

func Warnln(v ...interface{}) {
	Logger.Warnln(v...)
}

func Errorf(format string, v ...interface{}) {
	Logger.Errorf(format, v...)
}

func Errorln(v ...interface{}) {
	Logger.Errorln(v...)
}

func Fatalf(format string, v ...interface{}) {
	Logger.Fatalf(format, v...)
}

func Fatalln(v ...interface{}) {
	Logger.Fatalln(v...)
}

func Panicf(format string, v ...interface{}) {
	Logger.Panicf(format, v...)
}

func Panicln(v ...interface{}) {
	Logger.Panicln(v...)
}
