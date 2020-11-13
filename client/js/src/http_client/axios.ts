import Axios, { AxiosInstance, AxiosBasicCredentials, AxiosProxyConfig, AxiosError, AxiosResponse } from "axios";
import qs from "qs";
import { UnreachableCaseError } from "../internal/errors";
import { HttpClient, HttpRequest, HttpResponse, HttpRequestError, HttpResponseStatusError, normalizeHeaders } from ".";
import { console } from "../internal/util/console";

/**
 * HTTP client configuration to use axios HTTP client.
 *
 * Actually subset of axios's AxiosRequestConfig.
 */
export type HttpClientAxiosConfig = {
  /**
   * (Required) URL of the server (e.g. "https://dsps.example.com").
   * If DSPS server configured with pathPrefix, it should be included also (such as "https://dsps.example.com/path-prefix/").
   */
  baseURL: string;

  /** HTTP headers to send always (APIs may override some headers such as Authorization) */
  headers?: {
    [name: string]: string;
  };

  //
  // --- Network / IO ---
  //

  /**
   * Timeout in milliseconds, same as AxiosRequestConfig.
   * Note that long-polling request extends this value.
   */
  timeout?: number;
  /** Set to enable HTTP(S) proxy, same as AxiosRequestConfig. */
  proxy?: AxiosProxyConfig | false;

  /** (Node.js) http.Agent object to set client certificate, keepAlive and so on. Same as AxiosRequestConfig. */
  httpAgent?: any;
  /** (Node.js) https.Agent object to set client certificate, keepAlive and so on. Same as AxiosRequestConfig. */
  httpsAgent?: any;

  //
  // --- Security / Credentials ---
  //

  /** true to send cookies in CORS request, same as AxiosRequestConfig. */
  withCredentials?: boolean;
  /** Set to send BASIC auth credentials, same as AxiosRequestConfig. */
  auth?: AxiosBasicCredentials;
};

const defaultTimeout = 15 * 1000;

class HttpClientAxiosImpl implements HttpClient {
  public readonly isDspsHttpClient: true = true;

  private timeoutSec: number;

  private axios: AxiosInstance;

  constructor(config: HttpClientAxiosConfig) {
    this.timeoutSec = config.timeout ?? defaultTimeout;
    this.axios = Axios.create({
      maxRedirects: 5, // Follow redirects

      ...config,
      headers: normalizeHeaders({
        ...(config.headers ?? {}),
        Accept: "application/json",
      }),

      responseType: "text",
      transformResponse: (res) => res,
      paramsSerializer: qs.stringify,
      validateStatus: () => true, // Do not throw error regardless response status
    });
  }

  async request(req: HttpRequest): Promise<HttpResponse> {
    try {
      return this.parseResponse(
        req,
        await this.axios.request({
          method: req.method,
          url: req.path,

          headers: normalizeHeaders(req.headers ?? {}),
          params: {
            // eslint-disable-next-line @typescript-eslint/naming-convention
            _: `${Date.now()}-${Math.random()}`, // Cache buster parameter
            ...(req.queryParams ?? {}),
          },
          data: req.bodyJson,

          timeout: req.timeoutOffsetMs ? this.timeoutSec + req.timeoutOffsetMs : this.timeoutSec,
        })
      );
    } catch (e) {
      if (isAxiosError(e)) throw new HttpRequestError(`${e.message} (code: ${e.code ?? "(none)"})`);
      throw e;
    }
  }

  private parseResponse(req: HttpRequest, raw: AxiosResponse): HttpResponse {
    const contentType = raw.headers["content-type"];

    const res: HttpResponse = {
      status: raw.status,
      statusText: raw.statusText,
      headers: normalizeHeaders(raw.headers),
      text: typeof raw.data === "string" ? raw.data : undefined,
    };
    if (req.expectedStatusCodes.indexOf(res.status) === -1) throw new HttpResponseStatusError(req, res);
    if (res.text && /^application\/json($|;)/.test(contentType)) {
      try {
        // Must after response status assertion because JSON parsing error should not take precedence over it.
        res.json = JSON.parse(res.text);
      } catch (e) {
        throw new HttpRequestError(`Failed to not parse response body JSON: ${e}`);
      }
    }
    if (res.status >= 200 && res.status < 300) {
      switch (req.expected2xxResponseBody) {
        case null:
          break;
        case "json":
          if (typeof res.json === "undefined") throw new HttpRequestError(`Expected ${req.expected2xxResponseBody} body but not responded: ${req.method} ${req.path} (Status: ${res.status}, Content-Type: ${contentType ?? "(none)"})`);
          break;
        default:
          throw new UnreachableCaseError(req.expected2xxResponseBody);
      }
    }
    return res;
  }
}

export const HttpClientAxios: { new(config: HttpClientAxiosConfig): HttpClient } = HttpClientAxiosImpl;

const isAxiosError = (e: any): e is AxiosError => (e as AxiosError).isAxiosError; // https://github.com/axios/axios/issues/1415
