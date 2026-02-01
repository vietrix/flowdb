import { useEffect, useState } from "react";
import { ChevronRight, ChevronDown, Database, Table, Columns, Key, PanelLeftClose, PanelLeft } from "lucide-react";
import type { TreeNode } from "@/lib/types";

interface SidebarProps {
  data: TreeNode[];
  selectedItem: string | null;
  onSelectItem: (id: string, type: string) => void;
  collapsed: boolean;
  onToggleCollapse: () => void;
}

function TreeItem({ 
  node, 
  level, 
  selectedItem, 
  onSelectItem,
  expandedNodes,
  onToggleExpand 
}: { 
  node: TreeNode; 
  level: number;
  selectedItem: string | null;
  onSelectItem: (id: string, type: string) => void;
  expandedNodes: Set<string>;
  onToggleExpand: (id: string) => void;
}) {
  const hasChildren = node.children && node.children.length > 0;
  const isExpanded = expandedNodes.has(node.id);
  const isSelected = selectedItem === node.id;
  
  const getIcon = () => {
    switch (node.type) {
      case "connection":
        return <Database size={14} className="text-muted-foreground" />;
      case "namespace":
        return <Columns size={14} className="text-muted-foreground" />;
      case "table":
        return <Table size={14} className="text-muted-foreground" />;
      case "view":
        return <Table size={14} className="text-muted-foreground" />;
      case "index":
        return <Key size={14} className="text-muted-foreground" />;
      default:
        return null;
    }
  };

  return (
    <div>
      <div
        className={`tree-item ${isSelected ? 'active' : ''}`}
        style={{ paddingLeft: `${level * 12 + 8}px` }}
        onClick={() => {
          if (hasChildren) {
            onToggleExpand(node.id);
          }
          onSelectItem(node.id, node.type);
        }}
      >
        {hasChildren ? (
          isExpanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />
        ) : (
          <span style={{ width: 12 }} />
        )}
        {getIcon()}
        <span className="truncate">{node.name}</span>
      </div>
      
      {hasChildren && isExpanded && (
        <div>
          {node.children!.map((child) => (
            <TreeItem
              key={child.id}
              node={child}
              level={level + 1}
              selectedItem={selectedItem}
              onSelectItem={onSelectItem}
              expandedNodes={expandedNodes}
              onToggleExpand={onToggleExpand}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function Sidebar({ data, selectedItem, onSelectItem, collapsed, onToggleCollapse }: SidebarProps) {
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (data.length > 0) {
      setExpandedNodes((prev) => {
        if (prev.size > 0) return prev;
        return new Set([data[0].id]);
      });
    }
  }, [data]);

  const handleToggleExpand = (id: string) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  if (collapsed) {
    return (
      <div className="border-r flex flex-col" style={{ background: "hsl(var(--sidebar-bg))" }}>
        <button
          className="p-2 hover:bg-accent"
          onClick={onToggleCollapse}
          title="Expand sidebar"
        >
          <PanelLeft size={16} />
        </button>
      </div>
    );
  }

  return (
    <aside className="app-sidebar">
      <div className="panel-header">
        <span>Navigator</span>
        <button
          className="hover:bg-accent p-0.5 rounded"
          onClick={onToggleCollapse}
          title="Collapse sidebar"
        >
          <PanelLeftClose size={14} />
        </button>
      </div>
      
      <div className="flex-1 overflow-auto py-1">
        {data.map((node) => (
          <TreeItem
            key={node.id}
            node={node}
            level={0}
            selectedItem={selectedItem}
            onSelectItem={onSelectItem}
            expandedNodes={expandedNodes}
            onToggleExpand={handleToggleExpand}
          />
        ))}
      </div>
    </aside>
  );
}
