#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NETWORK="$ROOT/../fabric/fabric-samples/test-network"

CHANNEL="${CHANNEL:-mychannel}"
CHAINCODE="${CHAINCODE:-normative-p1}"

command -v jq >/dev/null || {
  echo "ERROR: jq is required"
  exit 1
}

cd "$NETWORK"

export PATH="$PWD/../bin:$PATH"
export FABRIC_CFG_PATH="$PWD/../config"

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=localhost:7051

export CORE_PEER_TLS_ROOTCERT_FILE="$PWD/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"

export CORE_PEER_MSPCONFIGPATH="$PWD/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp"

ORDERER_CA="$PWD/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"

ORG1_CA="$PWD/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"

ORG2_CA="$PWD/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"

payload() {
  local function="$1"
  shift

  jq -nc \
    --arg fn "PolicyContract:$function" \
    '{Args: ([$fn] + $ARGS.positional)}' \
    --args "$@"
}

evaluate() {
  local function="$1"
  shift

  peer chaincode query \
    -C "$CHANNEL" \
    -n "$CHAINCODE" \
    -c "$(payload "$function" "$@")"
}

submit() {
  local function="$1"
  shift

  peer chaincode invoke \
    -o localhost:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --tls \
    --cafile "$ORDERER_CA" \
    -C "$CHANNEL" \
    -n "$CHAINCODE" \
    --peerAddresses localhost:7051 \
    --tlsRootCertFiles "$ORG1_CA" \
    --peerAddresses localhost:9051 \
    --tlsRootCertFiles "$ORG2_CA" \
    --waitForEvent \
    -c "$(payload "$function" "$@")"
}

assert_query() {
  local expected="$1"
  local label="$2"
  local function="$3"
  shift 3

  local actual
  actual="$(evaluate "$function" "$@")"

  printf '%s: expected=%s observed=%s\n' \
    "$label" "$expected" "$actual"

  if [[ "$actual" != "$expected" ]]; then
    echo "FAIL: $label"
    exit 1
  fi

  echo "PASS: $label"
}

PROVIDER="D"
CONSUMER="C"
RESOURCE="maintenance-data"
PURPOSE="predictive-maintenance"
JOB="job-001"

echo "=== P1: INITIALISATION ==="

INITIAL_STATE="$ROOT/normative/bindings/p1-initial-state.json"

if [[ ! -f "$INITIAL_STATE" ]]; then
  echo "Missing initial-state file: $INITIAL_STATE"
  exit 1
fi

FACTS="$(jq -c . "$INITIAL_STATE")"
submit InitFacts "$FACTS"

echo
echo "=== P1: BEFORE APPROVAL ==="

assert_query \
  false \
  "Access before approval" \
  CanRequestAccess \
  "$CONSUMER" "$PROVIDER" \
  "$RESOURCE" "$PURPOSE" "$JOB"

echo
echo "=== P1: APPROVAL ==="

submit ApproveAdmission \
  "$PROVIDER" "$CONSUMER" \
  "$RESOURCE" "$PURPOSE" "$JOB"

assert_query \
  true \
  "Approval recorded" \
  HasFact \
  "admission-approved" \
  "$(jq -nc \
    --arg p "$PROVIDER" \
    --arg c "$CONSUMER" \
    --arg r "$RESOURCE" \
    --arg u "$PURPOSE" \
    --arg j "$JOB" \
    '[$p,$c,$r,$u,$j]')"

echo
echo "=== P1: AFTER APPROVAL ==="

assert_query \
  true \
  "Access after approval" \
  CanRequestAccess \
  "$CONSUMER" "$PROVIDER" \
  "$RESOURCE" "$PURPOSE" "$JOB"

echo
echo "=== P1: ADMISSION ==="

submit RequestAccess \
  "$CONSUMER" "$PROVIDER" \
  "$RESOURCE" "$PURPOSE" "$JOB"

assert_query \
  true \
  "Admission recorded" \
  HasFact \
  "admitted" \
  "$(jq -nc \
    --arg c "$CONSUMER" \
    --arg r "$RESOURCE" \
    --arg u "$PURPOSE" \
    --arg j "$JOB" \
    '[$c,$r,$u,$j]')"

echo
echo "=== P1: COMPLETED ==="
