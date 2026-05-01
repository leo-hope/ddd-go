package common

import (
	"errors"
	"fmt"
)

// ErrorCode is the interface for structured error codes.
type ErrorCode interface {
	Code() int
	Msg() string
}

type basicErrorCode struct {
	code int
	msg  string
}

func (b basicErrorCode) Code() int   { return b.code }
func (b basicErrorCode) Msg() string { return b.msg }

var (
	ErrSuccess     ErrorCode = basicErrorCode{200, "ok"}
	ErrParams      ErrorCode = basicErrorCode{400, "params.illegal"}
	ErrSystem      ErrorCode = basicErrorCode{500, "system.error"}
	ErrTimeout     ErrorCode = basicErrorCode{504, "timeout"}
	ErrConcurrency ErrorCode = basicErrorCode{600, "concurrency.conflict"}
)

// BizError is the Go equivalent of BizException.
// It is returned as an error value, never panicked.
type BizError struct {
	ErrorCode ErrorCode
	Detail    string
	Cause     error
}

func (e *BizError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.ErrorCode.Code(), e.ErrorCode.Msg(), e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.ErrorCode.Code(), e.ErrorCode.Msg())
}

func (e *BizError) Unwrap() error { return e.Cause }

func NewBizError(ec ErrorCode, detail string) *BizError {
	return &BizError{ErrorCode: ec, Detail: detail}
}

func NewBizErrorf(ec ErrorCode, format string, args ...any) *BizError {
	return &BizError{ErrorCode: ec, Detail: fmt.Sprintf(format, args...)}
}

func SysError(detail string) *BizError {
	return NewBizError(ErrSystem, detail)
}

// ConcurrencyConflictError is the Go equivalent of ConcurrencyConflictException.
type ConcurrencyConflictError struct {
	BizError
}

func NewConcurrencyConflict(msg string) *ConcurrencyConflictError {
	return &ConcurrencyConflictError{BizError{ErrorCode: ErrConcurrency, Detail: msg}}
}

// CheckRowsAffected returns a ConcurrencyConflictError if n != 1.
// Use after UPDATE/DELETE statements to detect optimistic-lock violations.
func CheckRowsAffected(n int, msg string) error {
	if n != 1 {
		return NewConcurrencyConflict(msg)
	}
	return nil
}

// IsBizError reports whether err is (or wraps) a *BizError.
func IsBizError(err error) (*BizError, bool) {
	var be *BizError
	if errors.As(err, &be) {
		return be, true
	}
	return nil, false
}

// IsConcurrencyConflict reports whether err is a *ConcurrencyConflictError.
func IsConcurrencyConflict(err error) bool {
	var ce *ConcurrencyConflictError
	return errors.As(err, &ce)
}
