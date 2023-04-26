#!/bin/bash

if [ -z "$ZCN_WASM_TESTS" ] 
then
  echo "WASM: SKIP DUE TO NOT SYSTEM-TESTS"
  exit 0
fi


echo "======================================================"
echo "SETTING WASM TESTS:"
echo "======================================================"
git clone -q https://github.com/0chain/gosdk.git
cd gosdk

if [[ -n "$ZCN_WASM_GOSDK" && "$ZCN_WASM_GOSDK" != "NONE" ]];
then
  git checkout $ZCN_WASM_GOSDK
fi

CGO_ENABLED=0 GOOS=js GOARCH=wasm go build -o ./wasmsdk/demo/zcn.wasm  ./wasmsdk

HTTPCODE=$(curl -s -o /dev/null -w "%{http_code}"  http://127.0.0.1:8080)
if test $HTTPCODE -eq 200; then
    echo "WASM: shutdown staled demo server"
    curl --silent http://127.0.0.1:8080/shutdown
fi
cd ./wasmsdk/demo && go build -o demo .
./demo &
sleep 3
cd ../../..

HTTPCODE=$(curl -s -o /dev/null -w "%{http_code}"  http://127.0.0.1:8080)
if test $HTTPCODE -eq 200; then
  echo "WASM DEMO SERVER IS RUNNING"
else
  echo "!!! WASM DEMO SERVER IS DOWN !!!"
  exit 1
fi