package response

import (
	"encoding/json"
	"net/http"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, status int, body Body) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Body{Code: 1, Msg: "success", Data: data})
}

func Error(w http.ResponseWriter, status int, msg string) {
	if status <= 0 {
		status = http.StatusOK
	}
	JSON(w, status, Body{Code: 0, Msg: msg})
}

func MethodNotAllowed(w http.ResponseWriter) {
	Error(w, http.StatusMethodNotAllowed, "方法不允许")
}
