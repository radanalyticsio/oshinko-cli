#/bin/bash

oc create configmap masterconfig --from-file=masterconfig

oc create configmap workerconfig --from-file=workerconfig

# This is a convenience so that if the driver application
# is launched from s2i it can be specified in the app launch
oc create configmap driverconfig --from-file=driverconfig

oc create configmap clusterconfig --from-literal=metrics.enable=true \
                                  --from-literal=scorpionstare.enable=true \
                                  --from-literal=sparkmasterconfig=masterconfig \
                                  --from-literal=sparkworkerconfig=workerconfig
