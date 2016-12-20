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
fi

if [ $1 = test ]; then
    TARGET=./tests
    GO_OPTIONS=-v
else
    TARGET=./cmd/oshinko
fi

#-instrument "$PROJECT/pkg/cmd/cli/cmd,$PROJECT/pkg/cmd/cli/cluster,$PROJECT/pkg/cmd/cli"
if [ $1 = debug ]; then
    godebug build  -o _output/oshinko-cli ./cmd/oshinko
fi

# this export is needed for the vendor experiment for as long as go version
# 1.5 is still in use.
export GO15VENDOREXPERIMENT=1

if [ $1 = build ]; then
    go $1 $GO_OPTIONS -ldflags \
    "-X $PROJECT/version.tag=$TAG -X $PROJECT/version.appName=$APP" \
    $OUTPUT_FLAG $TARGET
fi




