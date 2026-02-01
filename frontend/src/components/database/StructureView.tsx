import { Key, Hash } from "lucide-react";

interface ColumnDef {
  name: string;
  type: string;
  nullable: boolean;
  defaultValue: string | null;
  isPrimaryKey: boolean;
  isIndexed: boolean;
}

interface StructureViewProps {
  tableName: string;
  columns: ColumnDef[];
}

export function StructureView({ tableName, columns }: StructureViewProps) {
  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="px-3 py-2 border-b bg-secondary/30">
        <span className="text-sm font-medium">Structure: {tableName}</span>
      </div>
      
      <div className="flex-1 overflow-auto">
        <table className="data-table">
          <thead>
            <tr>
              <th style={{ width: 40 }}></th>
              <th>Column</th>
              <th>Type</th>
              <th>Nullable</th>
              <th>Default</th>
            </tr>
          </thead>
          <tbody>
            {columns.map((col) => (
              <tr key={col.name}>
                <td className="text-center">
                  {col.isPrimaryKey && (
                    <span title="Primary Key"><Key size={12} className="inline text-warning" /></span>
                  )}
                  {col.isIndexed && !col.isPrimaryKey && (
                    <span title="Indexed"><Hash size={12} className="inline text-muted-foreground" /></span>
                  )}
                </td>
                <td className="font-mono">{col.name}</td>
                <td className="font-mono text-muted-foreground">{col.type}</td>
                <td>{col.nullable ? "Yes" : "No"}</td>
                <td className="font-mono text-muted-foreground">
                  {col.defaultValue ?? <span className="italic">None</span>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
