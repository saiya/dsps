import { v4 as uuidv4 } from "uuid";
import { Channel, Message, Subscription, SubscriptionUnrecoverableError, subscriptionUnrecoverableErrorCodes } from "../client_interface";
import { HttpRequest, HttpRequestCanceledError, HttpRequestError, HttpResponse, HttpResponseStatusError } from "../http_client";
import { findErrorCode } from "./error_response";
import { UnreachableCaseError } from "./errors";
import { console } from "./util/console";
import { sleep } from "./util/sleep";
import { ClientInternals } from ".";

export type PollingMode = "long-polling" | "short-polling";
export type PollingIntervalMode = "no-messages" | "no-more-messages" | "paginated" | "error";
export const dedupWindowSizeMultiplier = 3; // Must be equal to or larger than 1

type SubscribeArgs = Parameters<Channel["subscribe"]>[0];
export type SubscriptionImplParams = {
  [P in keyof SubscribeArgs]-?: Exclude<SubscribeArgs[P], undefined | null>;
} & {
  pollingMode: PollingMode;
};

export class SubscriptionImpl implements Subscription {
  static generateSubscriberID(): string {
    return `s-${uuidv4()}`;
  }

  constructor(private client: ClientInternals, public readonly channelID: string, private params: Readonly<SubscriptionImplParams>) {
    if (!params.callback) throw new Error(`args.callback should be callable object but ${typeof params.callback} given`);
    if (!params.abnormalEndCallback) throw new Error(`args.abnormalEndCallback should be callable object but ${typeof params.callback} given`);
    if (params.subscriberID === "") throw new Error(`Subscriber ID could not be empty.`);
    if (params.bulkSize <= 0) throw new Error(`args.bulkSize must be larger than zero but ${params.bulkSize} given.`);
    if (params.pollingMode === "short-polling" && params.pollingIntervalSec <= 0) throw new Error(`To use ${params.pollingMode}, args.pollingIntervalSec must be larger than zero but ${params.pollingIntervalSec} given.`);
  }

  public get subscriptionID(): string {
    return this.params.subscriberID;
  }

  private stop = false;

  private closed = false;

  /** @internal */
  async init() {
    await this.createSubscription();
    this.startLongPollingLoop();
  }

  async close(): Promise<void> {
    if (this.closed) return;

    this.stopPolling();
    await this.deleteSubscription();
    this.closed = true;
  }

  private stopCurrentPollingApiCall: null | (() => void) = null;

  private dedup = new Dedup(this.params.bulkSize * dedupWindowSizeMultiplier);

  private stopPolling() {
    this.stop = true;
    if (this.stopCurrentPollingApiCall) this.stopCurrentPollingApiCall();
  }

  private startLongPollingLoop() {
    // eslint-disable-next-line no-void
    void (async () => {
      /* eslint-disable no-await-in-loop */
      while (!this.stop) {
        let intervalMode: PollingIntervalMode;
        try {
          const intervalModeOrFailure = await this.longPoolingCall({
            // eslint-disable-next-line @typescript-eslint/no-loop-func
            cancelable: (handler) => {
              this.stopCurrentPollingApiCall = handler;
            },
          });
          this.stopCurrentPollingApiCall = null;

          if (intervalModeOrFailure === "stopped") {
            break;
          } else if (SubscriptionUnrecoverableError.isInstance(intervalModeOrFailure)) {
            console.error("DSPS polling aborted due to unrecoverable error", intervalModeOrFailure);
            this.params.abnormalEndCallback(intervalModeOrFailure);
            break;
          } else {
            intervalMode = intervalModeOrFailure;
          }
        } catch (e) {
          intervalMode = "error";
          console.error("DSPS polling failed to due API failure, retrying...", e);
        }
        await sleep(Math.max(0, Math.ceil(this.pollingIntervalSec(intervalMode) * 1000)));
      }
    })();
  }

  private async longPoolingCall(args: { cancelable: (handler: () => void) => void }): Promise<PollingIntervalMode | "stopped" | SubscriptionUnrecoverableError> {
    let ackHandle: string | undefined;
    let mode: null | PollingIntervalMode | "stopped" | SubscriptionUnrecoverableError = null;
    try {
      const fetchStartAt = Date.now();
      const result = await this.pollingFetchMessages(args);
      const fetchSec = (Date.now() - fetchStartAt) / 1000;
      ackHandle = result.ackHandle;
      mode = result.mode;
      if (this.stop || typeof mode === "object" || mode === "error" || mode === "stopped") return mode;

      if (this.params.pollingMode === "long-polling" && mode === "no-messages" && this.params.longPollingSec > fetchSec) {
        console.info("Server returned empty messages without waiting longPollingSec (why?). Sleeping to prevent massive API call...");
        await sleep((this.params.longPollingSec - fetchSec) * 1000);
      }

      const messages = this.dedup.filter(result.messages); // After acknowledge failure, server returns same messages again.
      await this.callCallbacks(messages);
      return mode;
    } finally {
      if (ackHandle) {
        try {
          await this.ackMessages(ackHandle);
        } catch (ackError) {
          // Ignore this ackError and retry fetch + ack in next loop.
          this.client.eventTarget.onApiFailed(ackError);
          // Override resulted mode to prevent short interval.
          // eslint-disable-next-line no-unsafe-finally
          if (mode === "no-messages" || mode === "no-more-messages" || mode === "paginated") return "error";
        }
      }
    }
  }

