#!/bin/bash
DELTA_ENV=delta.env

echo "Δ Delta update"

cd delta-standalone
source $DELTA_ENV

echo "| pulling latest changes"
(cd delta && git pull && make build >/dev/null 2>&1)
(cd delta-dm && git pull && make build >/dev/null 2>&1)
(cd delta-nextjs-client && git pull && npm install && npm run build >/dev/null 2>&1)

DELTA_PID=$(ps -h -o pid -C delta)
DDM_PID=$(ps -h -o pid -C delta-dm)
UI_PID=$(ps -ef | grep '[d]elta-nextjs-client' | awk '{print $2}')

echo "| killing existing processes"
kill $DELTA_PID
kill $DDM_PID
kill $UI_PID


echo "| starting apps"
nohup ./delta/delta daemon --mode=standalone >/dev/null &
sleep 20
nohup ./delta-dm/delta-dm daemon --delta-auth=$DELTA_AUTH >/dev/null &
sleep 5
cd ./delta-nextjs-client && nohup npm run start >/dev/null &

echo "Δ Delta update complete!"