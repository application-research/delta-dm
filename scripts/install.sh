#!/bin/bash

echo "Δ  Setting up Delta"
mkdir delta-standalone
cd delta-standalone
DELTA_ENV=delta.env

echo "| cloning repositories"
# git clone https://github.com/application-research/delta.git
# git clone https://github.com/application-research/delta-dm.git
# git clone https://github.com/application-research/delta-nextjs-client.git

git clone git@github.com:application-research/delta.git >/dev/null 2>&1
git clone git@github.com:application-research/delta-dm.git >/dev/null 2>&1
git clone git@github.com:application-research/delta-nextjs-client.git >/dev/null 2>&1

echo "| building apps"
(cd delta && make build >/dev/null)
(cd delta-dm && make build >/dev/null)
(cd delta-nextjs-client && npm install && npm run build >/dev/null)

if [[ -f "$DELTA_ENV" ]]; then
    echo "| loading existing api key from .env"
    source $DELTA_ENV
else 
    echo "| obtaining a new api key"
    API_KEY=$(curl -s --location --request GET 'https://auth.estuary.tech/register-new-token' | jq -r .token)
    echo "| your DELTA API key is: $API_KEY"
    echo "| >> Please save this key for future requests"
    echo "| writing API key to env file"
    echo -e "set -a\nDELTA_AUTH=$API_KEY \nset +a" > $DELTA_ENV
    export DELTA_AUTH=$API_KEY
    source $DELTA_ENV
fi

echo "| starting apps"
nohup ./delta/delta daemon --mode=standalone >/dev/null &
sleep 20
nohup ./delta-dm/delta-dm daemon --delta-auth=$DELTA_AUTH >/dev/null &
sleep 5
cd ./delta-nextjs-client && nohup npm run start >/dev/null &

echo "Δ Delta set up complete!"