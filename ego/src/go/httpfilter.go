// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package ego

import (
	"context"

	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/envoy/datastatus"
	"github.com/grab/ego/ego/src/go/envoy/headersstatus"
	"github.com/grab/ego/ego/src/go/envoy/loglevel"
	"github.com/grab/ego/ego/src/go/envoy/trailersstatus"
	"github.com/grab/ego/ego/src/go/logger"
)

type HttpFilterFactory func(native envoy.GoHttpFilter) HttpFilter

type HttpFilter interface {
	StreamDecoderFilter
	StreamEncoderFilter
	OnPost(uint64)
	OnDestroy()
	Logger() logger.Logger
}

type StreamDecoderFilter interface {
	DecodeHeaders(envoy.RequestHeaderMap, bool) headersstatus.Type
	DecodeData(envoy.BufferInstance, bool) datastatus.Type
	DecodeTrailers(envoy.RequestTrailerMap) trailersstatus.Type
}

type StreamEncoderFilter interface {
	EncodeHeaders(envoy.ResponseHeaderMap, bool) headersstatus.Type
	EncodeData(envoy.BufferInstance, bool) datastatus.Type
}

type HttpFilterBase struct {
	Context context.Context
	Cancel  context.CancelFunc
	Native  envoy.GoHttpFilter
}

func (f *HttpFilterBase) Init(native envoy.GoHttpFilter) {
	f.Native = native
	f.Context, f.Cancel = context.WithCancel(context.Background())
}

func (f *HttpFilterBase) Logger() logger.Logger {
	return filterLogger{f.Native}
}

func (f *HttpFilterBase) Pin() {
	f.Native.Pin()
}

func (f *HttpFilterBase) Recover() {
	if err := recover(); err != nil {
		// TODO log error
	}
}

func (f *HttpFilterBase) Unpin() {
	f.Native.Unpin()
	f.Recover()
}

func (f *HttpFilterBase) OnDestroy() {
	f.Cancel()
}

func (f *HttpFilterBase) OnPost(tag uint64) {
}

func (f *HttpFilterBase) DecodeData(data envoy.BufferInstance, endStream bool) datastatus.Type {
	return datastatus.Continue
}

func (f *HttpFilterBase) DecodeTrailers(trailes envoy.RequestTrailerMap) trailersstatus.Type {
	return trailersstatus.Continue
}

func (f *HttpFilterBase) DecodeHeaders(headers envoy.RequestHeaderMap, endStream bool) headersstatus.Type {
	return headersstatus.Continue
}

func (f *HttpFilterBase) EncodeHeaders(headers envoy.ResponseHeaderMap, endStream bool) headersstatus.Type {
	return headersstatus.Continue
}

func (f *HttpFilterBase) EncodeData(envoy.BufferInstance, bool) datastatus.Type {
	return datastatus.Continue
}

type filterLogger struct {
	Native envoy.GoHttpFilter
}

func (l filterLogger) Trace(message string, data ...interface{}) {
	l.log(loglevel.Trace, message, data...)
}

func (l filterLogger) Debug(message string, data ...interface{}) {
	l.log(loglevel.Debug, message, data...)
}

func (l filterLogger) Info(message string, data ...interface{}) {
	l.log(loglevel.Info, message, data...)
}

func (l filterLogger) Warn(message string, data ...interface{}) {
	l.log(loglevel.Warn, message, data...)
}

func (l filterLogger) Error(message string, data ...interface{}) {
	l.log(loglevel.Error, message, data...)
}

func (l filterLogger) Critical(message string, data ...interface{}) {
	l.log(loglevel.Critical, message, data...)
}

func (l filterLogger) log(level loglevel.Type, message string, data ...interface{}) {
	l.Native.Log(level, logger.Render(message, data...))
}
