package error

import (
	"encoding/json"
	"net/http"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

//func ReturnJSONError(rw http.ResponseWriter, errorMessage string, statusCode int) {
//	rw.Header().Set("Content-Type", "application/json")
//	rw.WriteHeader(statusCode)
//
//	errorResponse := ErrorMessage{Error: errorMessage}
//	jsonResponse, err := json.Marshal(errorResponse)
//	if err != nil {
//		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
//		return
//	}
//
//	rw.Write(jsonResponse)
//}

func ReturnJSONError(w http.ResponseWriter, err interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}
