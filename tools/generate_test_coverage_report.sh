#!/bin/bash

# Clean up test logs for next run as caches are shared.
rm -rf bazel-out/k8-fastbuild/testlogs

bazel coverage //egofilters/...

COVERAGE_DIR=./generated
mkdir -p "${COVERAGE_DIR}"

COVERAGE_DATA="${COVERAGE_DIR}/coverage.dat"

echo "Merging coverage data..."
# Merge script supports only set.
echo "mode: set" > ${COVERAGE_DATA} && cat $(find -L bazel-out/k8-fastbuild/testlogs/ -name coverage.dat) | grep -v mode: | sort -r | \
awk '{if($1 != last) {print $0;last=$1}}' >> ${COVERAGE_DATA}

echo "Code coverage report..."
GOPATH=$(pwd)/GOPATH GO111MODULE=off go tool cover -html=generated/coverage.dat -o generated/coverage.html

echo "Code coverage summary..."
go tool cover -func=generated/coverage.dat

echo "Coverage report is at $(pwd)/${COVERAGE_DIR}/converage.html"

