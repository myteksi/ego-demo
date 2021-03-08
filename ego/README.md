This repository contains a shim for simplifying development of envoy filters using Golang.

Due to poorly understood build mechanics, this presently isn't a Bazel workspace in its own right.

Rather, it needs to be added as git submodule into a repository that also embeds envoy and provides the actual filter implementations.

In fact, this repository even has a reference to a directory in the embedding project (egofilters), see `src/go/internal/cgo/cutils.go`.

The reason for this is that the filters in the embedding project need to register themselves via package init() calls that would never be run as part of the cgo static library initialization, otherwise.

This is aggravated by the fact that Bazel's `rules_go` presently only allows one Cgo package.