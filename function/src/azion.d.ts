// LOCKED: Azion-specific type definitions
// Do not modify without understanding Azion Edge Function runtime

declare global {
  interface FetchEvent extends Event {
    request: Request;
    args?: Record<string, any>;
    respondWith(response: Response | Promise<Response>): void;
  }

  var _azionArgs: Record<string, any> | undefined;
  function setAzionArgs(args?: Record<string, any>): void;
}

export {};
