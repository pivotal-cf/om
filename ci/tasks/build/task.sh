#!/bin/bash -eu

if [ -z "$GITHUB_TOKEN" ]; then
    echo "GITHUB_TOKEN is required"
    exit 1
fi

export GOPATH="$PWD/go"
export OM_VERSION="$(cat om-version/version)"

cd go/src/github.com/pivotal-cf/om
go version
goreleaser release

