# DSPS server configuration

DSPS server can load YAML based configuration file.
To pass configuration file, use command line argument: `./dsps path-to-config-file.yml`

```yaml
# Configuration file example

storages:  # see ./storage/README.md for more info
  myRedis1:
    redis:
      singleNode: 'localhost:6379'

http:
  port: 3099
  realIpHeader: 'X-Forwarded-For'
  trustedProxyRanges:
    - 192.168.0.0/16
  discloseAuthRejectionDetail: false

logging:
  category:
    auth: WARN  # Set to INFO if you want auth rejection logs
    "*": INFO
  attributes:
    machineID: my-machine-id

channels:
  - regex: 'chat-room-(?P<id>\d+)'
    expire: 15m
    webhooks:
      - url: 'http://localhost:3001/you-got-message/room/{{.channel.id}}'

admin:
  auth:
    bearer:
      - 'my-api-key'
```

## Words & definitions used in this document

- `regex string` means YAML string that constructs [golang compatible regular expression](https://golang.org/pkg/regexp/) (e.g. `'chat-room-(?P<id>\d+)'`)
- `template string` means YAML string that follows [golang Template](https://golang.org/pkg/text/template) syntax (e.g. `'http://localhost:3001/room/{{.channel.id}}'`)
- `duration string` means YAML string that follows [golang ParseDuration](https://golang.org/pkg/time/#ParseDuration) syntax (e.g. `'1h30m'`)

## storages configuration block

You can configure storages under `storages` block.

```yaml
# ex. Simple (non-Cluster) Redis
storage:
  myRedisA:  # Keep in mind that this name ("myRedisA") should not be changed, otherwise causes data-loss.
    redis:
      singleNode: 'my-redis-server-host-1:6379'
  myRedisB:
    redis:
      singleNode: 'my-redis-server-host-2:6379'
```

Configuration detail depends on type of the storage, see [storage document](./storage/README.md) for detail.

**Heads Up** : DSPS uses on-memory storage if no configuration given, should change it for production use.

## http configuration block

Configuration items under `http`:

- `port` (number, default `3000`): TCP port to listen for HTTP requests
  - You can override this value with `--port` command line option
- `listen` (string, optional): Listen string (e.g. ":3000"), this option overrides `port`
  - With [some security software](https://forum.eset.com/topic/22080-mac-firewall-issue-after-update-to-684000/), you may need to specify local IP as such as `127.0.0.1:3000` due to MTIM proxy problem.
- `pathPrefix` (string, optional): Prefix to add all endpoints
  - e.g. If `pathPrefix` is `/foo/bar`, endpoint `/probe/readiness` is served as `/foo/bar/probe/readiness`
- <a name="ipheader"></a> `realIpHeader` (string, optional): HTTP header name contains reliable IP address of the client
  - Note that `admin.auth.networks` rely on this configuration.
- `trustedProxyRanges` (list of string, optional): List of CIDR notation that trusted proxy lives in
  - `realIpHeader` only accepts header values from those ranges.
  - By default or if empty list given, allow [RFC 1918](https://tools.ietf.org/html/rfc1918) ranges `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` and [RFC 4193](https://tools.ietf.org/html/rfc4193) range `fc00::/7` and also `127.0.0.0/8` ([RFC 1122](https://tools.ietf.org/html/rfc1122#section-3.2.1.3)), `169.254.0.0/16` ([RFC 3927](https://tools.ietf.org/html/rfc3927)), `::1/128` and `fe80::/10` ([RFC 4291](https://tools.ietf.org/html/rfc4291)).
- `discloseAuthRejectionDetail` (boolean, default `false`): Show detail reason of 403 to clients, **do not enable on production**
- `idleTimeout` (duration string, default `1h30m`): Max duration to keep idle connection, should be larger than keep-alive duration of clients/loadbalancer.
- `readTimeout` (duration string, default `10s`): Max duration to read request from clients.
- `writeTimeout` (duration string, default `60s`): Max duration since end of request header reading until request processing completion
  - Note that server automatically add `longPollingMaxTimeout` to this value
- `longPollingMaxTimeout` (duration string, default `30s`): Max duration of the [long-polling requests](./interface/subscribe/polling.md#polling-get).
- `gracefulShutdownTimeout` (duration string, default `5s`): Timeout to await end of running requests.
- <a name="defaultHeaders"></a> `defaultHeaders` (string to string map, optional): Always send those response headers
  - Server send some headers by default, you can disable them by setting empty string as a value.

## <a name="logging"></a> logging configuration block

Configuration items under `logging`:

- `category` (string to string map, optional) Log level threshold for each `category` attribute of the log entries.
  - Available thresholds are `DEBUG`, `INFO`, `WARN`
  - `"*"` category controls default threshold and it's default is `INFO`
  - `--debug` command line option overrides this config completely
- `attributes` (string to string map, optional): Attributes set to every log records
  - Useful to set machine ID etcetera.

## <a name="telemetry"></a> telemetry configuration block

Configure `telemetry` block to enable tracing and metrics.

```yaml
telemetry:
  ot:
    tracing:
      enable: true
      sampling: 0.003
      batch:
        maxQueueSize: 2048
        timeout: 5s
        batchSize: 512
      attributes:
        host.name: hostname-of-the-instance
        my.attribute: "foo bar"
    exporters:
      stdout:
        enable: true
```

Configuration items under `telemetry.ot.tracing` ([OpenTelemetry](https://opentelemetry.io/) tracing):

- `enable` (boolean, default `false`): true to enable OpenTelemetry
- `sampling` (floating number, default `1.0`): Sampling ratio to capture or not capture traces
  - Note that if tracing propagated from upstream, this ratio is not applied
- `batch.maxQueueSize` (number, default `2048`): On-memory buffer size to buffer tracing spans
- `batch.timeout` (duration string, default `5s`): Maximum duration to keep trace in buffer for bulk transmission
- `batch.batchSize` (number, default `512`): Number of traces to submit at once.
- `attributes`: (string to any map): Attributes of resource to add to traces
  - See [official resource semantic conventions](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/resource/semantic_conventions/README.md) document for standard naming
- `exporters`: Setup tracing exporters, see below.

Configuration items under `telemetry.ot.exporters.stdout`:

- `enable` (boolean, default `false`): true to output traces to stdout
- `quantiles` (list of numbers, default `0.5, 0.9, 0.99`): quantiles for metrics sampling

Configuration items under `telemetry.ot.exporters.gcp`:

- `enableTrace` (boolean, default `false`): true to output traces to GCP Cloud Trace
- `projectID` (string, default `""`): Set non-empty string to specify GCP Project ID

## <a name="sentry"></a> sentry configuration block

Configure `sentry` block to enable [Sentry](https://sentry.io/welcome/) error monitoring tool.

DSPS server sends events such as outgoing webhook failure to the Sentry when you enable it.

```yaml
# To enable Sentry, you need to set SENTRY_DSN environment variable.

sentry:  # Fine-tuning configuration example
  serverName: my-server-name
  environment: my-production

  tags:
    my_tag: value
  context:
    my_context: value

  sampleRate: 1.0
  ignoreErrors:
    - "something .+"
  disableStacktrace: false
  hideRequestData: false

  flushTimeout: 15s
```

To enable sentry, you must set `SENTRY_DSN` environment variable. It's value should be [DSN string of the Sentry server](https://docs.sentry.io/product/sentry-basics/dsn-explainer/).

Configuration items under `telemetry.ot.tracing` ([OpenTelemetry](https://opentelemetry.io/) tracing):

- `serverName` (string, default is hostname): Server name attribute to send to sentry
- `environment` (string, optional): [Environment name](https://docs.sentry.io/product/sentry-basics/environments/) send to sentry
- `tags` (string to string map, optional): Set [tag](https://docs.sentry.io/platforms/go/enriching-events/tags/) values
- `contexts` (string to string map, optional): Set [context](https://docs.sentry.io/platforms/go/enriching-events/context/) values
- `sampleRate` (number, default `1.0`): Ratio of the sampling, between `0.0` to `1.0`.
- `ignoreErrors` (list of regex string): If an event matches to one or more of the regex, ignore them
- `disableStacktrace` (boolean, default `false`): If true, omit stacktrace.
- `hideRequestData` (boolean, default `false`): If true, do not send request body data.
- `flushTimeout` (duration string, default `15s`): Timeout to flush sentry events on application shutdown.

## <a name="channels"></a> channels configuration block

You can configure channels under `channels` block.

Each channel configuration must have `regex` to match with name of a channel.
If multiple configuration match with a channel, server merges them.

If you configure one or more `channels` blocks, DSPS server will reject unmatched channel name.
If `channels` configuration is empty, DSPS server automatically define `.+` channel configuration to accept any channel name.

Configuration items under `channels[n]`:

- `regex` (regex string, required): regex string to match with name of a channel
  - Must match with *entire string* of the channel name (no need to write `^` nor `$`).
  - You can use named group/subexp (e.g. `(?P<id>\d+)`). In the channel configuration, captured value of the group is visible to template strings under `.channel` (e.g. `{{.channel.id}}`).
- `expire` (duration string, default `30m`): DSPS server may discard inactive subscribers & messages after this duration
  - Duration counts from last access of the subscriber or sent time of the message.
  - DSPS may not resend after this expiration duration, so that this value must be larger than client's polling period if you polling.
  - If multiple channel configuration matches to a channel, largest value wins.
  - If outgoing webhook is configured, expire value must be larger than maximum webhook time includes webhook timeout and retry interval

### <a name="outgoing-webhook"></a> channels.webhooks configuration block

You can configure outgoing webhook to send messages from DSPS server to any HTTP(S) services.

DSPS server calls given HTTP(S) endpoint for each incoming message.

See [outgoing webhook document](./outgoing-webhook.md) for more info.

```yaml
# Webhook with retry configuration example
channels:
  - regex: 'chat-room-(?P<id>\d+)'
    # Must be larger than final retry attempt time
    expire: 15m
    webhooks:
      - method: PUT
        url: 'http://localhost:3001/you-got-message/room/{{.channel.id}}'
        timeout: 30s
        connection:
          max: 1024
          maxIdleTime: 3m
        retry:
          # Enable 3 retries as below:
          # 1st retry: 3 * 1.5^0 ± 1.5 = 3    ± 1.5 [sec] after first webhook attempt
          # 2nd retry: 3 * 1.5^1 ± 1.5 = 4.5  ± 1.5 [sec] after 1st retry
          # 3rd retry: 3 * 1.5^2 ± 1.5 = 6.75 ± 1.5 [sec] after 1st retry
          count: 3
          interval: 3s
          intervalMultiplier: 1.5
          intervalJitter: 1s500ms
        headers:
          User-Agent: My DSPS server
          X-Chat-Room-ID: '{{.channel.id}}'
```

If there are multiple webhooks, DSPS server calls them concurrently. Configuration order of the webhooks has no meaning.

Configuration item under `channels[n].webhooks`:

- `method` (string, default `PUT`): HTTP method to send.
- `url` (template string, required): Full URL to send message
- `timeout` (duration string, default: `30s`): Timeout of the webhook call
- `connection.max` (integer, default: `1024`): Max connections between DSPS server and webhook target
  - This configuration also controls of `Transport.MaxIdleConnsPerHost`
- `connection.maxIdleTime` (duration string, default: `3m`): Max idle time to keep-alive connections
- `retry.force` (boolean, default: `false`): true to retry any errors (such as `404 Not Found`)
- `retry.count` (integer, default: `3`): 0 to disable retry, 1 to retry only once, ...
- `retry.interval` (duration string, default: `3s`): Retry base interval
- `retry.intervalMultiplier` (float, default: `1.5`): Exponential backoff factor, multiply to the previous interval
- `retry.intervalJitter` (duration string, default: `1s500ms`): Max range of the retry interval randomization, plus or minus to the resulted interval
- `headers` (string to template string map, optional): HTTP headers to set for each outgoing requests
- `maxRedirects` (number, default `10`): Max count of redirects to follow.

### <a name="jwt"></a> channels.jwt configuration block

To protect endpoints, can validate signed [JSON Web Tokens (JWT, RFC 7519)](https://jwt.io/).

```yaml
channels:
  - regex: 'chat-room-(?P<id>\d+)'
    jwt:
      iss:
        - https://issuer.example.com/issuer-url
      aud:
        - https://my-service.example.com/
      keys:
        RS256:
          - path/to/public-key-file.pem
      claims:
        chatroom: '{{.channel.id}}'
        role:
          - 'admin'
          - 'user'
      clockSkewLeeway: 5m
```

If this configuration present on the channel, clients must present valid JWT with `Authorization: Bearer <jwt>` request header for every API call.

In addition, you can revoke JWT with [revocation API](./interface/admin/revoke_jwt.md) if JWTs has `jti` (unique ID claim).

Configuration item under `channels[n].jwt`:

- `iss` (list of string, required): List of JWT issuers. `iss` claim of the JWT must exactly match with one of this list.
- `aud` (list of string, optional): List of JWT recipients. One or more value of the `aud` claim of the JWT must exactly match with one of this list.
- `keys` (map of string to string list, required): Key is JWT signing algorithm name such as `RS512`, value is list of file paths of signing key.
  - For RSA alg or ECDSA alg (such as `RS512`, `ES512`), the file should be PEM encoded x509 certificate that contains public key
  - For HMAC alg such as `HS512`, content of the file should be Base64 encoded key
  - For `none` alg, empty list is allowed (`none: []`)
    - `none` alg is easy way for testing purpose, but **do NOT use `none` on production**.
- `claims` (map of string to template string or list of template strings, optional): Validation rule of custom claims
  - For example, `foo: 'bar'` means JWT must have custom claim named `foo` with a value `bar`
  - You can use template string to validate value (e.g. `chatroom: '{{.channel.id}}'` means custom claim `chatroom` must match with `id` of `channels.regex`).
  - If value of JWT claim is boolean or number, validator convert them to string (e.g. `"true"`, `"3.14"`)
- `clockSkewLeeway` (duration string, default `5m`): When validate time-based claims such as `exp`, `nbf`, allow clock skew with this tolerance.

### <a name="admin"></a> `admin` configuration block

```yaml
admin:
  auth:
    networks:
      - 10.1.2.0/8
    bearer:
      - 'my-api-key'
```

Configuration item under `admin`:

- `auth.networks` (list of CIDR string, optional): List of CIDR IP ranges to accept admin API calls
  - By default or if empty list given, allow [RFC 1918](https://tools.ietf.org/html/rfc1918) ranges `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` and [RFC 4193](https://tools.ietf.org/html/rfc4193) range `fc00::/7` and also `127.0.0.0/8` ([RFC 1122](https://tools.ietf.org/html/rfc1122#section-3.2.1.3)), `169.254.0.0/16` ([RFC 3927](https://tools.ietf.org/html/rfc3927)), `::1/128` and `fe80::/10` ([RFC 4291](https://tools.ietf.org/html/rfc4291)).
- `auth.bearer` (list of string, optional): List of API keys required to call admin APIs
  - To call admin APIs, client need to send token as `Authorization: Bearer {token}` header
  - By default or if empty list given, server automatically generate random string on start.
