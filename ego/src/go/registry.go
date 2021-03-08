// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package ego

import (
	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/volatile"
)

type HttpFilterFactoryFactory interface {
	CreateFilterFactory(config envoy.GoHttpFilterConfig) (HttpFilterFactory, error)
	CreateRouteSpecificFilterConfig(config envoy.GoHttpFilterConfig) (interface{}, error)
}

var httpFilterFactoryFactories = map[string]HttpFilterFactoryFactory{}

func RegisterHttpFilter(name string, factory HttpFilterFactoryFactory) HttpFilterFactoryFactory {
	if _, found := httpFilterFactoryFactories[name]; found {
		// TODO: log error, make some noise
	}
	httpFilterFactoryFactories[name] = factory
	return factory
}

func GetHttpFilterFactoryFactory(name volatile.String) HttpFilterFactoryFactory {
	return httpFilterFactoryFactories[string(name)]
}
