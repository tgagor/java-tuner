#!/usr/bin/env bash
set -e
for dir in ; do
  go vet /
done
