import { getEnv } from './env';

export interface Config {
  goApiUrl: string;
  internalToken: string;
  cacheMaxAge: number;
  apiTimeoutMs: number;
}

export function getConfig(): Config {
  const goApiUrl = getEnv('GO_API_URL', 'http://localhost:8080');
  const internalToken = getEnv('INTERNAL_TOKEN', 'change-me');
  const cacheMaxAge = parseInt(getEnv('CACHE_MAX_AGE', '60'), 10);
  const apiTimeoutMs = parseInt(getEnv('API_TIMEOUT_MS', '5000'), 10);

  if (!goApiUrl) {
    throw new Error('GO_API_URL environment variable is required');
  }

  return {
    goApiUrl,
    internalToken,
    cacheMaxAge,
    apiTimeoutMs,
  };
}
