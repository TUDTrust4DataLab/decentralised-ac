package pap

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func policyKey(policyID string) string {
	return policyKeyPrefix + policyID
}

func resourcePolicyKey(resourceID string) string {
	return resourcePolicyKeyPrefix + resourceID
}

func getClientIdentity(
	ctx contractapi.TransactionContextInterface,
) (string, error) {

	identity := ctx.GetClientIdentity()

	id, err := identity.GetID()
	if err != nil {
		return "", fmt.Errorf("failed to obtain client identity: %w", err)
	}

	// Fabric's GetID() is commonly represented as a base64-encoded
	// X.509 subject/issuer identifier. Decode it when possible so the
	// stored audit value is human-readable.
	decoded, err := base64.StdEncoding.DecodeString(id)
	if err == nil {
		return string(decoded), nil
	}

	return id, nil
}

func getTransactionTimestamp(
	ctx contractapi.TransactionContextInterface,
) (string, error) {

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf(
			"failed to obtain transaction timestamp: %w",
			err,
		)
	}

	t := time.Unix(
		timestamp.Seconds,
		int64(timestamp.Nanos),
	).UTC()

	return t.Format(time.RFC3339Nano), nil
}
