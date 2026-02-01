"use client";

import { useEffect, useState } from "react";
import { AppShell } from "@/components/database/AppShell";
import { listHistory } from "@/lib/api";
import type { QueryHistoryEntry } from "@/lib/types";

export default function HistoryPage() {
  const [entries, setEntries] = useState<QueryHistoryEntry[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await listHistory();
        setEntries(data);
      } catch (err) {
        setError((err as Error).message || "Không tải được history.");
      }
    };
    load();
  }, []);

  return (
    <AppShell>
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-3 py-2 border-b bg-secondary/30">
          <div className="text-sm font-medium">Query History</div>
        </div>
        <div className="flex-1 overflow-auto">
          {error && <div className="p-3 text-sm text-destructive">{error}</div>}
          <table className="data-table">
            <thead>
              <tr>
                <th>Started</th>
                <th>Status</th>
                <th>Rows</th>
                <th>Duration</th>
                <th>Action</th>
                <th>Resource</th>
                <th>Statement Hash</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry) => (
                <tr key={entry.id}>
                  <td>{entry.startedAt}</td>
                  <td>{entry.status}</td>
                  <td>{entry.rowCount}</td>
                  <td>{entry.durationMs}ms</td>
                  <td>{entry.action}</td>
                  <td>{entry.resource}</td>
                  <td className="font-mono">{entry.statementHash}</td>
                </tr>
              ))}
              {entries.length === 0 && (
                <tr>
                  <td colSpan={7} className="text-center text-muted-foreground">
                    No history entries
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
