#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/get"
# This test validates the help commands and output text

# verify some default commands
os::cmd::expect_failure_and_text "_output/oshinko-cli get" "The token is not provided."
os::test::junit::declare_suite_end
