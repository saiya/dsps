package jwt

import (
	"fmt"
	"io/ioutil"

	jwtgo "github.com/dgrijalva/jwt-go/v4"

	"github.com/saiya/dsps/server/domain"
)

// ValidateVerificationKey validates plain-text representation of public key
func ValidateVerificationKey(alg domain.JwtAlg, keyFilePath string) error {
	_, err := LoadVerificationKey(alg, keyFilePath)
	return err
}

// LoadVerificationKey loads public key file
func LoadVerificationKey(alg domain.JwtAlg, keyFilePath string) (interface{}, error) {
	return LoadKey(alg, keyFilePath, false)
}

// LoadKey loads public/private key file
func LoadKey(alg domain.JwtAlg, keyFilePath string, privateKey bool) (interface{}, error) {
	if alg.IsNone() {
		return jwtgo.UnsafeAllowNoneSignatureType, nil
	}

	parserWrapper := func(f func(bytes []byte) (interface{}, error)) (interface{}, error) {
		bytes, err := ioutil.ReadFile(keyFilePath) //nolint:gosec // Only loads file specified by server configuration file
		if err != nil {
			return nil, fmt.Errorf(`failed to read JWT key file "%s": %w`, keyFilePath, err)
		}
		if len(bytes) == 0 {
			return nil, fmt.Errorf(`content of JWT key file "%s" is empty, expected non-empty file`, keyFilePath)
		}
		key, err := f(bytes)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse JWT key file "%s" for alg "%s": %w`, keyFilePath, alg, err)
		}
		return key, nil
	}
	if IsRSA(alg) {
		return parserWrapper(func(bytes []byte) (interface{}, error) {
			if privateKey {
				return jwtgo.ParseRSAPrivateKeyFromPEM(bytes)
			}
			return jwtgo.ParseRSAPublicKeyFromPEM(bytes)
		})
	}
	if IsECDSA(alg) {
		return parserWrapper(func(bytes []byte) (interface{}, error) {
			if privateKey {
				return jwtgo.ParseECPrivateKeyFromPEM(bytes)
			}
			return jwtgo.ParseECPublicKeyFromPEM(bytes)
		})
	}
	if IsHMAC(alg) {
		return parserWrapper(func(bytes []byte) (interface{}, error) { return bytes, nil })
	}
	return nil, fmt.Errorf("Unsupported JWT alg: %s", alg)
}
