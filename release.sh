#!/bin/bash

set -ex

if [[ ! $(type -P gox) ]]; then
    echo "Error: gox not found."
    echo "To fix: run 'go install github.com/mitchellh/gox@latest', and/or add \$GOPATH/bin to \$PATH"
    exit 1
fi

if [[ ! $(type -P gh) ]]; then
    echo "Error: github cli not found."
    exit 1
fi

VER=$1

if [[ -z $VER ]]; then
    echo "Need to specify version."
    exit 1
fi

PRE_ARG=
if [[ $VER =~ pre ]]; then
    PRE_ARG="--pre-release"
fi

git tag $VER

echo "Building $VER"
echo

gox -ldflags "-X main.version=$VER" -osarch="darwin/amd64 linux/amd64 windows/amd64 linux/arm64 darwin/arm64"

echo "* " > desc
echo "" >> desc

echo "$ sha1sum skeg_*" >> desc
sha1sum skeg_* >> desc
echo "$ sha256sum skeg_*" >> desc
sha256sum skeg_* >> desc
echo "$ md5sum skeg_*" >> desc
md5sum skeg_* >> desc

vi desc

git push --tags

sleep 2

gh release create $VER -t $VER -F desc skeg_*
