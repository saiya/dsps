# PUT `/channel/{channelID}/message/{messageID}`

Send message to the channel.

Note: To prevent data-loss, you should setup servers to use persistent [storage](../storage) type.

## Retry handling

You can retry this API with same `channelID` + `messageID`.

If you call this API twice with same ID, server delivers *any* one of message.

So that you should supply exactly same content with same ID.

## Request

### `channelID` parameter (required)

ChannelID to send a message.

See [channel creation API](./create_channel.md) for detail.

### `messageID` parameter (required)

Unique identifier of the message to send. This value must be unique within the channel.

This ID is used for retry handling, see "retry handling" section for detail.

### Request body (required, application/json)

Validation rule: must be valid JSON

Content of the message.

You can send any JSON.

## Response

Returns HTTP `200` with `application/json` response body if success.

Example:

```json
{
  "channelID": "cc457b533ad54a47b0facc44daf51ad8",
  "messageID": "my-first-message"
}
```

### `channelID` (string, always returned)

ChannelID of the channel you sent to, exactly same as request parameter.

### `messageID` (string, always returned)

ID of the message, exactly same as request parameter.
