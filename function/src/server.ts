// Development server for local testing
// In production, uses Azion Edge Function runtime

import { serve } from '@hono/node-server';
import app from './index';
import { getEnv } from './env';

const port = parseInt(getEnv('PORT', '3000'), 10);

serve({
  fetch: app.fetch,
  port,
}, (info) => {
  console.log(`✓ Server running on http://localhost:${info.port}`);
});
