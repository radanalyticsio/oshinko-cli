#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/scale"

os::cmd::try_until_text "_output/oshinko get" "There are no clusters in any projects. You can create a cluster with the 'create' command."

# make a cluster to scale
os::cmd::expect_success "_output/oshinko create abc --workers=1"
os::cmd::try_until_success "_output/oshinko get abc"

# scale
os::cmd::expect_success_and_text "_output/oshinko scale abc" "neither masters nor workers specified, cluster \"abc\" not scaled"
os::cmd::expect_failure_and_text "_output/oshinko scale abc --masters=2" "cluster configuration must have a master count of 0 or 1"
os::cmd::expect_success "_output/oshinko scale abc --workers=0 --masters=0"
os::cmd::expect_success "_output/oshinko scale abc --workers=2"
os::cmd::try_until_text "_output/oshinko get abc -o json" '"workerCount": 2'

os::test::junit::declare_suite_end
