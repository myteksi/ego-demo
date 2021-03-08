// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package http

import (
	"net/http"

	"github.com/grab/ego/egofilters/http/security/context"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HttpClientWithCtx interface {
	HttpClient
	DoWithTracing(ctx context.Context, req *http.Request, spanName string) (*http.Response, error)
}

type httpClientImpl struct {
	HttpClient
}

func (client httpClientImpl) DoWithTracing(ctx context.Context, req *http.Request, spanName string) (*http.Response, error) {
	if len(spanName) > 0 {
		span := ctx.ActiveSpan().SpawnChild(spanName)
		defer span.FinishSpan()

		spanHeaders := span.GetContext()
		for k, vals := range spanHeaders {
			if len(vals) > 0 {
				req.Header.Add(k, vals[0])
			}
		}
	}
	return client.Do(req)
}

func NewHttpClientWithCtx(client HttpClient) HttpClientWithCtx {
	return &httpClientImpl{client}
}
