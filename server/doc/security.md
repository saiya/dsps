# Security aspect of the DSPS server

## Use HTTPS (TLS)

In production, you should use TLS.

DSPS server itself does not support TLS, so that you need to use HTTPS capable loadbalancers or reverse proxy such as Nginx.

## Authorize HTTP endpoint

You can protect endpoints with JWT, see [channels.jwt configuration block](./config.md#jwt).

Also you can revoke JWT with [administration API](./interface/admin/revoke_jwt.md).

## Protect admin API

By default, server accepts admin API call from private IP addresses with randomly generated API key.

To configure it, see [`admin` configuration block](./config.md#admin).

Also if you run this server behind LoadBalancer, be sure to set [`http.ipheader` configuration item](./config.md#ipheader) if your LoadBalancer changes source IP of the packets.
Otherwise server could not check client's IP address due to LoadBalancer.

## Secure outgoing Webhook

Currently DSPS server only supports pre-configured outgoing webhook. So that you can control webhook destination by configuration.

To ensure webhook security, general outgoing HTTP security practices such as followings should be applied:

1. Use HTTPS (TLS)
2. Send webhook to only safe destinations
3. Do not send webhook to dynamic domain, domain name should be fixed

## HTTP response headers

DSPS server send some response headers by default but you can override them to more security.

For instance, you can enable `Strict-Transport-Security` by setting the header for each response.

Use [`http.defaultHeaders` configuration item](./config.md#defaultHeaders) to set your custom headers for all responses.
