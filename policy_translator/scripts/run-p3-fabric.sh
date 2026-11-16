
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NETWORK="$ROOT/../fabric/fabric-samples/test-network"

CHANNEL="${CHANNEL:-mychannel}"
CHAINCODE="${CHAINCODE:-normative-p3}"

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


P="D"
C="C"
R="maintenance-data"
U="predictive-maintenance"
J="job-001"
O="report-001"
OTHER="report-002"

ADMISSION_ARGS=("$P" "$C" "$R" "$U" "$J")
ACCESS_ARGS=("$C" "$P" "$R" "$U" "$J")
OUTPUT_ARGS=("$C" "$P" "$R" "$O" "$U" "$J")
APPROVAL_ARGS=("$P" "$C" "$R" "$O" "$U" "$J")
OTHER_ARGS=("$C" "$P" "$R" "$OTHER" "$U" "$J")

echo "=== P3: INITIALISATION ==="
FACTS="$(jq -c . \
  "$ROOT/normative/bindings/p3-initial-state.json")"
submit InitFacts "$FACTS"

echo "=== P3: ADMISSION ==="
submit ApproveAdmission "${ADMISSION_ARGS[@]}"
submit RequestAccess "${ACCESS_ARGS[@]}"

echo "=== P3: BEFORE OUTPUT REGISTRATION ==="
assert_query false \
  "Output not yet eligible" \
  CanApproveRelease "${APPROVAL_ARGS[@]}"

echo "=== P3: REGISTER OUTPUT ==="
submit RegisterOutput "${OUTPUT_ARGS[@]}"

assert_query true \
  "Output registered" \
  HasFact "produced-output" \
  "$(jq -nc --arg c "$C" --arg r "$R" \
    --arg o "$O" --arg j "$J" \
    '[$c,$r,$o,$j]')"

assert_query true \
  "Release approval now possible" \
  CanApproveRelease "${APPROVAL_ARGS[@]}"

echo "=== P3: BEFORE RELEASE APPROVAL ==="
assert_query false \
  "Release blocked before approval" \
  CanReleaseOutput "${OUTPUT_ARGS[@]}"

echo "=== P3: APPROVE RELEASE ==="
submit ApproveRelease "${APPROVAL_ARGS[@]}"

assert_query true \
  "Release approved" \
  CanReleaseOutput "${OUTPUT_ARGS[@]}"

assert_query false \
  "Other output not authorised" \
  CanReleaseOutput "${OTHER_ARGS[@]}"

echo "=== P3: RELEASE OUTPUT ==="
submit ReleaseOutput "${OUTPUT_ARGS[@]}"

assert_query true \
  "Release recorded" \
  HasFact "output-released" \
  "$(jq -nc --arg c "$C" --arg r "$R" \
    --arg o "$O" --arg u "$U" \
    --arg j "$J" '[$c,$r,$o,$u,$j]')"

echo "=== P3: COMPLETED ==="
