// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

import (
	logger "github.com/grab/ego/ego/src/go/logger"

	// This is a rather ugly. We can't use ego as a dependency of egofilters
	// because main() needs to be located in the same package as the cgo code,
	// and there can only be one package with cgo code (at least when building
	// with bazel).
	_ "github.com/grab/ego/egofilters"
)

func init() {

	logger.Init(nativeLogger{})

	// FIXME: To be super safe, we should probably say "hi" before accepting
	// CGo calls: https://github.com/golang/go/issues/15943#issuecomment-713153486
}

// main() is not used -- this is a static library.
func main() {

}
