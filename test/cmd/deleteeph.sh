#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
TEMPLATE="$(dirname "${BASH_SOURCE}")/../test.json"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/deleteeph"

# No clusters notice
os::cmd::try_until_text "_output/oshinko get" "No clusters found."

# Create clusters so we can look at them
os::cmd::expect_success "_output/oshinko create abc --workers=2"

# name required
os::cmd::expect_failure "_output/oshinko delete_eph"

# delete happens
os::cmd::expect_success "_output/oshinko create bob"
os::cmd::expect_success "_output/oshinko delete_eph bob"
os::cmd::expect_failure "_output/oshinko get bob"

# ephemeral flags
os::cmd::expect_failure_and_text "_output/oshinko delete_eph bob --app=sam-1" "Both --app and --app-status must be set"
os::cmd::expect_failure_and_text "_output/oshinko delete_eph bob --app-status=completed" "Both --app and --app-status must be set"
os::cmd::expect_failure_and_text "_output/oshinko delete_eph bob --app=sam-1 --app-status=wrong" "INVALID app-status value, only completed|terminated allowed"

oc create -f $TEMPLATE
oc new-app test-loop
os::cmd::try_until_success "oc get rc test-loop-dc-1"
os::cmd::expect_success_and_text "_output/oshinko create_eph -e cluster --app=test-loop-dc-1" 'ephemeral cluster cluster'

# replica count won't work for hack/test-cmd, so only do this test when we're started from run.sh
if [ "${USING_OPENSHIFT_INSTANCE:-false}" == true ]; then
    os::cmd::try_until_text 'oc get pod -l deploymentconfig=cluster-m --template="{{index .items 0 \"status\" \"phase\"}}"' "Running" $((5*minute))
    os::cmd::try_until_text 'oc get pod -l deploymentconfig=cluster-w --template="{{index .items 0 \"status\" \"phase\"}}"' "Running" $((5*minute))
    os::cmd::expect_failure_and_text "_output/oshinko delete_eph cluster --app=test-loop-dc-1 --app-status=terminated" "driver replica count > 0 \(or > 1 for completed app\)"
fi
os::cmd::expect_failure_and_text "_output/oshinko delete_eph cluster --app=someother --app-status=terminated" "cluster is not linked to app"

os::cmd::expect_success "_output/oshinko delete_eph cluster --app=test-loop-dc-1 --app-status=completed"

os::test::junit::declare_suite_end
