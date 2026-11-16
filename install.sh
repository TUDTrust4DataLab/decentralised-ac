#!/bin/bash
set -e
python3 -m venv .venv
source .venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt
cp .env.example .env

# clone fabric
cd fabric
git clone https://github.com/hyperledger/fabric-samples.git
bash install-fabric.sh -s -- -f 2.4.7 -c 1.5.5

cd ..