# oshinko-client-test

The code in this directory contains an example of how to use the oshinko-rest
client package. This application will simply connect to an instance of
oshinko-rest-server at 127.0.0.1:8080 and retrieve the data from the
server information endpoint.

## How to build

Run the following in this directory:

    $ go build

## How to run

Ensure that the oshinko-rest-server is running and listening at
127.0.0.1:8080. This can be accomplished with the following command:

    $ oshinko-rest-server --host 127.0.0.1 --port 8080

After the server is started, run the following command from this directory:

    $ ./oshinko-client-test

If everything has succeeded you will see something similar to this output:

    name: oshinko-rest-server
    version: 0.0.0

