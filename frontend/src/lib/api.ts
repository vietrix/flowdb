import type {
  UserProfile,
  Settings,
  Connection,
  Namespace,
  Entity,
  EntityInfo,
  BrowseResult,
  QueryStartResponse,
  QueryStreamResult,
  QueryHistoryEntry,
  AuditEntry,
  QueryApproval,
} from "@/lib/types";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "";

function buildUrl(path: string) {
  if (!API_BASE) {
    return path;
  }
  return `${API_BASE}${path}`;
}

function getCSRFToken() {
  if (typeof window === "undefined") return "";
  return localStorage.getItem("flowdb_csrf") || "";
}

export function setCSRFToken(token: string) {
  if (typeof window === "undefined") return;
  localStorage.setItem("flowdb_csrf", token);
}

async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const method = (init.method || "GET").toUpperCase();
  const headers = new Headers(init.headers || {});
  if (!headers.has("Content-Type") && init.body) {
    headers.set("Content-Type", "application/json");
  }
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    const csrf = getCSRFToken();
    if (csrf) {
      headers.set("X-CSRF-Token", csrf);
    }
  }
  const res = await fetch(buildUrl(path), {
    ...init,
    headers,
    credentials: "include",
  });
  if (res.status === 204) {
    return undefined as T;
  }
  const text = await res.text();
  let data: unknown = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
  }
  if (!res.ok) {
    const message = typeof data === "string" ? data : data?.message || res.statusText;
    throw new Error(message);
  }
  return data as T;
}

export async function login(username: string, password: string, mfaCode?: string) {
  const body: Record<string, string> = { username, password };
  if (mfaCode) {
    body.mfaCode = mfaCode;
  }
  const res = await apiFetch<{ id: string; username: string; csrfToken: string }>(
    "/api/v1/auth/login",
    { method: "POST", body: JSON.stringify(body) }
  );
  if (res?.csrfToken) {
    setCSRFToken(res.csrfToken);
  }
  return res;
}

export async function logout() {
  await apiFetch<void>("/api/v1/auth/logout", { method: "POST" });
}

export async function getMe() {
  return apiFetch<UserProfile & { settings?: Settings }>("/api/v1/auth/me");
}

export async function getSettings() {
  const raw = await apiFetch<unknown>("/api/v1/settings");
  return normalizeSettings(raw);
}

export async function updateSettings(settings: Settings) {
  const raw = await apiFetch<unknown>("/api/v1/settings", {
    method: "PUT",
    body: JSON.stringify({
      securityMode: settings.securityMode,
      flags: settings.flags,
      config: settings.config,
      ipAllowlist: settings.ipAllowlist,
    }),
  });
  return normalizeSettings(raw);
}

export async function updateSecurityMode(mode: string) {
  const raw = await apiFetch<unknown>("/api/v1/settings/security-mode", {
    method: "PUT",
    body: JSON.stringify({ mode }),
  });
  return normalizeSettings(raw);
}

export async function updateFlags(flags: Record<string, boolean>) {
  const raw = await apiFetch<unknown>("/api/v1/settings/flags", {
    method: "PUT",
    body: JSON.stringify({ flags }),
  });
  return normalizeSettings(raw);
}

export async function listConnections() {
  const raw = await apiFetch<unknown>("/api/v1/connections");
  return asArray(raw).map(normalizeConnection);
}

export async function getConnection(id: string) {
  const raw = await apiFetch<unknown>(`/api/v1/connections/${id}`);
  return normalizeConnection(raw);
}

