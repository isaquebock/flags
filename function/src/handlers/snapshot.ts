import type { Context } from 'hono';
import type { AppEnv, SnapshotResponse } from '../types';
import { fetchInternalSnapshot } from '../clients/api-client';
import { getConfig } from '../config';
import { error as logError } from '../utils/logger';

export async function snapshotHandler(c: Context<AppEnv>) {
  const clientId = c.get('clientId');
  const requestId = c.get('requestId');
  const config = getConfig();

  const internal = await fetchInternalSnapshot(clientId);

  if (internal.status === 404) {
    return c.json(
      {
        errors: [
          {
            status: 404,
            code: 'snapshot_not_found',
            message: `No flags found for client ${clientId}`,
          },
        ],
      },
      404
    );
  }

  if (internal.status >= 500) {
    logError('Failed to fetch snapshot from Go API', {
      requestId,
      clientId,
      status: internal.status,
      error: internal.error,
    });

    if (internal.status === 504) {
      return c.json(
        {
          errors: [
            {
              status: 504,
              code: 'go_api_timeout',
              message: 'Go API did not respond in time',
            },
          ],
        },
        504
      );
    }

    return c.json(
      {
        errors: [
          {
            status: 502,
            code: 'go_api_error',
            message: 'Failed to reach Go API',
          },
        ],
      },
      502
    );
  }

  if (!internal.body) {
    return c.json(
      {
        errors: [
          {
            status: 500,
            code: 'internal_error',
            message: 'Failed to parse Go API response',
          },
        ],
      },
      500
    );
  }

  const transformed: SnapshotResponse = {
    flags: Object.entries(internal.body.flags).reduce(
      (acc, [key, flag]) => {
        acc[key] = flag.enabled;
        return acc;
      },
      {} as Record<string, boolean>
    ),
    generated_at: internal.body.generated_at,
  };

  return c.json(transformed, 200, {
    'Cache-Control': `public, max-age=${config.cacheMaxAge}`,
    'Content-Type': 'application/json',
  });
}
