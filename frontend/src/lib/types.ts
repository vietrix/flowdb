export interface UserProfile {
  id: string;
  username: string;
  isAdmin: boolean;
  mfaEnabled: boolean;
}

export interface Settings {
  securityMode: string;
  flags: Record<string, boolean>;
  config: Record<string, unknown>;
  ipAllowlist: string[];
  updatedAt?: string;
}

export interface Connection {
  id: string;
  name: string;
  type: "postgres" | "mongodb" | string;
  host: string;
  port: number;
  database: string;
  username: string;
  tls: Record<string, unknown>;
  tags: Record<string, unknown>;
}

export interface TreeNode {
  id: string;
  name: string;
  type: "connection" | "namespace" | "table" | "view" | "index";
  namespace?: string;
  connectionId?: string;
  children?: TreeNode[];
}

export interface Namespace {
  name: string;
}

export interface Entity {
  name: string;
}

export interface EntityInfo {
  columns: { name: string; type: string }[];
  indexes: string[];
  stats: Record<string, unknown>;
}

export interface BrowseResult {
  columns: { name: string; type: string }[];
  rows?: unknown[][];
  docs?: Record<string, unknown>[];
}

export interface QueryStartResponse {
  queryId?: string;
  status: string;
  approvalId?: string;
}

export interface QueryStreamResult {
  columns: { key: string; label: string }[];
  rows: Record<string, unknown>[];
  executionTime: number;
  affectedRows: number;
}

export interface QueryHistoryEntry {
  id: string;
  userId: string | null;
  connectionId: string | null;
  statementHash: string;
  status: string;
  rowCount: number;
  durationMs: number;
  startedAt: string;
  endedAt?: string | null;
  action?: string;
  resource?: string;
}

export interface AuditEntry {
  id: string;
  eventType: string;
  actorId?: string | null;
  details: unknown;
  createdAt: string;
  prevHash?: string | null;
  hash?: string | null;
  errorId?: string | null;
}

export interface QueryApproval {
  id: string;
  connectionId: string;
  userId: string;
  statement: string;
  status: string;
  environment: string;
  createdAt: string;
}
