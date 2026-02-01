"use client";

import { useEffect, useState } from "react";
import { AppShell } from "@/components/database/AppShell";
import { approveApproval, denyApproval, listApprovals } from "@/lib/api";
import type { QueryApproval } from "@/lib/types";

export default function ApprovalsPage() {
  const [approvals, setApprovals] = useState<QueryApproval[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    try {
      const data = await listApprovals();
      setApprovals(data);
    } catch (err) {
      setError((err as Error).message || "Không tải được approvals.");
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleApprove = async (id: string) => {
    await approveApproval(id);
    await load();
  };

  const handleDeny = async (id: string) => {
    await denyApproval(id);
    await load();
  };

  return (
    <AppShell>
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-3 py-2 border-b bg-secondary/30">
          <div className="text-sm font-medium">Query Approvals</div>
        </div>
        <div className="flex-1 overflow-auto">
          {error && <div className="p-3 text-sm text-destructive">{error}</div>}
          <table className="data-table">
            <thead>
              <tr>
                <th>Connection</th>
                <th>User</th>
                <th>Environment</th>
                <th>Status</th>
                <th>Statement</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {approvals.map((approval) => (
                <tr key={approval.id}>
                  <td>{approval.connectionId}</td>
                  <td>{approval.userId}</td>
                  <td>{approval.environment}</td>
                  <td>{approval.status}</td>
                  <td className="font-mono">{approval.statement}</td>
                  <td className="space-x-1">
                    <button className="toolbar-button px-2" onClick={() => handleApprove(approval.id)}>
                      Approve
                    </button>
                    <button className="toolbar-button px-2" onClick={() => handleDeny(approval.id)}>
                      Deny
                    </button>
                  </td>
                </tr>
              ))}
              {approvals.length === 0 && (
                <tr>
                  <td colSpan={6} className="text-center text-muted-foreground">
                    No pending approvals
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
