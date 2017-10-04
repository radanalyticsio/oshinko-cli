#!/bin/bash
#
# This script is for running the oshinko-rest tests. It can be used to
# run either the unit tests, or the full client end-to-end tests. In the
# case of the latter, the tests assume that the execution shell has access
# to the `oc` command, and that a user with sufficient privileges has been
# authenticated.
#
# The unit tests are run by executing the standard Go language test command
# on the command line within the repository.
#
# The client tests are run by building a container within OpenShift based on
# a reference into the oshinko-rest source repository, then using the Go
# test command to invoke a series of tests that will run an oshinko-rest
# server and execute client calls against that server. The client tests will
# exercise the actual OpenShift and Kubernetes APIs to create, read, update,
# and destroy objects.
#
# As these two tests are radically different in execution and style, special
# care should be taken before running the client tests as they will execute
# commands against a running OpenShift installation.

# This script will exit if there is an error with any command, be sure that
# the user you are logged in as in OpenShift can perform the necessary
# cluster operations.
set -e

usage() {
    echo "usage: oshinko-test.sh [options]"
    echo
    echo "run tests for the oshinko-rest project."
    echo
    echo "required argument:"
    echo "  -t TEST       the name of the test to run {client, unit}"
    echo
    echo "optional arguments:"
    echo "  -h            show this help message"
    echo
    echo "client test arguments:"
    echo "  -p PROJECT    a project to run the tests in, created if it doesn't exist (default: current project)"
    echo "  -r REPO       the Git source repo to build the test from (default: https://github.com/radanalyticsio/oshinko-cli)"
    echo "  -b REF        a Git branch, tag, or commit to build the test from (default: master)"
    echo "  -n NAME       the Pod name for the client test (default: oshinko-tests)"
    echo "  -i URI        the URI for the internal registry which will store the built image, if not supplied the script will attempt to ascertain"
}

while getopts :t:p:r:b:n:i:h opt; do
    case $opt in
        t)
            REQUESTED_TEST=$OPTARG
            ;;
        p)
            PROJECT=$OPTARG
            ;;
        r)
            SOURCE_REPO=$OPTARG
            ;;
        b)
            SOURCE_REF=$OPTARG
            ;;
        n)
            POD_NAME=$OPTARG
            ;;
        i)
            REGISTRY=$OPTARG
            ;;
        h)
            usage
            exit 0
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            exit 1
            ;;
    esac
done

# Where this script lives ...
SCRIPT_DIR=$(readlink -f `dirname "${BASH_SOURCE[0]}"`)

case "$REQUESTED_TEST" in
    unit)
        # bring in the tag variable
        source $SCRIPT_DIR/common.sh
        # this export is needed for the vendor experiment for as long as go
        # version 1.5 is still in use.
        go get gopkg.in/check.v1
        export GO15VENDOREXPERIMENT=1
        go test -v -ldflags "$TAG_APPNAME_FLAGS" "github.com/radanalyticsio/oshinko-cli/rest/tests/unit"
        ;;

    client-local)
        # This test mode is for development and CI testing. It uses code from the
        # local git repository to run the oshinko rest client in the current project
        # on the local host. This is different from how the "client" test below
        # works, which runs the rest client in a pod in OpenShift after building
        # an image from the specified git repository.

        # As such, this mode needs a few environment variables set so that oshinko
        # knows what project it's in and how to authenticate to openshift.
        # Requires a current oc login. Does not require a serviceaccount.

        PROJECT=$(oc project -q)
        export OSHINKO_CLUSTER_NAMESPACE=$PROJECT
        export OSHINKO_KUBE_CONFIG=~/.kube/config
        set +e
        # These empty configmaps are needed for tests that look at reported config
        # Since they're empty they're not actually used, just reported back in status, this is fine
        oc create configmap clusterconfig
        oc create configmap masterconfig
        oc create configmap workerconfig
        set -e

        # bring in the tag variable
        source $SCRIPT_DIR/common.sh

        # this export is needed for the vendor experiment for as long as go
        # version 1.5 is still in use.
        export GO15VENDOREXPERIMENT=1
        go test -v -ldflags "$TAG_APPNAME_FLAGS" "github.com/radanalyticsio/oshinko-cli/rest/tests/client"
        ;;

    client-deployed)
        # bring in the tag variable
        source $SCRIPT_DIR/common.sh
        # this export is needed for the vendor experiment for as long as go
        # version 1.5 is still in use.
        export GO15VENDOREXPERIMENT=1
        go test -v -ldflags "$TAG_APPNAME_FLAGS" "github.com/radanalyticsio/oshinko-cli/rest/tests/client"
        ;;

    client)
        if which oc &> /dev/null
        then :
        else
            echo "Cannot find oc command, please check path to ensure it is installed"
            exit 1
        fi

        if [ -n "$PROJECT" ]
        then
            if oc new-project $PROJECT
            then :
            else
                oc project $PROJECT
            fi
        else
            PROJECT=$(oc project -q)
        fi

        if [ -z "$SOURCE_REPO" ]
        then
            SOURCE_REPO=https://github.com/radanalyticsio/oshinko-cli
        fi

        if [ -z "$SOURCE_REF" ]
        then
            SOURCE_REF=master
        fi

        if [ -z "$POD_NAME" ]
        then
            POD_NAME=oshinko-tests
        fi

        if [ -z "$REGISTRY" ]
        then
            REGISTRY=$(oc get service docker-registry -n default --template='{{index .spec.clusterIP}}:{{index .spec.ports 0 "port"}}')
        fi

        oc create sa oshinko -n $PROJECT
        oc policy add-role-to-user admin system:serviceaccount:$PROJECT:oshinko -n $PROJECT

        # These empty configmaps are needed for tests that look at reported config
        # Since they're empty they're not actually used, just reported back in status, this is fine
        oc create configmap clusterconfig
        oc create configmap masterconfig
        oc create configmap workerconfig

        oc process -f tools/oshinko-client-tests.yaml \
                   -v SOURCE_REPO=$SOURCE_REPO \
                   -v SOURCE_REF=$SOURCE_REF \
                   -v CLIENT_TEST_NAME=$POD_NAME \
                   -v CLIENT_TEST_IMAGE=$REGISTRY/$PROJECT/oshinko-client-tests-image:latest \
                   | oc create -n $PROJECT -f -

        # at this point, the build should be creating the image for the client
        # tests. the script will now loop until it is able to determine the
        # outcome of the tests. if nothing has happened within 10 minutes, it
        # will timeout.
        for i in {1..60}
        do
            TEST_STATUS=$(oc get pod ${POD_NAME} --template={{.status.phase}})
            case $TEST_STATUS in
                Succeeded)
                    EXIT_STATUS=0
                    ;;
                Failed)
                    EXIT_STATUS=1
                    ;;
                *)
                    echo "waiting on test pod, phase = $TEST_STATUS"
                    sleep 10
                    ;;
            esac
            if [ -n "$EXIT_STATUS" ]
            then
                oc logs $POD_NAME
                exit $EXIT_STATUS
            fi
        done
        echo "Error: client test timed out."
        exit 1
        ;;

    *)
        echo "Error: unrecognized test requested."
        usage
        exit 1
        ;;
esac
