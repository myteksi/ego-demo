// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AuthResponse(t *testing.T) {
	r := &responseContextImpl{}
	r.authResponse = AuthResponseOK()
	res := r.AuthResponse()
	assert.Equal(t, r.authResponse, res)
}

func Test_BodyReader(t *testing.T) {
	r := &responseContextImpl{}
	res := r.BodyReader()
	assert.Nil(t, res)
}

func Test_GoContext(t *testing.T) {
	r := &responseContextImpl{}
	res := r.GoContext()
	assert.Nil(t, res)
}

func Test_GetSecret(t *testing.T) {
	r := &responseContextImpl{}
	res := r.GetSecret("test")
	assert.Equal(t, "", res)
}

func Test_Callbacks(t *testing.T) {
	r := &responseContextImpl{}
	res := r.Callbacks()
	assert.Nil(t, res)
}

func Test_Headers(t *testing.T) {
	r := &responseContextImpl{}
	res := r.Headers()
	assert.Nil(t, res)
}

func Test_RequestHeaders(t *testing.T) {
	r := &responseContextImpl{}
	res := r.RequestHeaders()
	assert.Nil(t, res)
}

func Test_Logger(t *testing.T) {
	r := &responseContextImpl{}
	res := r.Logger()
	assert.Nil(t, res)
}

func Test_ActiveSpan(t *testing.T) {
	r := &responseContextImpl{}
	res := r.ActiveSpan()
	assert.Nil(t, res)
}
