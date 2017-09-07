#!/bin/bash
source $(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/version"

os::cmd::expect_success_and_text "_output/oshinko version" "oshinko"

source $(dirname "${BASH_SOURCE}")/../../sparkimage.sh
os::cmd::expect_success '_output/oshinko version | grep Default\ spark\ image:\ "$SPARK_IMAGE"'

os::test::junit::declare_suite_end
