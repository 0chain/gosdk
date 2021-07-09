#!/usr/bin/env bash

BASEDIR=$(pwd)

echo "Making mocks..."

cd $BASEDIR/zboxcore || exit
mockery --output=./mocks --all

cd $BASEDIR/zcncore || exit
mockery --output=./mocks --all

echo "Mocks files are generated."