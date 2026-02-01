import { ChevronDown, ChevronUp, Trash2 } from "lucide-react";

interface LogEntry {
  id: string;
  timestamp: string;
  type: "query" | "error" | "info";
  message: string;
  duration?: number;
}

interface BottomPanelProps {
  logs: LogEntry[];
  collapsed: boolean;
  onToggleCollapse: () => void;
  onClearLogs: () => void;
}

export function BottomPanel({ logs, collapsed, onToggleCollapse, onClearLogs }: BottomPanelProps) {
  return (
    <div className={`app-bottom-panel ${collapsed ? 'collapsed' : ''}`}>
      <div className="panel-header">
        <div className="flex items-center gap-2">
          <button
            className="hover:bg-accent p-0.5 rounded"
            onClick={onToggleCollapse}
          >
            {collapsed ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
          </button>
          <span>Query Log</span>
          <span className="text-muted-foreground font-normal">({logs.length})</span>
        </div>
        
        {!collapsed && (
          <button
            className="hover:bg-accent p-0.5 rounded"
            onClick={onClearLogs}
            title="Clear logs"
          >
            <Trash2 size={12} />
          </button>
        )}
      </div>
      
      {!collapsed && (
        <div className="flex-1 overflow-auto">
          {logs.length === 0 ? (
            <div className="p-3 text-sm text-muted-foreground">No queries logged</div>
          ) : (
            logs.map((log) => (
              <div
                key={log.id}
                className={`log-entry ${log.type === 'error' ? 'error' : ''} ${log.type === 'query' ? 'success' : ''}`}
              >
                <span className="text-muted-foreground">[{log.timestamp}]</span>
                {log.duration !== undefined && (
                  <span className="text-muted-foreground ml-2">({log.duration}ms)</span>
                )}
                <span className="ml-2">{log.message}</span>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
