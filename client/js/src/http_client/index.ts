/**
 * HTTP client abstract interface.
 * You can write your own implementation to use any HTTP client implementation.
 */
export type HttpClient = {
  readonly isDspsHttpClient: true;

  /**
   * @throws {HttpRequestError}
   */
  request(req: HttpRequest): Promise<HttpResponse>;
};

export const isHttpClient = (obj: any): obj is HttpClient => (obj as HttpClient).isDspsHttpClient ?? false;

export type HttpRequest = {
  method: "GET" | "PUT" | "DELETE";
  /**
   * Path from baseURL, no need to include pathPrefix.
   * Always starts with "/".
   */
  path: string;

  queryParams?: {
    [name: string]: string | undefined;
  };
  bodyJson?: {};
  headers?: {
    [name: string]: string | undefined;
  };

  expectedStatusCodes: number[];
  /** Note: `null` means to permit any response body, not rejects JSON. */
  expected2xxResponseBody: null | "json";

  /** If need to increase timeout setting (e.g. long-polling), set this value. */
  timeoutOffsetMs?: number;

  /** If function given, passes cancel function object. */
  cancelable?: (cancel: (message: string) => void) => void;
};

export type HttpResponse = {
  status: number;
  statusText: string;

  /** Response body JSON, if present */
  json?: any;
  /** Response body string, if present */
  text?: string;

  headers: {
    /** name is always lower-case. */
    [name: string]: string;
  };
};

export const normalizeHeaders = (headers: { [name: string]: string | undefined }): { [name: string]: string } =>
  Object.fromEntries(
    Object.entries(headers)
      .map(([name, value]): [string, string | undefined] => [name.toLowerCase(), value])
      .filter((pair): pair is [string, string] => typeof pair[1] !== "undefined")
  );

export class HttpRequestError extends Error {
  static isInstance(e: any): e is HttpRequestError {
    return (e as HttpRequestError).isHttpRequestError ?? false;
  }

  private isHttpRequestError: true = true;
}

export class HttpResponseStatusError extends HttpRequestError {
  static isInstance(e: any): e is HttpResponseStatusError {
    return (e as HttpResponseStatusError).isHttpResponseStatusError ?? false;
  }

  private isHttpResponseStatusError: true = true;

  constructor(public readonly request: HttpRequest, public readonly response: HttpResponse) {
    super(`Unexpected HTTP response status ${response.status} (${response.statusText}, body: ${response.text ?? "(none)"}) from ${request.method} ${request.path}`);
  }
}

export class HttpRequestCanceledError extends HttpRequestError {
  static isInstance(e: any): e is HttpRequestCanceledError {
    return (e as HttpRequestCanceledError).isHttpRequestCanceledError ?? false;
  }

  private isHttpRequestCanceledError: true = true;

  constructor(public readonly request: HttpRequest, public readonly detail: any) {
    super(`HTTP request ${request.method} ${request.path} has been canceled: ${detail}`);
  }
}
