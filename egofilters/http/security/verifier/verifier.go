// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package verifier

import (
	"github.com/grab/ego/ego/src/go/envoy"

	"github.com/grab/ego/egofilters/http/security/context"
)

// Verifier ...
type Verifier interface {
	Verify(context.RequestContext)
	WithBody() bool
}

// Signer ...
type Signer interface {
	// Clients have to check if SigningRequired before calling Sign.
	Sign(context.ResponseContext)
	SigningRequired(headers envoy.ResponseHeaderMap, authResp context.AuthResponse) bool
}
