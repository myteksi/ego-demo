// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package verifier

import (
	"github.com/grab/ego/egofilters/http/security/context"
)

type baseProvider struct {
}

func (v *baseProvider) Verify(ctx context.RequestContext) {}

func (v *baseProvider) WithBody() bool {
	return false
}
