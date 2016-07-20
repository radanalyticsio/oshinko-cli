#!/bin/bash

if [ -z $1 ]
then
    echo "for the moment, pass me an IP address to use in the exposed route for oshinko-web, for example:"
    echo "./oshinko-setup 10.16.40.70"
    exit
fi
WEBROUTEIP=$1

# install some stuff we need for building
rpm -qa | grep -qw git || sudo yum -y install git
rpm -qa | grep -qw golang || sudo yum -y install golang
rpm -qa | grep -qw make || sudo yum -y install make
rpm -qa | grep -qw docker || sudo yum -y install docker
rpm -qa | grep -qw wget || sudo yum -y install wget
rpm -qa | grep -qw tar || sudo yum -y install tar

############ get the oshinko repos and build the images

sudo systemctl start docker 

CURRDIR=`pwd`
export GOPATH=$CURRDIR/oshinko

SRCDIR=$CURRDIR/oshinko/src/github.com/redhatanalytics
mkdir -p $SRCDIR
cd $SRCDIR
if [ ! -d "oshinko-rest" ]; then
    git clone git@github.com:redhatanalytics/oshinko-rest
fi
if [ ! -d "oshinko-webui" ]; then
    git clone git@github.com:redhatanalytics/oshinko-webui
fi
if [ ! -d "openshift-spark" ]; then
    git clone git@github.com:redhatanalytics/openshift-spark
fi
if [ ! -d "oshinko-s2i" ]; then
    git clone git@github.com:redhatanalytics/oshinko-s2i
fi

cd $SRCDIR/oshinko-rest; sudo make image
cd $SRCDIR/oshinko-webui; sudo docker build -t oshinko-webui .
cd $SRCDIR/oshinko-s2i; make build

# this works but it can probably be smarter .. maybe hadoop doesn't
# have to download each time. Maybe we can check for current images? 
cd $SRCDIR/openshift-spark; sudo make build

########### get the origin image and run oc cluster up
########### this part can be replaced with some other openshift install recipe

if [ ! -d "openshift-spark" ]; then
    wget https://github.com/openshift/origin/releases/download/v1.3.0-alpha.2/openshift-origin-server-v1.3.0-alpha.2-983578e-linux-64bit.tar.gz
    tar -xvzf openshift-origin-server-v1.3.0-alpha.2-983578e-linux-64bit.tar.gz
    sudo cp openshift-origin-server-v1.3.0-alpha.2-983578e-linux-64bit/* /usr/bin
fi

sudo sed -i "s/# INSECURE_REGISTRY='--insecure-registry '/INSECURE_REGISTRY='--insecure-registry 172.30.0.0\/16'/" /etc/sysconfig/docker
sudo systemctl restart docker

# make sure your local host name can be resolved!
# put it in /etc/hosts if you have to, otherwise you will have no nodes
sudo oc cluster up

############

# Get the address of the docker registry so we can push our images to it
sudo oc login -u system:admin
sudo oc project default
REGISTRY=$(sudo oc get service docker-registry --template='{{index .spec.clusterIP}}:{{index .spec.ports 0 "port"}}')
ROUTERIP=$(sudo oc get service router --template='{{index .spec.clusterIP}}')

# Push to a default oshinko project for a default oshinko user
oc login -u oshinko -p oshinko
oc new-project oshinko

# Wait for the registry to be fully up
r=1
while [ $r -ne 0 ]; do
    sudo docker login -u oshinko -e "jack@jack.com" -p $(oc whoami -t) $REGISTRY
    r=$?
    sleep 1
done

sudo docker tag oshinko-rest-server $REGISTRY/oshinko/oshinko-rest-server
sudo docker push $REGISTRY/oshinko/oshinko-rest-server
sudo docker tag oshinko-webui $REGISTRY/oshinko/oshinko-webui
sudo docker push $REGISTRY/oshinko/oshinko-webui
sudo docker tag openshift-spark $REGISTRY/oshinko/openshift-spark
sudo docker push $REGISTRY/oshinko/openshift-spark
sudo docker tag daikon-pyspark $REGISTRY/oshinko/daikon-pyspark
sudo docker push $REGISTRY/oshinko/daikon-pyspark

# set up the oshinko service account
oc create sa oshinko                          # note, VV, first oshinko is the proj name :)
oc policy add-role-to-user admin system:serviceaccount:oshinko:oshinko -n oshinko

# process the standard oshinko template and launch it
cd $SRCDIR/oshinko-rest
oc process -f tools/server-ui-template.yaml \
OSHINKO_SERVER_IMAGE=$REGISTRY/oshinko/oshinko-rest-server \
OSHINKO_CLUSTER_IMAGE=$REGISTRY/oshinko/openshift-spark \
OSHINKO_WEB_IMAGE=$REGISTRY/oshinko/oshinko-webui \
OSHINKO_WEB_EXTERNAL_IP=mywebui.$WEBROUTEIP.xip.io > $CURRDIR/oshinko-template.json
oc create -f $CURRDIR/oshinko-template.json

# Add the s2I template
oc create -f $SRCDIR/oshinko-s2i/pyspark/pyspark.json
