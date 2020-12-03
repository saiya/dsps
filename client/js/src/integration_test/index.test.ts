import { v4 as uuidv4 } from "uuid";
import { Dsps, Channel, Message, SubscriptionUnrecoverableError } from "..";
import { sleep } from "../internal/util/sleep";

const dsps = new Dsps({
  http: {
    baseURL: process.env.DSPS_BASE_URL ?? "http://localhost:3000/",
  },
});

test("Publish + Subscribe", async () => {
  const channel = dsps.channel(randomChannelID());
  await withSubscription({ channel }, async ({ waitNewMessages }) => {
    const msg1 = await channel.publish(null, {
      hi: "hello!",
      nested: {
        value: Math.PI,
      },
    });
    expect((await waitNewMessages())[0]).toEqual(msg1);
  });
}, 3000);

const withSubscription = async (
  args: {
    channel: Channel;
    abnormalEndCallback?: (e: SubscriptionUnrecoverableError) => void;
  },
  h: (args: { waitNewMessages: (count?: number) => Promise<Message<any>[]> }) => Promise<void>
) => {
  let received: Message<any>[] = [];
  const subscription = await args.channel.subscribe({
    callback: async (msgs) => {
      received = [...received, ...msgs];
    },
    abnormalEndCallback: args.abnormalEndCallback ?? fail,
  });
  try {
    await h({
      waitNewMessages: async (count) => {
        const waitUntil = received.length + (count ?? 1);
        while (received.length < waitUntil) {
          await sleep(50); // eslint-disable-line no-await-in-loop
        }
        return received.slice(-(count ?? 1));
      },
    });
  } finally {
    await subscription.close();
  }
};

const randomChannelID = () => `ch-${uuidv4()}`;
