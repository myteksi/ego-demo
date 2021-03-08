// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package http

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contextmocks "github.com/grab/ego/egofilters/mock/gen/http/security/context"
	"github.com/grab/ego/egofilters/mock/gen/http/security/http"
	egomocks "github.com/grab/ego/ego/test/go/mock/gen/envoy"
)

func TestDoWithTracing(t *testing.T) {
	tcs := []struct {
		name           string
		spanName       string
		existingHeader http.Header

		spanContext    map[string][]string
		expectedHeader http.Header
	}{
		{
			name:     "add headers to http request",
			spanName: "span_name",

			spanContext: map[string][]string{"X-Tracing-Id": {"123"}},
			expectedHeader: http.Header{
				"X-Tracing-Id": []string{"123"},
			},
		},
		{
			name:           "add multiple headers to http request",
			spanName:       "span_name",
			existingHeader: http.Header{"X-Existing": {"1"}},

			spanContext: map[string][]string{"X-Tracing-Id": {"125"}, "X-Span-Id": {"124", "ignored"}},
			expectedHeader: http.Header{
				"X-Tracing-Id": []string{"125"},
				"X-Span-Id":    []string{"124"},
				"X-Existing":   []string{"1"},
			},
		},
		{
			name:     "should handle nil span context",
			spanName: "span_name",

			spanContext:    nil,
			expectedHeader: http.Header{},
		},
		{
			name:     "should neither spawn child span nor add headers to http request if span name is empty",
			spanName: "",

			spanContext:    map[string][]string{"X-Tracing-Id": {"123"}},
			expectedHeader: http.Header{},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &contextmocks.Context{}

			childSpan := &egomocks.Span{}
			if len(tc.spanName) > 0 {
				activeSpan := &egomocks.Span{}
				ctx.On("ActiveSpan").Return(activeSpan)

				childSpan := &egomocks.Span{}
				activeSpan.On("SpawnChild", tc.spanName).Return(childSpan)

				childSpan.On("GetContext").Return(tc.spanContext)
				childSpan.On("FinishSpan").Times(1)
			}

			httpClient := &mocks.HttpClient{}
			clientWithCtx := NewHttpClientWithCtx(httpClient)

			httpResponse := &http.Response{}
			httpErr := errors.New("an error")
			var actualReq *http.Request
			httpClient.On("Do", mock.AnythingOfType("*http.Request")).Run(func(args mock.Arguments) {
				actualReq = args[0].(*http.Request)
			}).Return(httpResponse, httpErr)

			req, _ := http.NewRequest(http.MethodPost, "http://test.com", nil)
			for k, v := range tc.existingHeader {
				req.Header[k] = v
			}

			actualResp, actualErr := clientWithCtx.DoWithTracing(ctx, req, tc.spanName)

			assert.Equal(t, httpResponse, actualResp)
			assert.Equal(t, httpErr, actualErr)

			childSpan.AssertExpectations(t)
			httpClient.AssertExpectations(t)

			assert.Equal(t, tc.expectedHeader, actualReq.Header)
		})
	}

}
