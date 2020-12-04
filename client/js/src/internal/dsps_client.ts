import { Channel, DspsClient } from "../client_interface";
import { defaultApiRetry, DspsClientConfig } from "../dsps_client_config";
import { HttpClient, HttpRequest, HttpRequestCanceledError, HttpResponse, isHttpClient } from "../http_client";
import { HttpClientAxios } from "../http_client/axios";
import { Retry, makeRetry } from "../retry";
import { ChannelImpl } from "./channel";
import { DspsClientEventTargetImpl } from "./event_target";
import { ClientInternals } from ".";

export class DspsClientImpl implements DspsClient, ClientInternals {
  /** @internal */
  public eventTarget: DspsClientEventTargetImpl = new DspsClientEventTargetImpl();

  private http: HttpClient;

  private apiRetry: Retry;

  constructor(config: DspsClientConfig) {
    const rawHttp = isHttpClient(config.http) ? config.http : new HttpClientAxios(config.http);
    this.http = new HttpClientWrapper(config, rawHttp);
    this.apiRetry = makeRetry(config.apiRetry ?? defaultApiRetry);
  }

  /**
   * Returns instance to interact with the channel.
   * Note that this method does not check validity & accessibility of the channel.
   */
  channel(channelID: string): Channel {
    return new ChannelImpl(this, channelID);
  }

  addEventListener(type: Parameters<DspsClientEventTargetImpl["addEventListener"]>[0], listener: Parameters<DspsClientEventTargetImpl["addEventListener"]>[1]): ReturnType<DspsClientEventTargetImpl["addEventListener"]> {
    return this.eventTarget.addEventListener(type, listener);
  }

  removeEventListener(type: Parameters<DspsClientEventTargetImpl["removeEventListener"]>[0], listener: Parameters<DspsClientEventTargetImpl["removeEventListener"]>[1]): void {
    this.eventTarget.removeEventListener(type, listener);
  }

  async apiCall(
    req: HttpRequest,
    handling?: {
      retry?: boolean;
    }
  ): Promise<HttpResponse> {
    try {
      if (handling?.retry) {
        return await this.apiRetry.perform(req.path, async () => this.http.request(req));
      }
      return await this.http.request(req);
    } catch (e) {
      if (!HttpRequestCanceledError.isInstance(e)) {
        this.eventTarget.onApiFailed(e);
      }
      throw e;
    }
  }
}

class HttpClientWrapper implements HttpClient {
  readonly isDspsHttpClient: true = true;

  constructor(private config: DspsClientConfig, private wrapped: HttpClient) {}

  /**
   * @throws {HttpRequestError}
   */
  async request(req: HttpRequest): Promise<HttpResponse> {
    return this.postprocess(req, await this.wrapped.request(this.preprocess(req)));
  }

  private preprocess(req: HttpRequest): HttpRequest {
    const headers = { ...(req.headers ?? {}) };
    if (this.config.jwt) headers.authorization = `Bearer ${this.config.jwt}`;
    return {
      ...req,
      headers,
    };
  }

  private postprocess(req: HttpRequest, res: HttpResponse): HttpResponse {
    return res;
  }
}
