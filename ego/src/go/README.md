# Layout / Packages

## src/cc/filter

An envoy filter dispatching to the go runtime.

## src/go

A handful of Golang packages (`ego`) abstracting the interaction between Go code
and the Envoy runtime.

### internal/cgo

This package exists mostly because we have trouble linking CGO exports from
multiple packages (<https://github.com/golang/go/issues/27150>). For the similar
reasons, it contains the (unused) `main()` function.

It contains all the CGO stubs for dispatching envoy filter callbacks to Go
handlers. Since we need to avoid circular imports, we wrap all call-backs to
the envoy runtime via interfaces declared in the [envoy](#envoy) package.

This also simplifies development as all of these dependencies can be mocked.

### envoy

This package abstracts all interactions with the C runtime. The interfaces are
instantiated by the [cgo](#cgo) package, which can't be imported by the filter
packages due to language and linker constraints.

### volatile

This package exists mostly because `volatile.String` looks better than
`envoy.VolatileString`.

### stub

This package contains mocks for some of the interfaces defined in the
[envoy](#envoy) package. This is useful for running unit tests without having
to build envoy.
