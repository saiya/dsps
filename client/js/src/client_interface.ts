/** Top-level interface of DSPS client. */
export type DspsClient = DspsClientEventTarget & {
  /**
   * Returns instance to interact with the channel.
   * Note that this method does not check validity & accessibility of the channel.
   */
  channel(channelID: string): Channel;
};

export type DspsClientEventTarget = {
  /**
   * Call given listener when any API/communication error occurs.
   * @returns Given listener itself
   */
  addEventListener(type: "apiFailed", listener: (e: any) => void): (...args: any[]) => void;
  /**
   * Call given listener when user-supplied callback function returns error.
   * @returns Given listener itself
   */
  addEventListener(type: "subscriptionCallbackError", listener: (info: SubscriptionCallbackErrorInfo) => void): (...args: any[]) => void;

  /** Reverse operation of {@link addEventListener} */
  removeEventListener(type: string, listener: (...args: any[]) => void): void;
};

export type Channel = {
  /**
   * Send message to this channel.
   * This method automatically generate uniqueID of the message.
   *
   * @param messageID Pass null to automatically generate ID.
   * @param content Must be JSON serializable object.
   */
  publish<T>(messageID: null | string, content: T): Promise<Message<T>>;

  /**
   * Receive messages, call given callback function.
   */
  subscribe(args: {
    /** If null or undefined, automatically generate ID. */
    subscriberID?: null | string;

    /**
     * Callback function to call when message received.
     *
     * If callback returns error (rejected Promise), this method discards messages.
     * To catch errors, use {@link DspsClient#addEventListener} with `"subscriptionCallbackError"` event type.
     */
    callback: (messages: Message<any>[]) => Promise<void>;

    /**
     * Callback function called when subscription abnormally aborted.
     * You should handle this and rescue this error (e.g. navigate to initial screen).
     */
    abnormalEndCallback: (e: SubscriptionUnrecoverableError) => void;

    /** Size of bulk message fetch (no guarantee, could receive more or less for each callback). */
    bulkSize?: number;

    /**
     * Timeout of long-polling, default is {@link defaultLongPollingSec}.
     * Set `0` to use short-polling.
     */
    longPollingSec?: number;
    /**
     * Interval of polling, default is {@link defaultLongPollingIntervalSec} or {@link defaultShortPollingIntervalSec}.
     * For long-polling, recommend to use smaller value for best response time.
     * If you use short-polling, this value MUST be larger than zero otherwise client rejects it.
     */
    pollingIntervalSec?: number;
    /**
     * Random offset to add or minus from pollingIntervalSec.
     * Default is {@link defaultLongPollingIntervalJitterSec} or {@link defaultShortPollingIntervalJitterSec}.
     * Note that this jitter is not applied just after "paginated" response that described in {@link pollingPagingIntervalSec}.
     */
    pollingIntervalJitterSec?: number;
    /**
     * If one or more messages returned in polling response, should immediately call to receive stacked (queued) messages.
     * Thus this interval applied just after received one or more messages.
     * Default is {@link defaultPollingPagingIntervalSec}.
     */
    pollingPagingIntervalSec?: number;
    /** Interval after polling API failure, default is {@link defaultPollingErrorIntervalSec} */
    pollingErrorIntervalSec?: number;
    /** Random value to add/minus from {@link pollingErrorIntervalSec}, default is {@link defaultPollingErrorIntervalJitterSec} */
    pollingErrorIntervalJitterSec?: number;
  }): Promise<Subscription>;
};

export const defaultLongPollingSec = 30;
export const defaultLongPollingIntervalSec = 0.05;
export const defaultLongPollingIntervalJitterSec = 0.1;
export const defaultShortPollingIntervalSec = 5;
export const defaultShortPollingIntervalJitterSec = 0.5;
export const defaultPollingPagingIntervalSec = 0.05;
export const defaultPollingErrorIntervalSec = 5.0;
export const defaultPollingErrorIntervalJitterSec = 2.5;

export type Message<T> = {
  readonly channelID: string;
  readonly messageID: string;
  readonly content: T;
};

export type Subscription = {
  readonly channelID: string;
  readonly subscriptionID: string;

  /** To stop this subscription, call this function. */
  close(): Promise<void>;
};

export type SubscriptionCallbackErrorInfo = {
  readonly error: any;

  readonly messages: Message<any>[];
  readonly channelID: string;
  readonly subscriberID: string;
};

/** Content of this list possibly change in future version. */
export const subscriptionUnrecoverableErrorCodes = ["dsps.auth.invalid-credentials", "dsps.auth.channel-forbidden", "dsps.storage.subscription-not-found", "dsps.storage.invalid-channel"] as const;

export class SubscriptionUnrecoverableError extends Error {
  /**
   * @param code Note that list of {@link subscriptionUnrecoverableErrorCodes} possibly change in future version.
   */
  constructor(message: string, public readonly code: null | typeof subscriptionUnrecoverableErrorCodes[number]) {
    super(message);
  }
}
