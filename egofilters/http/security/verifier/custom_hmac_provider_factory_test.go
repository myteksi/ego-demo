// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package verifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/grab/ego/egofilters/http/security/proto"
)

func TestCreateCustomHMACProvider(t *testing.T) {
	provider, err := CreateCustomHMACProvider(&pb.CustomHMACProvider{})

	require.Nil(t, err)
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.getCurrentTime())
}
