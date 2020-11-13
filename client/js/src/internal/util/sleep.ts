import globalThis from "./global_this";

export const sleep = async (milliseconds: number) =>
  new Promise<void>((resolve) => {
    globalThis.setTimeout(resolve, milliseconds);
  });
