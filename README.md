
[![Build Status](https://travis-ci.org/radanalyticsio/oshinko-cli.svg?branch=master)](https://travis-ci.org/radanalyticsio/oshinko-cli)

# Oshinko Cli

Command line interface for spark cluster management app

# Usage

You will need to login using `oc` command and then perform the following
commands to get a list of running spark clusters

```bash

make build
./_output/oshinko-cli  --insecure-skip-tls-verify=true --token=$(oc whoami -t) -o json

```
