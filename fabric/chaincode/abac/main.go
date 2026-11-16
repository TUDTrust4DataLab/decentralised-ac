package main

import (
	"log"

	"cipher/fabric/chaincode/abac/pap"
	"cipher/fabric/chaincode/abac/pdp"
	"cipher/fabric/chaincode/abac/pip"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	chaincode, err := contractapi.NewChaincode(
		pap.NewContract(),
		pip.NewContract(),
		pdp.NewContract(),
	)

	if err != nil {
		log.Panicf(
			"failed to create ABAC chaincode: %v",
			err,
		)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf(
			"failed to start ABAC chaincode: %v",
			err,
		)
	}
}
