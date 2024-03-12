#! /bin/bash

existing_pr=$(gh pr list --state all | grep "Upgrade Golang")
if [[ -n ${existing_pr} ]]; then
    echo "PR with title 'Upgrade Golang' already exists. Exiting..."
    exit 0
fi
echo "PR not found"





existing_pr=$(gh pr list --search "Upgrade Golang in:title" | grep "Upgrade Golang")
          if [[ -n ${existing_pr} ]]; then
            echo "PR with title 'Upgrade Golang' already exists. Exiting..."
            exit 0
          fi