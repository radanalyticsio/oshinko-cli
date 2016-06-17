# tools for use with oshinko-reset

## server-template.yaml

This is a template that can be processed for adding an oshinko-rest-server
image to OpenShift. It maintains the default 0.0.0.0:8080 host:port option
for the server.

To use this template, an oshinko-rest-server image and a spark image must
first be tagged into the OpenShift project. Then the template must be
processed with the locations of the images. Finally, the output of the
processed template can be used to create the service and pod.

## Service account

The oshinko-rest-server uses an "oshinko" service account to perform openshift
operations. This service account must be created and given the admin role in
*each* project which will use the oshinko-rest-server, for example:

Create a new project *myproject*:

    $ oc new-project myproject

Create the oshinko service account:

    $ more sa.json
    {
      "apiVersion": "v1",
      "kind": "ServiceAccount",
      "metadata": {
        "name": "oshinko"
      }
    }

    $ oc create -f sa.json

Assign the admin role:

    $ oc policy add-role-to-user admin system:serviceaccount:myproject:oshinko -n myproject

## Sample template usage

Example usage with an internal registry at 172.30.159.57:5000, and a project
named "myproject":

    $ oc process -f server-template.yaml -v OSHINKO_SERVER_IMAGE=172.30.159.57:5000/myproject/oshinko-rest-server, OSHINKO_CLUSTER_IMAGE=172.30.159.57:5000/myproject/openshift-spark > server-template.json
    $ oc create -f server-template.json
