#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/delete"

# No clusters notice
os::cmd::try_until_text "_output/oshinko get" "There are no clusters in any projects. You can create a cluster with the 'create' command."

# Create clusters so we can look at them
os::cmd::expect_success "_output/oshinko create abc --workers=2"

# name required
os::cmd::expect_failure "_output/oshinko delete"

# delete happens
os::cmd::expect_success "_output/oshinko create bob"
os::cmd::expect_success "_output/oshinko delete bob"
os::cmd::expect_failure "_output/oshinko get bob"

# ephemeral flags invalid
os::cmd::expect_failure_and_text "_output/oshinko delete bob --app=sam-1" "unknown flag"
os::cmd::expect_failure_and_text "_output/oshinko delete bob --app-status=completed" "unknown flag"

os::test::junit::declare_suite_end
