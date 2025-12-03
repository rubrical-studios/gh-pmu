#!/bin/bash
set -e

# Build script for gh-extension-precompile
# $1 is the output filename (e.g., dist/gh-pmu_v0.2.0_darwin-amd64)

go build -ldflags="-X github.com/scooter-indie/gh-pmu/cmd.version=${GH_RELEASE_TAG}" -o "$1"
