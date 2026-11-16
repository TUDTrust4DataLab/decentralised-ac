package pap

import (
	"encoding/json"
	"fmt"
	"time"

	"cipher/fabric/chaincode/abac/model"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PolicyHistoryEntry struct {
	TxID      string        `json:"txId"`
	Timestamp string        `json:"timestamp"`
	IsDelete  bool          `json:"isDelete"`
	Policy    *model.Policy `json:"policy,omitempty"`
}

func (c *Contract) GetPolicyHistory(
	ctx contractapi.TransactionContextInterface,
	policyID string,
) ([]PolicyHistoryEntry, error) {

	iterator, err := ctx.GetStub().GetHistoryForKey(
		policyKey(policyID),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve history for policy %s: %w",
			policyID,
			err,
		)
	}
	defer iterator.Close()

	var history []PolicyHistoryEntry

	for iterator.HasNext() {
		modification, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to read policy history: %w",
				err,
			)
		}

		entry := PolicyHistoryEntry{
			TxID:     modification.TxId,
			IsDelete: modification.IsDelete,
		}

		if modification.Timestamp != nil {
			t := time.Unix(
				modification.Timestamp.Seconds,
				int64(modification.Timestamp.Nanos),
			).UTC()

			entry.Timestamp = t.Format(time.RFC3339Nano)
		}

		if !modification.IsDelete && len(modification.Value) > 0 {
			var policy model.Policy

			if err := json.Unmarshal(
				modification.Value,
				&policy,
			); err != nil {
				return nil, fmt.Errorf(
					"failed to deserialise historical policy state: %w",
					err,
				)
			}

			entry.Policy = &policy
		}

		history = append(history, entry)
	}

	return history, nil
}
