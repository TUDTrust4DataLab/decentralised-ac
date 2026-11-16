package util

import (
	"crypto/x509"
	"encoding/base64"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type MockClientIdentity struct {
	ID string
}

func (m *MockClientIdentity) GetID() (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(m.ID)), nil
}

func (m *MockClientIdentity) GetMSPID() (string, error) {
	return "Org1MSP", nil
}

func (m *MockClientIdentity) GetAttributeValue(
	attrName string,
) (value string, found bool, err error) {
	return "", false, nil
}

func (m *MockClientIdentity) AssertAttributeValue(
	attrName string,
	attrValue string,
) error {
	return nil
}

func (m *MockClientIdentity) GetX509Certificate() (
	*x509.Certificate,
	error,
) {
	return nil, nil
}

type MockTransactionContext struct {
	contractapi.TransactionContextInterface
	Stub           *testStub
	clientIdentity *MockClientIdentity
}

func (m *MockTransactionContext) GetStub() shim.ChaincodeStubInterface {
	return m.Stub
}

func (m *MockTransactionContext) GetClientIdentity() cid.ClientIdentity {
	return m.clientIdentity
}

type testStub struct {
	*shimtest.MockStub
	txTimestamp *timestamp.Timestamp
}

func (s *testStub) GetTxTimestamp() (*timestamp.Timestamp, error) {
	return s.txTimestamp, nil
}

func StartTransaction(
	ctx *MockTransactionContext,
	txID string,
) {
	ctx.Stub.MockTransactionStart(txID)
}

func EndTransaction(
	ctx *MockTransactionContext,
	txID string,
) {
	ctx.Stub.MockTransactionEnd(txID)
}

func NewTestContext() *MockTransactionContext {
	Stub := &testStub{
		MockStub: shimtest.NewMockStub("abac-test", nil),
		txTimestamp: &timestamp.Timestamp{
			Seconds: time.Date(
				2026,
				time.October,
				5,
				12,
				0,
				0,
				0,
				time.UTC,
			).Unix(),
		},
	}

	return &MockTransactionContext{
		Stub: Stub,
		clientIdentity: &MockClientIdentity{
			ID: "x509::CN=policy-admin::CN=ca.org1.example.com",
		},
	}
}
