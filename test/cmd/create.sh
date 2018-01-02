#!/bin/bash
source "$(dirname "${BASH_SOURCE}")/../../hack/lib/init.sh"
trap os::test::junit::reconcile_output EXIT

os::test::junit::declare_suite_start "cmd/create"

# No clusters notice
os::cmd::try_until_text "_output/oshinko get" "There are no clusters in any projects. You can create a cluster with the 'create' command."

# name required
os::cmd::expect_failure "_output/oshinko create"

# General note -- at present, the master and worker counts in the included config object on get are "MasterCount" and "WorkerCount"
# the master and worker counts in the outer cluster status are "masterCount" and "workerCount"
# Likewise, SparkImage is from config and 'image' is in the outer cluster status

# default one worker / one master
os::cmd::expect_success "_output/oshinko create abc"
os::cmd::expect_success "_output/oshinko get abc -o yaml | grep 'WorkerCount: 1'"
os::cmd::expect_success "_output/oshinko get abc -o yaml | grep 'MasterCount: 1'"
# could still be creating so use 'until'
os::cmd::try_until_text "_output/oshinko get abc -o yaml" "workerCount: 1"
os::cmd::try_until_text "_output/oshinko get abc -o yaml" "masterCount: 1"
os::cmd::expect_success "_output/oshinko delete abc"

# workers flag
os::cmd::expect_success "_output/oshinko create def --workers=-1"
os::cmd::expect_success "_output/oshinko get def -o yaml | grep 'WorkerCount: 1'"
os::cmd::try_until_text "_output/oshinko get def -o yaml" "workerCount: 1"
os::cmd::expect_success "_output/oshinko delete def"

os::cmd::expect_success "_output/oshinko create ghi --workers=2"
os::cmd::expect_success "_output/oshinko get ghi -o yaml | grep 'WorkerCount: 2'"
os::cmd::try_until_text "_output/oshinko get ghi -o yaml" "workerCount: 2"
os::cmd::expect_success "_output/oshinko delete ghi"

os::cmd::expect_success "_output/oshinko create sam --workers=0"
os::cmd::expect_success "_output/oshinko get sam -o yaml | grep 'WorkerCount: 0'"
os::cmd::try_until_text "_output/oshinko get sam -o yaml" "workerCount: 0"
os::cmd::expect_success "_output/oshinko delete sam"

# masters flag
os::cmd::expect_success "_output/oshinko create jkl --masters=-1"
os::cmd::expect_success "_output/oshinko get jkl -o yaml | grep 'MasterCount: 1'"
os::cmd::try_until_text "_output/oshinko get jkl -o yaml" "masterCount: 1"
os::cmd::expect_success "_output/oshinko delete jkl"

os::cmd::expect_success "_output/oshinko create jill --masters=0"
os::cmd::expect_success "_output/oshinko get jill -o yaml | grep 'MasterCount: 0'"
os::cmd::try_until_text "_output/oshinko get jill -o yaml" "masterCount: 0"
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
os::cmd::expect_success "_output/oshinko delete sally"

# metrics
os::cmd::expect_success "_output/oshinko create klondike --metrics=true"
os::cmd::try_until_success "oc get service klondike-metrics"
os::cmd::try_until_text "oc log dc/klondike-m" "with jolokia metrics"
os::cmd::expect_success "_output/oshinko delete klondike"

os::cmd::expect_success "_output/oshinko create klondike0 --metrics=jolokia"
os::cmd::try_until_success "oc get service klondike0-metrics"
os::cmd::try_until_text "oc log dc/klondike0-m" "with jolokia metrics"
os::cmd::expect_success "_output/oshinko delete klondike0"

os::cmd::expect_success "_output/oshinko create klondike1 --metrics=prometheus"
os::cmd::try_until_success "oc get service klondike1-metrics"
os::cmd::try_until_text "oc log dc/klondike1-m" "with prometheus metrics"
os::cmd::expect_success "_output/oshinko delete klondike1"

os::cmd::expect_success "_output/oshinko create klondike2"
os::cmd::try_until_success "oc get service klondike2-ui"
os::cmd::expect_failure "oc get service klondike2-metrics"
os::cmd::expect_success "_output/oshinko delete klondike2"

os::cmd::expect_success "_output/oshinko create klondike3 --metrics=false"
os::cmd::try_until_success "oc get service klondike3-ui"
os::cmd::expect_failure "oc get service klondike3-metrics"
os::cmd::expect_success "_output/oshinko delete klondike3"

os::cmd::expect_failure_and_text "_output/oshinko create klondike4 --metrics=notgonnadoit" "must be 'true', 'false', 'jolokia', or 'prometheus'"

# exposeui
os::cmd::expect_success "_output/oshinko create charlie --exposeui=false"
os::cmd::expect_success_and_text "_output/oshinko get charlie" "charlie.*<no route>"
os::cmd::expect_success "_output/oshinko delete charlie"

