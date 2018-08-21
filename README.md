[![Build Status](https://travis-ci.org/radanalyticsio/oshinko-cli.svg?branch=master)](https://travis-ci.org/radanalyticsio/oshinko-cli)
# Oshinko Cli
Command line interface for spark cluster management application.

## Oshinko Application
The Oshinko application manages Apache Spark clusters on OpenShift.
The application consists of a REST server (Oshinko-rest) and a web UI. It is designed to run in an OpenShift project. This repository contains tools to launch the Oshinko application

## Using the Oshinko CLI
The easiest way to get started using the Oshinko CLI is to use a release version which can be found [here](https://github.com/radanalyticsio/oshinko-cli/releases). There are releases for linux32 or 64 and MacOS.	To use a release either untar/unzip the precompiled Oshinko binary and use this to call commands to create, delete or modify clusters.

``oshinko <command>``

## Creating a cluster
To create an oshinko cluster through the cli you use the command `create`
 ``oshinko create <cluster name>``

## Deleting a cluster
To delete a cluster you use the delete command:
``
oshinko delete <name of cluster>
``
## Scaling a cluster
You may need to scale a cluster out to provide more resources or scale down to save money to do this you can use the command:
``oshinko scale <name of cluster> [options] ``
Options include scaling the number of masters or workers.
## Further reading
For a full set of commands and parameters please use:
``oshinko help``
Please see the CONTRIBUTING.md and HACKING.md docs for more information about working with this codebase and the docs directory for more general information on usage.