  private async pollingFetchMessages(args: {
    cancelable: (handler: () => void) => void;
  }): Promise<{
    mode: PollingIntervalMode | "stopped" | SubscriptionUnrecoverableError;
    messages: Message<any>[];
    ackHandle?: string | undefined;
  }> {
    let res: HttpResponse;
    try {
      res = await this.fetchMessages(args);
    } catch (e) {
      if (this.stop) {
        // Note that error (such as 401 due to subscription end) could occur in this case but it is not problem.
        // Thus this code should take precedence.
        return { mode: "stopped", messages: [] };
      }
      if (HttpRequestCanceledError.isInstance(e)) return { mode: "no-messages", messages: [] };
      if (HttpRequestError.isInstance(e)) return { mode: "error", messages: [] }; // Error already reported to eventTarget by apiCall().
      if (SubscriptionUnrecoverableError.isInstance(e)) return { mode: e, messages: [] };
      throw e;
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

  //
  // --- API calls ---
  //

  private async createSubscription() {
    await this.client.apiCall({
      method: "PUT",
      path: `/channel/${this.channelID}/subscription/polling/${this.subscriptionID}`,
      expectedStatusCodes: [200],
      expected2xxResponseBody: "json",
    });
  }

  private async fetchMessages(args: { cancelable: (handler: () => void) => void }): Promise<HttpResponse> {
    const req: HttpRequest = {
      method: "GET",
      path: `/channel/${this.channelID}/subscription/polling/${this.subscriptionID}`,
      queryParams: {
        timeout: this.params.pollingMode === "long-polling" ? `${this.params.longPollingSec}s` : undefined,
        max: `${this.params.bulkSize}`,
      },
      expectedStatusCodes: [200, 401, 403, 404],
      expected2xxResponseBody: "json",
      timeoutOffsetMs: this.params.pollingMode === "long-polling" ? this.params.longPollingSec * 1000 : 0,
      cancelable: (cancel) => args.cancelable(() => cancel("Long polling canceled (onStopped)")),
    };
    const res = await this.client.apiCall(req, { retry: false });
    switch (res.status) {
      case 200:
        break;
      case 401: // Unauthorized (may caused by auth rejection)
      case 403: // Forbidden (may caused by forbidden channelID or auth rejection)
      case 404: {
        // Not found (may caused by deleted/expired subscriber/channel)
        const errorCode = findErrorCode(res, ...subscriptionUnrecoverableErrorCodes);
        if (errorCode) throw new SubscriptionUnrecoverableError(`${errorCode}`, errorCode);
        // Fall through because cannot find expected error code.
        // This error could be caused by non-DSPS party such as reverse proxy.
        // Treat it as recoverable error.
      }
      // falls through
      default: {
        // Includes fall through from 4xx handling
        const e = new HttpResponseStatusError(req, res);
        this.client.eventTarget.onApiFailed(e);
        throw e;
      }
    }
    return res;
  }

  private async deleteSubscription() {
    await this.client.apiCall({
      method: "DELETE",
      path: `/channel/${this.channelID}/subscription/polling/${this.subscriptionID}`,
      expectedStatusCodes: [200],
      expected2xxResponseBody: "json",
    });
  }

  private async ackMessages(ackHandle: string) {
    await this.client.apiCall({
      method: "DELETE",
      path: `/channel/${this.channelID}/subscription/polling/${this.subscriptionID}/message`,
      queryParams: { ackHandle },
      expectedStatusCodes: [204],
      expected2xxResponseBody: null,
    });
  }

  //
  // --- Utility functions ---
  //

  private pollingIntervalSec(intervalMode: PollingIntervalMode) {
    const jitter = Math.random() * 2 - 1; // -1.0 to +1.0
    switch (intervalMode) {
      case "error":
        return this.params.pollingErrorIntervalSec + jitter * this.params.pollingErrorIntervalJitterSec;
      case "paginated":
        return this.params.pollingPagingIntervalSec;
      case "no-messages":
      case "no-more-messages":
        return this.params.pollingIntervalSec + jitter * this.params.pollingIntervalJitterSec;
      default:
        throw new UnreachableCaseError(intervalMode);
    }
  }

  private async callCallbacks(messages: Message<any>[]) {
    if (messages.length === 0) return;
    try {
      await this.params.callback(messages);
    } catch (e) {
      this.client.eventTarget.onSubscriptionCallbackError({
        channelID: this.channelID,
        subscriberID: this.subscriptionID,
        messages,
        error: e,
      });
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
