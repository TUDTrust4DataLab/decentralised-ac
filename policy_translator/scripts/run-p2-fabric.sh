
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NETWORK="$ROOT/../fabric/fabric-samples/test-network"

CHANNEL="${CHANNEL:-mychannel}"
CHAINCODE="${CHAINCODE:-normative-p2}"

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
O="analysis-001"

# Arguments follow the generated action signatures.
ADMISSION_ARGS=("$P" "$C" "$R" "$U" "$J")
ACCESS_ARGS=("$C" "$P" "$R" "$U" "$J")
STEP_ARGS=("$C" "$P" "$R" "$U" "$J" "$O")
SUSPEND_ARGS=("$P" "$C" "$R" "$J")

echo "=== P2: INITIALISATION ==="

FACTS="$(jq -c . \
  "$ROOT/normative/bindings/p2-initial-state.json")"

submit InitFacts "$FACTS"

echo
echo "=== P2: ADMISSION ==="

submit ApproveAdmission "${ADMISSION_ARGS[@]}"
submit RequestAccess "${ACCESS_ARGS[@]}"

echo
echo "=== P2: OPERATION BEFORE SUSPENSION ==="

assert_query true \
  "Step initially permitted" \
  CanAuthoriseStep "${STEP_ARGS[@]}"

submit AuthoriseStep "${STEP_ARGS[@]}"

assert_query true \
  "Previous execution recorded" \
  HasFact "step-executed" \
  "$(jq -nc \
      --arg c "$C" --arg r "$R" \
      --arg j "$J" --arg o "$O" \
      '[$c,$r,$j,$o]')"

echo
echo "=== P2: SUSPENSION ==="

assert_query true \
  "Session can be suspended" \
  CanSuspendSession "${SUSPEND_ARGS[@]}"

submit SuspendSession "${SUSPEND_ARGS[@]}"

assert_query false \
  "Operational conditions terminated" \
  HasFact "operating-conditions-met" \
  "$(jq -nc \
      --arg p "$P" --arg r "$R" \
      --arg j "$J" \
      '[$p,$r,$j]')"

echo
echo "=== P2: AFTER SUSPENSION ==="

assert_query false \
  "New step no longer permitted" \
  CanAuthoriseStep "${STEP_ARGS[@]}"

assert_query true \
  "Previous execution remains recorded" \
  HasFact "step-executed" \
  "$(jq -nc \
      --arg c "$C" --arg r "$R" \
      --arg j "$J" --arg o "$O" \
      '[$c,$r,$j,$o]')"

echo
echo "=== P2: COMPLETED ==="
