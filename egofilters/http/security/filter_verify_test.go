// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package security

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/grab/ego/ego/src/go/envoy/datastatus"
	"github.com/grab/ego/ego/src/go/envoy/headersstatus"
	"github.com/grab/ego/ego/src/go/envoy/trailersstatus"
	"github.com/grab/ego/ego/src/go/volatile"

	context "github.com/grab/ego/egofilters/http/security/context"
	pb "github.com/grab/ego/egofilters/http/security/proto"
	"github.com/grab/ego/egofilters/http/security/verifier"

	verifiermocks "github.com/grab/ego/egofilters/mock/gen/http/security/verifier"
	envoymocks "github.com/grab/ego/ego/test/go/mock/gen/envoy"
)

func TestDecodeHeaders(t *testing.T) {

	tcs := []struct {
		// Set-up
		name                      string
		requestBodyRequired       bool
		endstream                 bool
		routeSpecificFilterConfig interface{}

		// Verify
		verifierCalled               bool
		expectedDecodeHeadersResult  headersstatus.Type
		expectedDecodeTrailersResult trailersstatus.Type
	}{
		{
			name:                "should call verifier and wait for response if body isn't required",
			requestBodyRequired: false,
			endstream:           true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_verifier"},
			},

			verifierCalled:               true,
			expectedDecodeHeadersResult:  headersstatus.StopAllIterationAndWatermark,
			expectedDecodeTrailersResult: trailersstatus.StopIteration,
		},

		{
			name:                "should wait for request body if required",
			requestBodyRequired: true,
			endstream:           false,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_verifier"},
			},

			verifierCalled:               false,
			expectedDecodeHeadersResult:  headersstatus.StopIteration,
			expectedDecodeTrailersResult: trailersstatus.Continue,
		},

		{
			name:                "shouldn't wait for request body if endstream is true",
			requestBodyRequired: true,
			endstream:           true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_verifier"},
			},

			verifierCalled:               true,
			expectedDecodeHeadersResult:  headersstatus.StopAllIterationAndWatermark,
			expectedDecodeTrailersResult: trailersstatus.StopIteration,
		},

		{
			name:                      "should continue if nil routeSpecificFilterConfig",
			requestBodyRequired:       true,
			endstream:                 true,
			routeSpecificFilterConfig: nil,

			verifierCalled:               false,
			expectedDecodeHeadersResult:  headersstatus.Continue,
			expectedDecodeTrailersResult: trailersstatus.Continue,
		},

		{
			name:                      "should continue if invalid type of routeSpecificFilterConfig",
			requestBodyRequired:       true,
			endstream:                 true,
			routeSpecificFilterConfig: "string instead of requirement",

			verifierCalled:               false,
			expectedDecodeHeadersResult:  headersstatus.Continue,
			expectedDecodeTrailersResult: trailersstatus.Continue,
		},

		{
			name:                "should continue if unknown provider",
			requestBodyRequired: true,
			endstream:           true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_unknown_provider"},
			},

			verifierCalled:               false,
			expectedDecodeHeadersResult:  headersstatus.Continue,
			expectedDecodeTrailersResult: trailersstatus.Continue,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			native := &envoymocks.GoHttpFilter{}

			wg := sync.WaitGroup{}
			native.On("Pin").Run(func(args mock.Arguments) {
				wg.Add(1)
			})

			secretProvider := &envoymocks.GenericSecretConfigProvider{}
			native.On("GenericSecretProvider").Return((secretProvider))
			secretProvider.On("Secret").Return(volatile.String("zzz"))

			native.On("Log", mock.Anything, mock.Anything)

			decoderCallbacks := &envoymocks.DecoderFilterCallbacks{}
			native.On("DecoderCallbacks").Return(decoderCallbacks)

			decoderCallbacks.On("ActiveSpan").Return(&envoymocks.Span{})

			route := &envoymocks.Route{}
			decoderCallbacks.On("Route").Return(route)

			native.On("ResolveMostSpecificPerGoFilterConfig", FilterID, route).Return(tc.routeSpecificFilterConfig)

			provider := &verifiermocks.Verifier{}
			config := &securityConfig{
				verifiers: map[string]verifier.Verifier{
					"my_verifier": provider,
				},
			}
			provider.On("WithBody").Return(tc.requestBodyRequired)

			filter := newSecurity(native, config)
			var ctx context.RequestContext
			provider.On("Verify", mock.Anything).Run(func(args mock.Arguments) {
				ctx = args[0].(context.RequestContext)
				defer wg.Done()
			})
			assert.NotNil(t, filter)

			headerMap := &envoymocks.RequestHeaderMap{}
			decodeHeadersResult := filter.DecodeHeaders(headerMap, tc.endstream)
			decodeTrailersResult := filter.DecodeTrailers(&envoymocks.RequestTrailerMap{})

			wg.Wait()

			// Continue if not route specific route config
			assert.Equal(t, tc.expectedDecodeHeadersResult, decodeHeadersResult)
			assert.Equal(t, tc.expectedDecodeTrailersResult, decodeTrailersResult)
			if tc.verifierCalled {
				provider.AssertCalled(t, "Verify", mock.Anything)
				// verify context passed to verifier
				assert.Equal(t, filter, ctx.Callbacks())
				assert.NotNil(t, ctx.GoContext())
				assert.Equal(t, decoderCallbacks.ActiveSpan(), ctx.ActiveSpan())
				assert.Equal(t, headerMap, ctx.Headers())
				assert.Nil(t, ctx.BodyReader())
				assert.NotNil(t, ctx.Logger())
			}
		})
	}
}

