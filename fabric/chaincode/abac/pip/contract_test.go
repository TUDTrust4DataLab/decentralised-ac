package pip

import (
	"encoding/json"
	"testing"

	"cipher/fabric/chaincode/abac/model"
	"cipher/fabric/chaincode/abac/util"
)

func testResource() model.Resource {
	return model.Resource{
		ID:          "ddm-resource-1",
		URI:         "/ddm/models/1",
		Owner:       "CIPHER",
		ContentHash: "sha256:abc123",
		Attributes: map[string]interface{}{
			"type": "ddm-model",
		},
	}
}

func resourceJSON(
	t *testing.T,
	resource model.Resource,
) string {

	t.Helper()

	data, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf(
			"failed to marshal resource: %v",
			err,
		)
	}

	return string(data)
}

func TestCreateResource(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	resource := testResource()

	ctx.Stub.MockTransactionStart(
		"tx-create-resource",
	)

	err := contract.CreateResource(
		ctx,
		resourceJSON(t, resource),
	)

	ctx.Stub.MockTransactionEnd(
		"tx-create-resource",
	)

	if err != nil {
		t.Fatalf(
			"CreateResource returned error: %v",
			err,
		)
	}

	stored, err := contract.GetResource(
		ctx,
		resource.ID,
	)

	if err != nil {
		t.Fatalf(
			"GetResource returned error: %v",
			err,
		)
	}

	if stored.ContentHash != "sha256:abc123" {
		t.Fatalf(
			"unexpected content hash: %s",
			stored.ContentHash,
		)
	}

	if stored.CreatedBy == "" {
		t.Fatal(
			"expected CreatedBy to be populated",
		)
	}

	if stored.CreatedAt == "" {
		t.Fatal(
			"expected CreatedAt to be populated",
		)
	}
}

func TestUpdateResource(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	resource := testResource()

	ctx.Stub.MockTransactionStart(
		"tx-create-resource",
	)

	err := contract.CreateResource(
		ctx,
		resourceJSON(t, resource),
	)

	ctx.Stub.MockTransactionEnd(
		"tx-create-resource",
	)

	if err != nil {
		t.Fatalf(
			"unexpected creation error: %v",
			err,
		)
	}

	resource.ContentHash = "sha256:def456"

	ctx.Stub.MockTransactionStart(
		"tx-update-resource",
	)

	err = contract.UpdateResource(
		ctx,
		resourceJSON(t, resource),
	)

	ctx.Stub.MockTransactionEnd(
		"tx-update-resource",
	)

	if err != nil {
		t.Fatalf(
			"UpdateResource returned error: %v",
			err,
		)
	}

	updated, err := contract.GetResource(
		ctx,
		resource.ID,
	)

	if err != nil {
		t.Fatalf(
			"GetResource returned error: %v",
			err,
		)
	}

	if updated.ContentHash != "sha256:def456" {
		t.Fatalf(
			"expected updated hash, got %s",
			updated.ContentHash,
		)
	}

	if updated.UpdatedBy == "" {
		t.Fatal(
			"expected UpdatedBy to be populated",
		)
	}
}
