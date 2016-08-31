#!/bin/bash

# This script is for deploying oshinko into an already running cluster.
# It assumes a few things:
# * you have the following images in your local docker registry:
#   * oshinko-rest-server
#   * oshinko-webui
#   * openshift-spark
#   * radanalytics-pyspark
# * you have a file named "server-ui-template.yaml" in the current directory
#
# Usage:
# $ oshinko-deploy.sh {route IP} {project name}
#
# route IP -- address to use in the exposed route information
# project name -- project to deploy oshinko into


while getopts :s:w:p:u:h opt; do
    case $opt in
        s)
            SPARK_IMAGE=$OPTARG
            ;;
        w)
            WEBROUTE=$OPTARG
            ;;
        p)
            PROJECT=$OPTARG
            ;;
        u)
            OS_USER=$OPTARG
            ;;
        h)
            echo "usage: oshinko-deploy.sh [-w HOSTNAME] [-s IMAGE] [-p PROJECT] [-u USER]"
            echo
            echo "deploy the oshinko suite into a running OpenShift cluster"
            echo
            echo "optional arguments:"
            echo "  -h            show this help message"
            echo "  -w HOSTNAME   hostname to use in exposed route to oshinko-web"
            echo "  -s IMAGE      spark docker image to use for clusters (default: openshift-spark)"
            echo "  -p PROJECT    OpenShift project name to install oshinko into (default: myproject)"
            echo "  -u USER       OpenShift user to run commands as (default: developer)"
            echo
            echo "  If -w is not set, the default route will be used based on routing suffix, etc set at installation"
            exit
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            exit
            ;;
    esac
done

if [ -z "$PROJECT" ]
then
    echo "project name not set, using default value"
    PROJECT=myproject
fi

if [ -z "$OS_USER" ]
then
    echo "user not set, using default value"
    OS_USER=developer
fi

oc login -u system:admin
oc project default
REGISTRY=$(oc get service docker-registry --no-headers=true | awk -F ' ' '{print $2":"$4}' | sed "s,/TCP$,,")

# reset back to the default development account
oc login -u $OS_USER
oc project $PROJECT

# Wait for the registry to be fully up
r=1
while [ $r -ne 0 ]; do
    docker login -u $(oc whoami) -e "jack@jack.com" -p $(oc whoami -t) $REGISTRY
    r=$?
    sleep 1
done

docker tag oshinko-rest-server $REGISTRY/$PROJECT/oshinko-rest-server
docker push $REGISTRY/$PROJECT/oshinko-rest-server
docker tag oshinko-webui $REGISTRY/$PROJECT/oshinko-webui
docker push $REGISTRY/$PROJECT/oshinko-webui
docker tag radanalytics-pyspark $REGISTRY/$PROJECT/radanalytics-pyspark
docker push $REGISTRY/$PROJECT/radanalytics-pyspark

if [ -z "$SPARK_IMAGE" ]; then
    docker tag openshift-spark $REGISTRY/$PROJECT/oshinko-spark
else
    docker tag $SPARK_IMAGE $REGISTRY/$PROJECT/oshinko-spark
fi
docker push $REGISTRY/$PROJECT/oshinko-spark

# set up the oshinko service account
oc create sa oshinko -n $PROJECT
oc policy add-role-to-user admin system:serviceaccount:$PROJECT:oshinko -n $PROJECT

# process the standard oshinko template and launch it
if [ -n "$WEBROUTE" ] ; then
    ROUTEVALUE=$WEBROUTE
fi

# process the standard oshinko template and launch it
oc process -f server-ui-template.yaml \
OSHINKO_SERVER_IMAGE=$REGISTRY/$PROJECT/oshinko-rest-server \
OSHINKO_CLUSTER_IMAGE=$REGISTRY/$PROJECT/oshinko-spark \
OSHINKO_WEB_IMAGE=$REGISTRY/$PROJECT/oshinko-webui \
OSHINKO_WEB_ROUTE_HOSTNAME=$ROUTEVALUE \
> oshinko-deploy-processed.json

oc create -f oshinko-deploy-processed.json -n $PROJECT
