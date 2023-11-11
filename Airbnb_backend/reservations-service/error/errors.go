package error

import (
	"encoding/json"
	"net/http"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

func ReturnJSONError(rw http.ResponseWriter, errorMessage string, statusCode int) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)

	errorResponse := ErrorMessage{Error: errorMessage}
	jsonResponse, err := json.Marshal(errorResponse)
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rw.Write(jsonResponse)
}
