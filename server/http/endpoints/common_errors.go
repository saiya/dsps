package endpoints

import (
	"github.com/saiya/dsps/server/domain"
)

// TODO: Move to authenticate module
var (
	ErrInvalidCredentials = domain.NewErrorWithCode("dsps.auth.invalid-credentials")
	ErrForbiddenChannel   = domain.NewErrorWithCode("dsps.auth.channel-forbidden")
)
