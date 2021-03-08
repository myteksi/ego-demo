// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package envoy

import (
	"io"

	pb "github.com/grab/ego/ego/src/cc/goc/proto"

	"github.com/grab/ego/ego/src/go/envoy/lifespan"
	"github.com/grab/ego/ego/src/go/envoy/loglevel"
	"github.com/grab/ego/ego/src/go/envoy/statetype"
	"github.com/grab/ego/ego/src/go/envoy/stats"
	"github.com/grab/ego/ego/src/go/volatile"
)

type GoHttpFilterConfig interface {
	// Settings returns a pointer to the underlying envoy
	// configuration data (a protobuf Any field). The data
	// must not be modified nor must references be retained
	// beyond the duration of the filter factory creation
	// call. If in doubt, please use Copy() on the result.
	Settings() volatile.Bytes
	Scope() Scope
}

type GoHttpFilter interface {
	Post(uint64)
	DecoderCallbacks() DecoderFilterCallbacks
	EncoderCallbacks() EncoderFilterCallbacks
	Pin()
	Unpin()
	Log(loglevel.Type, string)
	ResolveMostSpecificPerGoFilterConfig(name string, route Route) interface{}
	GenericSecretProvider() GenericSecretConfigProvider
}

type StreamFilterCallbacks interface {
	StreamInfo() StreamInfo
	Route() Route
	RouteExisting() bool
	ActiveSpan() Span
}

type DecoderFilterCallbacks interface {
	StreamFilterCallbacks
	ContinueDecoding()
	SendLocalReply(responseCode int, body string, headers map[string]string, details string)
	AddDecodedData(buffer BufferInstance, streamingFilter bool)
	DecodingBuffer() BufferInstance
	EncodeHeaders(responseCode int, headers *pb.ResponseHeaderMap, endStream bool)
}

type EncoderFilterCallbacks interface {
	StreamFilterCallbacks
	EncodingBuffer() BufferInstance
	AddEncodedData(buffer BufferInstance, streamingFilter bool)
	ContinueEncoding()
}

// BufferInstance is a proxy for Envoy::Buffer::Instance
//
// See //envoy/include/envoy/http/header_map.h
type BufferInstance interface {
	// Copy up to len(p) bytes of data from the buffer to p and return the
	// actual number of bytes retrieved.
	CopyOut(start uint64, p []byte) int

	// Retrieves all raw slices
	GetRawSlices() []volatile.Bytes

	// Retrieve the net total length of data stored in this buffer
	Length() uint64

	// NewReader is a Go convenience function
	NewReader(start uint64) io.Reader
}

// RequestHeaderMap is a proxy for Envoy::Http::RequestHeaderMap
//
// See //envoy/include/envoy/http/header_map.h
type HeaderMap interface {
	HeaderMapReadOnly
	headerMapUpdatable
}

type HeaderMapReadOnly interface {
	Get(name string) volatile.String
}

type headerMapUpdatable interface {
	AddCopy(name, value string)
	SetCopy(name, value string)
	AppendCopy(name, value string)
	Remove(name string)
}

type RequestOrResponseHeaderMap interface {
	RequestOrResponseHeaderMapReadOnly
	requestOrResponseHeaderMapUpdatable
}

type RequestOrResponseHeaderMapReadOnly interface {
	HeaderMapReadOnly
	ContentType() volatile.String
}

type requestOrResponseHeaderMapUpdatable interface {
	headerMapUpdatable
}

type RequestHeaderMap interface {
	RequestHeaderMapReadOnly
	requestHeaderMapUpdatable
}

type RequestHeaderMapReadOnly interface {
	RequestOrResponseHeaderMapReadOnly
	Path() volatile.String
	Method() volatile.String
	Authorization() volatile.String
	// There is no method with this name in Envoy.
	// It's a utilitity for Go filters to query headers by prefix.
	GetByPrefix(prefix string) map[string][]string
}

type requestHeaderMapUpdatable interface {
	requestOrResponseHeaderMapUpdatable
	SetPath(path string)
}

type RequestTrailerMap interface {
	RequestTrailerMapReadOnly
	requestTrailerMapUpdatable
}

type requestTrailerMapUpdatable interface {
	headerMapUpdatable
}

type RequestTrailerMapReadOnly interface {
	HeaderMapReadOnly
}

type ResponseHeaderMap interface {
	ResponseHeaderMapReadOnly
	responseHeaderMapUpdatable
}

type ResponseHeaderMapReadOnly interface {
	RequestOrResponseHeaderMapReadOnly
	Status() volatile.String
}

type responseHeaderMapUpdatable interface {
	requestOrResponseHeaderMapUpdatable
	SetStatus(status int)
}

type StreamInfo interface {
	FilterState() FilterState
	LastDownstreamTxByteSent() int64
	GetRequestHeaders() RequestHeaderMapReadOnly
	ResponseCode() int
	ResponseCodeDetails() volatile.String
}

type FilterState interface {
	SetData(name, value string, stateType statetype.Type, lifeSpan lifespan.Type)
	GetDataReadOnly(name string) (volatile.String, bool)
}

type Route interface {
	RouteEntry() RouteEntry
}

type RouteEntry interface {
	PathMatchCriterion() PathMatchCriterion
}

type PathMatchType int

const (
	PathMatchNone PathMatchType = iota
	PathMatchPrefix
	PathMatchExact
	PathMatchRegex
)

type PathMatchCriterion interface {
	MatchType() (PathMatchType, error)
	Matcher() (volatile.String, error)
}

type GenericSecretConfigProvider interface {
	Secret() volatile.String
}

// Scope envoy/include/envoy/stats/scope.h
type Scope interface {
	CounterFromStatName(name string) Counter
	GaugeFromStatName(name string, importMode stats.ImportMode) Gauge
	HistogramFromStatName(name string, unit stats.Unit) Histogram
}

// Counter envoy/include/envoy/stats/stats.h
type Counter interface {
	Add(amount uint64)
	Inc()
	Latch() uint64
	Reset()
	Value() uint64
}

// Gauge envoy/include/envoy/stats/stats.h
type Gauge interface {
	Add(amount uint64)
	Dec()
	Inc()
	Set(value uint64)
	Sub(amount uint64)
	Value() uint64
}

// Histogram envoy/include/envoy/stats/histogram.h
type Histogram interface {
	Unit() stats.Unit
	RecordValue(value uint64)
}

type Span interface {
	GetContext() map[string][]string
	SpawnChild(name string) Span
	FinishSpan()
}
