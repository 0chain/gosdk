#!/bin/bash

if [ -z "$RUNNER_NUMBER" ] 
then
  echo "WASM: SKIP DUE TO NOT SYSTEM-TESTS"
  exit 0
fi

echo "======================================================"
echo "STARTING WASM DEMO SERVER:"
echo "======================================================"

HTTPCODE=$(curl -s -o /dev/null -w "%{http_code}"  http://127.0.0.1:8080)
if test $HTTPCODE -eq 200; then
    echo "WASM: shutdown staled demo server"
    curl --silent http://127.0.0.1:8080/shutdown
fi
echo $(pwd)
cd ../demo 
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