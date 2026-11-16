
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NETWORK="$ROOT/../fabric/fabric-samples/test-network"

CHANNEL="${CHANNEL:-mychannel}"
CHAINCODE="${CHAINCODE:-normative-p4-compliant}"

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

MODE="${MODE:-compliant}"

P="D"
C="C"
R="maintenance-data"
U="predictive-maintenance"
J="job-001"

ADMISSION_ARGS=("$P" "$C" "$R" "$U" "$J")
ACCESS_ARGS=("$C" "$P" "$R" "$U" "$J")
CLOSE_ARGS=("$P" "$C" "$R" "$U" "$J")
SUBMIT_ARGS=("$C" "$P" "$R" "$J")
ACCEPT_ARGS=("$P" "$C" "$R" "$J")

fact_args() {
    jq -nc --args '$ARGS.positional' -- "$@"
}

# jq --args produces a JSON array of strings.
ADMISSION_FACT="$(fact_args "$C" "$R" "$U" "$J")"
DUTY_ARGS="$(fact_args "$C" "$P" "$R" "$J")"
SUBMITTED_ARGS="$(fact_args "$C" "$R" "$J")"
ACCEPTED_ARGS="$(fact_args "$P" "$C" "$R" "$J")"

case "$MODE" in
  compliant)
    INITIAL="$ROOT/normative/bindings/p4-compliant.json"
    ;;
  breach)
    INITIAL="$ROOT/normative/bindings/p4-breach.json"
    ;;
  *)
    echo "Unknown mode: $MODE"
    exit 1
    ;;
esac

if [[ "${RESUME:-false}" != "true" ]]; then
    echo "=== P4: INITIALISATION ($MODE) ==="

    FACTS="$(jq -c . "$INITIAL")"
    submit InitFacts "$FACTS"

    echo "=== P4: ADMISSION ==="

    submit ApproveAdmission "${ADMISSION_ARGS[@]}"
    submit RequestAccess "${ACCESS_ARGS[@]}"
else
    echo "=== P4: RESUMING AFTER ADMISSION ==="
fi

assert_query true \
  "Admission exists" \
  HasFact "admitted" "$ADMISSION_FACT"

echo "=== P4: CLOSE JOB ==="

if [[ "${SKIP_CLOSE:-false}" != "true" ]]; then
    submit CloseJob "${CLOSE_ARGS[@]}"
else
    echo "Skipping CloseJob: previously committed"
fi

assert_query false \
  "Admission terminated" \
  HasFact "admitted" "$ADMISSION_FACT"

assert_query true \
  "Retained duty created" \
  HasFact "deletion-evidence-duty" "$DUTY_ARGS"

if [[ "$MODE" == "compliant" ]]; then
    echo "=== P4: SUBMIT EVIDENCE ==="

    submit SubmitEvidence "${SUBMIT_ARGS[@]}"

    assert_query true \
      "Evidence submitted" \
      HasFact "deletion-evidence-submitted" \
      "$SUBMITTED_ARGS"

    assert_query true \
      "Duty still active" \
      HasFact "deletion-evidence-duty" \
      "$DUTY_ARGS"

    echo "=== P4: ACCEPT EVIDENCE ==="

    submit AcceptEvidence "${ACCEPT_ARGS[@]}"

    assert_query true \
      "Evidence accepted" \
      HasFact "deletion-evidence-accepted" \
      "$ACCEPTED_ARGS"

    assert_query false \
      "Duty terminated" \
      HasFact "deletion-evidence-duty" \
      "$DUTY_ARGS"

else
    echo "=== P4: DEADLINE ASSESSMENT ==="

    assert_query true \
      "Deadline assessment permitted" \
      CanAssessDeadline "${ACCEPT_ARGS[@]}"

    assert_query false \
      "No breach before assessment" \
      HasFact "deletion-deadline-breached" \
      "$SUBMITTED_ARGS"

    submit AssessDeadline "${ACCEPT_ARGS[@]}"

    assert_query true \
      "Deadline breach recorded" \
      HasFact "deletion-deadline-breached" \
      "$SUBMITTED_ARGS"

    assert_query true \
      "Duty remains active" \
      HasFact "deletion-evidence-duty" \
      "$DUTY_ARGS"
fi

echo "=== P4: COMPLETED ($MODE) ==="
