#!/bin/bash
source $(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/help"

os::cmd::expect_success '_output/oshinko help | grep -e "get\s*Get running spark clusters$"'
os::cmd::expect_success '_output/oshinko help | grep -e "delete\s*Delete spark cluster by name$"'
os::cmd::expect_success '_output/oshinko help | grep -e "create\s*Create new spark cluster$"'
os::cmd::expect_success '_output/oshinko help | grep -e "scale\s*Scale spark cluster by name$"'

# hidden commands
os::cmd::expect_success_and_not_text '_output/oshinko help' 'create_eph'
os::cmd::expect_success_and_not_text '_output/oshinko help' 'delete_eph'
os::cmd::expect_success_and_not_text '_output/oshinko help' 'get_eph'
os::cmd::expect_success_and_not_text '_output/oshinko help' 'configmap'

os::cmd::expect_success_and_text '_output/oshinko help create_eph' 'Create new spark cluster'
os::cmd::expect_success_and_text '_output/oshinko help delete_eph' 'Delete spark cluster by name'
os::cmd::expect_success_and_text '_output/oshinko help get_eph' 'Get running spark clusters'
os::cmd::expect_success_and_text '_output/oshinko help configmap' 'Lookup a configmap by name and print as json if it exists'

source $(dirname "${BASH_SOURCE}")/../../sparkimage.sh
os::cmd::expect_success '_output/oshinko help create | grep Default\ image\ is\ "$SPARK_IMAGE"'

os::test::junit::declare_suite_end
