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

godep go $1 -ldflags \
    "-X github.com/redhatanalytics/oshinko-rest/handlers.gitTag=$TAG -X github.com/redhatanalytics/oshinko-rest/handlers.appName=$APP" \
    $OUTPUT_FLAG ./cmd/oshinko-rest-server
