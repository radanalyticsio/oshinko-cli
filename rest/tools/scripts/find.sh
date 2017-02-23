#/bin/bash
# Lookup the cluster $1 using the oshinko-rest at $2
curl -H "Content-Type: application/json" -X GET $2/clusters/$1
