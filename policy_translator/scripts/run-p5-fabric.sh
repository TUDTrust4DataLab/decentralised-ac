
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NETWORK="$ROOT/../fabric/fabric-samples/test-network"

CHANNEL="${CHANNEL:-mychannel}"
CHAINCODE="${CHAINCODE:-normative-p5-acceptance}"

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


MODE="${MODE:-acceptance}"

P="D"
C="C"
H="R"
R="maintenance-data"
U="predictive-maintenance"
J="job-001"

fact_args() {
    jq -nc --args '$ARGS.positional' "$@"
}

ADMISSION_ARGS=("$P" "$C" "$R" "$U" "$J")
ACCESS_ARGS=("$C" "$P" "$R" "$U" "$J")
CLOSE_ARGS=("$P" "$C" "$R" "$U" "$J")
EVIDENCE_ARGS=("$C" "$P" "$R" "$J")
ASSESS_ARGS=("$P" "$C" "$R" "$J")

OBSERVATION_ARGS=("$C" "$P" "$R" "$J")
FINDING_ARGS=("$H" "$C" "$P" "$R" "$J")
HUMAN_DECISION_ARGS=("$P" "$C" "$H" "$R" "$J")

DUTY_FACT="$(fact_args "$C" "$P" "$R" "$J")"
FINDING_FACT="$(fact_args "$H" "$C" "$R" "$J")"
ACCEPTED_FACT="$(fact_args "$P" "$C" "$R" "$J")"
RESPONSE_FACT="$(fact_args "$P" "$C" "$R" "$J")"

case "$MODE" in
    acceptance)
        INITIAL="$ROOT/normative/bindings/p5-acceptance.json"
        ;;
    response)
        INITIAL="$ROOT/normative/bindings/p5-response.json"
        ;;
    *)
        echo "Unknown mode: $MODE"
        exit 1
        ;;
esac

echo "=== P5: INITIALISATION ($MODE) ==="

FACTS="$(jq -c . "$INITIAL")"
submit InitFacts "$FACTS"

echo "=== P5: ADMISSION AND CLOSURE ==="

submit ApproveAdmission "${ADMISSION_ARGS[@]}"
submit RequestAccess "${ACCESS_ARGS[@]}"
submit CloseJob "${CLOSE_ARGS[@]}"

assert_query true \
    "Deletion duty created" \
    HasFact "deletion-evidence-duty" "$DUTY_FACT"

echo "=== P5: HUMAN ASSESSMENT ==="

if [[ "$MODE" == "acceptance" ]]; then
    submit SubmitEvidence "${EVIDENCE_ARGS[@]}"
else
    submit AssessDeadline "${ASSESS_ARGS[@]}"
fi

assert_query false \
    "No finding before review" \
    HasFact "finding-recorded" "$FINDING_FACT"

submit ReportObservation "${OBSERVATION_ARGS[@]}"
submit RecordFinding "${FINDING_ARGS[@]}"

assert_query true \
    "Human finding recorded" \
    HasFact "finding-recorded" "$FINDING_FACT"

if [[ "$MODE" == "acceptance" ]]; then
    echo "=== P5: HUMAN ACCEPTANCE ==="

    assert_query true \
        "Human acceptance permitted" \
        CanAcceptDeletion "${HUMAN_DECISION_ARGS[@]}"

    submit AcceptDeletion "${HUMAN_DECISION_ARGS[@]}"

    assert_query true \
        "Evidence accepted" \
        HasFact "deletion-evidence-accepted" \
        "$ACCEPTED_FACT"

    assert_query false \
        "Duty terminated" \
        HasFact "deletion-evidence-duty" "$DUTY_FACT"

else
    echo "=== P5: CORRECTIVE RESPONSE ==="

    assert_query true \
        "Corrective response permitted" \
        CanAuthoriseResponse "${HUMAN_DECISION_ARGS[@]}"

    submit AuthoriseResponse "${HUMAN_DECISION_ARGS[@]}"

    assert_query true \
        "Response authorisation recorded" \
        HasFact "human-response-authorised" \
        "$RESPONSE_FACT"

    assert_query true \
        "Duty remains active" \
        HasFact "deletion-evidence-duty" "$DUTY_FACT"
fi

echo "=== P5: COMPLETED ($MODE) ==="
