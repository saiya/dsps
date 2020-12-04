import { HttpRequest, HttpResponse } from "../http_client";
import { DspsClientEventTargetImpl } from "./event_target";

/** @internal */
export interface ClientInternals {
  eventTarget: DspsClientEventTargetImpl;

  apiCall(
    req: HttpRequest,
    handling?: {
      retry?: boolean;
    }
  ): Promise<HttpResponse>;
}
