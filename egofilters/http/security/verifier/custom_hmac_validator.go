// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package verifier

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type hmacSignatureValidationResponse struct {
	Valid bool `json:"valid"`
}

type isValidHMACSignatureOpt func(resp *http.Response) (bool, error)

func isValidSignature(resp *http.Response) (bool, error) {
	if resp.StatusCode >= http.StatusInternalServerError {
		return false, fmt.Errorf("5xx (%v) status code", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return false, nil
	}

	result := hmacSignatureValidationResponse{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, fmt.Errorf("can't parse HMAC signature check response: %v", err)
	}

	return result.Valid, nil
}
