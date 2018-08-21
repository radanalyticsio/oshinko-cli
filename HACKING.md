# Collaborating on the CLI
Here are some instructions to help you build and

## Prerequisites for collaborating
The Oshinko CLI is a go program, therefore you will need the Go language compiler installed this can be found [here](https://golang.org/dl/), or if you are a mac user and have homebrew installed use this command:
``
brew install go --cross-compile-common
``

Fedora:
``
sudo dnf install golang
``

Once you have Go installed you will need to set the $GOPATH to your current Golang path. If you are not familiar Golang then you may not know that you need to set up a go directory and in that will be a folder called ``/src`` this will be where you keep your code. To run the Oshinko cli you must have this go folder as your $GOPATH and then enter your code into the /src folder here your code must have the directory structure: ``/go/src/github.com/radanalyticsio/oshinko-cli - this last folder is the git repository for the project. This file structure is necessary for the code to find the project itself.

## Building Oshinko CLI with MacOS

Some of the scripts used to build the cli use linux specific commands this can be worked around for MacOS
by using the following instructions.

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


