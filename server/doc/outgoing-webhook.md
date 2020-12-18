# Outgoing webhook

With outgoing webhook setup, DSPS server calls given HTTP(S) endpoint for each incoming message.

It is especially useful to deliver messages to your web API servers.

## Durability

Keep in mind that outgoing webhook is volatile operation.

In case of webhook destination failed to receive messages even for retries, DSPS server just give up.

If you want to keep messages certainly, consider to use [subscription API](./interface/subscribe) to pull messages from DSPS server rather than outgoing webhooks that push messages from DSPS server.

### Retry settings

DSPS server automatically retry outgoing webhook calls.

See [channels.webhooks configuration block](./config.md#outgoing-webhook) for how to tune retry.

## Outgoing webhook request

Outgoing webhook sends HTTP(S) request as described below.

### Request method

HTTP method of the request is `PUT` by default, can change on [channels.webhooks configuration block](./config.md#outgoing-webhook).

### Request headers

You can set HTTP headers freely on [channels.webhooks configuration block](./config.md#outgoing-webhook).

### Request body

Body of the outgoing request is `application/json` ([RFC 8259](https://tools.ietf.org/html/rfc8259)). [Text encoding of the JSON is UTF-8](https://tools.ietf.org/html/rfc8259#section-8.1).

Below describes JSON structure with using [TypeScript typing syntax](https://www.typescriptlang.org/docs/handbook/intro.html):

```ts
type OutgoingWebhookBody = {
  /** Fixed string that marks outgoing-webhook. */
  type: "dsps.channel.outgoing-webhook";

  /** ID of the channel */
  channelID: string;

  /** ID of the message, given by message sender. */
  messageID: string;

  /** Content of the message */
  content: any;
}
```

Note that webhook receiver MUST ignore unknown properties of the JSON.
Future version of DSPS server could put more information in the body.


## Outgoing webhook response

### HTTP status code

Webhook receiver should respond 2xx HTTP status code.

If response has following code, DSPS immediately give up: 

- 400 to 418, expect for 400, 404, 408, 409
  - Because misconfigured proxy tend to return 404 or 400, retry them.
- 426
- 431
- 451
- 501 ("Not Implemented" error)

Otherwise DSPS retries.

