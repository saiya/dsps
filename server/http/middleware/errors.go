package middleware

import (
	"github.com/saiya/dsps/server/domain"
)

var (
	// ErrAuthRejection : auth rejection
	ErrAuthRejection = domain.NewErrorWithCode("dsps.auth.rejected")
)
