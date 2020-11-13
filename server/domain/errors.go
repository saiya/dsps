package domain

// ErrorWithCode is an error interface with error code
type ErrorWithCode interface {
	Error() string
	Code() *string
}

// NewErrorWithCode is to make stateless ErrorWithCode instance.
func NewErrorWithCode(code string) ErrorWithCode {
	return &errorWithCode{code: code}
}

type errorWithCode struct {
	code string
}

func (e *errorWithCode) Error() string {
	return e.code
}

func (e *errorWithCode) Code() *string {
	code := e.code
	return &code
}
