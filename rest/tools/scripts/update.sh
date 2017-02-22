#/bin/bash
# Update the cluster $2 to have $1 worker nodes using the oshinko-rest at $3
curl -H "Content-Type: application/json" -X PUT -d '{"config": {"masterCount": 1, "workerCount": '$1'}, "name": "'$2'"}' $3/clusters/$2
