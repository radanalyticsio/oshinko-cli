#!/bin/sh
set -ex

TAG=`git describe --tags --abbrev=0 2> /dev/null | head -n1`
if [ -z $TAG ]; then
    TAG='0.0.0'
fi

godep go build -ldflags "-X handlers.gitTag=$TAG" -o _output/oshinko-rest-server ./cmd/oshinko-rest-server