export async function createConnection(payload: {
  name: string;
  type: string;
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  tls?: Record<string, unknown>;
  tags?: Record<string, unknown>;
}) {
  const raw = await apiFetch<unknown>("/api/v1/connections", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  return normalizeConnection(raw);
}

export async function updateConnection(
  id: string,
  payload: {
    name: string;
    type: string;
    host: string;
    port: number;
    database: string;
    username: string;
    password?: string;
    tls?: Record<string, unknown>;
    tags?: Record<string, unknown>;
  }
) {
  const raw = await apiFetch<unknown>(`/api/v1/connections/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
  return normalizeConnection(raw);
}

export async function deleteConnection(id: string) {
  return apiFetch<void>(`/api/v1/connections/${id}`, { method: "DELETE" });
}

export async function testConnection(id: string) {
  return apiFetch<{ status: string }>(`/api/v1/connections/${id}/test`, { method: "POST" });
}

export async function listNamespaces(connectionId: string) {
  return apiFetch<Namespace[]>(`/api/v1/connections/${connectionId}/namespaces`);
}

export async function listEntities(connectionId: string, ns: string) {
  return apiFetch<Entity[]>(
    `/api/v1/connections/${connectionId}/entities?ns=${encodeURIComponent(ns)}`
  );
}

export async function getEntityInfo(connectionId: string, ns: string, name: string) {
  return apiFetch<EntityInfo>(
    `/api/v1/connections/${connectionId}/entities/${encodeURIComponent(
      name
    )}/info?ns=${encodeURIComponent(ns)}`
  );
}

export async function browseEntity(
  connectionId: string,
  ns: string,
  name: string,
  page: number,
  pageSize: number
) {
  return apiFetch<BrowseResult>(
    `/api/v1/connections/${connectionId}/entities/${encodeURIComponent(
      name
    )}/browse?ns=${encodeURIComponent(ns)}&page=${page}&pageSize=${pageSize}`
  );
}

export async function startQuery(
  connectionId: string,
  statement: string,
  options?: { approvalId?: string; maxRows?: number; timeoutMs?: number }
) {
  return apiFetch<QueryStartResponse>(`/api/v1/connections/${connectionId}/query`, {
    method: "POST",
    body: JSON.stringify({
      statement,
      approvalId: options?.approvalId,
      maxRows: options?.maxRows,
      timeoutMs: options?.timeoutMs,
    }),
  });
}

function getWebSocketBase() {
  if (API_BASE) {
    return API_BASE.replace(/^http/, "ws");
  }
  if (typeof window !== "undefined") {
    return window.location.origin.replace(/^http/, "ws");
  }
  return "";
}

export function streamQuery(
  connectionId: string,
  queryId: string,
  handlers: {
    onSchema?: (columns: string[]) => void;
    onRow?: (row: unknown[] | Record<string, unknown>) => void;
    onEnd?: (rowCount: number, durationMs: number) => void;
    onError?: (message: string) => void;
  }
) {
  const wsBase = getWebSocketBase();
  const socket = new WebSocket(
    `${wsBase}/api/v1/connections/${connectionId}/query/${queryId}/stream`
  );
  socket.onmessage = (event) => {
    const payload = JSON.parse(event.data);
    switch (payload.type) {
      case "schema": {
        if (payload.columns) {
          handlers.onSchema?.(payload.columns.map((c: { name: string }) => c.name));
        } else if (payload.fields) {
          handlers.onSchema?.(payload.fields as string[]);
        }
        break;
      }
      case "rows": {
        for (const row of payload.rows || []) {
          handlers.onRow?.(row);
        }
        break;
      }
      case "end": {
        handlers.onEnd?.(payload.rowCount || 0, payload.durationMs || 0);
        socket.close();
        break;
      }
      case "error": {
        handlers.onError?.(payload.message || "error");
        socket.close();
        break;
      }
      default:
        break;
    }
  };
  return socket;
}

export async function runQuery(
  connectionId: string,
  statement: string
): Promise<{ status: string; result?: QueryStreamResult; approvalId?: string }> {
  const start = await startQuery(connectionId, statement);
  if (start.status !== "ready" || !start.queryId) {
    return { status: start.status, approvalId: start.approvalId };
  }
  return new Promise((resolve, reject) => {
    const rows: Record<string, unknown>[] = [];
    let columns: string[] = [];
    let durationMs = 0;
    let rowCount = 0;
    streamQuery(connectionId, start.queryId!, {
      onSchema: (cols) => {
        columns = cols;
      },
      onRow: (row) => {
        if (Array.isArray(row)) {
          const mapped: Record<string, unknown> = {};
          for (let i = 0; i < columns.length; i++) {
            mapped[columns[i]] = row[i];
          }
          rows.push(mapped);
        } else if (row && typeof row === "object") {
          rows.push(row as Record<string, unknown>);
          if (columns.length === 0) {
            columns = Object.keys(row as Record<string, unknown>);
          }
        }
        rowCount += 1;
      },
      onEnd: (count, duration) => {
        durationMs = duration;
        const result: QueryStreamResult = {
          columns: columns.map((c) => ({ key: c, label: c })),
          rows,
          executionTime: durationMs,
          affectedRows: count || rowCount,
        };
        resolve({ status: "ok", result });
      },
      onError: (message) => {
        reject(new Error(message));
      },
    });
  });
}

export async function listHistory(limit = 100, offset = 0) {
  const raw = await apiFetch<unknown>(`/api/v1/history?limit=${limit}&offset=${offset}`);
  return asArray(raw).map(normalizeHistory);
}

export async function listAudit(limit = 100, offset = 0) {
  const raw = await apiFetch<unknown>(`/api/v1/audit?limit=${limit}&offset=${offset}`);
  return asArray(raw).map(normalizeAudit);
}

export async function listApprovals() {
  const raw = await apiFetch<unknown>("/api/v1/approvals/pending");
  return asArray(raw).map(normalizeApproval);
}

export async function approveApproval(id: string) {
  return apiFetch<{ status: string }>(`/api/v1/approvals/${id}/approve`, { method: "POST" });
}

export async function denyApproval(id: string) {
  return apiFetch<{ status: string }>(`/api/v1/approvals/${id}/deny`, { method: "POST" });
}

function normalizeConnection(raw: unknown): Connection {
  const record = asRecord(raw);
  return {
    id: asString(record.id ?? record.ID),
    name: asString(record.name ?? record.Name),
    type: asString(record.type ?? record.Type),
    host: asString(record.host ?? record.Host),
    port: asNumber(record.port ?? record.Port),
    database: asString(record.database ?? record.Database),
    username: asString(record.username ?? record.Username),
    tls: asRecord(record.tls ?? record.TLS),
    tags: asRecord(record.tags ?? record.Tags),
  };
}

function normalizeSettings(raw: unknown): Settings {
  const record = asRecord(raw);
  return {
    securityMode: asString(record.securityMode ?? record.SecurityMode) || "basic",
    flags: asRecord(record.flags ?? record.Flags) as Record<string, boolean>,
    config: asRecord(record.config ?? record.Config),
    ipAllowlist: asStringArray(record.ipAllowlist ?? record.IPAllowlist),
    updatedAt: asString(record.updatedAt ?? record.UpdatedAt),
  };
}

function normalizeHistory(raw: unknown): QueryHistoryEntry {
  const record = asRecord(raw);
  return {
    id: asString(record.id ?? record.ID),
    userId: asString(record.userId ?? record.UserID) || null,
    connectionId: asString(record.connectionId ?? record.ConnectionID) || null,
    statementHash: asString(record.statementHash ?? record.StatementHash),
    status: asString(record.status ?? record.Status),
    rowCount: asNumber(record.rowCount ?? record.RowCount),
    durationMs: asNumber(record.durationMs ?? record.DurationMs),
    startedAt: asString(record.startedAt ?? record.StartedAt),
    endedAt: asString(record.endedAt ?? record.EndedAt),
    action: asString(record.action ?? record.Action),
    resource: asString(record.resource ?? record.Resource),
  };
}

function normalizeAudit(raw: unknown): AuditEntry {
  const record = asRecord(raw);
  return {
    id: asString(record.id ?? record.ID),
    eventType: asString(record.eventType ?? record.EventType),
    actorId: asString(record.actorId ?? record.ActorID) || null,
    details: record.details ?? record.Details,
    createdAt: asString(record.createdAt ?? record.CreatedAt),
    prevHash: asString(record.prevHash ?? record.PrevHash),
    hash: asString(record.hash ?? record.Hash),
    errorId: asString(record.errorId ?? record.ErrorID),
  };
}

function normalizeApproval(raw: unknown): QueryApproval {
  const record = asRecord(raw);
  return {
    id: asString(record.id ?? record.ID),
    connectionId: asString(record.connectionId ?? record.ConnectionID),
    userId: asString(record.userId ?? record.UserID),
    statement: asString(record.statement ?? record.Statement),
    status: asString(record.status ?? record.Status),
    environment: asString(record.environment ?? record.Environment),
    createdAt: asString(record.createdAt ?? record.CreatedAt),
  };
}

function asRecord(value: unknown): Record<string, unknown> {
  if (value && typeof value === "object") {
    return value as Record<string, unknown>;
  }
  return {};
}

function asArray(value: unknown): Record<string, unknown>[] {
  if (Array.isArray(value)) {
    return value as Record<string, unknown>[];
  }
  return [];
}

function asString(value: unknown): string {
  if (typeof value === "string") return value;
  if (value === undefined || value === null) return "";
  return String(value);
}

function asNumber(value: unknown): number {
  if (typeof value === "number") return value;
  if (typeof value === "string") return Number(value) || 0;
  return 0;
}

function asStringArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map((v) => asString(v));
  }
  if (typeof value === "string") {
    return value.split(",").map((v) => v.trim()).filter(Boolean);
  }
  return [];
}
