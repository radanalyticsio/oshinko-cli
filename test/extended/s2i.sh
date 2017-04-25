#!/bin/bash
#set -o xtrace
#
# Runs all standard extended tests against either an existing cluster (TEST_ONLY=1)
# or a standard started server.
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
source "${OS_ROOT}/test/extended/setup.sh"
os::test::extended::setup


trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "extended/s2i"

os::cmd::expect_success "oc whoami"

os::test::junit::declare_suite_end
