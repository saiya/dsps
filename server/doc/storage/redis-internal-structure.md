# Internal structure of Redis storage implementation

This documentation describes internal data structure of the Redis storage implementation.

To know how to use Redis storage, read [Redis storage usage documentation](./redis.md) rather than this document.

## Key-value I/O example scenario

Assume you created 1 channel named "chX" with a subscribers named "sA" and "sB" at t=1:

| Key       | Value | TTL        |
| --------- | ----- | ---------- |
| chX.clock | 0     | 1 + expire |
| chX.r.sA  | 0     | 1 + expire |
| chX.r.sB  | 0     | 1 + expire |

`expire` means message expiry duration of the channel setting.

And create "sB" at t=2:

| Key       | Value | TTL        |
| --------- | ----- | ---------- |
| chX.clock | 0     | 1 + expire |
| chX.r.sA  | 0     | 1 + expire |
| chX.r.sB  | 0     | 2 + expire |

Then publish a message ID "msg123" with content `{ "text": "hello!" }"` at t=3:

| Key            | Value                                                 | TTL        |
| -------------- | ----------------------------------------------------- | ---------- |
| chX.clock      | 1                                                     | 3 + expire |
| chX.r.sA       | 0                                                     | 1 + expire |
| chX.r.sB       | 0                                                     | 2 + expire |
| chX.m.1        | `{ "id": "msg123", "content": { "text": "hello!" } }` | 3 + expire |
| chX.mid.msg123 | 1                                                     | 3 + expire |

Publish operation increments clock of the channel (`chX.clock`), and put message with key `chX.m.{clock}` (in this example `chX.m.1`).

Also put `chX.mid.msg123` for deduplication, further publish operations look this key to dedup.

Now subscriber receives "msg123" when they fetch new messages.
subscriber look up messages that has larger clock than subscriber's clock (subscriber's clock = `chX.r.sA` or `chX.r.sB`).

At this timing, `receiptHandle` that sent to clients encodes clock=1 because subscriber receives messages until clock=1.

After "sA" subscriber deletes message with `receiptHandle` (that represents clock=1) at t=4:

| Key            | Value                                                 | TTL        |
| -------------- | ----------------------------------------------------- | ---------- |
| chX.clock      | 1                                                     | 3 + expire |
| chX.r.sA       | 1                                                     | 4 + expire |
| chX.r.sB       | 0                                                     | 2 + expire |
| chX.m.1        | `{ "id": "msg123", "content": { "text": "hello!" } }` | 3 + expire |
| chX.mid.msg123 | 1                                                     | 3 + expire |

On this deletion operation, subscriber "sA" advances it's clock to 1 because of the  `receiptHandle` that represents clock=1.

After deletion, subscriber "sA" will not receive "msg123" anymore because `chX.r.sA` is not smaller than clock of `chX.m.1`.

In contrast, "sB" still receive "msg123". 

## Inside of publish operation

As shown in above scenario, publish operation need some I/O to Redis:

1. `{channel}.mid.{message-id}` : Read this value to deduplication, and put with TTL
    - If this key already exists, exit publish operation because message ID duplicated (should be caused by retry).
2. `{channel}.clock` : Increment & increase TTL
3. `{channel}.m.{clock}` : Put with TTL

Above operations must be done atomic, otherwise subscribers may look inconsistent state.

To prevent this problem, Redis storage implementation uses Lua scripting to perform atomic operation. Because this operation is deterministic, script is compatible with Redis Cluster.

## Inside of fetch operation

Fetch operation simply iterate messages (`{channel}.m.{clock}` keys) that has clock *larger than* clock of the subscriber `{channel}.r.{subscriber}` and *equal to or smaller than* clock of the channel `{channel}.clock`.

Note that iteration must consider clock overflow (described later in this document).

Those I/O does not need atomicity.

## Inside of delete operation

Delete operation requires some I/O:

1. `{channel}.r.{subscriber}` : Read this value to know current position of the subscriber
2. `{channel}.clock` : Read this value to check validity of the given `receiptHandle`
    - `receiptHandle` must between current position of the subscriber and the clock, otherwise invalid.
    - Note that comparison must consider clock overflow (described later in this document).
3. `{channel}.r.{subscriber}` : Overwrite this value to advance subscriber's clock

Above operations must be done atomic. So that this operation also use Lua scripting.

## Clock overflow handling

Because this storage implementation uses Lua scripting, safe integer range is from `-(2^53 - 1)` (inclusive) to `2^53 - 1` (inclusive).
Lua represents all numerics with double precision float, so that `2^53 + 1 == 2^53` in Lua world. This is reason why avoid integers with absolute value larger than `2^53 - 1`.

To increment channel clock from `2^31 - 1` (largest safe integer), use clock value `-(2^53 - 1)` (smallest safe integer).

If subscriber's clock is smaller than channel's clock, it means clock rounding. 

For example, if subscriber's clock is `2^53 - 3` and channel's clock is `-(2^53 - 1) + 1`, there are unread messages on following clocks: 

1. `2^53 - 2` (subscriber's clock + 1, starting point of unread messages)
2. `2^53 - 1`
3. `-(2^53 - 1)`
4. `-(2^53 - 1) + 1` (channel's clock, latest message's clock in the channel)
