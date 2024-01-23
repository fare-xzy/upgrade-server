package util

import "net/http"

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(data any) *Result {
	return &Result{
		Code:    http.StatusOK,
		Message: "OK",
		Data:    data,
	}
}

func FailWithData(code int, data any) *Result {
	return &Result{
		Code:    code,
		Message: "FAIL",
		Data:    data,
	}
}

func FailWithMsg(code int, message string) *Result {
	return &Result{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

func Build(code int, message string, data any) *Result {
	return &Result{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
