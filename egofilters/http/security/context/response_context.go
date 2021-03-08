// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package context

import (
	gocontext "context"
	"io"

	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/logger"
)

// SignResponse Sign response object for a Callbacks.
type SignResponse struct {
	StatusCode   int
	HeadersToSet map[string]string
}

type ResponseCallbacks interface {
	OnCompleteSigning(SignResponse)
}

type ResponseContext interface {
	Context
	AuthResponse() AuthResponse
	GoContext() gocontext.Context
	BodyReader() io.Reader
	GetSecret(string) string
	Callbacks() ResponseCallbacks
	Headers() envoy.ResponseHeaderMap
	RequestHeaders() envoy.RequestHeaderMap
	Logger() logger.Logger
}

type responseContextImpl struct {
	goContext      gocontext.Context
	authResponse   AuthResponse
	headers        envoy.ResponseHeaderMap
	requestHeaders envoy.RequestHeaderMap
	bodyReader     io.Reader
	secrets        map[string]string
	callbacks      ResponseCallbacks
	logger         logger.Logger
	activeSpan     envoy.Span
}

func (c *responseContextImpl) AuthResponse() AuthResponse {
	return c.authResponse
}

func (c *responseContextImpl) BodyReader() io.Reader {
	return c.bodyReader
}

func (c *responseContextImpl) GoContext() gocontext.Context {
	return c.goContext
}

func (c *responseContextImpl) GetSecret(key string) string {
	return c.secrets[key]
}

func (c *responseContextImpl) Callbacks() ResponseCallbacks {
	return c.callbacks
}

func (c *responseContextImpl) Headers() envoy.ResponseHeaderMap {
	return c.headers
}

func (c *responseContextImpl) RequestHeaders() envoy.RequestHeaderMap {
	return c.requestHeaders
}

func (c *responseContextImpl) Logger() logger.Logger {
	return c.logger
}

func (c *responseContextImpl) ActiveSpan() envoy.Span {
	return c.activeSpan
}

func CreateResponseContext(callbacks ResponseCallbacks, goContext gocontext.Context, activeSpan envoy.Span,
	secrets map[string]string, authResponse AuthResponse, requestHeaders envoy.RequestHeaderMap, headers envoy.ResponseHeaderMap,
	bodyReader io.Reader, logger logger.Logger) ResponseContext {
	return &responseContextImpl{
		callbacks:      callbacks,
		goContext:      goContext,
		secrets:        secrets,
		authResponse:   authResponse,
		requestHeaders: requestHeaders,
		headers:        headers,
		bodyReader:     bodyReader,
		logger:         logger,
		activeSpan:     activeSpan,
	}
}
