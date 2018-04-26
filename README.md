[![Build Status](https://travis-ci.org/radanalyticsio/oshinko-cli.svg?branch=master)](https://travis-ci.org/radanalyticsio/oshinko-cli)

# Oshinko Cli
Command line interface for spark cluster management app

## Usage

You will need to login using the `oc` command and then perform the following
commands to get a list of running spark clusters

```bash

make build
./_output/oshinko-cli get --insecure-skip-tls-verify=true --token=$(oc whoami -t) -o json

```

# Oshinko Application

The oshinko application manages Apache Spark clusters on OpenShift.
The application consists of a REST server (oshinko-rest) and a web UI
and is designed to run in an OpenShift project.

This repository contains tools to launch the oshinko application
along with the source code for the oshinko REST server in the `rest`
subdirectory. The source code for the web UI is located in a different
repository.

## Deploying the oshinko application in the current project

For the most complete usage of oshinko-rest, we recommend installing the
entire oshinko suite using the `tools/oshinko-deploy.sh` script. It will pull the
latest upstream images from the radanalyticsio organization.

First log into an OpenShift installation as your user, then use this command
to deploy the oshinko application into the current project:

    $ ./tools/oshinko-deploy.sh -u $(oc whoami) -p $(oc project --short)

For more information on what you can do with the `oshinko-deploy.sh` script,
see the rest/HACKING doc.

## Building and running the oshinko-rest binary

To build the `oshinko-rest` binary simply run the `build` or `install` target
in the makefile in the `rest` subdirectory

**Example**

    $ cd rest
    $ make build

Assuming a successful build, the output will be stored in the `_output`
directory. For an `install` target, the binary will be placed in your
`$GOPATH/bin`.

## Building the oshinko-cli on MacOS

First of all you need to install coreutils - this can be done with brew:

```
    brew install coreutils
```

Once you have installed this make sure you have installed go and set your
$GOPATH to the root of your project directory.

After this you then run:

```
    make build
```
There is no need to set $GOROOT to run this project.

## Running oshinko-rest

For most functionality an OpenShift cluster will be needed, but the
application can be tested for basic operation without one.

**Example**

After building the binary a basic test can be performed as follows:

* start the server in a terminal

```
    $ _output/oshinko-rest-server --port 42000 --scheme http
    2016/07/14 16:41:00 Serving oshinko rest at http://127.0.0.1:42000
```

* in a second terminal run a small curl command against the server

```
    $ curl http://localhost:42000/
    {"application":{"name":"oshinko-rest-server","version":"0.1.0"}}
```

*The return value may be different depending on the version of the
server you have built*

### TLS

To start the server with TLS enabled, you will need to supply a certificate
file and a key file for the server to use. Once these files are created you
can start the server as follows to enable HTTPS access:

```
    $ _output/oshinko-rest-server --port 42000 --tls-port 42443 --tls-key keyfile.key --tls-certificate certificatefile.cert
    2016/09/28 12:10:47 Serving oshinko rest at http://127.0.0.1:42000
    2016/09/28 12:10:47 Serving oshinko rest at https://127.0.0.1:42443
```

At this point the server is ready to accept both HTTP and HTTPS requests. If
you would like to restrict access to **only** use TLS, add the
`--scheme https` flag to the command line as follows:

```
    $ _output/oshinko-rest-server --scheme https --tls-port 42443 --tls-key keyfile.key --tls-certificate certificatefile.cert
    2016/09/28 12:10:47 Serving oshinko rest at https://127.0.0.1:42443
```

## Further reading

Please see the rest/CONTRIBUTING and rest/HACKING docs for more information about
working with this codebase and the docs directory for more general information on usage.
