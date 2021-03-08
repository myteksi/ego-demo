// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package verifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grab/ego/ego/src/go/volatile"

	"github.com/grab/ego/egofilters/http/security/context"
	pb "github.com/grab/ego/egofilters/http/security/proto"

	httpmocks "github.com/grab/ego/egofilters/mock/gen/http/security/http"
	envoymocks "github.com/grab/ego/ego/test/go/mock/gen/envoy"
)

func TestSignRequired(t *testing.T) {

	tcs := []struct {
		name       string
		statusCode string
		config     *pb.CustomHMACProvider
		authResp   context.AuthResponse

		signRequired bool
	}{
		{
			name:       "should return true if everything is good",
			statusCode: "200",
			config: &pb.CustomHMACProvider{
				SignResp: true,
			},
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "123"},
			},

			signRequired: true,
		},
		{
			name:       "should return false if SignResp configuration is false",
			statusCode: "200",
			config: &pb.CustomHMACProvider{
				SignResp: false,
			},
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "123"},
			},

			signRequired: false,
		},
		{
			name:       "should return false if nil FilterState",
			statusCode: "200",
			config: &pb.CustomHMACProvider{
				SignResp: true,
			},
			authResp: context.AuthResponse{},

			signRequired: false,
		},
		{
			name:       "should return false if missing hmacUserIDSessionKey",
			statusCode: "200",
			config: &pb.CustomHMACProvider{
				SignResp: true,
			},
			authResp: context.AuthResponse{
				FilterState: map[string]string{},
			},

			signRequired: false,
		},
		{
			name:       "should return false if empty hmacUserIDSessionKey",
			statusCode: "200",
			config: &pb.CustomHMACProvider{
				SignResp: true,
			},
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: ""},
			},

			signRequired: false,
		},
		{
			name:       "should return false if statusCode (500) >= 500",
			statusCode: "500",
			config: &pb.CustomHMACProvider{
				SignResp: true,
			},
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "123"},
			},

			signRequired: false,
		},

		{
			name:       "should return false if statusCode (504) >= 500 ",
			statusCode: "504",
			config: &pb.CustomHMACProvider{
				SignResp: true,
			},
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "123"},
			},

			signRequired: false,
		},
		{
			name:       "should return false if invalid status code",
			statusCode: "invalid",
			config: &pb.CustomHMACProvider{
				SignResp: true,
			},
			authResp: context.AuthResponse{
				FilterState: map[string]string{hmacUserIDSessionKey: "123"},
			},

			signRequired: false,
		},
	}

	for _, val := range tcs {
		tc := val
		t.Run(tc.name, func(t *testing.T) {
			signer, err := createCustomHMACProvider(tc.config, &httpmocks.HttpClientWithCtx{}, nil, nil)
			require.Nil(t, err)

			responseHeader := &envoymocks.ResponseHeaderMap{}
			responseHeader.On("Status").Return(volatile.String(tc.statusCode))
			result := signer.SigningRequired(responseHeader, tc.authResp)

			assert.Equal(t, tc.signRequired, result)
		})
	}
}

func TestSigningRequiredReturnFalseIfNilResponseHeader(t *testing.T) {
	config := &pb.CustomHMACProvider{
		SignResp: true,
	}
	signer, err := createCustomHMACProvider(config, &httpmocks.HttpClientWithCtx{}, nil, nil)
	require.Nil(t, err)

	authResp := context.AuthResponse{
		FilterState: map[string]string{hmacUserIDSessionKey: "123"},
	}
	result := signer.SigningRequired(nil, authResp)

	assert.Equal(t, false, result)
}
