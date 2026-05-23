import type { Context } from 'hono';

export async function healthHandler(c: Context) {
  return c.json({ ok: true }, 200);
}
