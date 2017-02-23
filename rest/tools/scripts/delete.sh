#/bin/bash
# Delete the cluster $1 using the oshinko-rest at $2
curl -v -H "Content-Type: application/json" -X DELETE $2/clusters/$1
