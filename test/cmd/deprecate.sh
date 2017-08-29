#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/get"

# deprecation notice
os::cmd::expect_success_and_text "_output/oshinko-cli get" "The 'oshinko-cli' command has been deprecated."
os::cmd::expect_success_and_not_text "_output/oshinko get" "The 'oshinko-cli' command has been deprecated."

os::cmd::expect_success_and_text "_output/oshinko-cli create bob" "The 'oshinko-cli' command has been deprecated."
os::cmd::expect_success_and_not_text "_output/oshinko create bill" "The 'oshinko-cli' command has been deprecated."

os::cmd::expect_success_and_text "_output/oshinko-cli scale bob" "The 'oshinko-cli' command has been deprecated."
os::cmd::expect_success_and_not_text "_output/oshinko scale bob" "The 'oshinko-cli' command has been deprecated."

os::cmd::expect_success_and_text "_output/oshinko-cli delete bob" "The 'oshinko-cli' command has been deprecated."
os::cmd::expect_success_and_not_text "_output/oshinko delete bill" "The 'oshinko-cli' command has been deprecated."

os::test::junit::declare_suite_end
