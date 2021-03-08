// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package context

import (
	gocontext "context"
	"io"
	"net/http"

	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/logger"
)

type AuthStatus int

const (
	// TODO: Change AuthOK to another number as 0 is default value of int
	// and we don't want AuthOk by default.
	AuthOK     AuthStatus = 0
	AuthError  AuthStatus = 1
	AuthDenied AuthStatus = 3
)

const (
	FilterStatePrefix = "egodemo.security.ctx.session."
)

// Authentication response object for a Callbacks.
type AuthResponse struct {
	// Required call status.
	Status AuthStatus
	// Required http status used only on denied response.
	StatusCode int
	// Optional http body used only on denied response.
	Body string
	// Optional http headers used on either denied or ok responses.
	HeadersToRemove map[string]struct{}
	// Optional http headers used on either denied or ok responses.
	HeadersToSet map[string]string
	// Optional http headers used on either denied or ok responses.
	HeadersToAppend map[string]string
	// Filter State. FilterStatePrefix will be added to keys
	// before storing in FilterState
	FilterState map[string]string
}

func AuthResponseOK() AuthResponse {
	return AuthResponse{
		Status:     AuthOK,
		StatusCode: http.StatusOK,
	}
}

func AuthResponseUnauthorized() AuthResponse {
	return AuthResponse{
		Status:     AuthDenied,
		StatusCode: http.StatusUnauthorized,
	}
}

func AuthResponseDenied(statusCode int) AuthResponse {
	return AuthResponse{
		Status:     AuthDenied,
		StatusCode: statusCode,
	}
}

func AuthResponseError() AuthResponse {
	return AuthResponse{
		Status:     AuthError,
		StatusCode: http.StatusInternalServerError,
	}
}

type Callbacks interface {
	OnComplete(AuthResponse)
}

type RequestContext interface {
	Context
	Callbacks() Callbacks
	Headers() envoy.RequestHeaderMap
	GoContext() gocontext.Context
	BodyReader() io.Reader
	GetSecret(string) string
	Logger() logger.Logger
}

type requestContextImpl struct {
	callbacks  Callbacks
	goContext  gocontext.Context
	headers    envoy.RequestHeaderMap
	bodyReader io.Reader
	secrets    map[string]string
	activeSpan envoy.Span
	logger     logger.Logger
}

func (c *requestContextImpl) Callbacks() Callbacks {
	return c.callbacks
}

func (c *requestContextImpl) Headers() envoy.RequestHeaderMap {
	return c.headers
}

func (c *requestContextImpl) BodyReader() io.Reader {
	return c.bodyReader
}

func (c *requestContextImpl) GoContext() gocontext.Context {
	return c.goContext
}

func (c *requestContextImpl) GetSecret(key string) string {
	return c.secrets[key]
}

func (c *requestContextImpl) Logger() logger.Logger {
	return c.logger
}

func (c *requestContextImpl) ActiveSpan() envoy.Span {
	return c.activeSpan
}

func CreateRequestContext(callbacks Callbacks, goContext gocontext.Context, activeSpan envoy.Span,
	headers envoy.RequestHeaderMap, secrets map[string]string, bodyReader io.Reader, logger logger.Logger) RequestContext {
	return &requestContextImpl{
		callbacks:  callbacks,
		goContext:  goContext,
		headers:    headers,
		secrets:    secrets,
		logger:     logger,
		bodyReader: bodyReader,
		activeSpan: activeSpan,
	}
}
