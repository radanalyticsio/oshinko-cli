#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/get"
# This test validates the help commands and output text
os::cmd::expect_success "oc whoami"
os::cmd::expect_success "oc project default/127-0-0-1:28443/system:admin"
# verify some default commands
os::cmd::expect_success_and_text "_output/oshinko get" "There are no clusters in any projects. You can create a cluster with the 'create' command."
os::cmd::expect_success "_output/oshinko version"
os::cmd::expect_success "oc login -u oshinko -p password"
os::cmd::expect_success "oc new-project oshinko"
os::cmd::expect_success "oc get pods"
os::cmd::expect_success "oc whoami -t"
os::cmd::expect_success "oc whoami"

#create
os::cmd::expect_success "_output/oshinko create abc --workers=1"
VERBOSE=true os::cmd::expect_success "_output/oshinko get abc -o json"

#scale
os::cmd::expect_success_and_text "_output/oshinko scale abc --token=`oc whoami -t`" "neither masters nor workers specified, cluster \"abc\" not scaled"
os::cmd::expect_failure_and_text "_output/oshinko scale abc --masters=2 --token=`oc whoami -t`" "cluster configuration must have a master count of 0 or 1"
os::cmd::expect_success "_output/oshinko scale abc --workers=0 --masters=0 --token=`oc whoami -t`"
os::cmd::expect_success "_output/oshinko scale abc --workers=2 --token=`oc whoami -t`"
os::cmd::try_until_text "_output/oshinko get abc --token=`oc whoami -t` -o json" '"workerCount": 0' 2

# shared or ephemeral
os::cmd::expect_success_and_text "_output/oshinko-cli get abc" "<shared>"

# incomplete
os::cmd::expect_success "oc delete service abc-ui"
os::cmd::expect_success_and_text "_output/oshinko-cli get abc" "Incomplete"
os::cmd::expect_success_and_text "_output/oshinko-cli get abc" "<missing>"

#delete
os::cmd::expect_success "_output/oshinko delete abc --token=`oc whoami -t`"
os::cmd::expect_failure_and_text "_output/oshinko get abc --token=`oc whoami -t` -o json" "no such cluster 'abc'"

#flags
os::cmd::expect_failure_and_text "_output/oshinko get --token=`oc whoami -t` --verbose --namespace=bob" "Using project \"bob\""
os::cmd::expect_success_and_text "_output/oshinko get --token=`oc whoami -t` --verbose" "Using project \"oshinko\""

# deprecation notice
os::cmd::expect_success_and_text "_output/oshinko-cli get --token=`oc whoami -t`" "The 'oshinko-cli' command has been deprecated."
os::cmd::expect_success_and_not_text "_output/oshinko get --token=`oc whoami -t`" "The 'oshinko-cli' command has been deprecated."

os::cmd::expect_success "oc project default/127-0-0-1:28443/system:admin"
os::cmd::expect_success "oc delete ns oshinko"
os::test::junit::declare_suite_end
