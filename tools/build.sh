#!/bin/sh
set -ex

TAG=`git describe --tags --abbrev=0 2> /dev/null | head -n1`
if [ -z $TAG ]; then
    TAG='0.0.0'
fi

APP=oshinko-rest-server

if [ $1 = build ]; then
    OUTPUT_FLAG="-o _output/oshinko-rest-server"
fi

if [ $1 = test ]; then
    TARGET=./tests
    GO_OPTIONS=-v
else
    TARGET=./cmd/oshinko-rest-server
fi

# this export is needed for the vendor experiment for as long as go version
# 1.5 is still in use.
export GO15VENDOREXPERIMENT=1

godep go $1 $GO_OPTIONS -ldflags \
    "-X github.com/redhatanalytics/oshinko-rest/version.gitTag=$TAG -X github.com/redhatanalytics/oshinko-rest/version.appName=$APP" \
    $OUTPUT_FLAG $TARGET
