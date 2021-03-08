// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package security

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grab/ego/ego/test/go/mock"

	pb "github.com/grab/ego/egofilters/http/security/proto"

	envoymocks "github.com/grab/ego/ego/test/go/mock/gen/envoy"
)

func TestCreateSecurityConfig(t *testing.T) {
	tcs := []struct {
		name string
		// set-up
		pbConfig                        string
		failedToCreateAuthOkCounter     bool
		failedToCreateAuthDeniedCounter bool
		failedToCreateAuthErrorCounter  bool

		// verify
		hasError  bool
		verifiers map[string]string
		signers   map[string]string
	}{
		{
			name: "HMAC verifier",
			pbConfig: `
				providers: <
					key: "my_custom_hmac_provider"
					value: <
						custom_hmac_provider: <
							request_validation_url: "https://custom-auth.example.com/v1/hmacverify"
							service_key: "service_key"
							service_token: "service_token"
						>
					>
				>
			`,

			verifiers: map[string]string{"my_custom_hmac_provider": "*verifier.customHMACProvider"},
			signers:   map[string]string{"my_custom_hmac_provider": "*verifier.customHMACProvider"},
		},

		{
			name: "Can't create auth ok counter",
			pbConfig: `
				providers: <
					key: "my_custom_hmac_provider"
					value: <
						custom_hmac_provider: <
							request_validation_url: "https://custom-auth.example.com/v1/hmacverify"
							service_key: "service_key"
							service_token: "service_token"
						>
					>
				>
			`,
			failedToCreateAuthOkCounter: true,

			hasError: true,
		},

		{
			name: "Can't create auth denied counter",
			pbConfig: `
				providers: <
					key: "my_custom_hmac_provider"
					value: <
						custom_hmac_provider: <
							request_validation_url: "https://custom-auth.example.com/v1/hmacverify"
							service_key: "service_key"
							service_token: "service_token"
						>
					>
				>
			`,
			failedToCreateAuthDeniedCounter: true,

			hasError: true,
		},

		{
			name: "Can't create auth error counter",
			pbConfig: `
				providers: <
					key: "my_custom_hmac_provider"
					value: <
						custom_hmac_provider: <
							request_validation_url: "https://custom-auth.example.com/v1/hmacverify"
							service_key: "service_key"
							service_token: "service_token"
						>
					>
				>
			`,
			failedToCreateAuthErrorCounter: true,

			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			settings := &pb.Settings{}
			err := proto.UnmarshalText(tc.pbConfig, settings)
			require.Nil(t, err)

			configBytes, err := proto.Marshal(settings)
			require.Nil(t, err)

			scope := &envoymocks.Scope{}

			authOkCounter := &envoymocks.Counter{}
			authOkCounter.TestData().Set("name", "auth_ok")

			authDeniedCounter := &envoymocks.Counter{}
			authOkCounter.TestData().Set("name", "auth_denied")

			authErrorCounter := &envoymocks.Counter{}
			authErrorCounter.TestData().Set("name", "auth_error")

			stats := securityStats{
				authOK:     authOkCounter,
				authDenied: authDeniedCounter,
				authError:  authErrorCounter,
			}

			if tc.failedToCreateAuthOkCounter {
				scope.On("CounterFromStatName", "auth_ok").Return(nil)
			} else {
				scope.On("CounterFromStatName", "auth_ok").Return(authOkCounter)
			}
			if tc.failedToCreateAuthDeniedCounter {
				scope.On("CounterFromStatName", "auth_denied").Return(nil)
			} else {
				scope.On("CounterFromStatName", "auth_denied").Return(authDeniedCounter)
			}

			if tc.failedToCreateAuthErrorCounter {
				scope.On("CounterFromStatName", "auth_error").Return(nil)
			} else {
				scope.On("CounterFromStatName", "auth_error").Return(authErrorCounter)
			}

			gohttpConfig := &mock.GoHttpFilterConfig{ConfigBytes: configBytes, EnvoyScope: scope}

			securityConfig, err := createSecurityConfig(gohttpConfig)

			if tc.hasError {
				assert.NotNil(t, err)
			} else {
				require.Nil(t, err)

				scope.AssertExpectations(t)
				assert.Equal(t, stats, securityConfig.stats)

				for name, verifierType := range tc.verifiers {
					assert.Equal(t, verifierType, fmt.Sprintf("%v", reflect.TypeOf(securityConfig.verifiers[name])))
				}

				for name, verifierType := range tc.signers {
					assert.Equal(t, verifierType, fmt.Sprintf("%v", reflect.TypeOf(securityConfig.signers[name])))
				}
			}

		})
	}
}

func TestUnmashalError(t *testing.T) {
	scope := &envoymocks.Scope{}
	gohttpConfig := &mock.GoHttpFilterConfig{ConfigBytes: []byte("invalid_config"), EnvoyScope: scope}

	_, err := createSecurityConfig(gohttpConfig)
	assert.NotNil(t, err)
}
