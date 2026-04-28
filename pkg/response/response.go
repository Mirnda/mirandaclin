package response

import (
	"encoding/json"
	"net/http"
)

// Response é o envelope padrão de todas as respostas da API.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func write(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func OK(w http.ResponseWriter, message string, data any) {
	write(w, http.StatusOK, Response{Success: true, Message: message, Data: data})
}

func Created(w http.ResponseWriter, message string, data any) {
	write(w, http.StatusCreated, Response{Success: true, Message: message, Data: data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, status int, message string) {
	write(w, status, Response{Success: false, Message: message})
}
