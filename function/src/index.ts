import { Hono } from 'hono';
import { cors } from 'hono/cors';
import type { AppEnv } from './types';
import { complianceLoggerMiddleware } from './utils/logger';
import { requestIdMiddleware, timeoutMiddleware, bodyLimitMiddleware, secureHeadersMiddleware } from './middleware/security';
import { clientIdMiddleware } from './middleware/client-id';
import { snapshotHandler } from './handlers/snapshot';
import { healthHandler } from './handlers/health';

const app = new Hono<AppEnv>();

// Global middleware chain (order matters)
app.use(requestIdMiddleware);
app.use(complianceLoggerMiddleware);
app.use(timeoutMiddleware);
app.use(bodyLimitMiddleware);
app.use(secureHeadersMiddleware);
app.use(cors());

// Health check (no auth required)
app.get('/healthz', healthHandler);

// Feature flags snapshot (requires X-Client-Id)
app.get('/v1/snapshot', clientIdMiddleware, snapshotHandler);

export default app;
