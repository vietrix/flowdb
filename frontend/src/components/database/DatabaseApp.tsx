"use client";

import { useEffect, useMemo, useState } from "react";
import { Sidebar } from "./Sidebar";
import { Workspace } from "./Workspace";
import { BottomPanel } from "./BottomPanel";
import { listEntities, listNamespaces } from "@/lib/api";
import { useAppContext } from "@/lib/app-context";
import type { TreeNode } from "@/lib/types";

export function DatabaseApp() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [bottomPanelCollapsed, setBottomPanelCollapsed] = useState(false);
  const [treeData, setTreeData] = useState<TreeNode[]>([]);
  const [selectedItem, setSelectedItem] = useState<TreeNode | null>(null);
  const [logs, setLogs] = useState<
    { id: string; timestamp: string; type: "query" | "error" | "info"; message: string; duration?: number }[]
  >([]);

  const { connections, selectedConnectionId } = useAppContext();

  const selectedConnection = useMemo(
    () => connections.find((c) => c.id === selectedConnectionId) || null,
    [connections, selectedConnectionId]
  );

  const handleSelectItem = (id: string, type: string) => {
    const node = findNode(treeData, id);
    if (node) {
      setSelectedItem(node);
    }
  };

  const addLog = (entry: { type: "query" | "error" | "info"; message: string; duration?: number }) => {
    setLogs((prev) => [
      {
        id: `${Date.now()}`,
        timestamp: new Date().toLocaleTimeString("en-US", { hour12: false }).slice(0, 8),
        ...entry,
      },
      ...prev,
    ]);
  };

  const handleClearLogs = () => {
    setLogs([]);
  };

  useEffect(() => {
    const loadTree = async () => {
      if (!selectedConnectionId) {
        setTreeData([]);
        return;
      }
      const namespaces = await listNamespaces(selectedConnectionId);
      const namespaceNodes: TreeNode[] = [];
      for (const ns of namespaces) {
        const entities = await listEntities(selectedConnectionId, ns.name);
        const children = entities.map<TreeNode>((entity) => ({
          id: `${selectedConnectionId}:${ns.name}:${entity.name}`,
          name: entity.name,
          type: "table",
          namespace: ns.name,
          connectionId: selectedConnectionId,
        }));
        namespaceNodes.push({
          id: `${selectedConnectionId}:${ns.name}`,
          name: ns.name,
          type: "namespace",
          namespace: ns.name,
          connectionId: selectedConnectionId,
          children,
        });
      }
      setTreeData([
        {
          id: selectedConnectionId,
          name: selectedConnection?.name || "connection",
          type: "connection",
          connectionId: selectedConnectionId,
          children: namespaceNodes,
        },
      ]);
      if (!selectedItem) {
        const firstTable = namespaceNodes.flatMap((n) => n.children || [])[0];
        if (firstTable) {
          setSelectedItem(firstTable);
        }
      }
    };
    loadTree().catch(() => addLog({ type: "error", message: "Không thể tải schema." }));
  }, [selectedConnectionId, selectedConnection, selectedItem]);

  return (
    <div className="flex-1 flex overflow-hidden">
      <div className="flex-1 flex overflow-hidden">
        <Sidebar
          data={treeData}
          selectedItem={selectedItem?.id || null}
          onSelectItem={handleSelectItem}
          collapsed={sidebarCollapsed}
          onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
        />
        
        <div className="flex-1 flex flex-col overflow-hidden">
          <Workspace
            connectionId={selectedConnectionId}
            connectionType={selectedConnection?.type || null}
            selectedNamespace={selectedItem?.namespace || null}
            selectedEntity={selectedItem?.type === "table" || selectedItem?.type === "view" ? selectedItem.name : null}
            onLog={addLog}
          />
          
          <BottomPanel
            logs={logs}
            collapsed={bottomPanelCollapsed}
            onToggleCollapse={() => setBottomPanelCollapsed(!bottomPanelCollapsed)}
            onClearLogs={handleClearLogs}
          />
        </div>
      </div>
    </div>
  );
}

function findNode(tree: TreeNode[], id: string): TreeNode | null {
  for (const node of tree) {
    if (node.id === id) return node;
    if (node.children) {
      const found = findNode(node.children, id);
      if (found) return found;
    }
  }
  return null;
}
