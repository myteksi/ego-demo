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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/grab/ego/ego/src/go/logger"
	"github.com/grab/ego/ego/src/go/volatile"
	egomocks "github.com/grab/ego/ego/test/go/mock"

	"github.com/grab/ego/egofilters/http/security/context"
	pb "github.com/grab/ego/egofilters/http/security/proto"

	contextmocks "github.com/grab/ego/egofilters/mock/gen/http/security/context"
	httpmocks "github.com/grab/ego/egofilters/mock/gen/http/security/http"
	envoymocks "github.com/grab/ego/ego/test/go/mock/gen/envoy"
)

func TestHMACSignResponse(t *testing.T) {

	tcs := []struct {
		name                         string
		config                       *pb.CustomHMACProvider
		requestID                    string
		responseStatusCode           string
		responseBody                 string
		authResp                     context.AuthResponse
		customAuthResponseStatusCode int
		customAuthResponseBody       string
		httpErr                      error
		currentTime                  time.Time
		goCtx                        gocontext.Context

		headersToCustomAuth http.Header
		bodyToCustomAuth    string
		spanName            string
		signResult          context.SignResponse
	}{
		{
			name: "should call custom auth provider to sign response",
			config: &pb.CustomHMACProvider{
				ResponseSigningUrl: "http://custom-auth.example.com",
				ServiceToken:    "service_token",
				ServiceKey:      "service_key",
				TracingEnabled:  true,
			},
			requestID:          "request1",
			responseStatusCode: "200",
			responseBody:       "this is body from upstream",

			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "partner_id1"},
			},
			customAuthResponseBody:       `{"signature": "hmac_signature"}`,
			customAuthResponseStatusCode: http.StatusOK,
			currentTime:                  time.Date(2000, time.January, 1, 2, 3, 4, 5, time.UTC),
			goCtx:                        gocontext.WithValue(gocontext.Background(), "key", "val"),

			headersToCustomAuth: http.Header{
				"Authorization":            []string{"Token decrypted_service_key decrypted_service_token"},
				"X-Custom-Auth-Date":        []string{"Sat, 01 Jan 2000 02:03:04 GMT"},
				"X-Custom-Auth-Status-Code": []string{"200"},
				"X-Request-Id":             []string{"request1"},
			},
			bodyToCustomAuth: "this is body from upstream",
			spanName:         "custom_hmac_sign",
			signResult: context.SignResponse{
				HeadersToSet: map[string]string{"Date": "Sat, 01 Jan 2000 02:03:04 GMT", "X-Custom-Auth-Signature-HMAC-SHA256": "hmac_signature"},
			},
		},
		{
			name: "should call custom auth provider to sign response with different configuration",
			config: &pb.CustomHMACProvider{
				ResponseSigningUrl: "http://custom-auth-provider.example.com",
				ServiceToken:    "service_token2",
				ServiceKey:      "service_key2",
				TracingEnabled:  false,
			},
			requestID:          "request2",
			responseStatusCode: "400",
			responseBody:       "this is body from upstream 2",
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "partner_id2"},
			},
			customAuthResponseBody:       `{"signature": "hmac_signature2"}`,
			customAuthResponseStatusCode: http.StatusOK,
			currentTime:                  time.Date(2002, time.January, 2, 2, 3, 4, 5, time.UTC),
			goCtx:                        gocontext.WithValue(gocontext.Background(), "key2", "val2"),

			headersToCustomAuth: http.Header{
				"Authorization":             []string{"Token decrypted_service_key2 decrypted_service_token2"},
				"X-Custom-Auth-Date":        []string{"Wed, 02 Jan 2002 02:03:04 GMT"},
				"X-Custom-Auth-Status-Code": []string{"400"},
				"X-Request-Id":              []string{"request2"},
			},
			bodyToCustomAuth: "this is body from upstream 2",
			spanName:         "",
			signResult: context.SignResponse{
				HeadersToSet: map[string]string{"Date": "Wed, 02 Jan 2002 02:03:04 GMT", "X-Custom-Auth-Signature-HMAC-SHA256": "hmac_signature2"},
			},
		},
		{
			name: "should return 500 if can't parse ResponseSigningUrl",
			config: &pb.CustomHMACProvider{
				ResponseSigningUrl: ":zzz",
				ServiceToken:    "service_token",
				ServiceKey:      "service_key",
			},
			responseStatusCode: "200",
			responseBody:       "this is body from upstream",
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "partner_id1"},
			},
			customAuthResponseBody:       `{"signature": "hmac_signature"}`,
			customAuthResponseStatusCode: http.StatusOK,
			currentTime:                  time.Date(2000, time.January, 1, 2, 3, 4, 5, time.UTC),
			goCtx:                        gocontext.WithValue(gocontext.Background(), "key", "val"),

			signResult: context.SignResponse{
				StatusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "should return 500 if can't call custom auth provider",
			config: &pb.CustomHMACProvider{
				ResponseSigningUrl: "http://invalid.example.com",
				ServiceToken:    "service_token",
				ServiceKey:      "service_key",
			},
			responseStatusCode: "200",
			responseBody:       "this is body from upstream",
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "partner_id1"},
			},
			httpErr:     errors.New("can't call custom auth provider"),
			currentTime: time.Date(2000, time.January, 1, 2, 3, 4, 5, time.UTC),
			goCtx:       gocontext.WithValue(gocontext.Background(), "key", "val"),

			signResult: context.SignResponse{
				StatusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "should return 500 if nil Go context",
			config: &pb.CustomHMACProvider{
				ResponseSigningUrl: "http://something.com",
				ServiceToken:    "service_token",
				ServiceKey:      "service_key",
			},
			responseStatusCode: "200",
			responseBody:       "this is body from upstream",
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "partner_id1"},
			},
			currentTime: time.Date(2000, time.January, 2, 2, 3, 4, 5, time.UTC),
			goCtx:       nil,

			signResult: context.SignResponse{
				StatusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "should return 500 if nil response from custom auth provider",
			config: &pb.CustomHMACProvider{
				ResponseSigningUrl: "http://something.com",
				ServiceToken:    "service_token",
				ServiceKey:      "service_key",
			},
			responseStatusCode: "200",
			responseBody:       "this is body from upstream",
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "partner_id1"},
			},
			currentTime: time.Date(2000, time.January, 1, 2, 3, 4, 5, time.UTC),
			goCtx:       gocontext.WithValue(gocontext.Background(), "key", "val"),

			signResult: context.SignResponse{
				StatusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "should return 500 if can't parse response from custom auth provider",
			config: &pb.CustomHMACProvider{
				ResponseSigningUrl: "http://custom-auth.example.com",
				ServiceToken:    "service_token",
				ServiceKey:      "service_key",
			},
			responseStatusCode: "200",
			responseBody:       "this is body from upstream",
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "partner_id1"},
			},
			customAuthResponseBody:       ``,
			customAuthResponseStatusCode: http.StatusOK,
			currentTime:                  time.Date(2000, time.January, 1, 2, 3, 4, 5, time.UTC),
			goCtx:                        gocontext.WithValue(gocontext.Background(), "key", "val"),

			headersToCustomAuth: http.Header{
				"Authorization":            []string{"Token decrypted_service_key decrypted_service_token"},
				"X-Custom-Auth-Date":        []string{"Sat, 01 Jan 2000 02:03:04 GMT"},
				"X-Custom-Auth-Status-Code": []string{"200"},
				"X-Request-Id":             []string{""},
			},
			bodyToCustomAuth: "this is body from upstream",
			signResult: context.SignResponse{
				StatusCode: http.StatusInternalServerError,
			},
		},
		{
			name: "should not crash when custom auth provider return non-200 status code",
			config: &pb.CustomHMACProvider{
				ResponseSigningUrl: "http://custom-auth.example.com",
				ServiceToken:    "service_token",
				ServiceKey:      "service_key",
			},
			responseStatusCode: "200",
			responseBody:       "this is body from upstream",
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "partner_id1"},
			},
			customAuthResponseBody:       ``,
			customAuthResponseStatusCode: http.StatusBadRequest,
			currentTime:                  time.Date(2000, time.January, 1, 2, 3, 4, 5, time.UTC),
			goCtx:                        gocontext.WithValue(gocontext.Background(), "key", "val"),

			headersToCustomAuth: http.Header{
				"Authorization":            []string{"Token decrypted_service_key decrypted_service_token"},
				"X-Custom-Auth-Date":        []string{"Sat, 01 Jan 2000 02:03:04 GMT"},
				"X-Custom-Auth-Status-Code": []string{"200"},
				"X-Request-Id":             []string{""},
			},
			bodyToCustomAuth: "this is body from upstream",
			signResult: context.SignResponse{
				StatusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, val := range tcs {
		tc := val
		t.Run(tc.name, func(t *testing.T) {
			responseContext := &contextmocks.ResponseContext{}
			responseContext.On("GetSecret", tc.config.ServiceKey).Return("decrypted_" + tc.config.ServiceKey)
			responseContext.On("GetSecret", tc.config.ServiceToken).Return("decrypted_" + tc.config.ServiceToken)

			requestHeader := &envoymocks.RequestHeaderMap{}
			requestHeader.On("Get", requestIDHeader).Return(volatile.String(tc.requestID))
			responseContext.On("RequestHeaders").Return(requestHeader)

			responseHeader := &envoymocks.ResponseHeaderMap{}
			responseHeader.On("Status").Return(volatile.String(tc.responseStatusCode))
			responseContext.On("Headers").Return(responseHeader)

			responseContext.On("GoContext").Return(tc.goCtx)

			bodyReader := bytes.NewReader([]byte(tc.responseBody))
			responseContext.On("BodyReader").Return(bodyReader)

			responseContext.On("AuthResponse").Return(tc.authResp)

			callbacks := &contextmocks.ResponseCallbacks{}
			responseContext.On("Callbacks").Return(callbacks)

			// set-up Logger
			responseContext.On("Logger").Return(logger.NewLogger("CustomHMACLogger", egomocks.NativeLogger{}))

			// Set-up custom auth provider response
			var httpResponse *http.Response
			if tc.customAuthResponseStatusCode > 0 || tc.customAuthResponseBody != "" {
				httpResponse = &http.Response{StatusCode: tc.customAuthResponseStatusCode, Body: ioutil.NopCloser(strings.NewReader(tc.customAuthResponseBody))}
			}

			// set-up HttpClient
			httpClient := &httpmocks.HttpClientWithCtx{}

			var actualCustomAuthReq *http.Request
			httpClient.On("DoWithTracing", responseContext, mock.AnythingOfType("*http.Request"), tc.spanName).Run(func(args mock.Arguments) {
				actualCustomAuthReq = args[1].(*http.Request)
			}).Return(httpResponse, tc.httpErr)

			mockGetCurrentTime := func() time.Time {
				return tc.currentTime
			}
			signer, err := createCustomHMACProvider(tc.config, httpClient, mockGetCurrentTime, nil)
			require.Nil(t, err)

			var signResp context.SignResponse
			callbacks.On("OnCompleteSigning", mock.Anything).Run(func(args mock.Arguments) {
				signResp = args[0].(context.SignResponse)
			})

			signer.Sign(responseContext)

			if tc.headersToCustomAuth != nil || tc.bodyToCustomAuth != "" {
				require.NotNil(t, actualCustomAuthReq)
				assert.Equal(t, http.MethodPost, actualCustomAuthReq.Method)

				// verify URL
				assert.Equal(t, tc.config.ResponseSigningUrl, actualCustomAuthReq.URL.String())

				// verify custom auth provider request headers
				assert.Equal(t, tc.headersToCustomAuth, actualCustomAuthReq.Header)

				// verify custom auth provider request body
				actualCustomAuthReqBytes, _ := ioutil.ReadAll(actualCustomAuthReq.Body)
				assert.Equal(t, tc.bodyToCustomAuth, string(actualCustomAuthReqBytes))
			}

			assert.Equal(t, tc.signResult, signResp)

		})
	}
}
