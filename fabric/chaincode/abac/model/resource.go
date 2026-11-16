package model

type Resource struct {
	ID          string                 `json:"id"`
	URI         string                 `json:"uri"`
	Owner       string                 `json:"owner"`
	ContentHash string                 `json:"contentHash"`
	Attributes  map[string]interface{} `json:"attributes"`

	CreatedAt string `json:"createdAt"`
	CreatedBy string `json:"createdBy"`
	UpdatedAt string `json:"updatedAt"`
	UpdatedBy string `json:"updatedBy"`
}
