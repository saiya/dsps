package domain

// ErrorWithCode is an error interface with error code
type ErrorWithCode interface {
	Error() string
	Code() string
	Unwrap() error
}

type errorWithCode struct {
	code    string
	wrapped error
}

// NewErrorWithCode is to make stateless ErrorWithCode instance.
func NewErrorWithCode(code string) ErrorWithCode {
	return &errorWithCode{code: code}
}

// WrapErrorWithCode is to wrap error object with code
func WrapErrorWithCode(code string, err error) ErrorWithCode {
	return &errorWithCode{code: code, wrapped: err}
}

func (e *errorWithCode) Unwrap() error {
	return e.wrapped
}

func (e *errorWithCode) Error() string {
	if e.wrapped != nil {
		return e.wrapped.Error()
	}
	return e.code
}

func (e *errorWithCode) Code() string {
	return e.code
}