func TestDecodeData(t *testing.T) {
	tcs := []struct {
		name string
		// set-up
		endstream                 bool
		routeSpecificFilterConfig interface{}

		// verify
		expectedDataStatus datastatus.Type
		verifierCalled     bool
	}{
		{
			name:      "should call verifier if endstream is true",
			endstream: true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_verifier"},
			},

			expectedDataStatus: datastatus.StopIterationAndWatermark,
			verifierCalled:     true,
		},
		{
			name:      "should wait for more data if endstream is false",
			endstream: false,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_verifier"},
			},

			expectedDataStatus: datastatus.StopIterationAndBuffer,
			verifierCalled:     false,
		},
		{
			name:      "should continue if not waiting for request body",
			endstream: false,
			routeSpecificFilterConfig: pb.Requirement{
				// Invalid provider name will cause filter ignoring current request
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "unknown"},
			},

			expectedDataStatus: datastatus.Continue,
			verifierCalled:     false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			native := &envoymocks.GoHttpFilter{}

			wg := sync.WaitGroup{}
			native.On("Pin").Run(func(args mock.Arguments) {
				wg.Add(1)
			})

			secretProvider := &envoymocks.GenericSecretConfigProvider{}
			native.On("GenericSecretProvider").Return((secretProvider))
			secretProvider.On("Secret").Return(volatile.String("zzz"))

			native.On("Log", mock.Anything, mock.Anything)

			decoderCallbacks := &envoymocks.DecoderFilterCallbacks{}
			native.On("DecoderCallbacks").Return(decoderCallbacks)

			decoderCallbacks.On("ActiveSpan").Return(&envoymocks.Span{})

			route := &envoymocks.Route{}
			decoderCallbacks.On("Route").Return(route)

			native.On("ResolveMostSpecificPerGoFilterConfig", FilterID, route).Return(tc.routeSpecificFilterConfig)

			provider := &verifiermocks.Verifier{}
			config := &securityConfig{
				verifiers: map[string]verifier.Verifier{
					"my_verifier": provider,
				},
			}
			provider.On("WithBody").Return(true)

			filter := newSecurity(native, config)
			var ctx context.RequestContext
			provider.On("Verify", mock.Anything).Run(func(args mock.Arguments) {
				ctx = args[0].(context.RequestContext)
				defer wg.Done()
			})
			assert.NotNil(t, filter)

			headerMap := &envoymocks.RequestHeaderMap{}
			filter.DecodeHeaders(headerMap, false)

			bufferInstance := &envoymocks.BufferInstance{}
			decoderCallbacks.On("AddDecodedData", bufferInstance, true)

			fullBuffer := &envoymocks.BufferInstance{}

			bodyReader := strings.NewReader("zzz")
			fullBuffer.On("NewReader", uint64(0)).Return(bodyReader)
			decoderCallbacks.On("DecodingBuffer").Return(fullBuffer)

			decodeDataResult := filter.DecodeData(bufferInstance, tc.endstream)

			wg.Wait()

			// Continue if not route specific route config
			assert.Equal(t, tc.expectedDataStatus, decodeDataResult)
			if tc.verifierCalled {
				decoderCallbacks.AssertExpectations(t)

				provider.AssertCalled(t, "Verify", mock.Anything)
				// verify context passed to verifier
				assert.Equal(t, filter, ctx.Callbacks())
				assert.NotNil(t, ctx.GoContext())
				assert.Equal(t, decoderCallbacks.ActiveSpan(), ctx.ActiveSpan())
				assert.Equal(t, headerMap, ctx.Headers())
				assert.Equal(t, bodyReader, ctx.BodyReader())
				assert.NotNil(t, ctx.Logger())
			}
		})
	}
}

