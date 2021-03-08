// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package security

import (
	"errors"

	"github.com/golang/protobuf/proto"

	"github.com/grab/ego/ego/src/go/envoy"

	pb "github.com/grab/ego/egofilters/http/security/proto"
	"github.com/grab/ego/egofilters/http/security/verifier"
)

type securityStats struct {
	// TODO: add more metrics if needed
	authOK     envoy.Counter
	authDenied envoy.Counter
	authError  envoy.Counter
}

type securityConfig struct {
	verifiers map[string]verifier.Verifier
	signers   map[string]verifier.Signer
	stats     securityStats
}

var (
	// ErrUnsupportedProvider ...
	ErrUnsupportedProvider = errors.New("upsupported provider type")
	// ErrCannotCreateStats ...
	ErrCannotCreateStats = errors.New("can not create stats")
)

// Initialize together with filter at initial stage
func createSecurityConfig(native envoy.GoHttpFilterConfig) (*securityConfig, error) {
	settings := &pb.Settings{}
	bytes := native.Settings() // Volatile! Handle with care!
	if err := proto.Unmarshal([]byte(bytes), settings); err != nil {
		return nil, err
	}
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	verifiers := map[string]verifier.Verifier{}
	signers := map[string]verifier.Signer{}
	providers := settings.GetProviders()
	// TODO: use statsScope to create stats detail for every provider
	scope := native.Scope()
	authOK := scope.CounterFromStatName("auth_ok")
	authDenied := scope.CounterFromStatName("auth_denied")
	authError := scope.CounterFromStatName("auth_error")
	if authOK == nil || authDenied == nil || authError == nil {
		return nil, ErrCannotCreateStats
	}
	secStats := securityStats{
		authOK:     authOK,
		authDenied: authDenied,
		authError:  authError,
	}

	for k, v := range providers {
		switch v.GetProviderType().(type) {
		case *pb.Provider_CustomHmacProvider:
			hmacProvider, _ := verifier.CreateCustomHMACProvider(v.GetCustomHmacProvider())
			verifiers[k] = hmacProvider
			if hmacProvider != nil {
				signers[k] = hmacProvider
			}
		default:
			return nil, ErrUnsupportedProvider
		}

	}
	return &securityConfig{
		verifiers: verifiers,
		signers:   signers,
		stats:     secStats,
	}, nil
}

func (c *securityConfig) findProvider(requirement *pb.Requirement) (verifier.Verifier, verifier.Signer) {
	// TODO: only take care of single provider for now. This needs to be
	//       extended to handle requires_all & require_any to combine multiple
	//       auth-types.
	name := requirement.GetProviderName()
	if name == "" {
		return nil, nil
	}
	return c.verifiers[name], c.signers[name]
}
