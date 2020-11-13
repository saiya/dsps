# DSPS on-memory storage

On-memory is default but NOT production-use storage mode.

On-memory storage is designed for local testing purpose (e.g. suitable for CI automated test).
This storage stores all data on process's memory.
So that on-memory storage does not affect any local & remote resources.

This storage does NOT offer followings:

- Durability - lose data when server process ends
- Server redundancy - cannot share data across multiple server processes

## `storage.onmemory` configuration block

Currently no configuration item for this storage type.

```yaml
# Example to explicitly use on-memory storage
storage:
  myVolatileStorage:
     onmemory: {}
```
