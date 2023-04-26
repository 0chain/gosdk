#!/bin/bash

if [ -z "$RUNNER_NUMBER" ] 
then
  echo "WASM: SKIP DUE TO NOT SYSTEM-TESTS"
  exit 0
fi

echo "======================================================"
echo "STARTING WASM DEMO SERVER:"
echo "======================================================"

echo 
echo "> 1.build zcn.wasm"
export LIBVA_DRIVER_NAME=iHD
cd ..
CGO_ENABLED=0 GOOS=js GOARCH=wasm go build -o ./demo/zcn.wasm  .

echo

echo "> 2.build & start demo server"
HTTPCODE=$(curl -s -o /dev/null -w "%{http_code}"  http://127.0.0.1:8080)
if test $HTTPCODE -eq 200; then
    echo "WASM: shutdown staled demo server"
    curl --silent http://127.0.0.1:8080/shutdown
fi

cd ./demo && go build -o demo .
./demo &
sleep 3
cd ../

HTTPCODE=$(curl -s -o /dev/null -w "%{http_code}"  http://127.0.0.1:8080)
if test $HTTPCODE -eq 200; then
  echo "WASM DEMO SERVER IS RUNNING"
else
  echo "!!! WASM DEMO SERVER IS DOWN !!!"
  exit 1
fi

echo 
echo "> 3.cypress open"
cd ./cypress && CYPRESS_NETWORK_URL=$NETWORK_URL cypress open