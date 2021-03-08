// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package loglevel

// spdlog::level::level_enum
//
// see https://github.com/gabime/spdlog/blob/master/include/spdlog/common.h#L78

type Type int
const (
	Trace    Type = 0
	Debug    Type = 1
	Info     Type = 2
	Warn     Type = 3
	Error    Type = 4
	Critical Type = 5
)
