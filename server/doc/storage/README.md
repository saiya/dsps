# DSPS server storages

DSPS server needs storage to keep messages for (re)sending.

Following types are supported:

- [onmemory](./onmemory.md) : Default, but *not recommended for production*
- [redis](./redis.md) : Use Redis to store messages

See each documents for more detail.

To know how to configure DSPS server, see [configuration file documentation](../config.md).

## What storage used for

DSPS stores some information to the storage such as...

- Unconsumed messages queue of each subscribers
  - DSPS (re-)send messages until subscribers acknowledge it
- Set of [revoked JWT](../interface/admin/revoke_jwt.md)

## Multiple storages

You can configure multiple storages.

If multiple given, DSPS server write information into all storages thus it increase durability.

Caution: do not change storage ID after deployment, otherwise it may cause data lost.

```yaml
storage:
  myRedisA:  # <-- Do not modify this ID after deploy
    redis:
      singleNode: 'my-redis-server-host-1:6379'
  myRedisB:  # <-- Do not modify this ID after deploy
    redis:
      singleNode: 'my-redis-server-host-2:6379'
```
