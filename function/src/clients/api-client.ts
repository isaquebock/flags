import type { InternalSnapshot } from '../types';
import { getConfig } from '../config';

export interface ApiResult<T> {
  status: number;
  body?: T;
  error?: string;
}

export async function fetchInternalSnapshot(
  clientId: string
): Promise<ApiResult<InternalSnapshot>> {
  const config = getConfig();

  try {
    const response = await fetch(
      `${config.goApiUrl}/internal/snapshot`,
      {
        method: 'GET',
        headers: {
          'X-Internal-Token': config.internalToken,
          'X-Client-Id': clientId,
        },
        signal: AbortSignal.timeout(config.apiTimeoutMs),
      }
    );

    if (!response.ok) {
      return {
        status: response.status,
        error: `Go API returned ${response.status}`,
      };
    }

    const body = (await response.json()) as InternalSnapshot;
    return {
      status: response.status,
      body,
    };
  } catch (err) {
    if (err instanceof Error && err.name === 'AbortError') {
      return {
        status: 504,
        error: 'Timeout calling Go API',
      };
    }

    return {
      status: 502,
      error: `Failed to fetch from Go API: ${err instanceof Error ? err.message : String(err)}`,
    };
  }
}
