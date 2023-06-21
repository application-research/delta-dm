#!/bin/bash

# USAGE:
# ./edge-import.sh <edge-url> <ddm-url> <ddm-api-key> <dataset-id>
# EXAMPLE:
# ./edge-import.sh dev-hivemapper.edge.estuary.tech 10.32.32.20:1415 ESTXXXARY 1


# Input Arguments
EDGE_URL=$1
DDM_URL=$2
DDM_API_KEY=$3
DATASET_ID=$4

# Make the GET request
GET_RESULT=$(curl -s "https://$EDGE_URL/buckets/get-open")

echo $GET_RESULT

# Parse the result and prepare the POST data
POST_DATA=$(echo $GET_RESULT | jq -c '[.[] | {payload_cid: .payload_cid, commp: .piece_cid, padded_size: .piece_size, size: .size, content_location: ("https://'"$EDGE_URL"'" + .download_url)}]')

# Make the POST request and print the response
POST_RESULT=$(curl -s -X POST "http://$DDM_URL/api/v1/contents/$DATASET_ID" \
  -H "Authorization: Bearer $DDM_API_KEY" \
  -H "Content-Type: application/json" \
  -d "$POST_DATA")

# Print the response
echo $POST_RESULT