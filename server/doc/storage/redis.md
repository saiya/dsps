# DSPS Redis storage

DSPS servers can store & share messages on Redis.

Multiple servers can share same channels & messages with using this storage setup. Client can connect to any server, and any server receives all messages of the channel. So that you can use load balancers without any special care with this storage.

## `storage.redis` configuration block

To setup redis storage, write `redis` section under `storage` configuration block.

Note: to understand configuration file, see [configuration guite](../config.md).

### Redis endpoint configuration

Redis storage implementation requires Redis Clusters or Redis nodes.

Note that Redis or Redis cluster could lose data when reboot/failover occurs.
To prevent data loss, write multiple storage configurations to use multiple Redis-es.
See [Multiple Storages](./README.md#multiple-storage) section of storage configuration document.

Each `redis` configuration requires exactly one of followings:

- `singleNode` (string): `host:port` (e.g. `'localhost:6379'`) strings point Redis
- `cluster` (list of string): Cluster endpoint list that is list of `host:port` points seed nodes

You must supply `singleNode` xor `cluster`. Should not supply both. If you use Redis Cluster, supply `cluster`. If you use simple Redis, supply `singleNode`.

```yaml
# ex. Simple (non-Cluster) Redis
storage:
  myRedisA:  # Keep in mind not to change this name ("myRedisA") after first deployment, otherwise causes data-loss.
    redis:
      singleNode: 'my-redis-server-host-1:6379'
  myRedisB:
    redis:
      singleNode: 'my-redis-server-host-2:6379'
```

```yaml
# ex. Redis Clusters
storage:
  myCluster1:
    redis:
      cluster:
        # List of nodes of a cluster
        - 'a-node-of-cluster-1:6379'
        - 'another-node-of-cluster-1:6379'
  myCluster2:
    redis:
      cluster:
        # List of nodes of another cluster
        - 'a-node-of-cluster-2:6379'
        - 'another-node-of-cluster-2:6379'
```

### Other Redis storage options

Each Redis storage option can take additional options:

```yaml
# ex. db & timeout configurations
storage:
  myRedis1:
    redis:
      singleNode: 'my-redis-server-host-1:6379'
      db: 0  # database number of the Redis
      timeout:
        connect: 5s
        read: 5s
        write: 5s
  myRedis2:
    redis:
      cluster:
        # List of nodes of a cluster
        - 'a-node-of-cluster-1:6379'
        - 'another-node-of-cluster-1:6379'
      db: 1
      timeout:
        connect: 5s
        read: 5s
        write: 5s
```

Configuration items:

- `username` (string, default `""`): Username of Redis authentication
- `password` (string, default `""`): Password of Redis authentication
- `db` (number, optional, default `0`): Database number of the Redis
  - Note: ignored if using redis cluster because it does not support database number
- `scriptReloadInterval` (duration, default `5m`): Interval of [SCRIPT LOAD](https://redis.io/commands/script-load) to preload Redis lua scripts
- `timeout.connect` (duration, default `5s`): Timeout to connect to the Redis
- `timeout.read` (duration, default `5s`): Timeout to wait response from the Redis
- `timeout.write` (duration, default `5s`): Timeout of request write operation to the Redis
- `retry.count` (integer, default: `3`): 0 to disable retry, 1 to retry only once, ...
- `retry.interval` (duration string, default: `500ms`): Retry base interval
- `retry.intervalJitter` (duration string, default: `200ms`): Max range of the retry interval randomization, plus or minus to the interval
- `connection.max` (integer, default: `max(1024, NumCPU * 64)`): Max connections between DSPS server and the Redis
- `connection.min` (integer, default: `NumCPU * 16`): Minimum connections to keep-alive to reduce connect round-trip overhead
- `connection.maxIdleTime` (duration string, default: `5m`): Max idle time to keep-alive connections
