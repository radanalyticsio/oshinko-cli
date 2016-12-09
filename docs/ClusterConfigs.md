# Using named cluster configurations

Oshinko uses configmaps to store named cluster configurations.
This document describes how to add or edit named configurations and
how to use them with oshinko-rest.

## Named cluster configuration fields

A named cluster configuration can contain the following fields:

* mastercount -- the number of master nodes, currently limited to 1
* workercount -- the number of worker nodes
* sparkmasterconfig -- the name of a configmap that holds spark configuration files for the spark master
* sparkworkerconfig -- the name of a configmap that holds spark configuration files for the spark workers

A fully populated configmap called `myconfigs` might look like this:

    $ oc export configmap myconfigs
    apiVersion: v1
    data:
      sparkmasterconfig: master-config
      sparkworkerconfig: worker-config
      workercount: "4"
      mastercount: "1"
    kind: ConfigMap
    metadata:
      creationTimestamp: null
      name: myconfigs

A simple way to construct a configmap is to create it initially
empty and then edit it as shown below to add fields (for other methods of
creating configmaps, refer to the OpenShift documentation):

    $ oc create configmap mynewconfig

## Editing a named configuration

To modify a named configuration, simply edit the corresponding
configmap. A simple way to edit a configmap is to use the CLI:

    $ oc edit configmap mynewconfig

This will open an editor showing the contents of the configmap
as yaml. Note, if the configmap was created empty, the `data` section
will be missing and must be added to the yaml file.

Simply add fields to the data section or modify existing fields and
exit the editor.

From the OpenShift console configmaps may be edited
by going to `Resources -> other resources` and selecting `ConfigMap`
as the resource type.

## The default configuration

There is a default configuration named `default` which specifies a cluster
with one spark master and one spark worker. All cluster configurations
start with the values from `default` and then optionally update values.

Note, the `default` configuration itself can be modified in a project by
creating a configmap named `default`.

## Where configuration names may be used

The name of a configuration may be passed as the OSHINKO_NAMED_CONFIG
environment variable when a spark application is launched from the
oshinko templates. If no name is given, the `default` configuration
will be used.

The name of a configuration may also be passed in the json object
used to create or update a cluster through the oshinko-rest endpoint,
for example:

    $ curl -H "Content-Type: application/json" -X POST -d '{"name": "sam", "config": {"name": "mynewconfig"}}' http://oshinko-rest-host:8081/clusters
