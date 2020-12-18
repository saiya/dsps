package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

// AdminConfig represents administration feature settings
type AdminConfig struct {
	Auth AdminAuthConfig `json:"auth"`
}

// AdminAuthConfig represents auth settings of admin endpoints
type AdminAuthConfig struct {
	Networks     []domain.CIDR `json:"networks"`
	BearerTokens []string      `json:"bearer"`
}

func adminConfigDefault() *AdminConfig {
	return &AdminConfig{
		Auth: AdminAuthConfig{},
	}
}

// PostprocessAdminConfig cleanups user supplied config object.
func PostprocessAdminConfig(config *AdminConfig) error {
	if len(config.Auth.Networks) == 0 {
		config.Auth.Networks = make([]domain.CIDR, len(domain.PrivateCIDRs))
		copy(config.Auth.Networks, domain.PrivateCIDRs)
	}
	if len(config.Auth.BearerTokens) == 0 {
		config.Auth.BearerTokens = []string{generateAdminAuthRandomToken()}
	}
	return nil
}

var adminAuthRandomTokenOnce sync.Once
var adminAuthRandomToken string

func generateAdminAuthRandomToken() string {
	adminAuthRandomTokenOnce.Do(func() {
		uuid, err := uuid.NewRandom()
		if err != nil {
			panic(xerrors.Errorf("failed to generate UUID for random token: %w", err))
		}
		adminAuthRandomToken = strings.ReplaceAll(uuid.String(), "-", "")
		fmt.Fprintf(os.Stderr, `Generated random token for admin APIs: %s`+"\n", adminAuthRandomToken)
		fmt.Fprintf(os.Stderr, `Set auth.tokens configuration to use fixed token.`+"\n")
	})
	return adminAuthRandomToken
}
