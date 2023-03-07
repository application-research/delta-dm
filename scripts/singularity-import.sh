#!/bin/bash

 query = "{\"datasetName\": \"$1\"}"


mongoexport mongodb://localhost:27017 \
	--db=singularity \
	--collection=generationrequests \
	--fields="dataCid,pieceCid,carSize,pieceSize" \
	--query=$query \
	--out="singularity-out.json" 

