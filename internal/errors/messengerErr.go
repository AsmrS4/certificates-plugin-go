// internal/errors/errors.go
package errors

import "fmt"

type AppError struct {
	Status int
	Key    string
	Args   []interface{}
	Err    error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Key, e.Err)
	}
	return e.Key
}

func (e *AppError) Unwrap() error { return e.Err }

func New(key string, args ...interface{}) *AppError {
	return &AppError{Key: key, Args: args}
}

func Wrap(key string, err error, args ...interface{}) *AppError {
	return &AppError{Key: key, Args: args, Err: err}
}

const (
	KeyOrderNotFound         = "order_not_found"
	KeyOrderNotPending       = "order_not_pending"
	KeyOrderAlreadyCancelled = "order_cancelled"
	KeyInternalError         = "internal_error"
	KeyInvalidID             = "invalid_id"
	KeyAttachmentFailed      = "attachment_save_failed"
	KeyFileUploadFailed      = "file_upload_failed"
	KeyValidationFailed      = "validation_failed"
	KeyAccessDenied          = "access_denied"
	KeyTsuNotLinked          = "link_tsu_account_required"
	KeyUserNotFound          = "user_not_found"
)
