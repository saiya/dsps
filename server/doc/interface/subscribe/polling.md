# PUT `/channel/{channelID}/subscription/polling/{subscriberID}`

Create long polling subscriber.

Behind the scene, DSPS server saves messages for each polling subscribers. You **must create subscriber before messages you want to receive**. No guarantee whether you can / cannot receive messages that sent before subscriber creation.

## Retry handling

You can retry this API with same `channelID` + `subscriberID`.

This API success even if the subscriber already exists.

## Request

### `channelID` parameter (required)

ID of a channel to receive messages.

If not exists, automatically create channel.

### `subscriberID` parameter (required)

ID of the subscriber to create.

Subscriber ID must be unique within the channel.

Note: you can retry this API with the same subscriberID. DSPS server just returns `200` for duplicated requests and not create duplicated internal resources.

### Request body

No need to send request body to this API.

## Response

Returns HTTP `200` with `application/json` response body if success.

Example:

```json
{
  "channelID": "cc457b533ad54a47b0facc44daf51ad8",
  "subscriberID": "bf29dd67ced04692bf87500095396d9b"
}
```

### `channelID` (string, always returned)

ChannelID of the subscriber belongs to, exactly same as request parameter.

### `subscriberID` (string, always returned)

ID of the created subscriber, exactly same as request parameter.



# DELETE `/channel/{channelID}/subscription/polling/{subscriberID}`

Delete subscriber.

No guarantee whether you can / cannot receive messages that sent after subscriber deletion.

You should not create subscriber with the ID deleted before.
Should change subscriber ID to create new subscription, otherwise you may encounter confusing race-condition behavior.

## Retry handling

You can retry this API with same `channelID` + `subscriberID`.

This API success even if specified subscriber does not exists.

## Request

### `subscriberID` parameter (required)

ID of the subscriber.

### `channelID` parameter (required)

ID of the channel that the subscriber belongs to.

### Request body

No need to send request body to this API.

## Response

Returns HTTP `200` with `application/json` response body if success.

Note that this API returns `200` even if the subscriber does not exists (see "retry handling" section).

Example:

```json
{
  "channelID": "cc457b533ad54a47b0facc44daf51ad8",
  "subscriberID": "bf29dd67ced04692bf87500095396d9b"
}
```

### `channelID` (string, always returned)

ChannelID of the subscriber belongs to, exactly same as request parameter.

### `subscriberID` (string, always returned)

ID of the created subscriber, exactly same as request parameter.



# <a name="polling-get"></a> GET `/channel/{channelID}/subscription/polling/{subscriberID}?timeout={timeout}`

Receive messages with long polling.

If there are no messages, this API await for new message arrival with specified timeout.

**Important note**: This API does not remove received messages from the subscriber. You **MUST** acknowledge (DELETE) messages immediately after you successfully received message.

## Retry handling

You can retry this API.

Until you acknowledge messages from the subscriber, this API returns messages every time you call this API.

## Request

### `subscriberID` parameter (required)

ID of the subscriber.

You need to create subscription beforehand.

### `channelID` parameter (required)

ID of the channel that the subscriber belongs to.

### `timeout` parameter (optional but recommended)

With this option, server performs long-polling. Long polling is an API interface to await for new message arrival. Server await new message arrival if no messages available on the start of this API.

Note that we **strongly recommend long-polling rather than short-polling**, it offers low latency and system load efficiency.

Format of the duration is [golang ParseDuration](https://golang.org/pkg/time/#ParseDuration) syntax (e.g. `1h30m`).

### `max` parameter (optional, default `64`)

If there are so many new messages, server *may not* return specified count of messages.

Note that server could return some more messages than this value.

## Response

Returns HTTP `200` with `application/json` response body if success.

Example:

```javascript
{
  "channelID": "cc457b533ad54a47b0facc44daf51ad8",
  "messages": [
    {
      "messageID": "my-first-message",
      "content": /* any JSON */
    }
  ],
  "ackHandle": "B4CF3208,5139-4F71-B260,F7519680A886",
  "moreMessages": true
}
```

### `channelID` (string, always returned)

ChannelID of the channel you sent to, exactly same as request parameter.

### `message` (list, always returned)

List of messages received and not acknowledged yet.

Length of this list depends on following cases:

- 0 length: there are no messages
- 1 or more: there are some messages, list contains all or subset of the messages
  - Note that this list may not contain all of the messages.

### `message[n].messageID` (string, always returned)

ID of the message given by [message publish API](../publish.md).

### `message[n].content` (any JSON, always returned)

Content of the message given by [message publish API](../publish.md).

### `ackHandle` (string, returned if there are one or more messages)

A token to acknowledge (remove) received messages from the subscriber.

Use DELETE API (described below) to delete received messages with this token otherwise you will receive same messages again.

Note: `ackHandle` is not valid after any DELETE API call. You should hold only last `ackHandle` you received.

### `moreMessages` (boolean, always returned)

If there are more messages, true.


# DELETE `/channel/{channelID}/subscription/polling/{subscriberID}/message?ackHandle={ackHandle}`

Acknowledge (remove) received message from the subscriber.

You must acknowledge messages that successfully received.

Note:

- Do not confuse with `DELETE /channel/{channelID}/subscription/polling/{subscriberID}`, it deletes subscriber itself.
- With some storage type, deleted message could be re-sent even if you call this endpoint in rare case (e.g. due to data loss of Redis Cluster failover).
  - By design DSPS server put importance on prevent message lost (misfire) rather than duplicate message delivery.


## Retry handling

You can retry this API.

This API success even if specified messages had been already deleted from the subscriber.

## Request

### `subscriberID` parameter (required)

ID of the subscriber.

### `channelID` parameter (required)

ID of the channel that the subscriber belongs to.

### `ackHandle` parameter (required)

The string returned from the polling endpoint.

Do not pass old `ackHandle`, always use `ackHandle` of the latest polling.

### Request body

No need to send request body to this API.

## Response

Returns HTTP `204` (No Content) if success.
