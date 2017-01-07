package config

import "strings"

const (
	ErrorCodeBadRequest            = "C-BAD-REQUEST"
	ErrorCodeNotFound              = "C-NOT-FOUND"
	ErrorCodeBadRequestContentType = "C-BAD-CONTENT-TYPE"

	ErrorCodeServerError = "S-ERROR"

	ClientErrorCodePrefix = "C-"
	ServerErrorCodePrefix = "S-"
)

var codeToStatus = map[string]int{
	ErrorCodeBadRequest:            400,
	ErrorCodeNotFound:              404,
	ErrorCodeBadRequestContentType: 415,
}

var ErrorCodeToHTTPStatus = func(code string) int {
	if status, ok := codeToStatus[code]; ok {
		return status
	}
	if strings.HasPrefix(code, ClientErrorCodePrefix) {
		return 400
	}
	return 500
}
