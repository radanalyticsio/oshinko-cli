#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/flags"

PROJECT=$(oc project -q)

# namespace flag
# note on this first test, running against 'oc cluster up' returns a different code than running with hack/test-cmd.sh (go figure)
# so we ignore the code and just use try_until_text
os::cmd::try_until_text "_output/oshinko get --verbose --namespace=bob" "Using project \"bob\""
os::cmd::expect_success '_output/oshinko get --verbose | grep "Using project" | grep "$PROJECT"'

os::test::junit::declare_suite_end
