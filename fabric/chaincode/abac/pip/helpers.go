package pip

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const resourceKeyPrefix = "resource:"

func resourceKey(resourceID string) string {
	return resourceKeyPrefix + resourceID
}

func getClientIdentity(
	ctx contractapi.TransactionContextInterface,
) (string, error) {

	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf(
			"failed to obtain client identity: %w",
			err,
		)
	}

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
