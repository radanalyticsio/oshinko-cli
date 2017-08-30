#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/configmap"

os::cmd::expect_success "oc create configmap daniel --from-literal=mykey=myvalue"

# name required
os::cmd::expect_failure "_output/oshinko configmap"
os::cmd::expect_failure_and_text "_output/oshinko configmap nothere" 'configmaps "nothere" not found'

os::cmd::expect_success_and_text "_output/oshinko configmap daniel" '"mykey": "myvalue"'
os::cmd::expect_success_and_text "_output/oshinko configmap daniel -o json" '"mykey": "myvalue"'
os::cmd::expect_success_and_text "_output/oshinko configmap daniel -o yaml" 'mykey: myvalue'

os::test::junit::declare_suite_end
