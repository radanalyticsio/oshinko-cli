#/bin/bash

oc create configmap metricsconfig --from-file=metrics.properties

oc create configmap clusterconfig --from-literal=metrics.enable=true \
                                  --from-literal=scorpionstare.enable=true \
                                  --from-literal=sparkmasterconfig=metricsconfig
