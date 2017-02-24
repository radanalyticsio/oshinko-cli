#!/bin/bash

# install some stuff we need for building
rpm -qa | grep -qw git || sudo yum -y install git
rpm -qa | grep -qw golang || sudo yum -y install golang
rpm -qa | grep -qw make || sudo yum -y install make

CURRDIR=`pwd`
export GOPATH=$CURRDIR/oshinko
SRCDIR=$CURRDIR/oshinko/src/github.com/radanalyticsio
mkdir -p $SRCDIR
cd $SRCDIR
if [ ! -d "oshinko-cli" ]; then
    git clone git@github.com:radanalyticsio/oshinko-cli
fi
if [ ! -d "oshinko-s2i" ]; then
    git clone git@github.com:radanalyticsio/oshinko-s2i
fi

cd $SRCDIR/oshinko-cli/rest; make build
cd $SRCDIR/oshinko-cli/; make build
