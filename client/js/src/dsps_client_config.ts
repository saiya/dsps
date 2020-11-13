import { HttpClient } from "./http_client";
import { HttpClientAxiosConfig } from "./http_client/axios";
import { Retry, RetryConfig } from "./retry";

export type DspsClientConfig = {
  http: HttpClientAxiosConfig | HttpClient;

  /** Configuration of API retry (expect for polling). */
  apiRetry?: RetryConfig | Retry; // see defaultApiRetry for default settings.

  /** JSON web token (JWT - RFC 7519) to send for each request. */
  jwt?: string;
};

/*
 * Retry intervals before each retry:
 *   1st retry: 1.0 + 1.5^0 ± 0.5 = 0.5  to 1.5  sec
 *   2nd retry: 1.0 + 1.5^1 ± 0.5 = 2.0  to 3.0  sec
 *   3rd retry: 1.0 + 1.5^2 ± 0.5 = 2.75 to 3.75 sec
 */
export const defaultApiRetry: Readonly<RetryConfig> = {
  count: 3,
  intervalSec: 1.0,
  intervalMultiplier: 1.5,
  intervalJitterSec: 0.5,
};
