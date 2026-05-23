import type { Context, Next } from 'hono';
import type { AppEnv } from '../types';
import { logger } from '../utils/logger';

export async function requestIdMiddleware(c: Context<AppEnv>, next: Next) {
  const requestId = crypto.randomUUID();
  c.set('requestId', requestId);
  c.header('X-Request-Id', requestId);
  await next();
}

export async function timeoutMiddleware(c: Context, next: Next) {
  const timeout = 30000; // 30 seconds
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    c.env.signal = controller.signal;
    await next();
  } finally {
    clearTimeout(timeoutId);
  }
}

export async function bodyLimitMiddleware(c: Context, next: Next) {
  const maxSize = 8 * 1024; // 8 KB
  const request = c.req.raw;
  const contentLength = parseInt(request.headers.get('content-length') || '0', 10);

  if (contentLength > maxSize) {
    return c.json(
      {
        errors: [
          {
            status: 413,
            code: 'body_too_large',
            message: `Request body exceeds maximum size of ${maxSize} bytes`,
          },
        ],
      },
      413
    );
  }

  await next();
}

export async function secureHeadersMiddleware(c: Context, next: Next) {
  c.header('X-Content-Type-Options', 'nosniff');
  c.header('X-Frame-Options', 'DENY');
  c.header('X-XSS-Protection', '1; mode=block');
  c.header('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');

  await next();
}
