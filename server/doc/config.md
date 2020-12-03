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
  sourceIpHeader: 'X-Forwarded-For'
  showForbiddenDetail: false

logging:
  debug: false
  attributes:
    machineID: my-machine-id

channels:
  - regex: 'chat-room-(?P<id>\d+)'
    expire: 15m
    webhooks:
      - url: 'http://localhost:3001/you-got-message/room/{{.id}}'
```

## Words & definitions used in this document

- `Regex string` means YAML string that constructs [golang compatible regular expression](https://golang.org/pkg/regexp/) (e.g. `'chat-room-(?P<id>\d+)'`)
- `Template string` means YAML string that follows [golang Template](https://golang.org/pkg/text/template) syntax (e.g. `'http://localhost:3001/room/{{.id}}'`)
- `Duration string` means YAML string that follows [golang ParseDuration](https://golang.org/pkg/time/#ParseDuration) syntax (e.g. `'1h30m'`)

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

- `port` (number, optional, default `3000`): TCP port to listen for HTTP requests
  - You can override this value with `--port` command line option
- `listen` (string, optional): Listen string (e.g. ":3000"), this option overrides `port`
  - With [some security software](https://forum.eset.com/topic/22080-mac-firewall-issue-after-update-to-684000/), you may need to specify local IP as such as `127.0.0.1:3000` due to MTIM proxy problem.
- `pathPrefix` (string, optional): Prefix to add all endpoints
  - e.g. If `pathPrefix` is `/foo/bar`, endpoint `/probe/readiness` is served as `/foobar/probe/readiness`
- <a name="ipheader"></a> `sourceIpHeader` (string, optional, default null): HTTP header name contains reliable IP address of the client
  - Note that `admin.auth.networks` rely on this configuration.
- `showForbiddenDetail` (boolean, default `false`): Show detail reason of 403 to clients
- `longPollingMaxTimeout` (Duration string, optional, default `30s`): Max duration of the [long-polling requests](./interface/subscribe/polling.md#polling-get).
- `gracefulShutdownTimeout` (Duration string, optional, default `5s`): Timeout to await end of running requests.

## logging configuration block

Configuration items under `logging`:

- `debug` (boolean, optional, default `false`): Output DEBUG level logs
- `attributes` (string to string map, optional): Attributes set to every log records
  - Useful to set machine ID etcetera.

## <a name="channels"></a> channels configuration block

You can configure channels under `channels` block.

Each channel configuration must have `regex` to match with name of a channel.
If multiple configuration match with a channel, server merges them.

If you configure one or more `channels` blocks, DSPS server will reject unmatched channel name.
If `channels` configuration is empty, DSPS server automatically define `.*` channel configuration to accept any channel name.

Configuration items under `channels[n]`:

- `regex` (Regex string, required): Regex string to match with name of a channel
  - Must match with *entire string* of the channel name (no need to write `^` nor `$`).
  - You can use named group (e.g. `(?P<id>\d+)`). In the channel configuration, captured value of the group is visible to template strings.
- `expire` (Duration string, required): DSPS server may discard inactive subscribers & messages after this duration
  - Duration counts from last access of the subscriber or sent time of the message.
  - DSPS may not resend after this expiration duration, so that this value must be larger than client's polling period if you polling.
  - If multiple channel configuration matches to a channel, largest value wins.
  - If outgoing webhook is configured, expire value must be larger than maximum webhook time includes webhook timeout and retry interval

### channels.webhooks configuration block

You can configure outgoing webhook to send messages from DSPS server to any HTTP(S) services.

```yaml
# Webhook with retry configuration example
channels:
  - regex: 'chat-room-(?P<id>\d+)'
    message:
      # Must be larger than final retry attempt time
      expire: 15m
    webhooks:
      - url: 'http://localhost:3001/you-got-message/room/{{.id}}'
        retry:
          timeout: 30s
          # Enable 3 retries as below:
          # 1st retry: 3 * 1.5^0 ± 1.5 = 3    ± 1.5 [sec] after first webhook attempt
          # 2nd retry: 3 * 1.5^1 ± 1.5 = 4.5  ± 1.5 [sec] after 1st retry
          # 3rd retry: 3 * 1.5^2 ± 1.5 = 6.75 ± 1.5 [sec] after 1st retry
          count: 3
          interval: 3s
          intervalMultiplier: 1.5
          intervalJitter: 1s500ms
        headers:
          User-Agent: my DSPS server
          X-Chat-Room-ID: '{{.id}}'
