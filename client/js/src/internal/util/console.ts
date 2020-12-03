import globalThis from "./global_this";

export const console: {
  info: (msg: string, ...args: any[]) => void;
  error: (msg: string, ...args: any[]) => void;
} = {
  info: globalThis?.console?.info ?? globalThis?.console?.debug ?? (() => {}),
  error: globalThis?.console?.error ?? globalThis?.console?.debug ?? (() => {}),
};
