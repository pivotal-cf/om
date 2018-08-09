#!/bin/bash -eu

function main() {
  local cwd
  cwd="${1}"

  go run "${cwd}/../main.go" bake \
    --bosh-variables-directory "${cwd}/bosh-variables" \
    --embed "${cwd}/extra" \
    --forms-directory "${cwd}/forms" \
    --icon "${cwd}/icon.png" \
    --instance-groups-directory "${cwd}/instance-groups" \
    --jobs-directory "${cwd}/jobs" \
    --metadata "${cwd}/base.yml" \
    --migrations-directory "${cwd}/migrations" \
    --output-file "${cwd}/example-1.2.3-build.4.pivotal" \
    --properties-directory "${cwd}/properties" \
    --releases-directory "${cwd}/releases" \
    --runtime-configs-directory "${cwd}/runtime-configs" \
    --stemcell-tarball "${cwd}/stemcell.tgz" \
    --variable "some-variable=some-value" \
    --variables-file "${cwd}/variables.yml" \
    --version "1.2.3-build.4"
}

main "$(cd "$(dirname "${0}")" && pwd)"
