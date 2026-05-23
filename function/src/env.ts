// LOCKED: Environment variable bridge
// Checks Azion args first, then falls back to process.env (for Node dev)

let _envArgs: Record<string, any> = {};

export function setAzionArgs(args?: Record<string, any>) {
  _envArgs = args || {};
}

export function getEnv(key: string, defaultValue?: string): string {
  if (_envArgs && _envArgs[key]) {
    return String(_envArgs[key]);
  }
  if (typeof process !== 'undefined' && process.env) {
    return process.env[key] || defaultValue || '';
  }
  return defaultValue || '';
}
