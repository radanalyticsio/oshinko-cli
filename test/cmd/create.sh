#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/create"

# No clusters notice
os::cmd::try_until_text "_output/oshinko get" "There are no clusters in any projects. You can create a cluster with the 'create' command."

# name required
os::cmd::expect_failure "_output/oshinko create"

# default one worker / one master
os::cmd::expect_success "_output/oshinko create abc"
os::cmd::expect_success "_output/oshinko get abc -o yaml | grep 'WorkerCount: 1'"
os::cmd::expect_success "_output/oshinko get abc -o yaml | grep 'MasterCount: 1'"
os::cmd::expect_success "_output/oshinko delete abc"

# workers flag
os::cmd::expect_success "_output/oshinko create def --workers=-1"
os::cmd::expect_success "_output/oshinko get def -o yaml | grep 'WorkerCount: 1'"
os::cmd::expect_success "_output/oshinko delete def"

os::cmd::expect_success "_output/oshinko create ghi --workers=2"
os::cmd::expect_success "_output/oshinko get ghi -o yaml | grep 'WorkerCount: 2'"
os::cmd::expect_success "_output/oshinko delete ghi"

os::cmd::expect_success "_output/oshinko create sam --workers=0"
os::cmd::expect_success "_output/oshinko get sam -o yaml | grep 'WorkerCount: 0'"
os::cmd::expect_success "_output/oshinko delete sam"

# masters flag
os::cmd::expect_success "_output/oshinko create jkl --masters=-1"
os::cmd::expect_success "_output/oshinko get jkl -o yaml | grep 'MasterCount: 1'"
os::cmd::expect_success "_output/oshinko delete jkl"

os::cmd::expect_success "_output/oshinko create jill --masters=0"
os::cmd::expect_success "_output/oshinko get jill -o yaml | grep 'MasterCount: 0'"
os::cmd::expect_success "_output/oshinko delete jill"

os::cmd::expect_failure_and_text "_output/oshinko create mno --masters=2" "cluster configuration must have a master count of 0 or 1"

# workerconfig
os::cmd::expect_success "oc create configmap testmap"
os::cmd::expect_failure_and_text "_output/oshinko create mno --workerconfig=jack" "unable to find spark configuration 'jack'"
os::cmd::expect_success "_output/oshinko create mno --workerconfig=testmap"
os::cmd::expect_success "_output/oshinko delete mno"

# masterconfig
os::cmd::expect_failure_and_text "_output/oshinko create mno --masterconfig=jack" "unable to find spark configuration 'jack'"
os::cmd::expect_success "_output/oshinko create pqr --masterconfig=testmap"
os::cmd::expect_success "_output/oshinko delete pqr"

# create against existing cluster
os::cmd::expect_success "_output/oshinko create sally"
os::cmd::expect_failure_and_text "_output/oshinko create sally" "cluster 'sally' already exists"

# create against incomplete clusters
os::cmd::expect_success "oc delete service sally-ui"
os::cmd::expect_failure_and_text "_output/oshinko create sally" "cluster 'sally' already exists \(incomplete\)"

# exposeui
os::cmd::expect_success "_output/oshinko create charlie --exposeui=false" 
os::cmd::expect_success_and_text "_output/oshinko get charlie" "charlie.*<no route>"

# metrics
os::cmd::expect_success "_output/oshinko create klondike --metrics=true"
os::cmd::try_until_success "oc get service klondike-metrics"

os::cmd::expect_success "_output/oshinko create klondike2"
os::cmd::try_until_success "oc get service klondike2-ui"
os::cmd::expect_failure "oc get service klondike2-metrics"

os::cmd::expect_success "_output/oshinko create klondike3 --metrics=false"
os::cmd::try_until_success "oc get service klondike3-ui"
os::cmd::expect_failure "oc get service klondike3-metrics"

os::cmd::expect_failure_and_text "_output/oshinko create klondike4 --metrics=notgonnadoit" "must be a boolean"

# storedconfig
oc create configmap clusterconfig --from-literal=workercount=3 --from-literal=mastercount=0 
os::cmd::expect_failure_and_text "_output/oshinko create chicken --storedconfig=jack" "named config 'jack' does not exist"
os::cmd::expect_success "_output/oshinko create chicken --storedconfig=clusterconfig"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "WorkerCount: 3"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "MasterCount: 0"

os::cmd::expect_success "_output/oshinko create hawk --workers=1 --masters=1 --storedconfig=clusterconfig"
os::cmd::expect_success_and_text "_output/oshinko get hawk -o yaml" "WorkerCount: 1"
os::cmd::expect_success_and_text "_output/oshinko get hawk -o yaml" "MasterCount: 1"

# image
os::cmd::expect_success "_output/oshinko create cordial --image=someotherimage"

# flags for ephemeral not valid
os::cmd::expect_failure_and_text "_output/oshinko create mouse --app=bill" "unknown flag"
os::cmd::expect_failure_and_text "_output/oshinko create mouse -e" "unknown shorthand flag"
os::cmd::expect_failure_and_text "_output/oshinko create mouse --ephemeral=true" "unknown flag"

os::test::junit::declare_suite_end
