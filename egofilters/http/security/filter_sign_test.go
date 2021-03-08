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
	"github.com/grab/ego/ego/src/go/volatile"

	"github.com/grab/ego/egofilters/http/security/context"
	pb "github.com/grab/ego/egofilters/http/security/proto"
	"github.com/grab/ego/egofilters/http/security/verifier"

	verifiermocks "github.com/grab/ego/egofilters/mock/gen/http/security/verifier"
	envoymocks "github.com/grab/ego/ego/test/go/mock/gen/envoy"
)

func TestEncodeHeaders(t *testing.T) {

	tcs := []struct {
		name string
		// set-up
		signRequired              bool
		endstream                 bool
		routeSpecificFilterConfig interface{}

		// verify
		signCalled                  bool
		expectedEncodeHeadersResult headersstatus.Type
	}{
		{
			name:         "should wait for body if endstream is false",
			signRequired: true,
			endstream:    false,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_provider"},
			},

			signCalled:                  false,
			expectedEncodeHeadersResult: headersstatus.StopIteration,
		},
		{
			name:         "should sign response if endstream is true",
			signRequired: true,
			endstream:    true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_provider"},
			},

			signCalled:                  true,
			expectedEncodeHeadersResult: headersstatus.StopAllIterationAndWatermark,
		},
		{
			name:         "should continue if sign is not required",
			signRequired: false,
			endstream:    true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_provider"},
			},

			signCalled:                  false,
			expectedEncodeHeadersResult: headersstatus.Continue,
		},
		{
			name:         "should continue if can't find signer",
			signRequired: true,
			endstream:    true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_unknown_provider"},
			},

			signCalled:                  false,
			expectedEncodeHeadersResult: headersstatus.Continue,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			native := &envoymocks.GoHttpFilter{}

			wg := sync.WaitGroup{}
			native.On("Pin").Run(func(args mock.Arguments) {
				wg.Add(1)
			})
			native.On("Unpin").Run(func(args mock.Arguments) {
				wg.Done()
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

			encoderCallbacks := &envoymocks.EncoderFilterCallbacks{}
			native.On("EncoderCallbacks").Return(encoderCallbacks)

			encoderSpan := &envoymocks.Span{}
			encoderCallbacks.On("ActiveSpan").Return(encoderSpan)

			native.On("ResolveMostSpecificPerGoFilterConfig", FilterID, route).Return(tc.routeSpecificFilterConfig)

			signer := &verifiermocks.Signer{}

			config := &securityConfig{
				signers: map[string]verifier.Signer{
					"my_provider": signer,
				},
			}

			filter := newSecurity(native, config)
			assert.NotNil(t, filter)

			requestHeaders := &envoymocks.RequestHeaderMap{}
			filter.DecodeHeaders(requestHeaders, false)

			responseHeaderMap := &envoymocks.ResponseHeaderMap{}

			signer.On("SigningRequired", responseHeaderMap, mock.AnythingOfType("AuthResponse")).Return(tc.signRequired)
			var ctx context.ResponseContext
			signer.On("Sign", mock.Anything).Run(func(args mock.Arguments) {
				ctx = args[0].(context.ResponseContext)
			})

			result := filter.EncodeHeaders(responseHeaderMap, tc.endstream)

			assert.Equal(t, tc.expectedEncodeHeadersResult, result)
			wg.Wait()

			if tc.signCalled {
				signer.AssertCalled(t, "Sign", mock.Anything)

				// verify passing correct data to signer
				assert.NotNil(t, ctx.Callbacks())
				assert.Equal(t, responseHeaderMap, ctx.Headers())
				assert.Equal(t, filter, ctx.Callbacks())
				assert.NotNil(t, ctx.Logger())
				assert.NotNil(t, ctx.GoContext())
				assert.Equal(t, encoderCallbacks.ActiveSpan(), ctx.ActiveSpan())
				assert.Nil(t, ctx.BodyReader())
			}
		})
	}
}

