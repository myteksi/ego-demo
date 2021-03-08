// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

//#include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"fmt"
	"unsafe"

	ego "github.com/grab/ego/ego/src/go"
	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/logger"
	"github.com/grab/ego/ego/src/go/volatile"
)

type goHttpFilterConfig struct {
	settings volatile.Bytes
	scope    scope
}

func (c goHttpFilterConfig) Settings() volatile.Bytes {
	return c.settings
}

func (c goHttpFilterConfig) Scope() envoy.Scope {
	return c.scope
}

//export Cgo_GoHttpFilterFactory_Create
func Cgo_GoHttpFilterFactory_Create(factorySlot uint64, name *C.char, nameLen C.size_t,
	settings unsafe.Pointer, settingsLen C.size_t, scopePtr unsafe.Pointer) (result uint64) {
	log := logger.NewLogger("Cgo_GoHttpFilterFactory_Create", nativeLogger{})

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("panic recover with error", err))
			result = 0
		}
	}()

	factoryFactory := ego.GetHttpFilterFactoryFactory(CStrN(name, nameLen))
	if nil == factoryFactory {
		log.Error("can not find factory by name")
		return 0
	}

	cfg := goHttpFilterConfig{
		settings: CBytes(settings, settingsLen, settingsLen),
		scope: scope{
			ptr: scopePtr,
		},
	}
	factory, err := factoryFactory.CreateFilterFactory(cfg)
	if err != nil {
		log.Error(fmt.Sprintf("invoke CreateFilterFactory failed with error", err))
		return 0
	}
	if nil == factory {
		log.Error("invoke CreateFilterFactory without error but return nil")
		return 0
	}

	return TagHttpFilterFactory(factorySlot, factory)
}

//export Cgo_GoHttpFilterFactory_OnDestroy
func Cgo_GoHttpFilterFactory_OnDestroy(factoryTag uint64) {
	log := logger.NewLogger("Cgo_GoHttpFilterFactory_OnDestroy", nativeLogger{})
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("panic recover with error", err))
		}
	}()
	factory := RemoveHttpFilterFactory(factoryTag)
	if nil == factory {
		log.Error("invoke remove factory return nil")
		return
	}
}

// httpFilterFactories is a separate clutch for all our filter factories. This
// is a little bit overblown, indeed, but since factories are expected to live
// much longer than filters, this could result in severe hunk leakage.
//
var httpFilterFactories = &clutch{}

// Cgo_AcquireHttpFilterFactorySlot is the public proxy for
// httpFilterFactories.AcquireSlot
//
//export Cgo_AcquireHttpFilterFactorySlot
func Cgo_AcquireHttpFilterFactorySlot() uint64 {
	return httpFilterFactories.AcquireSlot()
}

// Cgo_ReleaseHttpFilterFactorySlot is the public proxy for
// httpFilterFactories.ReleaseSlot
//
//export Cgo_ReleaseHttpFilterFactorySlot
func Cgo_ReleaseHttpFilterFactorySlot(id uint64) {
	httpFilterFactories.ReleaseSlot(id)
}

// TagHttpFilterFactory is the public proxy for
// httpFilterFactories.TagItem
//
func TagHttpFilterFactory(slot uint64, factory ego.HttpFilterFactory) uint64 {
	return httpFilterFactories.TagItem(slot, factory)
}

// GetHttpFilterFactory is the public proxy for
// httpFilterFactories.GetItem
//
func GetHttpFilterFactory(tag uint64) ego.HttpFilterFactory {
	factory, _ := httpFilterFactories.GetItem(tag).(ego.HttpFilterFactory)
	return factory
}

// RemoveHttpFilterFactory is the public proxy for
// httpFilterFactories.RemoveItem
//
func RemoveHttpFilterFactory(tag uint64) ego.HttpFilterFactory {
	factory, _ := httpFilterFactories.RemoveItem(tag).(ego.HttpFilterFactory)
	return factory
}

//export Cgo_RouteSpecificFilterConfig_Create
func Cgo_RouteSpecificFilterConfig_Create(configSlot uint64, name *C.char, nameLen C.size_t, settings unsafe.Pointer, settingsLen C.size_t) (result uint64) {
	log := logger.NewLogger("Cgo_RouteSpecificFilterConfig_Create", nativeLogger{})

	factoryFactory := ego.GetHttpFilterFactoryFactory(CStrN(name, nameLen))
	if factoryFactory == nil {
		return 0
	}

	// TODO: Define new struct for route specific configuration?
	cfg := goHttpFilterConfig{
		settings: CBytes(settings, settingsLen, settingsLen),
	}

	config, err := factoryFactory.CreateRouteSpecificFilterConfig(cfg)
	if err != nil {
		log.Error(fmt.Sprintf("invoke CreateRouteSpecificFilterConfig failed with error", err))
		return 0
	}
	if nil == config {
		log.Error("invoke CreateRouteSpecificFilterConfig without error but return nil")
		return 0
	}

	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("panic recover with error", err))
			result = 0
		}
	}()

	return TagRouteSpecificFilterConfig(configSlot, config)
}

//export Cgo_RouteSpecificFilterConfig_dtor
func Cgo_RouteSpecificFilterConfig_dtor(configTag uint64) {
	log := logger.NewLogger("Cgo_RouteSpecificFilterConfig_dtor", nativeLogger{})
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("panic recover with error", err))
		}
	}()
	factory := RemoveRouteSpecificFilterConfig(configTag)
	if nil == factory {
		log.Error("invoke remove factory return nil")
		return
	}
}

var routeSpecificFilterConfigs = &clutch{}

// Cgo_AcquireRouteSpecificFilterConfigSlot is the public proxy for
// routeSpecificFilterConfigs.AcquireSlot
//
//export Cgo_AcquireRouteSpecificFilterConfigSlot
func Cgo_AcquireRouteSpecificFilterConfigSlot() uint64 {
	return routeSpecificFilterConfigs.AcquireSlot()
}

// Cgo_ReleaseRouteSpecificFilterConfigSlot is the public proxy for
// routeSpecificFilterConfigs.ReleaseSlot
//
//export Cgo_ReleaseRouteSpecificFilterConfigSlot
func Cgo_ReleaseRouteSpecificFilterConfigSlot(id uint64) {
	routeSpecificFilterConfigs.ReleaseSlot(id)
}

// TagRouteSpecificFilterConfig is the public proxy for
// routeSpecificFilterConfigs.TagItem
//
func TagRouteSpecificFilterConfig(slot uint64, config interface{}) uint64 {
	return routeSpecificFilterConfigs.TagItem(slot, config)
}

// GetRouteSpecificFilterConfig is the public proxy for
// routeSpecificFilterConfigs.GetItem
//
func GetRouteSpecificFilterConfig(tag uint64) interface{} {
	config := routeSpecificFilterConfigs.GetItem(tag)
	return config
}

// RemoveRouteSpecificFilterConfig is the public proxy for
// routeSpecificFilterConfigs.RemoveItem
//
func RemoveRouteSpecificFilterConfig(tag uint64) interface{} {
	config := routeSpecificFilterConfigs.RemoveItem(tag)
	return config
}
