#/bin/bash
# List all clusters using the oshinko-rest at $1
curl -H "Content-Type: application/json" -X GET $1/clusters
