// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AuthResponseOK(t *testing.T) {
	response := AuthResponseOK()
	assert.NotNil(t, response)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, AuthOK, response.Status)
}

func Test_AuthResponseUnauthorized(t *testing.T) {
	response := AuthResponseUnauthorized()
	assert.NotNil(t, response)
	assert.Equal(t, 401, response.StatusCode)
	assert.Equal(t, AuthDenied, response.Status)
}

func Test_AuthResponseDenied(t *testing.T) {
	response := AuthResponseDenied(401)
	assert.NotNil(t, response)
	assert.Equal(t, 401, response.StatusCode)
	assert.Equal(t, AuthDenied, response.Status)
}

func Test_AuthResponseError(t *testing.T) {
	response := AuthResponseError()
	assert.NotNil(t, response)
	assert.Equal(t, 500, response.StatusCode)
	assert.Equal(t, AuthError, response.Status)
}
