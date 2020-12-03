import { v4 as uuidv4 } from "uuid";
import { Channel, Message, defaultLongPollingSec, Subscription, defaultPollingPagingIntervalSec, defaultLongPollingIntervalJitterSec, defaultLongPollingIntervalSec, defaultShortPollingIntervalSec, defaultPollingErrorIntervalSec, defaultPollingErrorIntervalJitterSec, defaultShortPollingIntervalJitterSec, SubscriptionUnrecoverableError, subscriptionUnrecoverableErrorCodes } from "../client_interface";
import { HttpClient, HttpRequest, HttpResponse, HttpResponseStatusError } from "../http_client";
import { Retry } from "../retry";
import { findErrorCode } from "./error_response";
import { UnreachableCaseError } from "./errors";
import { DspsClientEventTargetImpl } from "./event_target";
import { console } from "./util/console";
import { sleep } from "./util/sleep";

const defaultPollingBulkSize = 32;
type PollingMode = "long-polling" | "short-polling";
type PollingIntervalMode = "no-messages" | "no-more-messages" | "paginated" | "error";

const dedupWindowSizeMultiplier = 3; // Must be equal to or larger than 1

export class ChannelImpl implements Channel {
  private channelID: string;

  private apiRetry: Retry;

  private http: HttpClient;

  private eventTarget: DspsClientEventTargetImpl;

  constructor(args: { channelID: string; apiRetry: Retry; http: HttpClient; eventTarget: DspsClientEventTargetImpl }) {
    this.channelID = args.channelID;
    this.apiRetry = args.apiRetry;
    this.http = args.http;
    this.eventTarget = args.eventTarget;
  }

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

