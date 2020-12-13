package jwt_test

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/jwt"
)

func TestKeyLoading(t *testing.T) {
	for _, testcase := range []struct {
		alg      domain.JwtAlg
		private  bool
		filename string
		errorMsg string
	}{
		// OK
		{alg: "ES512", private: true, filename: "ES512-test1-private.pem"},
		{alg: "ES512", private: false, filename: "ES512-test1-public.pem"},
		{alg: "HS256", private: false, filename: "HS256.rand"},
		{alg: "HS256", private: true, filename: "HS256.rand"},
		{alg: "RS256", private: true, filename: "RS256-2048bit-private.pem"},
		{alg: "RS256", private: false, filename: "RS256-4096bit-public.pem"},

		// Error
		{alg: "ES512", private: true, filename: "file-not-found", errorMsg: "no such file or directory"},
		{alg: "ES512", private: true, filename: "emptyfile", errorMsg: "expected non-empty file"},
		{alg: "ES512", private: true, filename: "HS256.rand", errorMsg: "Key must be PEM encoded PKCS1 or PKCS8 private key"},
	} {
		alg := testcase.alg
		assertError := func(err error) {
			if testcase.errorMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testcase.errorMsg)
			}
		}
		assertKey := func(key interface{}) {
			if testcase.errorMsg == "" {
				if IsECDSA(alg) {
					if testcase.private {
						assert.NotNil(t, key.(*ecdsa.PrivateKey))
					} else {
						assert.NotNil(t, key.(*ecdsa.PublicKey))
					}
				} else if IsRSA(alg) {
					if testcase.private {
						assert.NotNil(t, key.(*rsa.PrivateKey))
					} else {
						assert.NotNil(t, key.(*rsa.PublicKey))
					}
				} else if IsHMAC(alg) {
					assert.NotNil(t, key.([]byte))
				} else {
					assert.Fail(t, "Unknown alg "+string(alg))
				}
			} else {
				assert.Nil(t, key)
			}
		}

		filepath := "./testdata/" + testcase.filename
		if !testcase.private {
			assertError(ValidateVerificationKey(alg, filepath))

			pubkey, err := LoadVerificationKey(alg, filepath)
			assertError(err)
			assertKey(pubkey)
		}

		key, err := LoadKey(alg, filepath, testcase.private)
		assertError(err)
		assertKey(key)
	}
}
