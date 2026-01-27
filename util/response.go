package util

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func WriteResponseAsJSON(data interface{}, w http.ResponseWriter) {
	b, err := EncodeToJson(data)
	if err != nil {
		zap.S().Errorf("Error encoding dto: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		zap.S().Errorf("Error writing response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func EncodeToJson(data interface{}) ([]byte, error) {
	zap.S().Debugf("Parsing data to JSON")
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		zap.S().Errorf("Error marshalling %+v to JSON: %+v", data, err)
		return nil, err
	}
	return jsonBytes, nil
}
