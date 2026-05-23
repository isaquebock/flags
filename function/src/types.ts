import type { Context } from 'hono';

export type AppEnv = {
  Variables: {
    clientId: string;
    requestId: string;
  };
};

export type FlagFull = {
  enabled: boolean;
  description: string;
  created_at: string;
  updated_at: string;
};

export type InternalSnapshot = {
  schema_version: number;
  client_id: string;
  generated_at: string;
  flags: Record<string, FlagFull>;
};

export type SnapshotResponse = {
  flags: Record<string, boolean>;
  generated_at: string;
};

export type ApiError = {
  errors: Array<{
    status: number;
    code: string;
    message: string;
  }>;
};
