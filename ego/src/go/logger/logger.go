// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package logger

import (
	"fmt"

	"github.com/grab/ego/ego/src/go/envoy/loglevel"
)

// Logger is an interface to encourage developers using this interface
type Logger interface {
	Trace(message string, data ...interface{})
	Debug(message string, data ...interface{})
	Info(message string, data ...interface{})
	Warn(message string, data ...interface{})
	Error(message string, data ...interface{})
	Critical(message string, data ...interface{})
	// TODO: Fatal just like Critical but include throw exception on C, not implmeneted yet
}

// Configure as neeeded for your log aggregator...
var Render func(string, ...interface{}) string = func(message string, data ...interface{}) string {
	if 0 < len(data) {
		message += fmt.Sprintf(" %+v", data)
	}
	return message
}

// Data is an alias to make easier using Log function
type Data map[string]interface{}

// NewLogger with utility wrapper log level & tag
func NewLogger(tag string, native NativeLogger) Logger {
	finalNativeLogger := native
	if finalNativeLogger == nil {
		finalNativeLogger = defaultLogger
	}
	return envoyLogger{tag: tag, native: finalNativeLogger}
}

// NewDefaultLogger with utility wrapper log level & tag
func NewDefaultLogger(tag string) Logger {
	return envoyLogger{tag: tag, native: defaultLogger}
}

var defaultLogger NativeLogger

func Init(native NativeLogger) {
	defaultLogger = native
}

type envoyLogger struct {
	tag    string
	native NativeLogger
}

func (l envoyLogger) Trace(message string, data ...interface{}) {
	l.log(loglevel.Trace, l.tag, message, data...)
}

func (l envoyLogger) Debug(message string, data ...interface{}) {
	l.log(loglevel.Debug, l.tag, message, data...)
}

func (l envoyLogger) Info(message string, data ...interface{}) {
	l.log(loglevel.Info, l.tag, message, data...)
}

func (l envoyLogger) Warn(message string, data ...interface{}) {
	l.log(loglevel.Warn, l.tag, message, data...)
}

func (l envoyLogger) Error(message string, data ...interface{}) {
	l.log(loglevel.Error, l.tag, message, data...)
}

func (l envoyLogger) Critical(message string, data ...interface{}) {
	l.log(loglevel.Critical, l.tag, message, data...)
}

// Log is a wrapper for suggesting developers use it with log.Data
func (l envoyLogger) log(level loglevel.Type, tag, message string, data ...interface{}) {
	l.native.Log(level, tag, Render(message, data...))
}

// NativeLogger is an interface to C Logger
type NativeLogger interface {
	Log(level loglevel.Type, tag, message string)
}
