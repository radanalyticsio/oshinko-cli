#!/bin/sh
set -ex
TOP_DIR=$(readlink -f `dirname "$0"` | grep -o '.*/oshinko-cli')
. $TOP_DIR/sparkimage.sh

PROJECT='github.com/radanalyticsio/oshinko-cli'
TAG=`git describe --tags --abbrev=0 2> /dev/null | head -n1`
if [ -z $TAG ]; then
    TAG='0.0.0'
fi

GIT_COMMIT=`git log -n1 --pretty=format:%h`
TAG="${TAG}-${GIT_COMMIT}"

APP=oshinko-crd

CMD=$1

OUTPUT_DIR="_output"
OUTPUT_PATH="$OUTPUT_DIR/$APP"
OUTPUT_FLAG="-o $OUTPUT_PATH"
TARGET=./cmd/crd

# this export is needed for the vendor experiment for as long as go version
# 1.5 is still in use.
export GO15VENDOREXPERIMENT=1
if [ $CMD = build ]; then
    go build $GO_OPTIONS -ldflags \
    "-X $PROJECT/version.gitTag=$TAG -X $PROJECT/version.appName=$APP -X $PROJECT/version.sparkImage=$SPARK_IMAGE"\
    -o $OUTPUT_PATH $TARGET
fi
