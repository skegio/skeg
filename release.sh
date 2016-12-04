#!/bin/bash

set -e

if [[ ! $(type -P gox) ]]; then
    echo "Error: gox not found."
    echo "To fix: run 'go get github.com/mitchellh/gox', and/or add \$GOPATH/bin to \$PATH"
    exit 1
fi

if [[ ! $(type -P github-release) ]]; then
    echo "Error: github-release not found."
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

gox -ldflags "-X main.version $VER" -osarch="darwin/amd64 linux/amd64"

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

github-release release $PRE_ARG --user skegio --repo skeg --tag $VER --name $VER --description desc
github-release upload --user skegio --repo skeg --tag $VER --name skeg_darwin_amd64 --file skeg_darwin_amd64
github-release upload --user skegio --repo skeg --tag $VER --name skeg_linux_amd64 --file skeg_linux_amd64
