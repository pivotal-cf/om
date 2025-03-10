#!/bin/bash -eu

if [ -z "$GITHUB_TOKEN" ]; then
    echo "GITHUB_TOKEN is required"
    exit 1
fi

export OM_VERSION="$(cat om-version/version)"

cd om
git remote set-url origin https://github.com/pivotal-cf/om
git fetch -t -P -p

# Extract Go version from go.mod file
GO_VERSION=$(grep -E "^go [0-9]+\.[0-9]+(\.[0-9]+)?" go.mod | awk '{print $2}')
if [ -z "$GO_VERSION" ]; then
    echo "Failed to extract Go version from go.mod"
    exit 1
fi

echo "Using Go version $GO_VERSION from go.mod"

# Download and install the Go version specified in go.mod
wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
tar xf "go${GO_VERSION}.linux-amd64.tar.gz"
rm "go${GO_VERSION}.linux-amd64.tar.gz"
rm -rf /usr/local/go
mv go /usr/local

go version
goreleaser release

