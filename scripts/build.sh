#!/bin/sh
set -ex

go get github.com/renstrom/dedent
go get github.com/docker/go-connections/nat
go get github.com/ghodss/yaml

PROJECT='github.com/radanalyticsio/oshinko-cli'
TAG=`git describe --tags --abbrev=0 2> /dev/null | head -n1`
if [ -z $TAG ]; then
    TAG='0.0.0'
fi

GIT_COMMIT=`git log -n1 --pretty=format:%h`
TAG="${TAG}-${GIT_COMMIT}"

APP=oshinko

CMD=$1

OUTPUT_DIR="_output"
OUTPUT_PATH="$OUTPUT_DIR/$APP"
OUTPUT_FLAG="-o $OUTPUT_PATH"
TARGET=./cmd/oshinko

# this export is needed for the vendor experiment for as long as go version
# 1.5 is still in use.
export GO15VENDOREXPERIMENT=1
if [ $CMD = build ]; then
    go build $GO_OPTIONS -ldflags \
    "-X $PROJECT/version.gitTag=$TAG -X $PROJECT/version.appName=$APP" \
    -o $OUTPUT_PATH $TARGET
    if [ "$?" -eq 0 ]; then
       rm $OUTPUT_DIR/oshinko-cli || true
       ln -s ./oshinko $OUTPUT_DIR/oshinko-cli
    fi
fi
