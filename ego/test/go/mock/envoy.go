package mock

import (
	"github.com/grab/ego/ego/src/go/envoy"
	"github.com/grab/ego/ego/src/go/volatile"
)

type GoHttpFilterConfig struct {
	ConfigBytes []byte
	EnvoyScope  envoy.Scope
}

func (conf *GoHttpFilterConfig) Settings() volatile.Bytes {
	return volatile.Bytes(conf.ConfigBytes)
}

func (conf *GoHttpFilterConfig) Scope() envoy.Scope {
	return conf.EnvoyScope
}
