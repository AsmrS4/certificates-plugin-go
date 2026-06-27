package errors

import "fmt"

type HTTPError struct {
	Status  int
	Key     string
	Args    []interface{}
	Details map[string]interface{}
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Key, e.Err)
	}
	return e.Key
}

func (e *HTTPError) Unwrap() error { return e.Err }

func NewHTTP(status int, key string, args ...interface{}) *HTTPError {
	return &HTTPError{Status: status, Key: key, Args: args}
}

func NewHTTPWithDetails(status int, key string, details map[string]interface{}, args ...interface{}) *HTTPError {
	return &HTTPError{Status: status, Key: key, Args: args, Details: details}
}

func BadRequestWithDetails(key string, details map[string]interface{}, args ...interface{}) *HTTPError {
	return NewHTTPWithDetails(400, key, details, args...)
}

func WrapHTTP(status int, key string, err error, args ...interface{}) *HTTPError {
	return &HTTPError{Status: status, Key: key, Args: args, Err: err}
}

func BadRequest(key string, args ...interface{}) *HTTPError {
	return NewHTTP(400, key, args...)
}

func Unauthorized(key string, args ...interface{}) *HTTPError {
	return NewHTTP(401, key, args...)
}

func Forbidden(key string, args ...interface{}) *HTTPError {
	return NewHTTP(403, key, args...)
}

func NotFound(key string, args ...interface{}) *HTTPError {
	return NewHTTP(404, key, args...)
}

func InternalServer(key string, err error, args ...interface{}) *HTTPError {
	return WrapHTTP(500, key, err, args...)
}

const (
	KeyHTTPBadRequest     = "http_bad_request"
	KeyHTTPUnauthorized   = "http_unauthorized"
	KeyHTTPForbidden      = "http_forbidden"
	KeyHTTPNotFound       = "http_not_found"
	KeyHTTPInternalServer = "http_internal_server"
)
