#!/usr/bin/env bash
set -e -o pipefail

if [[ -f "go.sum" ]]; then
        govulncheck ./...
fi
