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

| #              | Msg/sec  | Msg delay [ms]<br />(incl. 20ms sleep *1) | Publish API TTFB [ms]    | Acknowledge API TTFB [ms] |
| -------------- | -------- | ----------------------------------------- | ------------------------ | ------------------------- |
| ` 50` ( `150`) | ` 406.7` | `22,  27,  29,   37`                      | `1.2,  2.9,  3.9,  10.7` | `0.8,  2.2,  3.0,   5.6`  |
| `100` ( `300`) | ` 811.6` | `22,  26,  29,   64`                      | `0.9,  2.8,  4.2,  30.3` | `0.7,  2.0,  2.9,   9.4`  |
| `200` ( `600`) | `1608.1` | `22,  33,  42,  329`                      | `2.0,  6.8, 10.3, 179.4` | `1.0,  4.1,  7.1,  16.6`  |
| `300` ( `900`) | `2278.2` | `23,  41,  62, 1470.3`                    | `1.6,  7.5, 12.3, 450.7` | `1.1,  5.3,  9.2,  22.9`  |
| `400` (`1200`) | `2636.6` | `36, 215, 326, 2616.1`                    | `7.9, 61.2, 90.3, 830.4` | `5.2, 49.5, 70.5, 115.9`  |

- `#` column shows total number of `channels (subscribers)`.
- `Msg/sec` column shows average count of received messages per second.
- Values of 4-tuple are `median, 90 percentile, 95 percentile, 99 percentile`.
- `TTFB` = [Time to first byte](https://en.wikipedia.org/wiki/Time_to_first_byte)

`*1`: Note that subscribers sleep `20` ms (`2 * 0.01` sec) for each API cycles in this scenario. `Msg delay` (duration from messsage JSON creation to received) contains the sleep durations.

## Raw data

See [result](./result) directory for k6 output files.

Look [../../experiment-local.sh](../../experiment-local.sh) and [../../loadtest.k6.js](../../loadtest.k6.js) for experiment detail.

## Steps to reproduction

1. Increase `nofile` soft ulimit and `kern.ipc.somaxconn` sysctl value (macOS default is to small)
2. Run `CHANNELS=nnn TEST_RAMPUP_SEC=0 TEST_DURATION_SEC=30 ../../experiment-local.sh . output-name` 
    - Replace `CHANNEL=nnn` to desired channel count
    - Replace `output-name` to test pattern name (directory name)
    - See [../../loadtest.k6.js](../../loadtest.k6.js) for more available options
