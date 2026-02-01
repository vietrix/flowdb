import { useCallback, useEffect, useRef, useState } from "react";
import { Play, Trash2, FlaskConical, Copy } from "lucide-react";
import type { QueryStreamResult } from "@/lib/types";

interface QueryEditorProps {
  query?: string;
  onQueryChange?: (query: string) => void;
  onExecute: (query: string) => Promise<{
    status: string;
    result?: QueryStreamResult;
    approvalId?: string;
  }>;
  onExplain?: (query: string) => Promise<Record<string, unknown> | null>;
  runToken?: number;
  explainToken?: number;
}

export function QueryEditor({
  query: queryProp,
  onQueryChange,
  onExecute,
  onExplain,
  runToken,
  explainToken,
}: QueryEditorProps) {
  const [query, setQuery] = useState(queryProp || "SELECT * FROM users LIMIT 10;");
  const [result, setResult] = useState<QueryStreamResult | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [running, setRunning] = useState(false);
  const [explainResult, setExplainResult] = useState<Record<string, unknown> | null>(null);
  const lastRunToken = useRef<number | null>(null);
  const lastExplainToken = useRef<number | null>(null);

  useEffect(() => {
    if (queryProp !== undefined && queryProp !== query) {
      setQuery(queryProp);
    }
  }, [queryProp, query]);

  const handleExecute = useCallback(async () => {
    setRunning(true);
    setStatusMessage(null);
    setExplainResult(null);
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
  }, [onExecute, query]);

  const handleExplain = useCallback(async () => {
    if (!onExplain) return;
    setRunning(true);
    setStatusMessage(null);
    setResult(null);
    try {
      const output = await onExplain(query);
      if (output) {
        setExplainResult(output);
      } else {
        setStatusMessage("Explain thất bại.");
      }
    } finally {
      setRunning(false);
    }
  }, [onExplain, query]);

  const handleClear = () => {
    setQuery("");
    setResult(null);
    setStatusMessage(null);
    setExplainResult(null);
  };

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(query);
      setStatusMessage("Đã copy query.");
    } catch {
      setStatusMessage("Không thể copy query.");
    }
  }, [query]);

  useEffect(() => {
    if (runToken !== undefined && runToken !== null && runToken !== lastRunToken.current) {
      lastRunToken.current = runToken;
      if (query.trim()) {
        handleExecute();
      }
    }
  }, [runToken, query, handleExecute]);

  useEffect(() => {
    if (explainToken !== undefined && explainToken !== null && explainToken !== lastExplainToken.current) {
      lastExplainToken.current = explainToken;
      if (query.trim()) {
        handleExplain();
      }
    }
  }, [explainToken, query, handleExplain]);

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="flex items-center gap-2 px-3 py-2 border-b bg-secondary/30">
        <button className="toolbar-button" onClick={handleExecute} disabled={running}>
          <Play size={12} />
          <span>{running ? "Running..." : "Run"}</span>
        </button>
        <button className="toolbar-button" onClick={handleExplain} disabled={running || !onExplain}>
          <FlaskConical size={12} />
          <span>Explain</span>
        </button>
        <button className="toolbar-button" onClick={handleCopy}>
          <Copy size={12} />
          <span>Copy</span>
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
          onChange={(e) => {
            setQuery(e.target.value);
            onQueryChange?.(e.target.value);
          }}
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
        ) : explainResult ? (
          <div className="flex-1 overflow-auto p-3">
            <pre className="text-xs bg-secondary/30 border rounded p-3 whitespace-pre-wrap">
              {JSON.stringify(explainResult, null, 2)}
            </pre>
          </div>
        ) : (
          <div className="flex-1 flex items-center justify-center text-muted-foreground text-sm">
            {statusMessage || "Run a query to see results"}
          </div>
        )}
      </div>
    </div>
  );
}
