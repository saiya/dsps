package jwt

import (
	"fmt"
	"regexp"

	"github.com/saiya/dsps/server/domain"
)

// ValidateAlg validates JWT alg
func ValidateAlg(alg domain.JwtAlg) error {
	if alg.IsNone() || IsRSA(alg) || IsECDSA(alg) || IsHMAC(alg) {
		return nil
	}
	return fmt.Errorf(`Unsupported JWT alg "%s"`, alg)
}

var algRSA = regexp.MustCompile(`^(RS|PS)[0-9]+$`) // RS: RSASSA-PKCS1, PS: RSASSA-PSS
var algECDSA = regexp.MustCompile(`^ES[0-9]+$`)
var algHMAC = regexp.MustCompile(`^HS[0-9]+$`)

// IsRSA returns true for RS*** algorithms
func IsRSA(alg domain.JwtAlg) bool {
	return algRSA.MatchString(string(alg))
}

// IsECDSA returns true for ES*** algorithms
func IsECDSA(alg domain.JwtAlg) bool {
	return algECDSA.MatchString(string(alg))
}

// IsHMAC returns true for HS*** algorithms
func IsHMAC(alg domain.JwtAlg) bool {
	return algHMAC.MatchString(string(alg))
}
