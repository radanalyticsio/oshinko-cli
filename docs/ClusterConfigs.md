# Using named cluster configurations

Oshinko uses a ConfigMap to store named cluster configurations.
This document describes how to add or edit named configurations and
how to use them with oshinko-rest.

## The oshinko-cluster-configs ConfigMap

The `tools/server-ui-template.yaml` creates a ConfigMap
named `oshinko-cluster-configs` which is read by oshinko-rest.
Any named cluster configuration defined in the ConfigMap
can be used to create or scale a cluster.

The default ConfigMap contains a single configuration named `small`.
The `small` configuration specifies a cluster that has three worker nodes.
To see what configurations are defined, use `oc export` in your project
after launching oshinko:

    $ oc export configmap oshinko-cluster-configs

    apiVersion: v1
    data:
      small.workercount: "3"
    kind: ConfigMap
    metadata:
      creationTimestamp: null
      labels:
        app: oshinko
      name: oshinko-cluster-configs

Named configurations are defined in the data section of the
ConfigMap. Currently `workercount` is the only parameter
which may be set for a configuration (`mastercount` may actually
be set but is constrained to a value of "1"). A parameter is set
using the name of the configuration followed by a dot and the name
of the parameter.

To add a configuration called `large` with a `workercount` of
ten, the ConfigMap would be modified to look like this:

    apiVersion: v1
    data:
      small.workercount: "3"
      large.workercount: "10"
    kind: ConfigMap
    metadata:
      creationTimestamp: null
      labels:
        app: oshinko
      name: oshinko-cluster-configs

## Editing oshinko-cluster-configs

The simplest way to edit oshinko-cluster-configs for a particular
project is to use the CLI:

    $ oc edit configmap oshinko-cluster-configs

From the OpenShift console oshinko-cluster-configs may be edited
by going to "Resources -> other resources" and selecting ConfigMap
as the resource type.

There may be a short delay before configuration changes are visible
to the oshinko-rest pod.

## The default configuration

There is a default configuration named `default` which specifies a cluster
with one spark master and one spark worker. All cluster configurations
start with the values from `default` and then optionally update values. So
the `small` configuration shown above inherits values from `default` and
then modifies its own `workercount` to be three.

Note, the `default` configuration itself can be modified in a project by
editing `oshinko-cluster-configs` and adding a definition for `default`.

## Where configuration names may be used

The name of a configuration may be passed as the OSHINKO_NAMED_CONFIG
environment variable when a spark application is launched from the
oshinko templates. If no name is given, the `default` configuration
will be used.

The name of a configuration may also be passed in the json object
used to create or update a cluster through the oshinko-rest endpoint,
for example:

    $ curl -H "Content-Type: application/json" -X POST -d '{"name": "sam", "config": {"name": "small"}}' http://oshinko-rest-host:8081/clusters
