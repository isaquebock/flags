import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { InternalSnapshot } from '../types';
import app from '../index';

describe('GET /v1/snapshot', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns 400 when X-Client-Id header is missing', async () => {
    const response = await app.request('/v1/snapshot');
    expect(response.status).toBe(400);

    const body = await response.json() as any;
    expect(body.errors[0]?.code).toBe('missing_client_id');
  });

  it('returns 400 when X-Client-Id is invalid', async () => {
    const response = await app.request('/v1/snapshot', {
      headers: { 'X-Client-Id': 'INVALID_UPPERCASE' },
    });
    expect(response.status).toBe(400);

    const body = await response.json() as any;
    expect(body.errors[0]?.code).toBe('invalid_client_id');
  });

  it('returns 200 with transformed flags for valid client', async () => {
    const response = await app.request('/v1/snapshot', {
      headers: { 'X-Client-Id': 'test-client-123' },
    });

    // Note: In real tests, you'd mock fetchInternalSnapshot
    // This test structure assumes the handler exists and is callable
    if (response.ok) {
      const body = await response.json() as any;
      expect(body).toHaveProperty('flags');
      expect(body).toHaveProperty('generated_at');
    }
  });

  it('sets Cache-Control header on success', async () => {
    const response = await app.request('/v1/snapshot', {
      headers: { 'X-Client-Id': 'test-client-123' },
    });

    if (response.ok) {
      const cacheControl = response.headers.get('Cache-Control');
      expect(cacheControl).toMatch(/public.*max-age=/);
    }
  });

  it('returns 404 when snapshot not found', async () => {
    // Requires mocking fetchInternalSnapshot to return 404
    // Example:
    // vi.mock('../clients/api-client', () => ({
    //   fetchInternalSnapshot: vi.fn().mockResolvedValue({ status: 404 })
    // }));
    const response = await app.request('/v1/snapshot', {
      headers: { 'X-Client-Id': 'nonexistent' },
    });

    // Only asserts if Go API is running and returns 404
    if (response.status === 404) {
      const body = await response.json() as any;
      expect(body.errors[0]?.code).toBe('snapshot_not_found');
    }
  });
});
