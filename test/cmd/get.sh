#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/get"

# No clusters notice
os::cmd::try_until_text "_output/oshinko get" "There are no clusters in any projects. You can create a cluster with the 'create' command."

# Create clusters so we can look at them
os::cmd::expect_success "_output/oshinko create abc --workers=2"
os::cmd::expect_success "_output/oshinko create def --workers=1"

# json and yaml output
os::cmd::try_until_text "_output/oshinko get abc -o json" '"WorkerCount": 2'
os::cmd::try_until_text "_output/oshinko get def -o yaml" 'WorkerCount: 1'

# pods vs nopods
os::cmd::try_until_text "_output/oshinko get abc -o json" '"pods"'
os::cmd::expect_success_and_not_text "_output/oshinko get abc -o json --nopods" '"pods"'

# get all
os::cmd::expect_success_and_text "_output/oshinko get" "abc"
os::cmd::expect_success_and_text "_output/oshinko get" "def"
 
# check for columns
os::cmd::expect_success '_output/oshinko-cli get abc | grep -e "^abc\s*[012]\s*spark://abc:7077\s*http://abc-ui:8080\s*abc-ui-route.*\s*Running\s*<shared>$"'

# incomplete clusters
os::cmd::expect_success "oc delete service abc-ui"
os::cmd::expect_success_and_text "_output/oshinko-cli get abc" "Incomplete"
os::cmd::expect_success_and_text "_output/oshinko-cli get abc" "<missing>"

# no such cluster
os::cmd::expect_failure_and_text "_output/oshinko get nothere" "no such cluster 'nothere'"

# verbose
os::cmd::expect_success_and_text "_output/oshinko get -v" "You have access to the following projects"
os::cmd::expect_success_and_text "_output/oshinko get --verbose" "You have access to the following projects"
os::cmd::expect_success_and_not_text "_output/oshinko get" "You have access to the following projects"
os::cmd::expect_success_and_not_text "_output/oshinko get --verbose=false" "You have access to the following projects"

# flags for ephemeral not valid
os::cmd::expect_failure_and_text "_output/oshinko get --app=bill" "unknown flag"

os::test::junit::declare_suite_end
