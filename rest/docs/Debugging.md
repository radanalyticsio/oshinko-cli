# Advanced cluster/app debugging

Where can things go wrong?

-   [Image building/tagging/pushing](#image-building)
-   [Oshinko-rest](#running-oshinko-rest)
-   [Oshinko-web](#running-oshinko-web)
-   [Running jobs](#running-spark-jobs)


## Image building

In general, each of the sub projects have a README.md file. They contain
additional instructions that may be needed in order to successfully
build each of the projects.

When using ‘make image’ to generate the docker images that will become
the Oshinko services, if you receive the following: **Cannot connect to the
Docker daemon. Is 'docker -d' running on this host?**

Be sure that docker is indeed running on the host. If it is, it may be
configured to require root/sudo access. If that is the case, you may
need to run \`sudo make image\` to get the images to build properly.

When using make to build oshinko-s2i, you receive the message **cannot
find package "github.com/go-openapi/runtime/client"**.

Oshinko-s2i requires golang to be installed and for your GOPATH to be
configured appropriately.

Doing docker push <openshift registry>/<project>/<image> complains about
“**No credentials**” or just stalls after saying that “**The push refers to a repository...**”

Be sure that your are logged in to OpenShift and the docker registry in question:

    $ oc whoami
    $ docker login -u <username> -e <anything that looks like an email address> -p $(oc whoami -t) <ip:port of docker registry>

Try pushing again

## Running oshinko-rest

If oshinko-rest fails to start (or fails with permissions errors in the
logs), you may want to check to be sure that a service account has been
created for your project and that it has the admin role.

Check for service account:

    $ oc get sa
    $ oc describe sa <service account name>

Create service account if it's missing:

    $ oc create sa <service accnount name>

Elevate permissions:

    $ oc policy add-role-to-user admin system:serviceaccount:<project>:<service account name> -n <project>

## Running oshinko-web

If oshinko-web is unable to list or deploy spark clusters, you may want
to check the logs for the oshinko-webui pod.  If you see `ENOTFOUND`,
you first want to be sure that oshinko-rest is up and running.  If it
is up and running, you may have a skydns issue where you are unable
to resolve dns references of services by name.

As a debugging test, you can try the following from a terminal inside
the pod running oshinko-webui:

    $ curl http://<oshinko rest service name>:8080

If that command returns a list of information about the oshinko-rest service,
then oshinko-rest is running and skydns is working correctly.

## Running Spark jobs

After you have a spark cluster running, you may try using `oc rsh` to
directly log in to a worker pod and run a spark job from the command
line.

If you see the following message: 
`Exception in thread "main" java.io.IOException: failure to login`
you may need to set the SPARK_USER environment variable before running your job.
You can try the following to run the sample SparkPi job:

    $ SPARK_USER=anyname spark-submit --master spark://<cluster name>:7077  --class org.apache.spark.examples.SparkPi /opt/spark/examples/jars/spark-examples_2.11-2.0.0-preview.jar 10

****
