# Load test experiment on macbook

## Environment

- MacBook Pro (mac OS 10.15.7, mem 16 GB, Core i7-8557U 8 threads)
  - [k6 (load test tool)](https://k6.io/) v0.29.0
  - DSPS server (Darwin/amd64 build, source code revision `?????`)
  - 2x [redis:6.0.9](https://hub.docker.com/_/redis) server processes

## Test scenario

- Load test script runs... 
  - `1` publisher per channel
  - `3` subscribers per channel
- Publisher publishes `3` messages with `1` second interval
- Subscriber call message acknowledge API after `0.01` sec, then polling next message after `0.01` sec sleep.

For detail, see [../../loadtest.k6.js](../../loadtest.k6.js).

## Key results

Unit of `ww/xx/yy/zz` is `med/90 percentile/95 percentile/99 percentile` in milliseconds.

`#` column shows total number of `channels (subscribers)`.

| #        | Msg delay | Publish API TTFB |
| -------- | --------- | ---------------- |
| 50 (150) | ...       | ...              |

Note that subscribers sleep `2 * 0.01` sec between cycles in this scenario. Message delay contains the sleep durations.

## Raw data

See [result](./result) directory for k6 output files.

Look [../../experiment-local.sh](../../experiment-local.sh) and [../../loadtest.k6.js](../../loadtest.k6.js) for experiment detail.

## Steps to reproduction

1. Increase `nofile` soft ulimit and `kern.ipc.somaxconn` sysctl value (macOS default is to small)
2. Run `CHANNELS=nnn TEST_RAMPUP_SEC=0 TEST_DURATION_SEC=30 ../../experiment-local.sh . output-name` 
    - Replace `CHANNEL=nnn` to desired channel count
    - Replace `output-name` to test pattern name (directory name)
    - See [../../loadtest.k6.js](../../loadtest.k6.js) for more available options
