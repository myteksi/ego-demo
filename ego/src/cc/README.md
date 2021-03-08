# Envoy Golang Filter Demo

This project demonstrates the linking of additional HTTP filters implemented in Go with the Envoy binary.

A new filter `add-header` which adds a HTTP header is introduced.

Integration tests demonstrating the filter's end-to-end behavior are
also provided.

## Building

To build the Envoy static binary:

1. `git submodule update --init`
2. `bazel build //:envoy`

## Testing

To run the `add-header` integration test:

`bazel test //src/cc/filter/http/addheader:integration_test`

# Layout / Packages

# cgo

This package provides two libraries: one for supporting upcalls to Go (`cgo`) and one for supporting downcalls from Go (`native`).

In order to avoid circular dependencies, the `cgo` library must not be used as dependency for the Golang packages. Rather, these need to depend on the `native` package only.

Both libraries use the file `envoy.h`: one indirectly (via Go code) and one directly (included by the downcall proxies).

# filter