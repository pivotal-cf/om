#!/bin/bash -eu

if [ -z "$GITHUB_TOKEN" ]; then
    echo "GITHUB_TOKEN is required"
    exit 1
fi

export OM_VERSION="$(cat om-version/version)"

cd om
go version
goreleaser release

