// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package verifier

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/grab/ego/egofilters/http/security/proto"
)

func TestIsValidSignature(t *testing.T) {
	tcs := []struct {
		name     string
		response *http.Response

		valid bool
		err   error
	}{
		{
			name: "should return valid if get 200 status code and valid response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"Valid": true}`)),
			},

			valid: true,
		},
		{
			name: "should return invalid if get 200 status code, but invalid response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"Valid": false}`)),
			},

			valid: false,
		},
		{
			name: "should return valid if get 401 status code, but valid response",
			response: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(strings.NewReader(`{"Valid": true}`)),
			},

			valid: true,
		},
		{
			name: "should return invalid if get 401 status code and invalid response",
			response: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(strings.NewReader(`{"Valid": false}`)),
			},

			valid: false,
		},
		{
			name: "should return invalid if get 401 status code and empty response",
			response: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			},

			valid: false,
		},
		{
			name: "should return error if status code >= 500",
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			},

			err: errors.New("5xx (500) status code"),
		},
		{
			name: "should return error if status code >= 500",
			response: &http.Response{
				StatusCode: http.StatusGatewayTimeout,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			},

			err: errors.New("5xx (504) status code"),
		},
		{
			name: "should return error if can't parse body",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("")),
			},

			err: errors.New("can't parse HMAC signature check response: EOF"),
		},
		{
			name: "should return invalid if get 4xx status",
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(strings.NewReader(`{"Valid": true}`)),
			},

			valid: false,
		},
		{
			name: "should return invalid if get 3xx status",
			response: &http.Response{
				StatusCode: http.StatusPermanentRedirect,
				Body:       ioutil.NopCloser(strings.NewReader(`{"Valid": true}`)),
			},

			valid: false,
		},
		{
			name: "should return invalid if get 2xx (exclude 200 and 204) status",
			response: &http.Response{
				StatusCode: http.StatusAccepted,
				Body:       ioutil.NopCloser(strings.NewReader(`{"Valid": true}`)),
			},

			valid: false,
		},
	}

	for _, tmp := range tcs {
		tc := tmp
		t.Run(tc.name, func(t *testing.T) {
			provider, err := CreateCustomHMACProvider(&pb.CustomHMACProvider{})
			require.Nil(t, err)

			valid, err := provider.isValidSignature(tc.response)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.valid, valid)
		})
	}
}
