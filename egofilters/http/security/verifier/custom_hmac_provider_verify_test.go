// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package verifier

import (
	"bytes"
	gocontext "context"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/grab/ego/ego/src/go/logger"
	"github.com/grab/ego/ego/src/go/volatile"
	egomocks "github.com/grab/ego/ego/test/go/mock"

	"github.com/grab/ego/egofilters/http/security/context"
	pb "github.com/grab/ego/egofilters/http/security/proto"

	"github.com/grab/ego/egofilters/mock/gen/http/security/context"
	contextmocks "github.com/grab/ego/egofilters/mock/gen/http/security/context"
	httpmocks "github.com/grab/ego/egofilters/mock/gen/http/security/http"
	envoymocks "github.com/grab/ego/ego/test/go/mock/gen/envoy"
)

func TestHMACInvalidToken(t *testing.T) {

	tcs := []struct {
		name                string
		authorizationHeader string
	}{
		{name: "empty authorization header", authorizationHeader: ""},
		{name: "authorization header with only 1 part", authorizationHeader: "abc"},
		{name: "authorization header with more than 2 parts", authorizationHeader: "abc:def:ghc"},
	}

	for _, val := range tcs {
		tc := val
		t.Run(tc.name, func(t *testing.T) {
			ctx := &contextmocks.RequestContext{}

			callbacks := &mocks.Callbacks{}
			callbacks.On("OnComplete", context.AuthResponseUnauthorized())
			ctx.On("Callbacks").Return(callbacks)

			requetHeaderMap := &envoymocks.RequestHeaderMap{}
			requetHeaderMap.On("Authorization").Return(volatile.String(tc.authorizationHeader))
			ctx.On("Headers").Return(requetHeaderMap)

			ctx.On("Logger").Return(logger.NewLogger("CustomHMACLogger", egomocks.NativeLogger{}))

			provider, _ := CreateCustomHMACProvider(&pb.CustomHMACProvider{})

			provider.Verify(ctx)

			ctx.AssertExpectations(t)
			callbacks.AssertExpectations(t)
			requetHeaderMap.AssertExpectations(t)
		})
	}
}

