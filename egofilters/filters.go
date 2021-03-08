// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package egofilters

import (
	ego "github.com/grab/ego/ego/src/go"

	"github.com/grab/ego/egofilters/http/getheader"
	"github.com/grab/ego/egofilters/http/security"
)

func init() {
	ego.RegisterHttpFilter("getheader", getheader.CreatFactoryFactory())
	ego.RegisterHttpFilter("security", security.CreateFactoryFactory())
}
