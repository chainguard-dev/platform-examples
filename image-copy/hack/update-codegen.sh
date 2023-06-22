#!/usr/bin/env bash

# Copyright 2022 Chainguard, Inc.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT_DIR=$PWD/$(dirname "$0")/..

echo === Tidying up for Golang
go mod tidy -compat=1.17

echo === Generating for Golang
go generate ./...

# Make sure our dependencies are up-to-date
${REPO_ROOT_DIR}/hack/update-deps.sh
