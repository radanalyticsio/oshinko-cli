#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT


function check_oc_version {
    vers=$(oc version | grep "oc v" | cut -d' ' -f2- | tr -d v)
    major=$(echo $vers | cut -d'.' -f1)
    minor=$(echo $vers | cut -d'.' -f2)
    if [ $((major)) -lt $1 ]; then
        return 1
    fi
    if [ $((major)) -gt $1 ]; then
        return 0
    fi
    if [ $((minor)) -ge $2 ]; then
        return 0
    fi
    return 1
}

#set +e
#check_oc_version 1 5
#res=$?
#set -e

os::test::junit::declare_suite_start "cmd/scale"

# General note -- at present, the master and worker counts in the included config object on get are "MasterCount" and "WorkerCount"
# the master and worker counts in the outer cluster status are "masterCount" and "workerCount"

os::cmd::try_until_text "_output/oshinko get" "There are no clusters in any projects. You can create a cluster with the 'create' command."

# make a cluster to scale
os::cmd::expect_success "_output/oshinko create abc"
os::cmd::try_until_success "_output/oshinko get abc"
os::cmd::expect_success_and_text "_output/oshinko get abc -o yaml" "WorkerCount: 1"
os::cmd::expect_success_and_text "_output/oshinko get abc -o yaml" "MasterCount: 1"
# could still be creating so use 'until'
os::cmd::try_until_text "_output/oshinko get abc -o yaml" "workerCount: 1"
os::cmd::try_until_text "_output/oshinko get abc -o yaml" "masterCount: 1"

# scale
os::cmd::expect_success_and_text "_output/oshinko scale abc" "neither masters nor workers specified, cluster \"abc\" not scaled"
os::cmd::expect_failure_and_text "_output/oshinko scale abc --masters=2" "cluster configuration must have a master count of 0 or 1"
os::cmd::expect_success "_output/oshinko scale abc --workers=0 --masters=0"
os::cmd::expect_success_and_text "_output/oshinko get abc -o yaml" "WorkerCount: 0"
os::cmd::expect_success_and_text "_output/oshinko get abc -o yaml" "MasterCount: 0"
os::cmd::try_until_text "_output/oshinko get abc -o yaml" "workerCount: 0"
os::cmd::try_until_text "_output/oshinko get abc -o yaml" "masterCount: 0"

os::cmd::expect_success "_output/oshinko scale abc --workers=2"
os::cmd::expect_success_and_text "_output/oshinko get abc -o yaml" "MasterCount: 0"
os::cmd::expect_success_and_text "_output/oshinko get abc -o json" '"WorkerCount": 2'
os::cmd::try_until_text "_output/oshinko get abc -o yaml" "workerCount: 2"
os::cmd::expect_success_and_text "_output/oshinko get abc -o yaml" "masterCount: 0"

os::cmd::expect_success "_output/oshinko scale abc --masters=1"
os::cmd::expect_success_and_text "_output/oshinko get abc -o yaml" "MasterCount: 1"
os::cmd::expect_success_and_text "_output/oshinko get abc -o json" '"WorkerCount": 2'
os::cmd::expect_success_and_text "_output/oshinko get abc -o yaml" "workerCount: 2"
os::cmd::try_until_text "_output/oshinko get abc -o yaml" "masterCount: 1"

os::test::junit::declare_suite_end
