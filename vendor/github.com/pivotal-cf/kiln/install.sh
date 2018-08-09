#!/bin/bash -exu

function main() {
  local cwd
  cwd="${1}"

  pushd "${cwd}" > /dev/null
    go install -ldflags "-X main.version=$(git rev-parse HEAD)"
  popd > /dev/null
}

main "$(cd "$(dirname "${0}")" && pwd)"
