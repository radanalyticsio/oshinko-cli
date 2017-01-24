#!/bin/bash

./oshinko-cli create $1 --storedconfig=clusterconfig --insecure-skip-tls-verify=true --token=$(oc whoami -t)

