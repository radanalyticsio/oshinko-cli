#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/geteph"

# No clusters notice
os::cmd::try_until_text "_output/oshinko get_eph" "There are no clusters in any projects. You can create a cluster with the 'create' command."

# Create clusters so we can look at them
os::cmd::expect_success "_output/oshinko create abc --workers=2"
os::cmd::expect_success "_output/oshinko create def --workers=1"

# json and yaml output
os::cmd::try_until_text "_output/oshinko get_eph abc -o json" '"WorkerCount": 2'
os::cmd::try_until_text "_output/oshinko get_eph abc -o yaml" 'WorkerCount: 2'

# pods vs nopods
os::cmd::try_until_text "_output/oshinko get_eph abc -o json" '"pods"'
os::cmd::expect_success_and_not_text "_output/oshinko get_eph abc -o json --nopods" '"pods"'

# get all
os::cmd::try_until_text "_output/oshinko get_eph" "abc"
os::cmd::try_until_text "_output/oshinko get_eph" "def"
 
# check for columns
os::cmd::expect_success '_output/oshinko-cli get_eph abc | grep -e "^abc\s*[012]\s*spark://abc:7077\s*http://abc-ui:8080\s*abc-ui-route.*\s*Running\s*<shared>$"'

# app flag
os::cmd::expect_failure_and_text "_output/oshinko get_eph --app=bill" "no cluster found for app 'bill'"
os::cmd::expect_success_and_text "_output/oshinko create_eph -e effy --app=abc-m-1" 'ephemeral cluster "effy" created'
os::cmd::expect_success_and_text "_output/oshinko get_eph --app=abc-m-1" "effy.*abc-m-1"



# incomplete clusters
os::cmd::expect_success "oc delete service abc-ui"
os::cmd::expect_success_and_text "_output/oshinko-cli get_eph abc" "Incomplete"
os::cmd::expect_success_and_text "_output/oshinko-cli get_eph abc" "<missing>"

# no such cluster
os::cmd::expect_failure_and_text "_output/oshinko get_eph nothere" "no such cluster 'nothere'"

# verbose
os::cmd::expect_success_and_text "_output/oshinko get_eph -v" "You have access to the following projects"
os::cmd::expect_success_and_not_text "_output/oshinko get_eph" "You have access to the following projects"
os::cmd::expect_success_and_not_text "_output/oshinko get" "You have access to the following projects"
os::cmd::expect_success_and_not_text "_output/oshinko get --verbose=false" "You have access to the following projects"

os::test::junit::declare_suite_end
