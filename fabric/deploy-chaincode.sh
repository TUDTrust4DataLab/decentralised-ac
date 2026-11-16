#!/bin/bash

cd fabric-samples/test-network

./network.sh deployCC \
  -ccn abac \
  -ccp ../../chaincode/abac \
  -ccl go