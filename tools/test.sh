#!/bin/sh
set -e

. tools/common.sh

usage() {
    echo
    echo "Usage: $(basename $0) <command>"
    echo "Run a suite of tests, where <command> is one of the following:"
    echo "  all -- run all tests"
    echo "  unit -- run the unittest package"
    echo "  client -- run the clienttest package"
    echo
}

CLIENTTEST="github.com/radanalyticsio/oshinko-rest/tests/client"
UNITTEST="github.com/radanalyticsio/oshinko-rest/tests/unit"

case "$1" in
    all)
        TEST_PACKAGES="$CLIENTTEST $UNITTEST"
        ;;
    client)
        TEST_PACKAGES="$CLIENTTEST"
        ;;
    unit)
        TEST_PACKAGES="$UNITTEST"
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

go test -v -ldflags "$TAG_APPNAME_FLAGS" $TEST_PACKAGES
