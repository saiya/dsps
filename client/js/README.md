# DSPS (Durable & Simple PubSub) JavaScript client

[DSPS](../.../README.md) client library for JavaScript environments such as web browser or Node.js.

## 3 minutes to start

1. `npm add dsps` or `yarn add dsps`
2. Copy & edit sample code below:

```js
import { Dsps } from "dsps";

// Create DSPS client
const dsps = new Dsps({
  http: {
    // You can use absolute URL (e.g. "https://dsps.examplem.com/path-prefix") or relative URL
    baseURL: `url-to-dsps`,
  },
});

// Send message
dsps.channel("my-channel").publish(
  null, // null means automatically generate messageID
  { hi: "Hello!" } // Content of the message, OK to pass any JSON
);

// Subscribe messages
const subscription = dsps.channel("my-channel").subscribe({
  callback: async (messages) => {
    console.log("I got messages", messages); // Do anything you want
  },
  abnormalEndCallback: (e) => {
    // Please handle this error (e.g. navigate to login screen with error message)
    console.error(`Subscription abnormally ended due to ${e.code} (;﹏;)`, e);
  },
});

// (Optional) If you lose interest in new messages, call close() method to stop subscription
subscription.end();
```

## Error handling

To handle errors, you can register listeners to the dsps client:

```js
const dsps = new Dsps({
  /* ... */
});

// Listen API call failures:
dsps.addEventListener("apiFailed", (e) => {
  // Do anything you want...
});

// Listen errors thrown by your subscription callbacks:
dsps.addEventListener("subscriptionCallbackError", (info) => {
  // Do anything you want...
});
```

To check full list of listeners, see [type definition of the client interface](./src/client_interface.ts).

## FAQ

### Where is `@types` package for TypeScript?

Package includes `.d.ts` type definition files. No need to install type package additionally.

### What configuration/customization supported?

Check [type definition of the config object](./src/dsps_client_config.ts) for the full list of configuration properties.
