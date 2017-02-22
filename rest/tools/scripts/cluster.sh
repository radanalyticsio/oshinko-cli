#/bin/bash
# Create a cluster named $2 with $1 workers using the oshinko-rest at $3
curl -H "Content-Type: application/json" -X POST -d '{"name": "'$2'", "config": {"workerCount": '$1', "masterCount": 1}}' $3/clusters

