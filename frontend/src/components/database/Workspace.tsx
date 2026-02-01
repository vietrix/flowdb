import { useCallback, useEffect, useMemo, useState } from "react";
import { DataViewer } from "./DataViewer";
import { QueryEditor } from "./QueryEditor";
import { StructureView } from "./StructureView";
import { VisualQueryBuilder } from "./VisualQueryBuilder";
import { browseEntity, explainQuery, getEntityInfo, runQuery } from "@/lib/api";

type TabType = "data" | "query" | "visual" | "structure";

interface WorkspaceProps {
  connectionId: string | null;
  connectionType: string | null;
  selectedNamespace: string | null;
  selectedEntity: string | null;
  onLog: (entry: { type: "query" | "error" | "info"; message: string; duration?: number }) => void;
}

export function Workspace({ connectionId, connectionType, selectedNamespace, selectedEntity, onLog }: WorkspaceProps) {
  const [activeTab, setActiveTab] = useState<TabType>("data");
  const [currentPage, setCurrentPage] = useState(1);
  const [columns, setColumns] = useState<{ key: string; label: string }[]>([]);
  const [rows, setRows] = useState<Record<string, unknown>[]>([]);
  const [totalRows, setTotalRows] = useState(0);
  const [queryDraft, setQueryDraft] = useState("SELECT * FROM users LIMIT 10;");
  const [runToken, setRunToken] = useState(0);
  const [explainToken, setExplainToken] = useState(0);
  const [structure, setStructure] = useState<
    { name: string; type: string; nullable: boolean; defaultValue: string | null; isPrimaryKey: boolean; isIndexed: boolean }[]
  >([]);

  const tableName = selectedEntity || "table";
  const canLoad = Boolean(connectionId && selectedNamespace && selectedEntity);
  const qualifiedName = useMemo(() => {
    if (!selectedEntity) return "table";
    if (connectionType === "mongodb") return selectedEntity;
    if (selectedNamespace) return `${selectedNamespace}.${selectedEntity}`;
    return selectedEntity;
  }, [connectionType, selectedEntity, selectedNamespace]);

  const handleExecuteQuery = async (queryText: string) => {
    if (!connectionId) return { status: "no_connection" };
    try {
      const result = await runQuery(connectionId, queryText);
      if (result.status === "ok" && result.result) {
        onLog({ type: "query", message: queryText, duration: result.result.executionTime });
      } else if (result.status === "pending_approval") {
        onLog({ type: "info", message: `Pending approval: ${result.approvalId || ""}` });
      }
      return result;
    } catch (err) {
      onLog({ type: "error", message: (err as Error).message || "Query failed" });
      return { status: "error" };
    }
  };

  const handleExplain = useCallback(async (queryText: string) => {
    if (!connectionId) return null;
    try {
      const result = await explainQuery(connectionId, queryText);
      onLog({ type: "info", message: "Explain completed." });
      return result;
    } catch (err) {
      onLog({ type: "error", message: (err as Error).message || "Explain thất bại." });
      return null;
    }
  }, [connectionId, onLog]);

  const loadData = useCallback(async (page: number) => {
    if (!canLoad) return;
    try {
      const res = await browseEntity(connectionId!, selectedNamespace!, selectedEntity!, page, 50);
      const cols = res.columns?.map((c) => ({ key: c.name, label: c.name })) || [];
      if (res.rows && res.rows.length > 0) {
        const mapped = res.rows.map((row) => {
          const record: Record<string, unknown> = {};
          for (let i = 0; i < cols.length; i++) {
            record[cols[i].key] = row[i];
          }
          return record;
        });
        setRows(mapped);
        setColumns(cols);
        setTotalRows(mapped.length);
      } else if (res.docs && res.docs.length > 0) {
        const docCols = Object.keys(res.docs[0] || {}).map((k) => ({ key: k, label: k }));
        setColumns(docCols);
        setRows(res.docs);
        setTotalRows(res.docs.length);
      } else {
        setRows([]);
        setColumns(cols);
        setTotalRows(0);
      }
    } catch (err) {
      onLog({ type: "error", message: (err as Error).message || "Browse thất bại." });
    }
  }, [canLoad, connectionId, selectedEntity, selectedNamespace, onLog]);

  const loadStructure = useCallback(async () => {
    if (!canLoad) return;
    try {
      const info = await getEntityInfo(connectionId!, selectedNamespace!, selectedEntity!);
      const mapped = (info.columns || []).map((col) => ({
        name: col.name,
        type: col.type,
        nullable: true,
        defaultValue: null,
        isPrimaryKey: false,
        isIndexed: (info.indexes || []).some((idx) => idx.includes(col.name)),
      }));
      setStructure(mapped);
    } catch (err) {
      onLog({ type: "error", message: (err as Error).message || "Load structure thất bại." });
    }
  }, [canLoad, connectionId, selectedEntity, selectedNamespace, onLog]);

  useEffect(() => {
    if (activeTab === "data") {
      loadData(currentPage);
    }
  }, [activeTab, currentPage, loadData]);

  useEffect(() => {
    if (activeTab === "structure") {
      loadStructure();
    }
  }, [activeTab, loadStructure]);

  useEffect(() => {
    if (!canLoad) {
      setRows([]);
      setColumns([]);
      setTotalRows(0);
      setStructure([]);
    }
  }, [canLoad]);

  useEffect(() => {
    if (!selectedEntity) return;
    if (connectionType === "mongodb") {
      setQueryDraft(
        JSON.stringify({ find: selectedEntity, filter: {}, limit: 100 }, null, 2)
      );
    } else {
      setQueryDraft(`SELECT * FROM ${qualifiedName} LIMIT 100;`);
    }
  }, [connectionType, qualifiedName, selectedEntity]);

  const tabs: { id: TabType; label: string }[] = [
    { id: "data", label: "Data" },
    { id: "query", label: "Query" },
    { id: "visual", label: "Visual Builder" },
    { id: "structure", label: "Structure" },
  ];

  return (
    <main className="app-workspace">
      <div className="flex border-b bg-secondary/30">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            className={`tab-button ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.label}
          </button>
        ))}
      </div>
      
      {activeTab === "data" && (
        <DataViewer
          tableName={tableName}
          columns={columns}
          rows={rows}
          currentPage={currentPage}
          totalPages={Math.max(1, Math.ceil(totalRows / 50))}
          totalRows={totalRows}
          onPageChange={setCurrentPage}
          onRefresh={() => loadData(currentPage)}
          onOpenQuery={() => {
            setQueryDraft(
              connectionType === "mongodb"
                ? JSON.stringify({ find: tableName, filter: {}, limit: 100 }, null, 2)
                : `SELECT * FROM ${qualifiedName} LIMIT 100;`
            );
            setActiveTab("query");
          }}
          onOpenVisual={() => setActiveTab("visual")}
          onExplain={() => {
            setQueryDraft(
              connectionType === "mongodb"
                ? JSON.stringify({ find: tableName, filter: {}, limit: 100 }, null, 2)
                : `SELECT * FROM ${qualifiedName} LIMIT 100;`
            );
            setActiveTab("query");
            setExplainToken((prev) => prev + 1);
          }}
        />
      )}
      
      {activeTab === "query" && (
        <QueryEditor
          query={queryDraft}
          onQueryChange={setQueryDraft}
          onExecute={handleExecuteQuery}
          onExplain={handleExplain}
          runToken={runToken}
          explainToken={explainToken}
        />
      )}

      {activeTab === "visual" && (
        <VisualQueryBuilder
          dialect={connectionType === "mongodb" ? "mongodb" : "sql"}
          onInsertQuery={(queryText) => {
            setQueryDraft(queryText);
            setActiveTab("query");
          }}
          onRunQuery={(queryText) => {
            setQueryDraft(queryText);
            setActiveTab("query");
            setRunToken((prev) => prev + 1);
          }}
          onExplainQuery={(queryText) => {
            setQueryDraft(queryText);
            setActiveTab("query");
            setExplainToken((prev) => prev + 1);
          }}
        />
      )}
      
      {activeTab === "structure" && (
        <StructureView tableName={tableName} columns={structure} />
      )}
    </main>
  );
}
