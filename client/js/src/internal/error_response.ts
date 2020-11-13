import { HttpResponse } from "../http_client";

export function findErrorCode<T extends string>(res: HttpResponse, ...codes: readonly T[]): null | T {
  if (!res.text) return null;

  let json: any;
  try {
    json = JSON.parse(res.text);
  } catch (e) {
    return null;
  }

  const i = codes.indexOf(json.code);
  if (i === -1) return null;
  return codes[i] as T;
}
