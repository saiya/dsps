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

String exactly equal to the value of [`jti` (JWT ID) claim](https://tools.ietf.org/html/rfc7519#section-4.1.7) identifies the JWT you want to revoke.

### `exp` parameter (required, integer)

Expiration of the deny list entry added by this API call.
To prevent bloat of storage, [storages](../../storage) may forget expired revocation records after this period passes.

This value should be equal to or larger than [`exp` (expiration time) claim](https://tools.ietf.org/html/rfc7519#section-4.1.4) of the JWT you want to revoke.

Format of this parameter is exactly same as `exp` claim of JWT, number of seconds from `1970-01-01T00:00:00Z` without leap seconds.

### Request body

No need to send request body to this API.

## Response

Returns HTTP `200` with `application/json` response body if success.

Example:

```json
{
  "jti": "id-of-the-JWT-to-revoke",
  "exp": 1300819380
}
```

### `jti` (string, always returned)

This is exactly same value you specified.

### `exp` (integer, always returned)

This is exactly same value you specified.
