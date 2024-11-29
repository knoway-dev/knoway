package openai

import (
	"encoding/json"

	"github.com/samber/lo"
)

type JSONPatchOperation string

const (
	JSONPatchOperationAdd     JSONPatchOperation = "add"
	JSONPatchOperationRemove  JSONPatchOperation = "remove"
	JSONPatchOperationReplace JSONPatchOperation = "replace"
)

type JSONPatchOperationObject struct {
	Operation JSONPatchOperation `json:"op"`
	Path      string             `json:"path"`
	Value     any                `json:"value,omitempty"`
}

func NewPatches(operations ...*JSONPatchOperationObject) []byte {
	return lo.Must(json.Marshal(operations))
}

func NewReplace(path string, to any) *JSONPatchOperationObject {
	return &JSONPatchOperationObject{
		Operation: JSONPatchOperationReplace,
		Path:      path,
		Value:     to,
	}
}

func NewAdd(path string, value any) *JSONPatchOperationObject {
	return &JSONPatchOperationObject{
		Operation: JSONPatchOperationAdd,
		Path:      path,
		Value:     value,
	}
}

func NewRemove(path string) *JSONPatchOperationObject {
	return &JSONPatchOperationObject{
		Operation: JSONPatchOperationRemove,
		Path:      path,
	}
}
