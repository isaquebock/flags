import type { Context, Next } from 'hono';

interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
  requestId?: string;
  method?: string;
  path?: string;
  status?: number;
  durationMs?: number;
  clientId?: string;
}

function log(entry: LogEntry) {
  console.log(JSON.stringify(entry));
}

export function info(message: string, data?: Record<string, any>) {
  log({
    timestamp: new Date().toISOString(),
    level: 'info',
    message,
    ...data,
  });
}

export function error(message: string, data?: Record<string, any>) {
  log({
    timestamp: new Date().toISOString(),
    level: 'error',
    message,
    ...data,
  });
}

export async function complianceLoggerMiddleware(c: Context, next: Next) {
  const requestId = crypto.randomUUID();
  const method = c.req.method;
  const path = c.req.path;
  const startTime = performance.now();

  const request = c.req.raw;
  await next();

  const durationMs = Math.round(performance.now() - startTime);
  const status = c.res.status;

  log({
    timestamp: new Date().toISOString(),
    level: 'info',
    message: 'request.completed',
    requestId,
    method,
    path,
    status,
    durationMs,
  });
}
