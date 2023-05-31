#!/bin/bash

# Usage: ./singularity-import.sh <SINGULARITY datasetName> <DELTA datasetID> <deltaToken>
query="{\"datasetName\": \"$1\", \"pieceSize\": { \"\$gt\": 0 }}"
DELTA_TOKEN="Authorization: Bearer $3"

mongoexport --uri='mongodb://localhost:27017' \
	--db=singularity \
	--collection=generationrequests \
	--fields="dataCid,pieceCid,carSize,pieceSize" \
	--jsonArray \
	--query="$query" \
	--out="singularity-out.json" \

echo "Importing dataset to DDM. Please wait..."

cat singularity-out.json | jq . |
res="$(curl -X POST -d @- \
  "http://127.0.0.1:1314/api/v1/contents/$2?import_type=singularity" \
  -H "$DELTA_TOKEN" \
  -H "Content-Type: application/json" \
  2>/dev/null )"

rm singularity-out.json

echo "Done importing CIDs to dataset"
