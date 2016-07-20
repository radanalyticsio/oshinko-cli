#!/bin/sh
set -e

. tools/common.sh

usage() {
    echo
    echo "Usage: $(basename $0) <command>"
    echo "Commands for building the project, where <command> is one of the following:"
    echo "  build -- build the project, storing in _output/"
    echo "  install -- build and install binaries"
    echo
}

case "$1" in
    build)
        OUTPUT_FLAG="-o _output/oshinko-rest-server"
        TARGET=./cmd/oshinko-rest-server
        ;;
    install)
        TARGET=./cmd/oshinko-rest-server
        ;;
    *)
        usage
        exit 0
        ;;
esac


# this export is needed for the vendor experiment for as long as go version
# 1.5 is still in use.
export GO15VENDOREXPERIMENT=1

set -x

go $1 -ldflags "$TAG_APPNAME_FLAGS" $OUTPUT_FLAG $TARGET
