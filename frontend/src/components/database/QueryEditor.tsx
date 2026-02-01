import { useState } from "react";
import { Play, Trash2 } from "lucide-react";
import type { QueryStreamResult } from "@/lib/types";

interface QueryEditorProps {
  onExecute: (query: string) => Promise<{
    status: string;
    result?: QueryStreamResult;
    approvalId?: string;
  }>;
}

export function QueryEditor({ onExecute }: QueryEditorProps) {
  const [query, setQuery] = useState("SELECT * FROM users LIMIT 10;");
  const [result, setResult] = useState<QueryStreamResult | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [running, setRunning] = useState(false);

  const handleExecute = async () => {
    setRunning(true);
    setStatusMessage(null);
    try {
      const res = await onExecute(query);
      if (res.status === "ok" && res.result) {
        setResult(res.result);
      } else if (res.status === "pending_approval") {
        setResult(null);
        setStatusMessage(`Chờ phê duyệt: ${res.approvalId || ""}`);
      } else if (res.status === "no_connection") {
        setResult(null);
        setStatusMessage("Chưa chọn kết nối.");
      } else if (res.status === "error") {
        setResult(null);
        setStatusMessage("Query thất bại.");
      }
    } finally {
      setRunning(false);
    }
  };

  const handleClear = () => {
    setQuery("");
    setResult(null);
    setStatusMessage(null);
  };

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="flex items-center gap-2 px-3 py-2 border-b bg-secondary/30">
        <button className="toolbar-button" onClick={handleExecute} disabled={running}>
          <Play size={12} />
          <span>{running ? "Running..." : "Run"}</span>
        </button>
        <button className="toolbar-button" onClick={handleClear}>
          <Trash2 size={12} />
          <span>Clear</span>
        </button>
      </div>
      
      <div className="h-40 border-b">
        <textarea
          className="query-editor"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Enter SQL query..."
          spellCheck={false}
        />
      </div>
      
      <div className="flex-1 overflow-hidden flex flex-col">
        {result ? (
          <>
            <div className="px-3 py-1.5 border-b bg-secondary/30 text-sm">
              <span className="text-muted-foreground">
                {result.rows.length} rows returned in {result.executionTime}ms
              </span>
            </div>
            
            <div className="flex-1 overflow-auto">
              <table className="data-table">
                <thead>
                  <tr>
                    {result.columns.map((col) => (
                      <th key={col.key}>{col.label}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {result.rows.map((row, idx) => (
                    <tr key={idx}>
                      {result.columns.map((col) => (
                        <td key={col.key}>
                          {row[col.key] === null ? (
                            <span className="text-muted-foreground italic">NULL</span>
                          ) : (
                            String(row[col.key])
                          )}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-muted-foreground text-sm">
            {statusMessage || "Run a query to see results"}
          </div>
        )}
      </div>
    </div>
  );
}
