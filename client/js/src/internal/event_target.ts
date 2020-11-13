import { DspsClientEventTarget, SubscriptionCallbackErrorInfo } from "../client_interface";
import { UnreachableCaseError } from "./errors";
import { console } from "./util/console";

export class DspsClientEventTargetImpl implements DspsClientEventTarget {
  private apiFailedHandlers: ((e: any) => void)[] = [];

  private subscriptionCallbackErrorHandlers: ((info: SubscriptionCallbackErrorInfo) => void)[] = [];

  addEventListener(type: "apiFailed" | "subscriptionCallbackError", listener: (e: any) => void): (...args: any[]) => void {
    switch (type) {
      case "apiFailed":
        this.apiFailedHandlers.push(listener);
        return listener;
      case "subscriptionCallbackError":
        this.subscriptionCallbackErrorHandlers.push(listener);
        return listener;
      default:
        throw new UnreachableCaseError(type);
    }
  }

  removeEventListener(type: string, listener: (...args: any[]) => void): void {
    function removeListElement<T>(list: T[], item: T) {
      const index = list.indexOf(item);
      if (index === -1) return;
      list.splice(index, 1);
    }
    switch (type) {
      case "apiFailed":
        removeListElement(this.apiFailedHandlers, listener);
        break;
      case "subscriptionCallbackError":
        removeListElement(this.subscriptionCallbackErrorHandlers, listener);
        break;
      default:
        break; // Do nothing.
    }
  }

  onApiFailed(e: any) {
    this.onEvent("error", "apiFailed", this.apiFailedHandlers, e);
  }

  onSubscriptionCallbackError(info: SubscriptionCallbackErrorInfo) {
    this.onEvent("error", "subscriptionCallbackError", this.subscriptionCallbackErrorHandlers, info);
  }

  private onEvent<T>(logLevel: null | "error", name: string, handlers: ((arg: T) => void)[], arg: T) {
    switch (logLevel) {
      case "error":
        console.error(`DSPS client event: ${name}`, arg);
        break;
      default:
        break;
    }

    for (const handler of handlers) {
      try {
        handler(arg);
      } catch (e) {
        console.error(`One of "${name}" event handler resulted in error. Event handler should not throw any error.`, e);
      }
    }
  }
}
