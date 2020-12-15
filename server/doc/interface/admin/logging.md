# PUT `/admin/log/level?category={category}&level={level}`

Change logging threshold.

## Retry handling

Because this API is idempotent, You can retry this API.

## Request

### `category` parameter (required, string)

Specify logging category, same as [key of "category" property of "logging" configuration file section](../../config.md#logging).

Special value `*` means default logging level.

### `level` parameter (required, string)

Logging level threshold, same as [value of "category" property of "logging" configuration file section](../../config.md#logging) (e.g. `INFO`).

### Request body

No need to send request body to this API.

## Response

Returns HTTP `204` (no content) if success.