os::cmd::expect_success "_output/oshinko create charlie2 --exposeui=true"
os::cmd::expect_success_and_text "_output/oshinko get charlie2" "charlie2-ui-route"
os::cmd::expect_success "_output/oshinko delete charlie2"

os::cmd::expect_success "_output/oshinko create charlie3"
os::cmd::expect_success_and_text "_output/oshinko get charlie3" "charlie3-ui-route"
os::cmd::expect_success "_output/oshinko delete charlie3"

os::cmd::expect_failure_and_text "_output/oshinko create charlie4 --exposeui=notgonnadoit" "must be a boolean"

# storedconfig
oc create configmap masterconfig
oc create configmap workerconfig
oc create configmap clusterconfig \
--from-literal=workercount=3 \
--from-literal=mastercount=0 \
--from-literal=sparkmasterconfig=masterconfig \
--from-literal=sparkworkerconfig=workerconfig \
--from-literal=exposeui=false \
--from-literal=metrics=true \
--from-literal=sparkimage=myimage
os::cmd::expect_failure_and_text "_output/oshinko create chicken --storedconfig=jack" "named config 'jack' does not exist"
os::cmd::expect_success "_output/oshinko create chicken --storedconfig=clusterconfig"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "WorkerCount: 3"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "MasterCount: 0"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "ExposeWebUI: \"false\""
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "Metrics: \"true\""
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "Name: clusterconfig"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "SparkImage: myimage"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "image: myimage"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "SparkMasterConfig: masterconfig"
os::cmd::expect_success_and_text "_output/oshinko get chicken -o yaml" "SparkWorkerConfig: workerconfig"
os::cmd::try_until_text "_output/oshinko get chicken -o yaml" "workerCount: 3"
os::cmd::try_until_text "_output/oshinko get chicken -o yaml" "masterCount: 0"
os::cmd::expect_success "_output/oshinko delete chicken"

oc create configmap clusterconfig2 --from-literal=metrics=jolokia
os::cmd::expect_success "_output/oshinko create chicken2 --storedconfig=clusterconfig2"
os::cmd::expect_success_and_text "_output/oshinko get chicken2 -o yaml" "Metrics: jolokia"
os::cmd::expect_success "_output/oshinko delete chicken2"

oc create configmap clusterconfig3 --from-literal=metrics=prometheus
os::cmd::expect_success "_output/oshinko create chicken3 --storedconfig=clusterconfig3"
os::cmd::expect_success_and_text "_output/oshinko get chicken3 -o yaml" "Metrics: prometheus"
os::cmd::expect_success "_output/oshinko delete chicken3"

os::cmd::expect_success "_output/oshinko create egg"
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "WorkerCount: 1"
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "MasterCount: 1"
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "ExposeWebUI: \"true\""
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "Metrics: \"false\""
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "Name: \"\""
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "SparkImage: radanalyticsio/openshift-spark"
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "image: radanalyticsio/openshift-spark"
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "SparkMasterConfig: \"\""
os::cmd::expect_success_and_text "_output/oshinko get egg -o yaml" "SparkWorkerConfig: \"\""
os::cmd::try_until_text "_output/oshinko get egg -o yaml" "workerCount: 1"
os::cmd::try_until_text "_output/oshinko get egg -o yaml" "masterCount: 1"
os::cmd::expect_success "_output/oshinko delete egg"

os::cmd::expect_success "_output/oshinko create hawk --workers=1 --masters=1 --storedconfig=clusterconfig"
os::cmd::expect_success_and_text "_output/oshinko get hawk -o yaml" "WorkerCount: 1"
os::cmd::expect_success_and_text "_output/oshinko get hawk -o yaml" "MasterCount: 1"
os::cmd::try_until_text "_output/oshinko get hawk -o yaml" "workerCount: 1"
os::cmd::try_until_text "_output/oshinko get hawk -o yaml" "masterCount: 1"
os::cmd::expect_success "_output/oshinko delete hawk"

# image
os::cmd::expect_success "_output/oshinko create cordial --image=someotherimage"
os::cmd::expect_success_and_text "_output/oshinko get cordial -o yaml" "SparkImage: someotherimage"
os::cmd::expect_success_and_text "_output/oshinko get cordial -o yaml" "image: someotherimage"
os::cmd::expect_success "_output/oshinko delete cordial"

# flags for ephemeral not valid
os::cmd::expect_failure_and_text "_output/oshinko create mouse --app=bill" "unknown flag"
os::cmd::expect_failure_and_text "_output/oshinko create mouse -e" "unknown shorthand flag"
os::cmd::expect_failure_and_text "_output/oshinko create mouse --ephemeral=true" "unknown flag"

os::test::junit::declare_suite_end
