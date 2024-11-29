package openai

import (
	"bytes"
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

func modifyBodyAndParsed(buffer *bytes.Buffer, patches ...*JSONPatchOperationObject) (*bytes.Buffer, map[string]any, error) {
	patch, err := jsonpatch.DecodePatch(NewPatches(patches...))
	if err != nil {
		return nil, nil, err
	}

	patched, err := patch.Apply(buffer.Bytes())
	if err != nil {
		return nil, nil, err
	}

	buffer = bytes.NewBuffer(patched)

	var newParsed map[string]any

	err = json.Unmarshal(patched, &newParsed)
	if err != nil {
		return nil, nil, err
	}

	return buffer, newParsed, nil
}
