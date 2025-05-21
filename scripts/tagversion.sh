#!/bin/bash
set -e

export GO111MODULE=on

SED_BINARY=sed
if [[ $(uname -s) == "Darwin" ]]; then SED_BINARY=gsed; fi

prefix=go.atatus.com/agent
version=$(${SED_BINARY} 's@^\s*AgentVersion = "\(.*\)"$@\1@;t;d' version.go)
modules=$(for dir in $(./scripts/moduledirs.sh); do (cd $dir && go list -m); done | grep ${prefix}/)

echo "# Create tags"
for m in "" $modules; do
    p=$(echo $m | ${SED_BINARY} "s@^${prefix}/\(.\{0,\}\)@\1/@")
    tag="${p}v${version}"
    echo "📌 Creating tag: $tag"
    git tag -a "$tag" -m "v${version}"
done

echo
echo "# Push tags"
git push --all
git push --tag
echo
