#/bin/bash
# Create a cluster named $1 with $2 workers using the oshinko-rest at $3
curl -H "Content-Type: application/json" -X POST -d '{"name": "'$2'", "config": {"sparkMasterConfig": "mysparkconfig", "sparkWorkerConfig": "mysparkconfig", "workerCount": '$1', "masterCount": 1}}' $3/clusters

