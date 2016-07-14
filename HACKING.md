## Running outside of openshift

It's possible to run oshinko-rest locally; this can be a
big help during development and debugging.

To do this make a file to setup the env, for instance:

    $ more oshinko-env
    export GOPATH=/home/oshinko-fork/
    export OSHINKO_CLUSTER_NAMESPACE="spark"
    export OSHINKO_KUBE_CONFIG="/home/user/.kube/config"
    export OSHINKO_CLUSTER_IMAGE="myrepo/openshift-spark"

This is a convenience that sets the GOPATH, tells oshinko-rest what openshift
project it will be running in, gives it a path to a valid kube config file
that it will use for communication with the openshift server, and tells it
what docker image from an accessible repo to use for creating spark clusters.
(These env values are usually handled via the template that launches
oshinko-rest in openshift).

**Note**, you must also log in to the appropriate openshift user
account before running oshinko-rest with this setup:

    $ oc login -u myuser
    $ source oshinko-env

Now you're ready to run:

    $ cd /path/to/oshinko_binary
    $ oshinko-rest

## Debugging with delve

Delve can be found at https://github.com/derekparker/delve with
instructions for building/installing.

Assuming that delve is installed on your local system, and you are
set up to run oshinko-rest locally as described above in
[Running outside of openshift](#running-outside-of-openshift), here is a simple
overview of how to get going with delve to debug oshinko-rest.

Delve can handle vendoring but you must set the flag in the env:

    $ export GO15VENDOREXPERIMENT=1

Set up your env as you would for a local run and navigate to the directory
containing the *main.go* file for oshinko-rest, for example:

    $ oc login -u myuser
    $ source oshinko-env
    $ cd /home/oshinko-rest/src/github.com/redhatanalytics/oshinko-rest/cmd/oshinko-rest-server
    $ ls
    main.go

Simply invoke delve in the directory to start debugging:

    $ dlv debug
    Type 'help' for list of commands.
    (dlv) continue
    2016/07/12 14:45:50 Serving oshinko rest at http://127.0.0.1:38016

To pass arguments to oshinko-rest, the invocation will look something like:

    $ dlv debug -- --port=32344

Breakpoints can be set by file:line or by package.function, as long as it
is unambiguous. For example:

    (dlv) break  /home/oshinko-rest/src/github.com/redhatanalytics/oshinko-rest/handlers/server.go:12
    Breakpoint 1 set at 0x5f3ebc for github.com/redhatanalytics/oshinko-rest/handlers.ServerResponse() /home/oshinko-rest/src/github.com/redhatanalytics/oshinko-rest/handlers/server.go:12

    (dlv) break main.main
    Breakpoint 1 set at 0x4020eb for main.main() ./main.go:17

That's it! Use **help** for commands like (c)ontinue, (s)tep, (n)ext, etc