func TestEncodeData(t *testing.T) {

	tcs := []struct {
		name                      string
		endstream                 bool
		signRequired              bool
		routeSpecificFilterConfig interface{}

		signCalled               bool
		expectedEncodeDataResult datastatus.Type
	}{
		{
			name:         "should wait for full body if endstream is false",
			endstream:    false,
			signRequired: true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_provider"},
			},

			signCalled:               false,
			expectedEncodeDataResult: datastatus.StopIterationAndBuffer,
		},
		{
			name:         "should continue if sign is not required",
			endstream:    false,
			signRequired: false,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_provider"},
			},

			signCalled:               false,
			expectedEncodeDataResult: datastatus.Continue,
		},
		{
			name:         "should sign response if endstream is true",
			endstream:    true,
			signRequired: true,
			routeSpecificFilterConfig: pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_provider"},
			},

			signCalled:               true,
			expectedEncodeDataResult: datastatus.StopIterationAndWatermark,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			native := &envoymocks.GoHttpFilter{}

			wg := sync.WaitGroup{}
			native.On("Pin").Run(func(args mock.Arguments) {
				wg.Add(1)
			})
			native.On("Unpin").Run(func(args mock.Arguments) {
				wg.Done()
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

			encoderCallbacks := &envoymocks.EncoderFilterCallbacks{}
			native.On("EncoderCallbacks").Return(encoderCallbacks)

			encoderSpan := &envoymocks.Span{}
			encoderCallbacks.On("ActiveSpan").Return(encoderSpan)

			native.On("ResolveMostSpecificPerGoFilterConfig", FilterID, route).Return(tc.routeSpecificFilterConfig)

			signer := &verifiermocks.Signer{}
			signer.On("SigningRequired", mock.Anything, mock.Anything).Return(tc.signRequired)
			var ctx context.ResponseContext
			signer.On("Sign", mock.Anything).Run(func(args mock.Arguments) {
				ctx = args[0].(context.ResponseContext)
			})
			config := &securityConfig{
				signers: map[string]verifier.Signer{
					"my_provider": signer,
				},
			}

			filter := newSecurity(native, config)
			assert.NotNil(t, filter)

			_ = filter.DecodeHeaders(&envoymocks.RequestHeaderMap{}, false)

			responseHeaderMap := &envoymocks.ResponseHeaderMap{}
			filter.EncodeHeaders(responseHeaderMap, false)

			data := &envoymocks.BufferInstance{}
			encoderCallbacks.On("AddEncodedData", data, true)

			fullBuffer := &envoymocks.BufferInstance{}
			bodyReader := strings.NewReader("zzz")
			fullBuffer.On("NewReader", uint64(0)).Return(bodyReader)
			encoderCallbacks.On("EncodingBuffer").Return(fullBuffer)

			encodeDataResult := filter.EncodeData(data, tc.endstream)

			assert.Equal(t, tc.expectedEncodeDataResult, encodeDataResult)
			wg.Wait()

			if tc.signCalled {
				signer.AssertCalled(t, "Sign", mock.Anything)
				encoderCallbacks.AssertExpectations(t)

				// verify passing correct data to signer
				assert.NotNil(t, ctx.Callbacks())
				assert.Equal(t, responseHeaderMap, ctx.Headers())
				assert.Equal(t, filter, ctx.Callbacks())
				assert.NotNil(t, ctx.Logger())
				assert.NotNil(t, ctx.GoContext())
				assert.Equal(t, encoderCallbacks.ActiveSpan(), ctx.ActiveSpan())
				assert.Equal(t, bodyReader, ctx.BodyReader())
			}
		})
	}
}

func TestOnCompleteSigning(t *testing.T) {

	tcs := []struct {
		name         string
		signResponse context.SignResponse

		statusCode     int
		responeHeaders map[string]string
	}{
		{
			name: "should continue encoding",
		},
		{
			name: "should set status code",
			signResponse: context.SignResponse{
				StatusCode: 200,
			},

			statusCode: 200,
		},
		{
			name: "should set header",
			signResponse: context.SignResponse{
				HeadersToSet: map[string]string{
					"x-header": "value",
				},
			},

			responeHeaders: map[string]string{
				"x-header": "value",
			},
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

			decoderCallbacks.On("ActiveSpan").Return(&envoymocks.Span{})

			route := &envoymocks.Route{}
			decoderCallbacks.On("Route").Return(route)

			encoderCallbacks := &envoymocks.EncoderFilterCallbacks{}
			native.On("EncoderCallbacks").Return(encoderCallbacks)

			native.On("ResolveMostSpecificPerGoFilterConfig", FilterID, route).Return(pb.Requirement{
				RequiresType: &pb.Requirement_ProviderName{ProviderName: "my_provider"},
			})

			signer := &verifiermocks.Signer{}
			signer.On("SigningRequired", mock.Anything, mock.Anything).Return(true)
			signer.On("Sign", mock.Anything)
			config := &securityConfig{
				signers: map[string]verifier.Signer{
					"my_provider": signer,
				},
			}

			filter := newSecurity(native, config)
			assert.NotNil(t, filter)

			filter.DecodeHeaders(&envoymocks.RequestHeaderMap{}, false)

			responseHeaderMap := &envoymocks.ResponseHeaderMap{}
			filter.EncodeHeaders(responseHeaderMap, false)

			callback := filter.(context.ResponseCallbacks)

			native.On("Post", signPost).Run(func(args mock.Arguments) {
				filter.OnPost(signPost)
			})

			encoderCallbacks.On("ContinueEncoding")
			if tc.statusCode > 0 {
				responseHeaderMap.On("SetStatus", tc.statusCode)
			}
			for k, v := range tc.responeHeaders {
				responseHeaderMap.On("SetCopy", k, v)
			}

			callback.OnCompleteSigning(tc.signResponse)

			encoderCallbacks.AssertExpectations(t)
			responseHeaderMap.AssertExpectations(t)
		})
	}
}
