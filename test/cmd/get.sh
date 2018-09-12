#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/get"

# No clusters notice
os::cmd::try_until_text "_output/oshinko get" "No clusters found."
os::cmd::try_until_text "_output/oshinko get -o json" "\[\]"
os::cmd::try_until_text "_output/oshinko get -o yaml" "\[\]"



# Create clusters so we can look at them
os::cmd::expect_success "_output/oshinko create abc --workers=2"
os::cmd::expect_success "_output/oshinko create def --workers=1"

# json and yaml output
os::cmd::try_until_text "_output/oshinko get abc -o json" '"WorkersCount": 2'
os::cmd::try_until_text "_output/oshinko get def -o yaml" 'WorkersCount: 1'

# pods vs nopods
os::cmd::try_until_text "_output/oshinko get abc -o json" '"Pods"'
os::cmd::expect_success_and_not_text "_output/oshinko get abc -o json --nopods" '"Pods"'

# get all
os::cmd::expect_success_and_text "_output/oshinko get" "abc"
os::cmd::expect_success_and_text "_output/oshinko get" "def"

# check for columns
os::cmd::expect_success '_output/oshinko-cli get -d abc | grep -e "^abc\s*[012]\s*spark://abc:7077\s*http://abc-ui:8080\s*http://abc-ui-route.*\s*Running\s*<shared>$"'
# incomplete clusters	# incomplete clusters
os::cmd::expect_success "oc delete service abc-ui"
os::cmd::expect_success_and_text "_output/oshinko-cli get -d abc" "Incomplete"
os::cmd::expect_success_and_text "_output/oshinko-cli get -d abc" "<missing>"

# no such cluster
os::cmd::expect_failure_and_text "_output/oshinko get nothere" "no such cluster 'nothere'"

# check for no cluster but return json/yaml
os::cmd::expect_success_and_text "_output/oshinko-cli get nemo -o json" "{}"
os::cmd::expect_success_and_text "_output/oshinko-cli get nemo -o yaml" "{}"



# flags for ephemeral not valid
os::cmd::expect_failure_and_text "_output/oshinko get --app=bill" "unknown flag"

os::test::junit::declare_suite_end