func TestOnComplete(t *testing.T) {
	type localReplyData struct {
		StatusCode int
		Body       string
		Header     map[string]string
	}

	tcs := []struct {
		name string
		// set-up
		authResp context.AuthResponse

		// verify
		localReply            *localReplyData
		increaseOkCounter     bool
		increaseErrCounter    bool
		increaseDeniedCounter bool
		filterState           map[string]string
	}{
		{
			name: "verify successfully",

			authResp: context.AuthResponse{
				Status:          context.AuthOK,
				HeadersToRemove: map[string]struct{}{"header-to-remove": {}},
				HeadersToSet:    map[string]string{"header-to-set": "val1"},
				HeadersToAppend: map[string]string{"header-to-append": "val2"},
				FilterState:     map[string]string{"state1": "val1"},
			},

			// HeadersToRemove, HeadersToSet and HeadersToAppend will be remove, set and append to headers to upstream respectly
			increaseOkCounter: true,

			// hardcode keys to prevent regressions in target environment
			filterState: map[string]string{"egodemo.security.ctx.session.state1": "val1"},
		},

		{
			name: "verify with error response",

			authResp: context.AuthResponse{
				StatusCode:      401,
				Status:          context.AuthError,
				Body:            "this is an error",
				HeadersToRemove: map[string]struct{}{"header-to-remove": {}},
				HeadersToSet:    map[string]string{"header-to-set": "val1"},
				HeadersToAppend: map[string]string{"header-to-append": "val2"},
			},

			localReply: &localReplyData{
				StatusCode: 401,
				Header:     map[string]string{"header-to-set": "val1"},
				Body:       "this is an error",
			},
			increaseErrCounter: true,
		},

		{
			name: "verify with denied response",

			authResp: context.AuthResponse{
				StatusCode:      500,
				Body:            "this is an denied error",
				Status:          context.AuthDenied,
				HeadersToRemove: map[string]struct{}{"header-to-remove": {}},
				HeadersToSet:    map[string]string{"header-to-set": "val1"},
				HeadersToAppend: map[string]string{"header-to-append": "val2"},
			},

			localReply: &localReplyData{
				StatusCode: 500,
				Header:     map[string]string{"header-to-set": "val1"},
				Body:       "this is an denied error",
			},
			increaseDeniedCounter: true,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			native := &envoymocks.GoHttpFilter{}

			secretProvider := &envoymocks.GenericSecretConfigProvider{}
			native.On("GenericSecretProvider").Return((secretProvider))
			secretProvider.On("Secret").Return(volatile.String("zzz"))

			native.On("Log", mock.Anything, mock.Anything)

			decoderCallbacks := &envoymocks.DecoderFilterCallbacks{}
			native.On("DecoderCallbacks").Return(decoderCallbacks)

			decoderCallbacks.On("ContinueDecoding")
			native.On("Unpin")

			provider := &verifiermocks.Verifier{}

			authOkStats := &envoymocks.Counter{}
			authDeniedStats := &envoymocks.Counter{}
			authErrorStats := &envoymocks.Counter{}

			config := &securityConfig{
				verifiers: map[string]verifier.Verifier{
					"my_verifier": provider,
				},
				stats: securityStats{
					authOK:     authOkStats,
					authDenied: authDeniedStats,
					authError:  authErrorStats,
				},
			}

			// set-up headermap
			decoderCallbacks.On("ActiveSpan").Return(&envoymocks.Span{})

			route := &envoymocks.Route{}
			decoderCallbacks.On("Route").Return(route)

			native.On("ResolveMostSpecificPerGoFilterConfig", FilterID, route).Return(pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_verifier"},
			})

			wg := sync.WaitGroup{}
			native.On("Pin").Run(func(args mock.Arguments) {
				wg.Add(1)
			})
			provider.On("Verify", mock.Anything).Run(func(args mock.Arguments) {
				defer wg.Done()
			})

			headerMap := &envoymocks.RequestHeaderMap{}
			// end set-up headermap

			if tc.increaseOkCounter {
				authOkStats.On("Inc")
			}

			if tc.increaseErrCounter {
				authErrorStats.On("Inc")
			}

			if tc.increaseDeniedCounter {
				authDeniedStats.On("Inc")
			}

			provider.On("WithBody").Return(true)

			filter := newSecurity(native, config)
			assert.NotNil(t, filter)

			native.On("Post", authPost).Run(func(args mock.Arguments) {
				filter.OnPost(authPost)
			})

			_ = filter.DecodeHeaders(headerMap, false)

			// set-up header map
			filterState := &envoymocks.FilterState{}
			if tc.localReply != nil {
				decoderCallbacks.On("SendLocalReply", tc.localReply.StatusCode, tc.localReply.Body, tc.localReply.Header, mock.Anything)
			} else {
				headerMap.On("Remove", "header-to-remove")
				headerMap.On("SetCopy", "header-to-set", "val1")
				headerMap.On("AppendCopy", "header-to-append", "val2")
				streamInfo := &envoymocks.StreamInfo{}
				decoderCallbacks.On("StreamInfo").Return(streamInfo)

				streamInfo.On("FilterState").Return(filterState)

				for k, v := range tc.filterState {
					filterState.On("SetData", k, v, mock.Anything, mock.Anything)
				}

			}

			callback := filter.(context.Callbacks)
			assert.NotNil(t, callback)

			callback.OnComplete(tc.authResp)
			native.AssertCalled(t, "Post", authPost)
			native.AssertCalled(t, "Unpin")

			// verify metrics
			authErrorStats.AssertExpectations(t)
			authOkStats.AssertExpectations(t)
			authDeniedStats.AssertExpectations(t)

			headerMap.AssertExpectations(t)
			filterState.AssertExpectations(t)

			if tc.localReply != nil {
				decoderCallbacks.AssertCalled(t, "SendLocalReply", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}

}
