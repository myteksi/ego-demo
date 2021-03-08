# Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
#
# Use of this source code is governed by the Apache License 2.0 that can be
# found in the LICENSE file

# Yes, it's all about "-framework Security". This is a bit tricky because
# `envoy_cc_test` sits on `linkopts` without recourse.

#load("@rules_python//python:defs.bzl", "py_binary")
load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_library", "cc_test")
#load("@rules_fuzzing//fuzzing:cc_defs.bzl", "fuzzing_decoration")
#load(":envoy_binary.bzl", "envoy_cc_binary")
load("@envoy//bazel:envoy_library.bzl", "tcmalloc_external_deps")
load(
    "@envoy//bazel:envoy_internal.bzl",
    "envoy_copts",
    "envoy_external_dep_path",
    "envoy_linkstatic",
    "envoy_select_force_libcpp",
    "envoy_stdlib_deps",
    "tcmalloc_external_dep",
)

# copy & paste from @envoy//bazel:envoy_test.bzl with small changes
def _envoy_cc_test_infrastructure_library(
        name,
        srcs = [],
        hdrs = [],
        data = [],
        external_deps = [],
        deps = [],
        repository = "",
        tags = [],
        include_prefix = None,
        copts = [],
        **kargs):
    # Add implicit tcmalloc external dependency(if available) in order to enable CPU and heap profiling in tests.
    deps += tcmalloc_external_deps(repository)

    native.cc_library(
        name = name,
        srcs = srcs,
        hdrs = hdrs,
        data = data,
        copts = envoy_copts(repository, test = True) + copts,
        testonly = 1,
        deps = deps + [envoy_external_dep_path(dep) for dep in external_deps] + [
            envoy_external_dep_path("googletest"),
        ],
        tags = tags,
        include_prefix = include_prefix,
        alwayslink = 1,
        linkstatic = envoy_linkstatic(),
        **kargs
    )

# copy & paste from @envoy//bazel:envoy_test.bzl with small changes
def _envoy_test_linkopts():
    return select({
        "@envoy//bazel:apple": [
            # See note here: https://luajit.org/install.html
            "-pagezero_size 10000",
            "-image_base 100000000",
            "-framework Security",
        ],
        "@envoy//bazel:windows_x86_64": [
            "-DEFAULTLIB:advapi32.lib",
            "-DEFAULTLIB:ws2_32.lib",
            "-WX",
        ],

        # TODO(mattklein123): It's not great that we universally link against the following libs.
        # In particular, -latomic and -lrt are not needed on all platforms. Make this more granular.
        "//conditions:default": ["-pthread", "-lrt", "-ldl"],
    }) + envoy_select_force_libcpp([], ["-lstdc++fs", "-latomic"])

# copy & paste from @envoy//bazel:envoy_test.bzl with small changes
def ego_cc_test(
        name,
        srcs = [],
        data = [],
        # List of pairs (Bazel shell script target, shell script args)
        repository = "",
        external_deps = [],
        deps = [],
        tags = [],
        args = [],
        copts = [],
        shard_count = None,
        coverage = True,
        local = False,
        size = "medium",
        flaky = False):
    if coverage:
        coverage_tags = tags + ["coverage_test_lib"]
    else:
        coverage_tags = tags

    _envoy_cc_test_infrastructure_library(
        name = name + "_lib_internal_only",
        srcs = srcs,
        data = data,
        external_deps = external_deps,
        deps = deps + ["@envoy//test/test_common:printers_includes"],
        repository = repository,
        tags = coverage_tags,
        copts = copts,
        # Allow public visibility so these can be consumed in coverage tests in external projects.
        visibility = ["//visibility:public"],
    )
    if coverage:
        coverage_tags = tags + ["coverage_test"]
    native.cc_test(
        name = name,
        copts = envoy_copts(repository, test = True) + copts,
        linkopts = _envoy_test_linkopts(),
        linkstatic = envoy_linkstatic(),
        malloc = tcmalloc_external_dep(repository),
        deps = envoy_stdlib_deps() + [
            ":" + name + "_lib_internal_only",
            "@envoy//test:main",
        ],
        # from https://github.com/google/googletest/blob/6e1970e2376c14bf658eb88f655a054030353f9f/googlemock/src/gmock.cc#L51
        # 2 - by default, mocks act as StrictMocks.
        args = args + ["--gmock_default_mock_behavior=2"],
        tags = coverage_tags,
        local = local,
        shard_count = shard_count,
        size = size,
        flaky = flaky,
    )
