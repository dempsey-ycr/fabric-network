/// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package logging implements a logging infrastructure for Go. It supports
// different logging backends like syslog, file and memory. Multiple backends
// can be utilized with different log levels per backend and logger.
package logging

import (
	"fmt"
	// stdlog "log"
	"os"
)

var l = MustGetLogger("IHT")

type Password string

func (p Password) Redacted() interface{} {
	return Redact(string(p))
}

func SetLogModel(debug, nc bool) {
	backendByQueue := NewQueueLogBackend()
	b := SetBackend(backendByQueue)
	if debug {
		b.SetLevel(DEBUG, "")
	} else {
		b.SetLevel(INFO, "")
	}
}

// Fatal is equivalent to l.Critical(fmt.Sprint()) followed by a call to os.Exit(1).
func Fatal(args ...interface{}) {
	s := fmt.Sprint(args...)
	l.log(CRITICAL, "%s", s)
	os.Exit(1)
}

// Fatalf is equivalent to l.Critical followed by a call to os.Exit(1).
func Fatalf(format string, args ...interface{}) {
	l.log(CRITICAL, format, args...)
	os.Exit(1)
}

// Panic is equivalent to l.Critical(fmt.Sprint()) followed by a call to panic().
func Panic(args ...interface{}) {
	s := fmt.Sprint(args...)
	l.log(CRITICAL, "%s", s)
	panic(s)
}

// Panicf is equivalent to l.Critical followed by a call to panic().
func Panicf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	l.log(CRITICAL, "%s", s)
	panic(s)
}

// Critical logs a message using CRITICAL as log level.
func Critical(format string, args ...interface{}) {
	l.log(CRITICAL, format, args...)
}

// Error logs a message using ERROR as log level.
func Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Warning logs a message using WARNING as log level.
func Warning(format string, args ...interface{}) {
	l.log(WARNING, format, args...)
}

// Notice logs a message using NOTICE as log level.
func Notice(format string, args ...interface{}) {
	l.log(NOTICE, format, args...)
}

// Info logs a message using INFO as log level.
func Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Debug logs a message using DEBUG as log level.
func Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func Sql(format string, args ...interface{}) {
	l.log(SQL, format, args...)
}

func Redis(format string, args ...interface{}) {
	l.log(REDIS, format, args...)
}