    const res = await this.apiCall({
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
    const { callback, abnormalEndCallback } = args;
    if (!callback) throw new Error(`args.callback should be callable object but ${typeof callback} given`);
    if (!abnormalEndCallback) throw new Error(`args.abnormalEndCallback should be callable object but ${typeof callback} given`);

    const subscriptionID = args.subscriberID ?? generateSubscriberID();
    if (subscriptionID === "") throw new Error(`Subscriber ID could not be empty.`);

    const longPollingSec = args.longPollingSec ?? defaultLongPollingSec;
    const pollingMode: PollingMode = longPollingSec === 0 ? "short-polling" : "long-polling";
    const pollingIntervalSec = args.pollingIntervalSec ?? (pollingMode === "long-polling" ? defaultLongPollingIntervalSec : defaultShortPollingIntervalSec);
    if (pollingMode === "short-polling" && pollingIntervalSec <= 0) throw new Error(`To use ${pollingMode}, args.pollingIntervalSec must be larger than zero but ${pollingIntervalSec} given.`);

    const bulkSize = args.bulkSize ?? defaultPollingBulkSize;
    if (bulkSize <= 0) throw new Error(`args.bulkSize must be larger than zero but ${bulkSize} given.`);

    await this.createSubscription(subscriptionID);
    const { stopPolling } = this.longPollingLoop({
      subscriberID: subscriptionID,
      callback,
      abnormalEndCallback,
      bulkSize,
      pollingMode,
      longPollingSec,
      pollingIntervalSec,
      pollingIntervalJitterSec: args.pollingIntervalJitterSec ?? pollingMode === "long-polling" ? defaultLongPollingIntervalJitterSec : defaultShortPollingIntervalJitterSec,
      pollingPagingIntervalSec: args.pollingPagingIntervalSec ?? defaultPollingPagingIntervalSec,
      pollingErrorIntervalSec: args.pollingErrorIntervalSec ?? defaultPollingErrorIntervalSec,
      pollingErrorIntervalJitterSec: args.pollingErrorIntervalJitterSec ?? defaultPollingErrorIntervalJitterSec,
    });

    let closed = false;
    return {
      channelID: this.channelID,
      subscriptionID,
      close: async () => {
        if (closed) return;
        stopPolling();
        await this.deleteSubscription(subscriptionID);
        closed = true;
      },
    };
  }

  private async createSubscription(subscriptionID: string) {
    await this.apiCall({
      method: "PUT",
      path: `/channel/${this.channelID}/subscription/polling/${subscriptionID}`,
      expectedStatusCodes: [200],
      expected2xxResponseBody: "json",
    });
  }

  private async deleteSubscription(subscriptionID: string) {
    await this.apiCall({
      method: "DELETE",
      path: `/channel/${this.channelID}/subscription/polling/${subscriptionID}`,
      expectedStatusCodes: [200],
      expected2xxResponseBody: "json",
    });
  }

  private longPollingLoop(
    args: (Omit<Parameters<ChannelImpl["longPoolingCall"]>[0], "isStopped" | "dedup"> & Parameters<ChannelImpl["pollingIntervalSec"]>[0]) & {
      abnormalEndCallback: (e: SubscriptionUnrecoverableError) => void;
      bulkSize: number;
    }
  ): {
    stopPolling: () => void;
  } {
    let stop = false;
    const isStopped = () => stop;
    const dedup = new Dedup(args.bulkSize * dedupWindowSizeMultiplier);
    // eslint-disable-next-line no-void
    void (async () => {
      /* eslint-disable no-await-in-loop */
      while (!isStopped()) {
        let intervalMode: PollingIntervalMode;
        try {
          const intervalModeOrFailure = await this.longPoolingCall({ ...args, dedup, isStopped });
          if (intervalModeOrFailure === "stopped") {
            break;
          } else if (typeof intervalModeOrFailure === "object") {
            console.error("DSPS polling aborted due to unrecoverable error", intervalModeOrFailure.abortedWith);
            args.abnormalEndCallback(intervalModeOrFailure.abortedWith);
            break;
          } else {
            intervalMode = intervalModeOrFailure;
          }
        } catch (e) {
          intervalMode = "error";
          console.error("DSPS polling failed to due API failure, retrying...", e);
        }
        await sleep(Math.max(0, Math.ceil(this.pollingIntervalSec(args, intervalMode) * 1000)));
      }
    })();
    return {
      stopPolling: () => {
        stop = true;
      },
    };
  }

  private async longPoolingCall(
    args: Parameters<ChannelImpl["longPollingFetchMessages"]>[0] & {
      isStopped: () => boolean;
      callback: (message: Message<any>[]) => Promise<void>;
      dedup: Dedup;
      subscriberID: string;
      pollingMode: PollingMode;
      longPollingSec: number;
    }
  ): Promise<PollingIntervalMode | "stopped" | { "abortedWith": any }> {
    let ackHandle: string | undefined;
    let mode: null | PollingIntervalMode | "stopped" | { "abortedWith": any } = null;
    try {
      const fetchStartAt = Date.now();
      const result = await this.longPollingFetchMessages(args);
      const fetchSec = (Date.now() - fetchStartAt) / 1000;
      ackHandle = result.ackHandle;
      mode = result.mode;
      if (typeof mode === "object" || mode === "error" || mode === "stopped") return mode;

      if (args.pollingMode === "long-polling" && mode === "no-messages" && args.longPollingSec > fetchSec) {
        console.info("Server returned empty messages without waiting longPollingSec (why?). Sleeping to prevent massive API call...");
        await sleep((args.longPollingSec - fetchSec) * 1000);
      }

      // After acknowledge failure, server returns same messages again.
      const messages = args.dedup.filter(result.messages);
      if (messages.length > 0) {
        try {
          await args.callback(messages);
        } catch (e) {
          this.eventTarget.onSubscriptionCallbackError({
            channelID: this.channelID,
            subscriberID: args.subscriberID,
            messages,
            error: e,
          });
        }
      }
      return mode;
    } finally {
      if (ackHandle) {
        try {
          await this.apiCall({
            method: "DELETE",
            path: `/channel/${this.channelID}/subscription/polling/${args.subscriberID}/message`,
            queryParams: { ackHandle },
            expectedStatusCodes: [204],
            expected2xxResponseBody: null,
          });
        } catch (ackError) {
          // Ignore this ackError and retry fetch + ack in next loop.
          this.eventTarget.onApiFailed(ackError);
          // Override resulted mode to prevent short interval.
          // eslint-disable-next-line no-unsafe-finally
          if (mode === "no-messages" || mode === "no-more-messages" || mode === "paginated") return "error";
        }
      }
    }
  }

  private async longPollingFetchMessages(args: {
    isStopped: () => boolean;
    subscriberID: string;
    bulkSize: number;
    pollingMode: PollingMode;
    longPollingSec: number;
  }): Promise<{
    mode: PollingIntervalMode | "stopped" | { "abortedWith": any };
    messages: Message<any>[];
    ackHandle?: string | undefined;
  }> {
    const { pollingMode, longPollingSec } = args;
    const req: HttpRequest = {
      method: "GET",
      path: `/channel/${this.channelID}/subscription/polling/${args.subscriberID}`,
      queryParams: {
        timeout: pollingMode === "long-polling" ? `${longPollingSec}s` : undefined,
        max: `${args.bulkSize}`,
      },

      expectedStatusCodes: [200, 401, 403, 404],
      expected2xxResponseBody: "json",

      timeoutOffsetMs: pollingMode === "long-polling" ? longPollingSec * 1000 : 0,
    };
    const res = await this.apiCall(req, { retry: false });
    if (args.isStopped()) {
      // Note that error (such as 401 due to subscription end) could occur in this case but it is not problem.
      // Thus this code should take precedence before handlePollingErrorResponse().
      return { mode: "stopped", messages: [] };
    }
    switch (res.status) {
      case 200:
        break;
      case 401: // Unauthorized (may caused by auth rejection)
      case 403: // Forbidden (may caused by forbidden channelID or auth rejection)
      case 404: {
        // Not found (may caused by deleted/expired subscriber/channel)
        const errorCode = findErrorCode(res, ...subscriptionUnrecoverableErrorCodes);
        if (errorCode) {
          return {
            mode: {
              abortedWith: new SubscriptionUnrecoverableError(`${errorCode}`, errorCode),
            },
            messages: [],
          };
        }
        // Fall through because cannot find expected error code.
        // This error could be caused by non-DSPS party such as reverse proxy.
        // Treat it as recoverable error.
      }
      // falls through
      default:
        // Includes fall through from 4xx handling
        this.eventTarget.onApiFailed(new HttpResponseStatusError(req, res));
        return { mode: "error", messages: [] };
    }
    const body = res.json as {
      messages: { messageID: string; content: any }[];
      ackHandle?: string; // Returned only if messages > 0
      moreMessages: boolean;
    };
    let mode: "no-messages" | "no-more-messages" | "paginated";
    if (body.messages.length === 0) {
      mode = "no-messages";
    } else {
      mode = body.moreMessages ? "paginated" : "no-more-messages";
    }
    return {
      mode,
      messages: body.messages.map((msg) => ({ ...msg, channelID: this.channelID })),
      ackHandle: body.ackHandle,
    };
  }

  private pollingIntervalSec(
    args: {
      pollingIntervalSec: number;
      pollingIntervalJitterSec: number;
      pollingPagingIntervalSec: number;
      pollingErrorIntervalSec: number;
      pollingErrorIntervalJitterSec: number;
    },
    intervalMode: PollingIntervalMode
  ) {
    const jitter = Math.random() * 2 - 1; // -1.0 to +1.0
    switch (intervalMode) {
      case "error":
        return args.pollingErrorIntervalSec + jitter * args.pollingErrorIntervalJitterSec;
      case "paginated":
        return args.pollingPagingIntervalSec;
      case "no-messages":
      case "no-more-messages":
        return args.pollingIntervalSec + jitter * args.pollingIntervalJitterSec;
      default:
        throw new UnreachableCaseError(intervalMode);
    }
  }

  private async apiCall(
    req: HttpRequest,
    handling?: {
      retry?: boolean;
    }
  ): Promise<HttpResponse> {
    try {
      if (handling?.retry) {
        return await this.apiRetry.perform(req.path, async () => this.http.request(req));
      }
      return this.http.request(req);
    } catch (e) {
      this.eventTarget.onApiFailed(e);
      throw e;
    }
  }
}

class Dedup {
  private ids: string[] = [];

  constructor(private windowSize: number) {}

  filter(messages: Message<any>[]): Message<any>[] {
    const result: Message<any>[] = [];
    const nextIds = [...this.ids];
    for (const msg of messages) {
      // Assume windowSize is not so large, permit O(n^2) for simplicity.
      if (this.ids.indexOf(msg.messageID) !== -1) continue;
      nextIds.push(msg.messageID);
      result.push(msg);
    }
    this.ids = nextIds.slice(Math.max(0, nextIds.length - this.windowSize));
    return result;
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

function generateSubscriberID(): string {
  return `s-${uuidv4()}`;
}
