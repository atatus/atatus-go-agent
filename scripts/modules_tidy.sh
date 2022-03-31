#!/usr/bin/env bash
set -e

for dir in $(scripts/moduledirs.sh); do
    (cd $dir && go mod tidy)
done
