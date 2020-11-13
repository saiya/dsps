# DSPS (Durable & Simple PubSub) server

## What is DSPS ?

DSPS is a PubSub system that provides following advantages:

- Durable message passing (no misfire)
- Simple messaging interface (even `curl` is enough to communicate with the DSPS server)

Read [DSPS README](../README.md) first to grasp.

## Recommended production setup

1. Should setup proper `storage` (see [storage configuration document](./doc/storage/README.md))
    - **default `onmemory` storage is not suitable for production use**
2. Should run multiple servers to keep high availability
3. Should increase file descriptor limit

## DSPS server configuration

DSPS server can load configuration file.

To pass configuration file, use command line argument: `./dsps path-to-config-file.yml`

See [configuration document](./doc/config.md) to how to write the configuration file.

## Message persistency

DSPS stores messages to ensure durability (message resending, deduplication, ...).

See [storage document](./doc/storage/README.md) to available storage implementations.

## To develop DSPS server locally

DSPS server requires following tools for development:

- GNU make
- go (see [go.mod](./go.mod) for desired version)
- [golangci-lint](https://golangci-lint.run/) to run lint locally
