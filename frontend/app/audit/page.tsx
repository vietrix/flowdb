"use client";

import { useEffect, useState } from "react";
import { AppShell } from "@/components/database/AppShell";
import { listAudit } from "@/lib/api";
import type { AuditEntry } from "@/lib/types";

export default function AuditPage() {
  const [entries, setEntries] = useState<AuditEntry[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await listAudit();
        setEntries(data);
      } catch (err) {
        setError((err as Error).message || "Không tải được audit log.");
      }
    };
    load();
  }, []);

  return (
    <AppShell>
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-3 py-2 border-b bg-secondary/30">
          <div className="text-sm font-medium">Audit Log</div>
        </div>
        <div className="flex-1 overflow-auto">
          {error && <div className="p-3 text-sm text-destructive">{error}</div>}
          <table className="data-table">
            <thead>
              <tr>
                <th>Time</th>
                <th>Event</th>
                <th>Actor</th>
                <th>Error ID</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry) => (
                <tr key={entry.id}>
                  <td>{entry.createdAt}</td>
                  <td>{entry.eventType}</td>
                  <td>{entry.actorId || "-"}</td>
                  <td className="font-mono">{entry.errorId || "-"}</td>
                </tr>
              ))}
              {entries.length === 0 && (
                <tr>
                  <td colSpan={4} className="text-center text-muted-foreground">
                    No audit entries
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </AppShell>
  );
}
