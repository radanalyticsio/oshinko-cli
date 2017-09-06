#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/version"

os::cmd::expect_success_and_text "_output/oshinko version" "oshinko"

SPARK_IMAGE=$(grep -m 1 SPARK_IMAGE= $(dirname "${BASH_SOURCE}")/../../scripts/build.sh | cut -d '=' -f 2 | sed 's/"//g')
os::cmd::expect_success '_output/oshinko version | grep Default\ spark\ image:\ "$SPARK_IMAGE"'

os::test::junit::declare_suite_end
