# Quick load test experiment on cloud infrastructure

INFORMATIONAL PURPOSES ONLY: system performance may vary due to various reasons, this data does NOT guarantee anything.

## Environment

- (DSPS server) ALB + AWS Fargate 1.3.0
  - Spread across 3 AZs
  - DSPS source code revision is `0107797b2aba844c1718954e15275c5d22fadb82`
- (Storage) 2x AWS ElastiCache Redis 6.0.5 on `cache.m5.large`
  - Pick up different AZ to averaging AZ difference
- (Load test server) AWS EC2 `c5.4xlarge` server
  - [k6 (load test tool)](https://k6.io/) v0.29.0

See each test scenario for spec of Fargate or Redis.

## Test scenario

- Load test script runs... 
  - `1` publisher per channel
  - `3` subscribers per channel
- Publisher publishes `3` messages with `1` second interval
- Subscriber call message acknowledge API after `0.01` sec, then polling next message after `0.01` sec sleep.

For detail, see [../../loadtest.k6.js](../../loadtest.k6.js).

To warm up enough, each scenario starts with a few channels then increase.

## Key results

### Summary

DSPS server shows CPU intensively in this scenario. 
It consumes almost `1 ms/request` Fargate vCPU time in average for each publish/fetch/acknowledge API calls.
Load test shows good scaling result in proportion to total vCPU count, so that scale out works well.

DSPS also consumes Redis CPU resource, it consumes almost `0.1 ms/request` vCPU time (of `cache.m5.large` node) for each API calls.
If expected workload exceeds this rate, need to use Redis Cluster or horizontally split DSPS servers by channel to scale out.

### 5 vCPU scenario

- DSPS server: 5x Fargate tasks
  - 1 vCPU
  - 2 GB memory (more than necessary, due to Fargate limitation)
- Storage: 2x `cache.m5.large` ElastiCache servers
  - 2 vCPU, 6.38 GB memory (more than necessary)
    - Using `m5` instance just because `t3` shows unstable performance

Saturated at `5300` HTTP req/sec, means each HTTP request consumes 1 milliseconds. 

| #               | msg/sec  | HTTP req/sec | Msg delay [ms]<br>(incl. 20ms sleep `*1`) | Publish API TTFB [ms] | Acknowledge API TTFB [ms] | Fargate CPU Usage | Redis CPU Usage (`*2`) |
| --------------- | -------- | ------------ | ----------------------------------------- | --------------------- | ------------------------- | ----------------- | ---------------------- |
| `100` (`300`)   | `762.8`  | `1274.9`     | `28, 36, 37`                              | `6.8, 7.8, 8.1`       | `4.8, 5.5, 6.2`           | 36-37%            | 6%                     |
| `200` (`600`)   | `1524.3` | `2547.2`     | `28, 37, 38`                              | `6.8, 7.8, 8.2`       | `4.8, 5.7, 6.3`           | 64-65%            | 10-11%                 |
| `300` (`900`)   | `2284.0` | `3814.8`     | `29, 38, 41`                              | `6.8, 8.2, 11.2`      | `4.8, 6.3, 8.4`           | 84-89%            | 14-17%                 |
| `400` (`1200`)  | `3035.1` | `5113.7`     | `32, 69, 84`                              | `7.2, 27.7, 42.3`     | `5.2, 23.4, 37.0`         | 100%              | 17-21%                 |
| `500` (`1500`)  | `3075.8` | `5384.1`     | `95, 731, 920`                            | `7.7, 567.7, 616.8`   | `5.5, 316.7, 392.0`       | 100%              | 18-20%                 |
| `600` (`1800`)  | `3118.6` | `5337.0`     | `248, 1170, 1504`                         | `7.7, 958.9, 1019.5`  | `5.5, 531.1, 613.2`       | 100%              | 17-19%                 |
| `700` (`2100`)  | `3240.3` | `5421.3`     | `389, 1430, 1854`                         | `8.0, 1223.8, 1292.7` | `6.2, 658.4, 766.3`       | 100%              | 17-18%                 |
| `800` (`2400`)  | `3341.1` | `5390.5`     | `658, 1974, 2504`                         | `7.7, 1618.0, 1723.7` | `5.5, 834.7, 973.7`       | 100%              | 17-19%                 |
| `900` (`2700`)  | `3439.7` | `5386.1`     | `835, 2343, 2999`                         | `7.7, 1911.6, 2069.8` | `5.6, 990.7, 1149.0`      | 100%              | 17-18%                 |
| `1000` (`3000`) | `3515.0` | `5394.6`     | `951, 2602, 3329`                         | `8.0, 2164.0, 2362.0` | `6.1, 1165.7, 1292.0`     | 100%              | 18-28%                 |

- Values of 3-tuple are `median, 90 percentile, 95 percentile`.
- `TTFB` = [Time to first byte](https://en.wikipedia.org/wiki/Time_to_first_byte)
- `#` column shows total number of `channels (subscribers)`.
- `msg/sec` column shows average count of received messages per second.
- `HTTP req/sec` column shows HTTP request rate to DSPS server (note that this experiment runs only 1 server process).

`*1`: Note that subscribers sleep `20` ms (`2 * 0.01` sec) for each API cycles in this scenario. `Msg delay` (duration from messsage JSON creation to received) contains the sleep durations.
`*2`: Because Redis is single thread system, it could not utilize CPU resource of `cache.m5.large` (2 vCPU) over 50%

### 10 vCPU scenario

- DSPS server: 10x Fargate tasks
  - 1 vCPU
  - 2 GB memory (more than necessary, due to Fargate limitation)
- Storage: 2x `cache.m5.large` ElastiCache servers
  - 2 vCPU, 6.38 GB memory (more than necessary)
    - Using `m5` instance just because `t3` shows unstable performance

Saturated between `8955.7` and `10328.6` HTTP req/sec, means each request consumes 1 milliseconds. 

| #               | msg/sec  | HTTP req/sec | Msg delay [ms]<br>(incl. 20ms sleep `*1`) | Publish API TTFB [ms] | Acknowledge API TTFB [ms] | Fargate CPU Usage | Redis CPU Usage (`*2`) |
| --------------- | -------- | ------------ | ----------------------------------------- | --------------------- | ------------------------- | ----------------- | ---------------------- |
| `100` (`300`)   | `764.5`  | `1277.8`     | `28, 36, 37`                              | `6.8, 7.8, 8.1`       | `5.0, 6.0, 6.3`           | 23-25%            | 6-7%                   |
| `200` (`600`)   | `1525.5` | `2549.5`     | `28, 36, 37`                              | `6.8, 7.8, 8.2`       | `4.9, 5.9, 6.3`           | 41-42%            | 11%                    |
| `300` (`900`)   | `2287.7` | `3823.5`     | `28, 36, 37`                              | `6.7, 7.9, 8.2`       | `4.9, 5.9, 6.4`           | 54-56%            | 14-16%                 |
| `400` (`1200`)  | `3048.4` | `5094.2`     | `28, 36, 38`                              | `6.9, 8.0, 8.4`       | `4.9, 6.2, 6.5`           | 69%               | 19-21%                 |
| `500` (`1500`)  | `3807.0` | `6359.5`     | `28, 37, 38`                              | `6.9, 8.1, 8.7`       | `4.9, 6.2, 6.7`           | 79-82%            | 21-25%                 |
| `600` (`1800`)  | `4567.6` | `7636.4`     | `28, 38, 44`                              | `6.9, 8.7, 13.2`      | `5.0, 6.6, 9.8`           | 91-93%            | 27-29%                 |
| `700` (`2100`)  | `5330.7` | `8955.7`     | `29, 50, 63`                              | `7.0, 15.1, 26.7`     | `5.1, 11.4, 21.0`         | 98-99%            | 29-32%                 |
| `800` (`2400`)  | `6007.6` | `10328.6`    | `37, 105, 138`                            | `7.4, 54.9, 73.8`     | `5.4, 42.8, 59.4`         | 100%              | 33-36%                 |
| `900` (`2700`)  | `5857.0` | `10199.6`    | `44, 720, 853`                            | `7.3, 112.8, 728.1`   | `5.4, 78.5, 397.3`        | 100%              | 32-34%                 |
| `1000` (`3000`) | `5923.6` | `10124.8`    | `58, 1059, 1196`                          | `7.4, 254.6, 1085.7`  | `5.4, 155.3, 555.7`       | 100%              | 32-35%                 |


### 20 vCPU scenario

- DSPS server: 20x Fargate tasks
  - 1 vCPU
  - 2 GB memory (more than necessary, due to Fargate limitation)
- Storage: 2x `cache.m5.large` ElastiCache servers
  - 2 vCPU, 6.38 GB memory (more than necessary)
    - Using `m5` instance just because `t3` shows unstable performance

Redis server exhausts it's single CPU core spec (50% cpu usage `*1`) between `10174.1` and `11416.5`  HTTP req/sec, means each request consumes 0.1 milliseconds. 

| #               | msg/sec  | HTTP req/sec | Msg delay [ms]<br>(incl. 20ms sleep `*1`) | Publish API TTFB [ms] | Acknowledge API TTFB [ms] | Fargate CPU Usage | Redis CPU Usage (`*2`) |
| --------------- | -------- | ------------ | ----------------------------------------- | --------------------- | ------------------------- | ----------------- | ---------------------- |
| `100` (`300`)   | `763.1`  | `1275.5`     | `28, 36, 37`                              | `6.8, 7.9, 8.1`       | `4.9, 5.7, 6.2`           | 32%               | 8-9%                   |
| `200` (`600`)   | `1523.0` | `2545.4`     | `28, 36, 37`                              | `6.9, 7.9, 8.1`       | `4.9, 5.7, 6.2`           | 41-43%            | 13-15%                 |
| `300` (`900`)   | `2287.1` | `3822.2`     | `28, 36, 37`                              | `6.9, 7.9, 8.1`       | `4.9, 5.8, 6.2`           | 50-55%            | 18-20%                 |
| `400` (`1200`)  | `3044.7` | `5088.2`     | `28, 36, 37`                              | `6.8, 7.8, 8.1`       | `4.9, 5.8, 6.3`           | 62-63%            | 23-25%                 |
| `500` (`1500`)  | `3804.7` | `6358.7`     | `28, 36, 38`                              | `6.8, 7.9, 8.3`       | `4.9, 5.9, 6.4`           | 72-73%            | 27-31%                 |
| `600` (`1800`)  | `4570.9` | `7639.1`     | `28, 36, 38`                              | `6.9, 7.9, 8.3`       | `4.9, 5.9, 6.4`           | 79-80%            | 31-35%                 |
| `700` (`2100`)  | `5329.7` | `8908.3`     | `28, 37, 39`                              | `6.9, 8.1, 8.6`       | `5.0, 6.1, 6.7`           | 80-85%            | 32-39%                 |
| `800` (`2400`)  | `6078.8` | `10174.1`    | `29, 39, 49`                              | `7.0, 8.9, 13.6`      | `5.0, 6.7, 10.4`          | 80-86%            | 36-43%                 |
| `900` (`2700`)  | `6792.9` | `11416.5`    | `30, 55, 78`                              | `7.2, 15.4, 28.1`     | `5.2, 11.7, 22.2`         | 86-88%            | 42-48%                 |
| `1000` (`3000`) | `7394.4` | `12488.7`    | `32, 80, 130`                             | `7.3, 28.8, 51.0`     | `5.3, 21.3, 39.3`         | 88-91%            | 45-48%                 |


## Raw data

See `result-*` directories for k6 output files.

Look [../../experiment-local.sh](../../experiment-local.sh) and [../../loadtest.k6.js](../../loadtest.k6.js) for experiment detail.

## Steps to reproduction

1. Make AWS environment with ALB, Fargate service (DSPS server), ElastiCache Redis instances, and EC2 instance to run k6.
2. Increase `nofile` ulimit of the EC2 instance
3. Run `CHANNELS=nnn TEST_RAMPUP_SEC=30 TEST_DURATION_SEC=180 BASE_URL='https://url-to-the-dsps' ../../experiment-remote.sh . output-name` 
    - Replace `CHANNEL=nnn` to desired channel count
    - Replace `output-name` to test pattern name (directory name)
    - Replace `https://url-to-the-dsps` to the base URL of the DSPS, should point to ALB
    - See [../../loadtest.k6.js](../../loadtest.k6.js) for more available options
