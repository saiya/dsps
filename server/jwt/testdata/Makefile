.PHONY: all pubkeys jwts
all: pubkeys jwts emptyfile

emptyfile:
	touch $@

pubkeys: HS256.rand RS256-2048bit-public.pem RS256-4096bit-public.pem ES512-test1-public.pem ES512-test2-public.pem
HS256.rand:
	openssl rand 64 > $@

RS256-2048bit-private.pem:
	openssl genrsa 2048 -outform PEM > $@
RS256-2048bit-public.pem: RS256-2048bit-private.pem
	openssl rsa -in $< -pubout -outform PEM -out $@

RS256-4096bit-private.pem:
	openssl genrsa 4096 -outform PEM > $@
RS256-4096bit-public.pem: RS256-4096bit-private.pem
	openssl rsa -in $< -pubout -outform PEM -out $@

ES512-test1-private.pem:
	openssl ecparam -genkey -name secp521r1 -noout -out $@
ES512-test1-public.pem: ES512-test1-private.pem
	openssl ec -in $< -pubout -outform PEM -out $@

ES512-test2-private.pem:
	openssl ecparam -genkey -name secp521r1 -noout -out $@
ES512-test2-public.pem: ES512-test2-private.pem
	openssl ec -in $< -pubout -outform PEM -out $@

#
# Requires `go install github.com/dgrijalva/jwt-go/cmd/jwt`
#
# UNIX epoch used in JWTs (UTC):
# 	946688400	= Saturday, January 1, 2000 1:00:00 AM
#	978310800	= Monday, January 1, 2001 1:00:00 AM
# 	4070912400	= Thursday, January 1, 2099 1:00:00 AM
# 
jwts: RS256-2048bit.jwt RS256-2048bit-expired.jwt
RS256-2048bit.jwt: RS256-2048bit-private.pem
	echo '{ "iat": 946688400, "exp": 4070912400, "iss": "https://example.com/issuer", "aud": "https://example.com/audience", "sub": "3AA7AD8B-FE18-4A6B-87FC-1A4878993AC1" }' | jwt -key $< -alg RS256 -sign - > $@
RS256-2048bit-expired.jwt: RS256-2048bit-private.pem
	echo '{ "iat": 946688400, "exp": 978310800, "iss": "https://example.com/issuer", "aud": "https://example.com/audience", "sub": "3AA7AD8B-FE18-4A6B-87FC-1A4878993AC1" }' | jwt -key $< -alg RS256 -sign - > $@
