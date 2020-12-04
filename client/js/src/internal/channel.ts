import { v4 as uuidv4 } from "uuid";
import { Channel, Message, defaultLongPollingSec, Subscription, defaultPollingPagingIntervalSec, defaultLongPollingIntervalJitterSec, defaultLongPollingIntervalSec, defaultShortPollingIntervalSec, defaultPollingErrorIntervalSec, defaultPollingErrorIntervalJitterSec, defaultShortPollingIntervalJitterSec, defaultPollingBulkSize } from "../client_interface";
import { SubscriptionImpl, PollingMode } from "./subscription";
import { ClientInternals } from ".";

export class ChannelImpl implements Channel {
  constructor(private client: ClientInternals, private channelID: string) {}

  async publish<T>(messgageID: null | string, content: T): Promise<Message<T>> {
    if (!isSerializableValue(content)) throw new Error(`Given message content is not JSON serializable (type = ${typeof content}): ${content}`);

    const messageID = messgageID ?? generateMessageID();
    if (messgageID === "") throw new Error("MessageID could not be empty.");

    let json: string;
    try {
      json = JSON.stringify(content);
    } catch (e) {
      throw new Error(`Cannot JSON serialize given message content: ${e}`);
    }

    const res = await this.client.apiCall({
      method: "PUT",
      path: `/channel/${this.channelID}/message/${messageID}`,
      bodyJson: json,
      expectedStatusCodes: [200],
      expected2xxResponseBody: "json",
    });
    return {
      ...(res.json as { channelID: string; messageID: string }),
      content,
    };
  }

  async subscribe(args: Parameters<Channel["subscribe"]>[0]): Promise<Subscription> {
    const longPollingSec = args.longPollingSec ?? defaultLongPollingSec;
    const pollingMode: PollingMode = longPollingSec === 0 ? "short-polling" : "long-polling";
    const sbsc = new SubscriptionImpl(this.client, this.channelID, {
      subscriberID: args.subscriberID ?? SubscriptionImpl.generateSubscriberID(),
      callback: args.callback,
      abnormalEndCallback: args.abnormalEndCallback,

      pollingMode,
      bulkSize: args.bulkSize ?? defaultPollingBulkSize,
      longPollingSec,
      pollingIntervalSec: args.pollingIntervalSec ?? (pollingMode === "long-polling" ? defaultLongPollingIntervalSec : defaultShortPollingIntervalSec),
      pollingIntervalJitterSec: args.pollingIntervalJitterSec ?? pollingMode === "long-polling" ? defaultLongPollingIntervalJitterSec : defaultShortPollingIntervalJitterSec,
      pollingPagingIntervalSec: args.pollingPagingIntervalSec ?? defaultPollingPagingIntervalSec,
      pollingErrorIntervalSec: args.pollingErrorIntervalSec ?? defaultPollingErrorIntervalSec,
      pollingErrorIntervalJitterSec: args.pollingErrorIntervalJitterSec ?? defaultPollingErrorIntervalJitterSec,
    });
    await sbsc.init();
    return sbsc;
  }
}

function isSerializableValue(content: any): boolean {
  switch (typeof content) {
    case "undefined":
    case "function":
    case "bigint": // RFC 7159 allows implementations to limit number range, and JSON.stringify(...) does not support BigInt.
    case "symbol":
      return false;
    default:
      return true;
  }
}

function generateMessageID(): string {
  return `msg-${uuidv4()}`;
}
