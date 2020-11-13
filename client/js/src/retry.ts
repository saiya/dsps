import { console } from "./internal/util/console";
import { sleep } from "./internal/util/sleep";

export type Retry = {
  perform<T>(taskDescription: string, task: () => Promise<T>): Promise<T>;
};

export type RetryConfig = {
  count: number;
  intervalSec: number;
  intervalMultiplier: number;
  intervalJitterSec: number;
};

class RetryImpl implements Retry {
  constructor(private config: RetryConfig) {}

  async perform<T>(taskDescription: string, task: () => Promise<T>): Promise<T> {
    const { intervalSec, intervalMultiplier, intervalJitterSec } = this.config;

    // eslint-disable-next-line no-constant-condition, no-plusplus
    for (let nextAttempt = 1; true; nextAttempt++) {
      try {
        return await task(); // eslint-disable-line no-await-in-loop
      } catch (e) {
        if (nextAttempt > this.config.count) throw e;

        const intervalMs = Math.ceil(1000 * Math.max(0, intervalSec * intervalMultiplier ** nextAttempt + (Math.random() * 2 - 1.0) * intervalJitterSec));
        console.info(`Will retry after ${intervalMs}ms (${nextAttempt}/${this.config.count}): ${taskDescription}`, e);
        await sleep(intervalMs); // eslint-disable-line no-await-in-loop
      }
    }
  }
}

export function makeRetry(config: Retry | RetryConfig): Retry {
  return config && isRetry(config) ? config : new RetryImpl(config);
}

function isRetry(config: Retry | RetryConfig): config is Retry {
  return typeof (config as Retry).perform === "function";
}
