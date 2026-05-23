import type { Context, Next } from 'hono';
import type { AppEnv } from '../types';
import { z } from 'zod';

const clientIdSchema = z
  .string()
  .min(1, 'Client ID is required')
  .max(128, 'Client ID must be at most 128 characters')
  .regex(
    /^[a-z0-9][a-z0-9-_]{0,127}$/,
    'Client ID must start with lowercase letter or digit, contain only lowercase letters, digits, hyphens, and underscores'
  );

export async function clientIdMiddleware(c: Context<AppEnv>, next: Next) {
  const clientId = c.req.header('x-client-id');

  if (!clientId) {
    return c.json(
      {
        errors: [
          {
            status: 400,
            code: 'missing_client_id',
            message: 'X-Client-Id header is required',
          },
        ],
      },
      400
    );
  }

  const result = clientIdSchema.safeParse(clientId);

  if (!result.success) {
    return c.json(
      {
        errors: [
          {
            status: 400,
            code: 'invalid_client_id',
            message: result.error.errors[0]?.message || 'Invalid client ID',
          },
        ],
      },
      400
    );
  }

  c.set('clientId', result.data);
  await next();
}
