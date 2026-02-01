import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, RefreshCw } from "lucide-react";

interface Column {
  key: string;
  label: string;
}

interface DataViewerProps {
  tableName: string;
  columns: Column[];
  rows: Record<string, unknown>[];
  currentPage: number;
  totalPages: number;
  totalRows: number;
  onPageChange: (page: number) => void;
  onRefresh: () => void;
}

export function DataViewer({
  tableName,
  columns,
  rows,
  currentPage,
  totalPages,
  totalRows,
  onPageChange,
  onRefresh,
}: DataViewerProps) {
  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-3 py-2 border-b bg-secondary/30">
        <div className="text-sm">
          <span className="font-medium">{tableName}</span>
          <span className="text-muted-foreground ml-2">({totalRows} rows)</span>
        </div>
        
        <button className="toolbar-button" onClick={onRefresh}>
          <RefreshCw size={12} />
          <span>Refresh</span>
        </button>
      </div>
      
      <div className="flex-1 overflow-auto">
        <table className="data-table">
          <thead>
            <tr>
              {columns.map((col) => (
                <th key={col.key}>{col.label}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, idx) => (
              <tr key={idx}>
                {columns.map((col) => (
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
      
      <div className="flex items-center justify-between px-3 py-2 border-t bg-secondary/30">
        <div className="text-sm text-muted-foreground">
          Page {currentPage} of {totalPages}
        </div>
        
        <div className="flex items-center gap-1">
          <button
            className="toolbar-button px-2"
            onClick={() => onPageChange(1)}
            disabled={currentPage === 1}
          >
            <ChevronsLeft size={14} />
          </button>
          <button
            className="toolbar-button px-2"
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage === 1}
          >
            <ChevronLeft size={14} />
          </button>
          <button
            className="toolbar-button px-2"
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage === totalPages}
          >
            <ChevronRight size={14} />
          </button>
          <button
            className="toolbar-button px-2"
            onClick={() => onPageChange(totalPages)}
            disabled={currentPage === totalPages}
          >
            <ChevronsRight size={14} />
          </button>
        </div>
      </div>
    </div>
  );
}
