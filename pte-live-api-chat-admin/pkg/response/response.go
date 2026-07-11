package response

import (
	"encoding/json"
	"net/http"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func JSON(w http.ResponseWriter, status int, code int, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Body{Code: code, Msg: msg, Data: data})
}

func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, 1, "success", data)
}

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, -1, msg, map[string]interface{}{})
}

func MethodNotAllowed(w http.ResponseWriter) {
	Error(w, http.StatusMethodNotAllowed, "method not allowed")
}
