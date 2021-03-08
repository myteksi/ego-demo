// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package security

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/grab/ego/ego/src/go/volatile"
	egomock "github.com/grab/ego/ego/test/go/mock"

	pb "github.com/grab/ego/egofilters/http/security/proto"

	envoymocks "github.com/grab/ego/ego/test/go/mock/gen/envoy"
)

func TestCreateFilterFactoryReturnsError(t *testing.T) {
	factoryFactory := CreateFactoryFactory()

	gohttpConfig := &egomock.GoHttpFilterConfig{ConfigBytes: []byte("invalid_config")}
	_, err := factoryFactory.CreateFilterFactory(gohttpConfig)

	assert.NotNil(t, err)
}

func TestCreateFilterSuccessfully(t *testing.T) {
	pbConfig := `
		providers: <
			key: "my_custom_hmac_provider"
			value: <
				custom_hmac_provider: <
					request_validation_url: "https://hmac.example.com/validate"
					service_key: "me"
					service_token: "top-secret!"
				>
			>
		>
	`
	settings := &pb.Settings{}
	err := proto.UnmarshalText(pbConfig, settings)
	require.Nil(t, err)

	configBytes, err := proto.Marshal(settings)
	require.Nil(t, err)

	scope := &envoymocks.Scope{}
	scope.On("CounterFromStatName", "auth_ok").Return(&envoymocks.Counter{})
	scope.On("CounterFromStatName", "auth_denied").Return(&envoymocks.Counter{})
	scope.On("CounterFromStatName", "auth_error").Return(&envoymocks.Counter{})

	factoryFactory := CreateFactoryFactory()

	gohttpConfig := &egomock.GoHttpFilterConfig{ConfigBytes: configBytes, EnvoyScope: scope}
	factory, err := factoryFactory.CreateFilterFactory(gohttpConfig)

	require.Nil(t, err)
	require.NotNil(t, factory)

	native := &envoymocks.GoHttpFilter{}

	secretProvider := &envoymocks.GenericSecretConfigProvider{}
	native.On("GenericSecretProvider").Return((secretProvider))
	secretProvider.On("Secret").Return(volatile.String("secret"))

	native.On("Log", mock.Anything, mock.Anything)

	filter := factory(native)
	assert.NotNil(t, filter)
}

func TestCreateRouteSpecificConfigReturnsError(t *testing.T) {
	factoryFactory := CreateFactoryFactory()

	gohttpConfig := &egomock.GoHttpFilterConfig{ConfigBytes: []byte("invalid_config")}
	_, err := factoryFactory.CreateRouteSpecificFilterConfig(gohttpConfig)

	assert.NotNil(t, err)
}

func TestCreateRouteSpecificSuccessfully(t *testing.T) {
	pbConfig := `
		provider_name: "something"
	`
	settings := &pb.Requirement{}
	err := proto.UnmarshalText(pbConfig, settings)
	require.Nil(t, err)

	configBytes, err := proto.Marshal(settings)
	require.Nil(t, err)

	factoryFactory := CreateFactoryFactory()

	gohttpConfig := &egomock.GoHttpFilterConfig{ConfigBytes: configBytes}
	config, err := factoryFactory.CreateRouteSpecificFilterConfig(gohttpConfig)

	require.Nil(t, err)
	assert.NotNil(t, config)
}

func TestCreateRouteSpecificReturnsValidationError(t *testing.T) {
	pbConfig := `
		requires_any: <
			requirements: <
			>
		>
	`
	settings := &pb.Requirement{}
	err := proto.UnmarshalText(pbConfig, settings)
	require.Nil(t, err)

	configBytes, err := proto.Marshal(settings)
	require.Nil(t, err)

	factoryFactory := CreateFactoryFactory()

	gohttpConfig := &egomock.GoHttpFilterConfig{ConfigBytes: configBytes}
	_, err = factoryFactory.CreateRouteSpecificFilterConfig(gohttpConfig)

	require.NotNil(t, err)
}
