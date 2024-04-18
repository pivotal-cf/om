#!/bin/bash -eu

if [ -z "$GITHUB_TOKEN" ]; then
    echo "GITHUB_TOKEN is required"
    exit 1
fi

export OM_VERSION="$(cat om-version/version)"

cd om
git remote set-url origin https://github.com/pivotal-cf/om
git fetch -t -P -p

# Kludge to get this thing buillt by go 1.22.2
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
tar xf go1.22.2.linux-amd64.tar.gz
rm -rf /usr/local/go
mv go /usr/local

go version
goreleaser release

