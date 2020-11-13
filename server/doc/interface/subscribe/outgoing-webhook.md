# Outgoing webhook subscription

DSPS supports outgoing webhook deliver messages to any HTTP(S) services.

To configure outgoing webhook, configure `channels.webhooks` section of the [server configuration file](../../config.md#channels-webhooks-configuration-block).

DSPS currently does not support dynamic `webhook` configuration change to make [security matters simple](../../security.md#outgoing-webhook).
