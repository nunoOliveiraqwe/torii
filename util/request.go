package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func DecodeJSONBody[T any](r *http.Request) (*T, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("request body is empty")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			zap.S().Errorf("error closing request body: %v", err)
		}
	}(r.Body)
	data, err := ParseJSONWithLimit[T](r.Body, 1024*64)
	if err != nil {
		if err == io.EOF {
			return data, fmt.Errorf("request body is empty")
		}
		return data, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return data, nil
}

func ParseJSONWithLimit[T any](r io.Reader, limit int64) (*T, error) {
	limitedReader := io.LimitReader(r, limit)
	var result T
	decoder := json.NewDecoder(limitedReader)
	if err := decoder.Decode(&result); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("empty or incomplete JSON data")
		}
		return nil, err
	}

	return &result, nil
}
