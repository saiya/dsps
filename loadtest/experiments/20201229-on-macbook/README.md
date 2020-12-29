# Quick load test experiment on single macbook

## Environment

- MacBook Pro (mac OS 10.15.7, mem 16 GB, Core i7-8557U 8 threads)
  - [k6 (load test tool)](https://k6.io/) v0.29.0
  - DSPS server (Darwin/amd64 build, source code revision `804907ea773b8cbeba55fee4594be2bf79b68a43`)
  - 2x [redis:6.0.9](https://hub.docker.com/_/redis) server processes

## Test scenario

- Load test script runs... 
  - `1` publisher per channel
  - `3` subscribers per channel
- Publisher publishes `3` messages with `1` second interval
- Subscriber call message acknowledge API after `0.01` sec, then polling next message after `0.01` sec sleep.

For detail, see [../../loadtest.k6.js](../../loadtest.k6.js).

## Key results

Keep in mind: everything runs on single machine in this experiment (**load testing tool occupies most of CPU so that it should cause latency**).

| #              | msg/sec  | HTTP req/sec | Msg delay [ms]<br>(incl. 20ms sleep *1) | Publish API TTFB [ms]    | Acknowledge API TTFB [ms] | Total CPU Usage<br>Sys + User |
| -------------- | -------- | ------------ | --------------------------------------- | ------------------------ | ------------------------- | ----------------------------- |
| ` 50` ( `150`) | ` 406.7` | ` 689.1`     | `22,  27,  29,   37`                    | `1.2,  2.9,  3.9,  10.7` | `0.8,  2.2,  3.0,   5.6`  | 10% + 15-18%                  |
| `100` ( `300`) | ` 811.6` | `1374.8`     | `22,  26,  29,   64`                    | `0.9,  2.8,  4.2,  30.3` | `0.7,  2.0,  2.9,   9.4`  | 10% + 20%                     |
| `200` ( `600`) | `1608.1` | `2714.1`     | `22,  33,  42,  329`                    | `2.0,  6.8, 10.3, 179.4` | `1.0,  4.1,  7.1,  16.6`  | 15% + 35-50%                  |
| `300` ( `900`) | `2278.2` | `3838.5`     | `23,  41,  62, 1470.3`                  | `1.6,  7.5, 12.3, 450.7` | `1.1,  5.3,  9.2,  22.9`  | 20% + 50-60%                  |
| `400` (`1200`) | `2636.6` | `4524.1`     | `36, 215, 326, 2616.1`                  | `7.9, 61.2, 90.3, 830.4` | `5.2, 49.5, 70.5, 115.9`  | 30% + 60-70%                  |

- Values of 4-tuple are `median, 90 percentile, 95 percentile, 99 percentile`.
- `TTFB` = [Time to first byte](https://en.wikipedia.org/wiki/Time_to_first_byte)
- `#` column shows total number of `channels (subscribers)`.
- `msg/sec` column shows average count of received messages per second.
- `HTTP req/sec` column shows HTTP request rate to DSPS server (note that this experiment runs only 1 server process).
- `Total CPU Usage` column shows CPU usage of the machine, 100% means all CPU cores running in top gear.

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
