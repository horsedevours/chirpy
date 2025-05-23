package main

import (
	"encoding/json"
	"io"
)

func unmarshalRequestBody[T any](body io.ReadCloser, v *T) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}
