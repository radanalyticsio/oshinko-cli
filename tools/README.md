# tools for use with oshinko-reset

## server-template.yaml

This is a template that can be processed for adding an oshinko-rest-server
image to OpenShift. It maintains the default 0.0.0.0:8080 host:port option
for the server.

To use this template, an oshinko-rest-server image must first be tagged into
the OpenShift project. Then the template must be processed with the location
for the image. Finally, the output of the processed template can be used
to create the service and pod.

Example usage with an internal registry at 172.30.159.57:5000, and a project
named "myproject":

    $ oc process -f server-template.yaml -v OSHINKO_SERVER_IMAGE=172.30.159.57:5000/myproject/oshinko-rest-server > server-template.json
    $ oc create -f server-template.json
