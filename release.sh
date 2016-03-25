#!/bin/bash

if [[ ! $(type -P gox) ]]; then
    echo "Error: gox not found."
    echo "To fix: run 'go get github.com/mitchellh/gox', and/or add \$GOPATH/bin to \$PATH"
    exit 1
fi

VER=$1

git tag $VER

gox -ldflags "-X main.version $VER" -osarch="darwin/amd64 linux/amd64"

echo "$ sha1sum skeg_*"
sha1sum skeg_*
echo "$ sha256sum skeg_*"
sha256sum skeg_*
echo "$ md5sum skeg_*"
md5sum skeg_*