func TestHMACVerify(t *testing.T) {
	var tcs = []struct {
		name           string
		settings       pb.CustomHMACProvider
		goCtx          gocontext.Context
		requestHeaders http.Header
		requestBody    string
		customAuthResp *http.Response
		httpErr        error
		validSignature bool
		validationErr  error

		headersToCustomAuth http.Header
		bodyToCustomAuth    string
		spanName            string
		headersToSet        map[string]string
		headersToRemove     map[string]struct{}
		filterState         map[string]string
		authResponse        context.AuthResponse
	}{
		{
			name: "should call custom auth provider with correct request headers and body",
			settings: pb.CustomHMACProvider{
				RequestValidationUrl:     "https://example.com/xyz",
				ServiceKey:               "service_key",
				ServiceToken:             "service_token",
				GeneratedUpstreamHeaders: []string{"x-custom-userid"},
				TracingEnabled:           true,
			},
			goCtx: gocontext.WithValue(gocontext.Background(), "key", "val"),
			requestHeaders: http.Header{
				"User-Agent":    {""},
				"X-Request-Id":  {"request-id1"},
				"Authorization": {"partner_id1:signature1"},
				":path":         {"/path1"},
				":method":       {"GET"},
				"Content-Type":  {"application/json"},
				"Date":          {"2015/1/1"},
			},
			requestBody:    "request_body1",
			customAuthResp: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader(""))},
			validSignature: true,

			headersToCustomAuth: http.Header{
				"Cache-Control":          {"no-cache"},
				"Content-Type":           {"application/json"},
				"Authorization":          {"Token decrypted_service_key decrypted_service_token"},
				"X-Request-Id":           {"request-id1"},
				"X-Custom-Auth-Date":      {"2015/1/1"},
				"X-Custom-Auth-Path":      {"/path1"},
				"X-Custom-Auth-Signature": {"signature1"},
				"X-Custom-Auth-Verb":      {"GET"},
			},
			bodyToCustomAuth: "request_body1",
			spanName:         "custom_hmac_verify",
			authResponse:     context.AuthResponseOK(),
			headersToSet: map[string]string{
				"x-custom-userid": "partner_id1",
			},
			headersToRemove: map[string]struct{}{
				"x-custom-userid": {},
			},
			filterState: map[string]string{
				"UserID":    "partner_id1",
			},
		},

		{
			name: "should only add supported generated_upstream_headers to requests to upstreams",
			settings: pb.CustomHMACProvider{
				RequestValidationUrl:     "https://custom-auth.example.com/xyz",
				ServiceKey:               "service_key",
				ServiceToken:             "service_token",
				GeneratedUpstreamHeaders: []string{"x-custom-unknown"},
				TracingEnabled:           false,
			},
			goCtx: gocontext.WithValue(gocontext.Background(), "key", "val"),
			requestHeaders: http.Header{
				"User-Agent":    {""},
				"X-Request-Id":  {"request-id1"},
				"Authorization": {"partner_id1:signature1"},
				":path":         {"/path1"},
				":method":       {"GET"},
				"Content-Type":  {"application/json"},
				"Date":          {"2015/1/1"},
			},
			requestBody:    "request_body1",
			spanName:       "",
			customAuthResp: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader(""))},
			validSignature: true,

			headersToCustomAuth: http.Header{
				"Cache-Control":          {"no-cache"},
				"Content-Type":           {"application/json"},
				"Authorization":          {"Token decrypted_service_key decrypted_service_token"},
				"X-Request-Id":           {"request-id1"},
				"X-Custom-Auth-Date":      {"2015/1/1"},
				"X-Custom-Auth-Path":      {"/path1"},
				"X-Custom-Auth-Signature": {"signature1"},
				"X-Custom-Auth-Verb":      {"GET"},
			},
			bodyToCustomAuth: "request_body1",
			authResponse:     context.AuthResponseOK(),
			headersToSet:     map[string]string{},
			headersToRemove:  map[string]struct{}{},
			filterState: map[string]string{
				"UserID":    "partner_id1",
			},
		},

		{
			name: "should return 500 status code if can't parse URL",
			settings: pb.CustomHMACProvider{
				RequestValidationUrl: ":zzz",
				ServiceKey:   "service_key2",
				ServiceToken: "service_token2",
			},
			goCtx: gocontext.WithValue(gocontext.Background(), "key", "val"),
			requestHeaders: http.Header{
				"User-Agent":    {""},
				"X-Request-Id":  {"request-id1"},
				"Authorization": {"partner_id1:signature1"},
				":path":         {"/path1"},
				":method":       {"GET"},
				"Content-Type":  {"application/json"},
				"Date":          {"2015/1/1"},
			},

			authResponse: context.AuthResponseError(),
		},

		{
			name: "should return 500 status code if nill Go context",
			settings: pb.CustomHMACProvider{
				RequestValidationUrl: "http://something.com",
				ServiceKey:   "service_key2",
				ServiceToken: "service_token2",
			},
			requestHeaders: http.Header{
				"User-Agent":    {""},
				"X-Request-Id":  {"request-id1"},
				"Authorization": {"partner_id1:signature1"},
				":path":         {"/path1"},
				":method":       {"GET"},
				"Content-Type":  {"application/json"},
				"Date":          {"2015/1/"},
			},

			authResponse: context.AuthResponseError(),
		},

		{
			name: "should return 500 status code if can't call custom auth provider",
			settings: pb.CustomHMACProvider{
				RequestValidationUrl: "https://example.com/xyz",
				ServiceKey:   "service_key1",
				ServiceToken: "service_token1",
			},
			goCtx:       gocontext.WithValue(gocontext.Background(), "key", "val"),
			requestBody: "request_body1",
			requestHeaders: http.Header{
				"User-Agent":    {""},
				"X-Request-Id":  {"request-id1"},
				"Authorization": {"partner_id1:signature1"},
				":path":         {"/path1"},
				":method":       {"GET"},
				"Content-Type":  {"application/json"},
				"Date":          {"2015/1/1"},
			},
			httpErr: errors.New("can't call custom auth provider"),

			headersToCustomAuth: http.Header{
				"Cache-Control":          {"no-cache"},
				"Content-Type":           {"application/json"},
				"Authorization":          {"Token decrypted_service_key1 decrypted_service_token1"},
				"X-Request-Id":           {"request-id1"},
				"X-Custom-Auth-Date":      {"2015/1/1"},
				"X-Custom-Auth-Path":      {"/path1"},
				"X-Custom-Auth-Signature": {"signature1"},
				"X-Custom-Auth-Verb":      {"GET"},
			},
			bodyToCustomAuth: "request_body1",

			authResponse: context.AuthResponseError(),
		},

		{
			name: "should return 500 status code if nil response from custom auth provider",
			settings: pb.CustomHMACProvider{
				RequestValidationUrl: "https://example.com/xyz",
				ServiceKey:   "service_key1",
				ServiceToken: "service_token1",
			},
			goCtx:       gocontext.WithValue(gocontext.Background(), "key", "val"),
			requestBody: "request_body1",
			requestHeaders: http.Header{
				"User-Agent":    {""},
				"X-Request-Id":  {"request-id1"},
				"Authorization": {"partner_id1:signature1"},
				":path":         {"/path1"},
				":method":       {"GET"},
				"Content-Type":  {"application/json"},
				"Date":          {"2015/1/1"},
			},

			headersToCustomAuth: http.Header{
				"Cache-Control":          {"no-cache"},
				"Content-Type":           {"application/json"},
				"Authorization":          {"Token decrypted_service_key1 decrypted_service_token1"},
				"X-Request-Id":           {"request-id1"},
				"X-Custom-Auth-Date":      {"2015/1/1"},
				"X-Custom-Auth-Path":      {"/path1"},
				"X-Custom-Auth-Signature": {"signature1"},
				"X-Custom-Auth-Verb":      {"GET"},
			},
			bodyToCustomAuth: "request_body1",

			authResponse: context.AuthResponseError(),
		},

		{
			name: "should return 500 if validator returns error",
			settings: pb.CustomHMACProvider{
				RequestValidationUrl:     "https://example.com/xyz",
				ServiceKey:               "service_key",
				ServiceToken:             "service_token",
				GeneratedUpstreamHeaders: []string{"x-custom-userid"},
			},
			goCtx: gocontext.WithValue(gocontext.Background(), "key", "val"),
			requestHeaders: http.Header{
				"User-Agent":    {""},
				"X-Request-Id":  {"request-id1"},
				"Authorization": {"partner_id1:signature1"},
				":path":         {"/path1"},
				":method":       {"GET"},
				"Content-Type":  {"application/json"},
				"Date":          {"2015/1/1"},
			},
			requestBody:    "request_body1",
			customAuthResp: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader(""))},
			validationErr:  errors.New("unknown error"),

			headersToCustomAuth: http.Header{
				"Cache-Control":          {"no-cache"},
				"Content-Type":           {"application/json"},
				"Authorization":          {"Token decrypted_service_key decrypted_service_token"},
				"X-Request-Id":           {"request-id1"},
				"X-Custom-Auth-Date":      {"2015/1/1"},
				"X-Custom-Auth-Path":      {"/path1"},
				"X-Custom-Auth-Signature": {"signature1"},
				"X-Custom-Auth-Verb":      {"GET"},
			},

			bodyToCustomAuth: "request_body1",
			authResponse:     context.AuthResponseError(),
		},

		{
			name: "should return 401 if validator return false",
			settings: pb.CustomHMACProvider{
				RequestValidationUrl:     "https://example.com/xyz",
				ServiceKey:               "service_key",
				ServiceToken:             "service_token",
				GeneratedUpstreamHeaders: []string{"x-custom-userid"},
			},
			goCtx: gocontext.WithValue(gocontext.Background(), "key", "val"),
			requestHeaders: http.Header{
				"User-Agent":    {""},
				"X-Request-Id":  {"request-id1"},
				"Authorization": {"partner_id1:signature1"},
				":path":         {"/path1"},
				":method":       {"GET"},
				"Content-Type":  {"application/json"},
				"Date":          {"2015/1/1"},
			},
			requestBody:    "request_body1",
			customAuthResp: &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(strings.NewReader(""))},
			validSignature: false,

			headersToCustomAuth: http.Header{
				"Cache-Control":          {"no-cache"},
				"Content-Type":           {"application/json"},
				"Authorization":          {"Token decrypted_service_key decrypted_service_token"},
				"X-Request-Id":           {"request-id1"},
				"X-Custom-Auth-Date":      {"2015/1/1"},
				"X-Custom-Auth-Path":      {"/path1"},
				"X-Custom-Auth-Signature": {"signature1"},
				"X-Custom-Auth-Verb":      {"GET"},
			},

			bodyToCustomAuth: "request_body1",
			authResponse:     context.AuthResponseUnauthorized(),
		},
	}

	for _, val := range tcs {
		tc := val
		t.Run(tc.name, func(t *testing.T) {
			ctx := &contextmocks.RequestContext{}
			ctx.On("GoContext").Return(tc.goCtx)
			ctx.On("GetSecret", tc.settings.ServiceKey).Return("decrypted_" + tc.settings.ServiceKey)
			ctx.On("GetSecret", tc.settings.ServiceToken).Return("decrypted_" + tc.settings.ServiceToken)

			callbacks := &mocks.Callbacks{}
			ctx.On("Callbacks").Return(callbacks)

			bodyReader := bytes.NewReader([]byte(tc.requestBody))
			ctx.On("BodyReader").Return(bodyReader)

			// set-up HttpClient
			httpClient := &httpmocks.HttpClientWithCtx{}

			var actualCustomAuthReq *http.Request
			httpClient.On("DoWithTracing", ctx, mock.AnythingOfType("*http.Request"), tc.spanName).Run(func(args mock.Arguments) {
				actualCustomAuthReq = args[1].(*http.Request)
			}).Return(tc.customAuthResp, tc.httpErr)

			// set-up HeaderMap
			headerMap := &envoymocks.RequestHeaderMap{}
			ctx.On("Headers").Return(headerMap)

			if len(tc.requestHeaders.Get("Authorization")) > 0 {
				headerMap.On("Authorization").Return(volatile.String(tc.requestHeaders.Get("Authorization")))
			}

			if len(tc.requestHeaders.Get(":path")) > 0 {
				headerMap.On("Path").Return(volatile.String(tc.requestHeaders.Get(":path")))
			}

			if len(tc.requestHeaders.Get(":method")) > 0 {
				headerMap.On("Method").Return(volatile.String(tc.requestHeaders.Get(":method")))
			}

			if len(tc.requestHeaders.Get("Content-Type")) > 0 {
				headerMap.On("ContentType").Return(volatile.String(tc.requestHeaders.Get("Content-Type")))
			}

			if len(tc.requestHeaders.Get("Date")) > 0 {
				headerMap.On("Date").Return(volatile.String(tc.requestHeaders.Get("Date")))
			}

			for k, vals := range tc.requestHeaders {
				for _, val := range vals {
					headerMap.On("Get", k).Return(volatile.String(val))
				}
			}
			// return emtpy if no x-request-id in headers
			headerMap.On("Get", "X-Request-Id").Return(volatile.String(""))

			// set-up Logger
			ctx.On("Logger").Return(logger.NewLogger("CustomHMACLogger", egomocks.NativeLogger{}))

			var validatorParam *http.Response
			isValid := func(resp *http.Response) (bool, error) {
				validatorParam = resp
				return tc.validSignature, tc.validationErr
			}

			provider, _ := createCustomHMACProvider(&tc.settings, httpClient, getCurrentTime, isValid)

			var authResp context.AuthResponse
			callbacks.On("OnComplete", mock.Anything).Run(func(args mock.Arguments) {
				authResp = args[0].(context.AuthResponse)
			})

			// Call verify function
			provider.Verify(ctx)

			// verify WithBody is always true
			assert.Equal(t, provider.WithBody(), true)

			ctx.AssertCalled(t, "GetSecret", tc.settings.ServiceKey)
			ctx.AssertCalled(t, "GetSecret", tc.settings.ServiceToken)
			if tc.headersToCustomAuth != nil || tc.bodyToCustomAuth != "" || tc.httpErr != nil {
				require.NotNil(t, actualCustomAuthReq)
				assert.Equal(t, http.MethodPost, actualCustomAuthReq.Method)

				// verify custom auth provider request headers
				assert.Equal(t, tc.headersToCustomAuth, actualCustomAuthReq.Header)

				// verify custom auth provider request body
				actualCustomAuthReqBytes, _ := ioutil.ReadAll(actualCustomAuthReq.Body)

				assert.Equal(t, tc.bodyToCustomAuth, string(actualCustomAuthReqBytes))
			}

			// Verify params to validator
			assert.Same(t, tc.customAuthResp, validatorParam)

			// verify response
			expectedResponse := tc.authResponse
			expectedResponse.HeadersToSet = tc.headersToSet
			expectedResponse.HeadersToRemove = tc.headersToRemove
			expectedResponse.FilterState = tc.filterState
			assert.Equal(t, expectedResponse, authResp)
		})
	}
}
