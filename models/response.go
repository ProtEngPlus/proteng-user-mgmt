package models

type HttpResponseOK struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type HttpResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type HttpResponseErrorWithData struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Error   string      `json:"error"`
}

type RetryAfter struct {
	RetryAfterSeconds int `json:"retry_after_seconds"`
}
