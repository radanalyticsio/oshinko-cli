#/bin/bash

oc create configmap masterconfig --from-file=masterconfig

oc create configmap workerconfig --from-file=workerconfig

oc create configmap clusterconfig --from-literal=metrics.enable=true \
                                  --from-literal=scorpionstare.enable=true \
                                  --from-literal=sparkmasterconfig=masterconfig \
                                  --from-literal=sparkworkerconfig=workerconfig