```

If there are multiple webhooks, DSPS server calls them concurrently. Configuration order of the webhooks has no meaning.

Configuration item under `channels[n].webhooks`:

- `url` (Template string, required): Full URL to send message
- `timeout` (Duration string, default: `30s`): Timeout of the webhook call
- `connection.max` (integer, default: `1024`): Max connections between DSPS server and webhook target
  - This configuration also controls of `Transport.MaxIdleConnsPerHost`
- `connection.maxIdleTime` (Duration string, default: `3m`): Max idle time to keep-alive connections
- `retry.force` (boolean, default: `false`): true to retry any errors (such as `404 Not Found`)
- `retry.count` (integer, default: `3`): 0 to disable retry, 1 to retry only once, ...
- `retry.interval` (Duration string, default: `3s`): Retry base interval
- `retry.intervalMultiplier` (float, default: `1.5`): Exponential backoff factor, multiply to the previous interval
- `retry.intervalJitter` (Duration string, default: `1s500ms`): Max range of the retry interval randomization, plus or minus to the resulted interval
- `headers` (string to template string map, optional): HTTP headers to set for each outgoing requests

### <a name="jwt"></a> channels.jwt configuration block

To protect endpoints, can validate signed [JSON Web Tokens (JWT, RFC 7519)](https://jwt.io/).

```yaml
channels:
  - regex: 'chat-room-(?P<id>\d+)'
    jwt:
      alg: RS256
      iss:
        - https://issuer.example.com/issuer-url
      keys:
        - |-
          -----BEGIN PUBLIC KEY-----
          MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAlgY7fpgGEKqGaoUc1O9K
          CdytNmBa7P1DWfA8QWFE042yn/dBLW8M+uWqsvD/pDWaSDNfEgY6J8nyKZ7DMps6
          E1TJNBkZ7/4TDVpmsIE8vqK/bhTz5SYTnLyMd2Wh7Yy+uUOk6XTR2Ade9ysHPD5U
          mmFBzQX2r+S25lpRUHXmSGl7cYiTbWmI2JVTId3agHR1jqZ1EeWDorEZ3HF7hExl
          pKXa0vZaMoK2mvzHOhaNPn57BNqcXfzLVYjny1br7qJOHgMBW+AwCbb7yE+aRsur
          WQEc6XyhbFG443Sb6tHvbiROg2nTXu1Pq0ZaB90mytpm0Md+p0QI0mqizhbOD3d3
          Lf10Zj86nlvT4dKbWwZHfrh9oiR9tLGgCyUtVQYhgv7BehdLnpJmxxaohteLJHon
          PfzIKqOY24OmteqAML7+G8gbrRIXMS8aTvPJvJ3XT51QD+61CMwExMWXz1CTXlc3
          tSZ0nx8hquPI9C/B9AIlnk0lgKNmq+A2aU98OnSlTPqsdZo3xr4PPMthiNr/dfEq
          HsijJ3dq9pwaO9t0xKti+Hd9ic/IqUH2OyT0Nw36f/MvDBAILF8SVimSKnEaQI04
          5AME2BK5WZiwL47SqZIWTNUglhyPEZCZ2tFJYHZHFSW6AbnDWAxYKBuDE7MB+t/u
          Y4XfEnmCs8dK48LUuB+IgF8CAwEAAQ==
          -----END PUBLIC KEY-----
      claims:
        chatroom: '{{.id}}'
```

If this configuration present on the channel, clients must present valid JWT with `Authorization: Bearer <jwt>` request header for every API call.

In addition, you can revoke JWT with [revocation API](./interface/admin/revoke_jwt.md) if JWTs has `jti` (unique ID claim).

Configuration item under `channels[n].jwt`:

- `alg` (string, required): Acceptable JWT signing algorithm such as `RS256`
  - `none` alg is easy way for testing purpose, but do not accept it on production.
- `iss` (list of string, required): Acceptable JWT issuers
- `keys` (list of string, required): key of signing
  - For `none` alg, this configuration does not have meaning (can be empty list)
  - For HMAC alg such as `HS256`, set Base64 encoded key here
  - For RSA alg such as `RS256` or ECDSA alg such as `ES256`, set PEM encoded x509 certificate that contains public key
- `claims` (string to Template string map, optional): Validation rule of custom claims
  - For example, `foo: 'bar'` means JWT must have custom claim named `foo` with a value `bar`
  - You can use template string to validate value (e.g. `chatroom: '{{.id}}'` means custom claim `chatroom` must match with `id` of `channels.regex`).

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
  - By default or if empyt list given, allow [RFC 1918](https://tools.ietf.org/html/rfc1918) ranges `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` and [RFC 4193](https://tools.ietf.org/html/rfc4193) range `fc00::/7`.
- `auth.bearer` (list of string, optional): List of `Authorization: Bearer` request header value
  - By default or if empyt list given, server automatically generate random string on start.
  - Client must send one of specified value.
