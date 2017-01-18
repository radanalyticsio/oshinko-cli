#!/bin/sh
set -ex

go get github.com/renstrom/dedent
go get github.com/docker/go-connections/nat
go get github.com/ghodss/yaml

PROJECT='github.com/redhatanalytics/oshinko-cli'
TAG=`git describe --tags --abbrev=0 2> /dev/null | head -n1`
if [ -z $TAG ]; then
    TAG='0.0.0'
fi

GIT_COMMIT=`git log -n1 --pretty=format:%h`
TAG="${GIT_TAG}-${GIT_COMMIT}"

APP=oshinko-cli

if [ $1 = build ]; then
    OUTPUT_FLAG="-o _output/oshinko-cli"
elif [ $1 = build-extended ]; then
    OUTPUT_FLAG="-o _output/oshinko-clix"
fi

if [ $1 = test ]; then
    TARGET=./tests
    GO_OPTIONS=-v
elif [ $1 = build ]; then
    TARGET=./cmd/oshinko
else
    TARGET=./cmd/extended
fi

# this export is needed for the vendor experiment for as long as go version
# 1.5 is still in use.
export GO15VENDOREXPERIMENT=1

if [ $1 != test ]; then
    go build $GO_OPTIONS -ldflags \
    "-X $PROJECT/version.tag=$TAG -X $PROJECT/version.appName=$APP" \
    $OUTPUT_FLAG $TARGET
fi




