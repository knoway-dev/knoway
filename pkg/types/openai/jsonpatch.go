package openai

import (
	"encoding/json"

	"github.com/samber/lo"
)

type JSONPatchOperation string

const (
	JSONPatchOperationReplace JSONPatchOperation = "replace"
)

type JSONPatchOperationObject struct {
	Operation JSONPatchOperation `json:"op"`
	Path      string             `json:"path"`
	Value     string             `json:"value,omitempty"`
}

func NewPatches(operations ...*JSONPatchOperationObject) []byte {
	return lo.Must(json.Marshal(operations))
}

func NewReplace(path string, to string) *JSONPatchOperationObject {
	return &JSONPatchOperationObject{
		Operation: JSONPatchOperationReplace,
		Path:      path,
		Value:     to,
	}
}
