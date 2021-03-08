// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package getheader

import (
	"github.com/golang/protobuf/proto"

	pb "github.com/grab/ego/egofilters/http/getheader/proto"
	ego "github.com/grab/ego/ego/src/go"
	"github.com/grab/ego/ego/src/go/envoy"
)

type factory struct {
}

func (f factory) CreateFilterFactory(native envoy.GoHttpFilterConfig) (ego.HttpFilterFactory, error) {
	settings := pb.Settings{}
	bytes := native.Settings() // Volatile! Handle with care!
	if err := proto.Unmarshal([]byte(bytes), &settings); err != nil {
		return nil, err
	}
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	return func(native envoy.GoHttpFilter) ego.HttpFilter {
		return newGetHeaderFilter(&settings, native)
	}, nil
}

// CreateRouteSpecificFilterConfig ...
func (f factory) CreateRouteSpecificFilterConfig(native envoy.GoHttpFilterConfig) (interface{}, error) {
	return struct{}{}, nil
}

// CreatFactoryFactory ...
func CreatFactoryFactory() ego.HttpFilterFactoryFactory {
	return factory{}
}
