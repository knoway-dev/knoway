package utils

import (
	"bytes"
	"encoding/json"
	"io"

	"k8s.io/client-go/util/jsonpath"
)

func GetByJSONPath[T any](input any, template string) T {
	var empty T

	j := jsonpath.New("document")
	j.AllowMissingKeys(true)

	err := j.Parse(template)
	if err != nil {
		return empty
	}

	buffer := new(bytes.Buffer)

	err = j.Execute(buffer, input)
	if err != nil {
		return empty
	}

	return FromStringOrEmpty[T](buffer.String())
}

func ReadAsJSONWithClose(readCloser io.ReadCloser) (*bytes.Buffer, map[string]any, error) {
	defer func() {
		_ = readCloser.Close()
	}()

	buffer, jsonMap, err := ReadAsJSON(readCloser)
	if err != nil {
		return buffer, jsonMap, err
	}

	return buffer, jsonMap, nil
}

func ReadAsJSON(reader io.Reader) (*bytes.Buffer, map[string]any, error) {
	buffer := new(bytes.Buffer)
	jsonMap := make(map[string]any)

	_, err := io.Copy(buffer, reader)
	if err != nil {
		return buffer, jsonMap, err
	}

	err = json.Unmarshal(buffer.Bytes(), &jsonMap)
	if err != nil {
		return buffer, jsonMap, err
	}

	return buffer, jsonMap, nil
}

func FromMap[T any, MK comparable, MV any](m map[MK]MV) (*T, error) {
	if m == nil {
		return nil, nil
	}

	if len(m) == 0 {
		return nil, nil
	}

	var initial T

	bs, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bs, &initial)
	if err != nil {
		return nil, err
	}

	return &initial, nil
}
