#!/bin/sh
HOST=${OSHINKO_SERVER_HOST:-0.0.0.0}
PORT=${OSHINKO_SERVER_PORT:-8080}

/go/bin/oshinko-rest-server --host $HOST --port $PORT --scheme http
