# PUT `/admin/jwt/revoke?jti={jti}&exp={exp}`

Revoke JWT.
This API adds specified JWT to deny list.

See [channels.jwt configuration block](../../config.md#jwt) document how you can use JWT validation to protect endpoints.

Note: To prevent data-loss, you should setup servers to use persistent [storage](../../storage) type.

## Retry handling

You can retry this API.

This API success even if specified JWT has already been revoked.

## Request

### `jti` parameter (required, string)

This value must be exactly equal to [`jti` (JWT ID) claim](https://tools.ietf.org/html/rfc7519#section-4.1.7) of the JWT you want to revoke.

### `exp` parameter (required, integer)

This value must be equal to or larger than [`exp` (expiration time) claim](https://tools.ietf.org/html/rfc7519#section-4.1.4) of the JWT you want to revoke.

Value of the `exp` claim is number of seconds from 1970-01-01T00:00:00Z UTC without leap seconds.

Note: to prevent bloat of storage, [storage implementation](../../storage) may automatically delete expired revocation records from revocation list.

### Request body

No need to send request body to this API.

## Response

Returns HTTP `200` with `application/json` response body if success.

Example:

```json
{
  "exp": 1300819380,
  "jti": "id-of-the-JWT-to-revoke"
}
```

### `exp` (integer, always returned)

This is exactly same value you specified.

### `jti` (string, always returned)

This is exactly same value you specified.
