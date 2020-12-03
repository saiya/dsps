# DSPS - Durable & Simple PubSub

[![MIT License](https://img.shields.io/badge/LICENSE-MIT-brightgreen)](./LICENSE)
[![Server Test](https://github.com/saiya/dsps/workflows/Server%20Test/badge.svg?1)](https://github.com/saiya/dsps/actions?query=workflow%3A%22Server+Test%22)
[![Codecov](https://codecov.io/gh/saiya/dsps/branch/main/graph/badge.svg?token=DSSOWMB60X)](https://codecov.io/gh/saiya/dsps)
[![Go Report Card](https://goreportcard.com/badge/github.com/saiya/dsps?1)](https://goreportcard.com/report/github.com/saiya/dsps)
[![DockerHub saiya/dsps](https://img.shields.io/badge/dockerhub-saiya%2Fdsps-blue)](https://hub.docker.com/r/saiya/dsps/tags)

DSPS is a PubSub system that provides following advantages:

- Durable message passing (no misfire)
- Simple messaging interface (even `curl` is enough to communicate with the DSPS server)

DSPS supports message buffering, resending, deduplication, ordering, etc. to secure your precious messages.

DSPS server supports intuitive interfaces such as HTTP short polling, long polling, outgoing webhook, etc.

Note that DSPS does **NOT** aim to provide followings:

- Very low latency message passing
  - DSPS suppose milliseconds latency tolerant use-case
- Too massive message flow rate comparing to your storage spec
  - DSPS temporary stores messages to resend message
- Warehouse to keep long-living message
  - DSPS aim to provide message passing, not archiving message


# 3 minutes to getting started with DSPS

```sh
# Download & run DSPS server
docker run -i -p 3000:3000/tcp saiya/dsps:latest

#
# ... Open another terminal window to run following tutorial ...
#

CHANNEL="my-channel"
SUBSCRIBER="my-subscriber"

# Create a HTTP polling subscriber.
curl -w "\n" -X PUT "http://localhost:3099/channel/${CHANNEL}/subscription/polling/${SUBSCRIBER}"

# Publish message to the channel.
curl -w "\n" -X PUT -H "Content-Type: application/json" \
  -d '{ "hello": "Hi!" }' \
  "http://localhost:3099/channel/${CHANNEL}/message/my-first-message"

# Receive messages with HTTP long-polling.
# In this example, this API immediately returns
# because the subscriber already have been received a message.
curl -w "\n" -X GET "http://localhost:3099/channel/${CHANNEL}/subscription/polling/${SUBSCRIBER}?timeout=30s&max=64"

ACK_HANDLE="<< set string returned in the above API response >>"

# Cleanup received messages from the subscriber.
curl -i -X DELETE \
  "http://localhost:3099/channel/${CHANNEL}/subscription/polling/${SUBSCRIBER}/message?ackHandle=${ACK_HANDLE}"
```

Tips: see [server interface documentation](./server/doc/interface) for more API interface detail.

## Message resending - you have to DELETE received messages

You may notice that your receive same messages every time you GET the subscriber endpoint. Because DSPS resend messages until you explicitly delete it to prevent message loss due to network/client error.

The way to delete message depends on the [subscriber type](./server/doc/interface/subscribe/README.md). For example, HTTP polling subscriber (used in above example) supports HTTP DELETE method to remove messages from the subscriber.

# To know more

- [Detail of the DSPS server](./server/README.md)
  - Before running DSPS in production, recommend to look this document
- [API interface of the DSPS server](./server/doc/interface)
- [JavaScript / TypeScript client](./client/js/README.md)
- [Security & Authentication](./server/doc/security.md)
