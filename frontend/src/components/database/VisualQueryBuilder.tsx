import { useState, useCallback, useRef, useEffect, useMemo } from "react";
import { Plus, Play, Table, Filter, Columns, SortAsc, Combine, X, Copy, FlaskConical, Sparkles } from "lucide-react";

interface QueryBlock {
  id: string;
  type: "table" | "filter" | "select" | "sort" | "join";
  position: { x: number; y: number };
  data: Record<string, string>;
}

interface Connection {
  id: string;
  from: string;
  to: string;
}

interface VisualQueryBuilderProps {
  dialect?: "sql" | "mongodb";
  onInsertQuery?: (query: string) => void;
  onRunQuery?: (query: string) => void;
  onExplainQuery?: (query: string) => void;
}

const blockTypes = [
  { type: "table", label: "Table", icon: Table, color: "bg-blue-500" },
  { type: "filter", label: "Filter", icon: Filter, color: "bg-amber-500" },
  { type: "select", label: "Select", icon: Columns, color: "bg-green-500" },
  { type: "sort", label: "Sort", icon: SortAsc, color: "bg-purple-500" },
  { type: "join", label: "Join", icon: Combine, color: "bg-rose-500" },
] as const;

interface BlockNodeProps {
  block: QueryBlock;
  isSelected: boolean;
  onSelect: (id: string) => void;
  onDrag: (id: string, position: { x: number; y: number }) => void;
  onDelete: (id: string) => void;
  onStartConnect: (id: string) => void;
  onEndConnect: (id: string) => void;
  onUpdateData: (id: string, data: Record<string, string>) => void;
}

function BlockNode({ 
  block, 
  isSelected, 
  onSelect, 
  onDrag, 
  onDelete,
  onStartConnect,
  onEndConnect,
  onUpdateData
}: BlockNodeProps) {
  const blockType = blockTypes.find(b => b.type === block.type);
  const Icon = blockType?.icon || Table;
  const [isDragging, setIsDragging] = useState(false);
  const dragOffset = useRef({ x: 0, y: 0 });

  const handleMouseDown = (e: React.MouseEvent) => {
    if ((e.target as HTMLElement).closest('.block-input, .block-connector')) return;
    e.preventDefault();
    setIsDragging(true);
    dragOffset.current = {
      x: e.clientX - block.position.x,
      y: e.clientY - block.position.y
    };
    onSelect(block.id);
  };

  useEffect(() => {
    if (!isDragging) return;
    
    const handleMouseMove = (e: MouseEvent) => {
      onDrag(block.id, {
        x: e.clientX - dragOffset.current.x,
        y: e.clientY - dragOffset.current.y
      });
    };

    const handleMouseUp = () => setIsDragging(false);

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging, block.id, onDrag]);

  const renderBlockContent = () => {
    switch (block.type) {
      case "table":
        return (
          <input
            className="block-input w-full px-2 py-1 text-xs border rounded bg-background"
            placeholder="Table name..."
            value={block.data.tableName || ""}
            onChange={(e) => onUpdateData(block.id, { ...block.data, tableName: e.target.value })}
          />
        );
      case "filter":
        return (
          <input
            className="block-input w-full px-2 py-1 text-xs border rounded bg-background"
            placeholder="WHERE condition..."
            value={block.data.condition || ""}
            onChange={(e) => onUpdateData(block.id, { ...block.data, condition: e.target.value })}
          />
        );
      case "select":
        return (
          <input
            className="block-input w-full px-2 py-1 text-xs border rounded bg-background"
            placeholder="Columns (*, col1, col2...)"
            value={block.data.columns || ""}
            onChange={(e) => onUpdateData(block.id, { ...block.data, columns: e.target.value })}
          />
        );
      case "sort":
        return (
          <input
            className="block-input w-full px-2 py-1 text-xs border rounded bg-background"
            placeholder="ORDER BY column..."
            value={block.data.orderBy || ""}
            onChange={(e) => onUpdateData(block.id, { ...block.data, orderBy: e.target.value })}
          />
        );
      case "join":
        return (
          <div className="space-y-1">
            <input
              className="block-input w-full px-2 py-1 text-xs border rounded bg-background"
              placeholder="Join table..."
              value={block.data.joinTable || ""}
              onChange={(e) => onUpdateData(block.id, { ...block.data, joinTable: e.target.value })}
            />
            <input
              className="block-input w-full px-2 py-1 text-xs border rounded bg-background"
              placeholder="ON condition..."
              value={block.data.onCondition || ""}
              onChange={(e) => onUpdateData(block.id, { ...block.data, onCondition: e.target.value })}
            />
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div
      className={`absolute flex flex-col rounded-lg shadow-lg border-2 transition-shadow ${
        isSelected ? "border-primary shadow-xl" : "border-border"
      } bg-card`}
      style={{ 
        left: block.position.x, 
        top: block.position.y,
        minWidth: 180,
        cursor: isDragging ? 'grabbing' : 'grab'
      }}
      onMouseDown={handleMouseDown}
    >
      {/* Header */}
      <div className={`flex items-center gap-2 px-3 py-2 rounded-t-md text-white ${blockType?.color}`}>
        <Icon size={14} />
        <span className="text-xs font-medium flex-1">{blockType?.label}</span>
        <button 
          className="hover:bg-white/20 rounded p-0.5"
          onClick={() => onDelete(block.id)}
        >
          <X size={12} />
        </button>
      </div>
      
      {/* Content */}
      <div className="p-2">
        {renderBlockContent()}
      </div>

      {/* Connectors */}
      <div
        className="block-connector absolute left-1/2 -top-2 w-4 h-4 rounded-full bg-primary border-2 border-background cursor-pointer transform -translate-x-1/2 hover:scale-125 transition-transform"
        onMouseUp={() => onEndConnect(block.id)}
        title="Drop connection here"
      />
      <div
        className="block-connector absolute left-1/2 -bottom-2 w-4 h-4 rounded-full bg-primary border-2 border-background cursor-pointer transform -translate-x-1/2 hover:scale-125 transition-transform"
        onMouseDown={(e) => {
          e.stopPropagation();
          onStartConnect(block.id);
        }}
        title="Drag to connect"
      />
    </div>
  );
}

export function VisualQueryBuilder({
  dialect = "sql",
  onInsertQuery,
  onRunQuery,
  onExplainQuery,
}: VisualQueryBuilderProps) {
  const [blocks, setBlocks] = useState<QueryBlock[]>([
    { id: "1", type: "table", position: { x: 100, y: 100 }, data: { tableName: "users" } },
    { id: "2", type: "select", position: { x: 100, y: 250 }, data: { columns: "*" } },
  ]);
  const [connections, setConnections] = useState<Connection[]>([
    { id: "conn-1", from: "1", to: "2" }
  ]);
  const [selectedBlock, setSelectedBlock] = useState<string | null>(null);
  const [connectingFrom, setConnectingFrom] = useState<string | null>(null);
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 });
  const canvasRef = useRef<HTMLDivElement>(null);

  const addBlock = (type: QueryBlock["type"]) => {
    const newBlock: QueryBlock = {
      id: Date.now().toString(),
      type,
      position: { x: 200 + Math.random() * 100, y: 150 + Math.random() * 100 },
      data: {}
    };
    setBlocks(prev => [...prev, newBlock]);
    setSelectedBlock(newBlock.id);
  };

  const updateBlockPosition = useCallback((id: string, position: { x: number; y: number }) => {
    setBlocks(prev => prev.map(b => b.id === id ? { ...b, position } : b));
  }, []);

  const updateBlockData = useCallback((id: string, data: Record<string, string>) => {
    setBlocks(prev => prev.map(b => b.id === id ? { ...b, data } : b));
  }, []);

  const deleteBlock = (id: string) => {
    setBlocks(prev => prev.filter(b => b.id !== id));
    setConnections(prev => prev.filter(c => c.from !== id && c.to !== id));
    if (selectedBlock === id) setSelectedBlock(null);
  };

  const startConnect = (id: string) => {
    setConnectingFrom(id);
  };

  const endConnect = (id: string) => {
    if (connectingFrom && connectingFrom !== id) {
      const exists = connections.some(c => c.from === connectingFrom && c.to === id);
      if (!exists) {
        setConnections(prev => [...prev, { id: `conn-${Date.now()}`, from: connectingFrom, to: id }]);
      }
    }
    setConnectingFrom(null);
  };

  const handleCanvasMouseMove = (e: React.MouseEvent) => {
    if (connectingFrom && canvasRef.current) {
      const rect = canvasRef.current.getBoundingClientRect();
      setMousePos({ x: e.clientX - rect.left, y: e.clientY - rect.top });
    }
  };

  const handleCanvasMouseUp = () => {
    setConnectingFrom(null);
  };

  const generateSQL = useCallback(() => {
    const parts: string[] = [];
    const tableBlocks = blocks.filter(b => b.type === "table");
    tableBlocks.forEach(table => {
      if (table.data.tableName) {
        parts.push(`FROM ${table.data.tableName}`);
      }
    });
    const selectBlocks = blocks.filter(b => b.type === "select");
    if (selectBlocks.length > 0 && selectBlocks[0].data.columns) {
      parts.unshift(`SELECT ${selectBlocks[0].data.columns}`);
    } else {
      parts.unshift("SELECT *");
    }
    const filterBlocks = blocks.filter(b => b.type === "filter");
    filterBlocks.forEach(filter => {
      if (filter.data.condition) {
        parts.push(`WHERE ${filter.data.condition}`);
      }
    });
    const joinBlocks = blocks.filter(b => b.type === "join");
    joinBlocks.forEach(join => {
      if (join.data.joinTable && join.data.onCondition) {
        parts.push(`JOIN ${join.data.joinTable} ON ${join.data.onCondition}`);
      }
    });
    const sortBlocks = blocks.filter(b => b.type === "sort");
    sortBlocks.forEach(sort => {
      if (sort.data.orderBy) {
        parts.push(`ORDER BY ${sort.data.orderBy}`);
      }
    });
    return parts.join("\n");
  }, [blocks]);

  const generateMongo = useCallback(() => {
    const table = blocks.find(b => b.type === "table")?.data.tableName || "collection";
    const filter = blocks.find(b => b.type === "filter")?.data.condition || "";
    const select = blocks.find(b => b.type === "select")?.data.columns || "";
    const sort = blocks.find(b => b.type === "sort")?.data.orderBy || "";
    const doc: Record<string, unknown> = {
      find: table,
      filter: filter ? { $where: filter } : {},
    };
    if (select) {
      doc.projection = select.split(",").reduce<Record<string, number>>((acc, key) => {
        const k = key.trim();
        if (k) acc[k] = 1;
        return acc;
      }, {});
    }
    if (sort) {
      doc.sort = { [sort]: 1 };
    }
    return JSON.stringify(doc, null, 2);
  }, [blocks]);

  const queryPreview = useMemo(() => {
    return dialect === "mongodb" ? generateMongo() : generateSQL();
  }, [dialect, generateMongo, generateSQL]);

  const getBlockCenter = (block: QueryBlock, isBottom: boolean) => {
    return {
      x: block.position.x + 90, // half of minWidth
      y: block.position.y + (isBottom ? 80 : 0) // approximate height
    };
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex flex-wrap items-center gap-2 p-2 border-b bg-secondary/30">
        <span className="text-xs text-muted-foreground mr-2">Blocks:</span>
        {blockTypes.map(({ type, label, icon: Icon, color }) => (
          <button
            key={type}
            className={`toolbar-button text-xs ${color} text-white hover:opacity-90`}
            onClick={() => addBlock(type)}
          >
            <Icon size={12} />
            {label}
          </button>
        ))}
        <div className="flex-1" />
        <div className="flex flex-wrap items-center gap-2">
          <button
            className="toolbar-button"
            onClick={() => navigator.clipboard.writeText(queryPreview)}
          >
            <Copy size={12} />
            Copy
          </button>
          <button
            className="toolbar-button bg-primary text-primary-foreground"
            onClick={() => onInsertQuery?.(queryPreview)}
            disabled={!onInsertQuery}
          >
            <Sparkles size={12} />
            Insert
          </button>
          <button
            className="toolbar-button"
            onClick={() => onExplainQuery?.(queryPreview)}
            disabled={!onExplainQuery}
          >
            <FlaskConical size={12} />
            Explain
          </button>
          <button
            className="toolbar-button bg-primary text-primary-foreground"
            onClick={() => onRunQuery?.(queryPreview)}
            disabled={!onRunQuery}
          >
            <Play size={12} />
            Run
          </button>
        </div>
      </div>

      <div className="flex flex-col lg:flex-row flex-1 overflow-hidden">
        <div className="flex-1 relative overflow-hidden border-r">
          <div
            ref={canvasRef}
            className="absolute inset-0 overflow-auto"
            style={{ background: "hsl(var(--workspace-bg))" }}
            onMouseMove={handleCanvasMouseMove}
            onMouseUp={handleCanvasMouseUp}
          >
            <svg className="absolute inset-0 w-full h-full pointer-events-none opacity-30">
              <defs>
                <pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse">
                  <path d="M 20 0 L 0 0 0 20" fill="none" stroke="currentColor" strokeWidth="0.5" className="text-muted-foreground" />
                </pattern>
              </defs>
              <rect width="100%" height="100%" fill="url(#grid)" />
            </svg>

            <svg className="absolute inset-0 w-full h-full pointer-events-none">
              <defs>
                <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                  <polygon points="0 0, 10 3.5, 0 7" className="fill-primary" />
                </marker>
              </defs>
              {connections.map(conn => {
                const fromBlock = blocks.find(b => b.id === conn.from);
                const toBlock = blocks.find(b => b.id === conn.to);
                if (!fromBlock || !toBlock) return null;
                const from = getBlockCenter(fromBlock, true);
                const to = getBlockCenter(toBlock, false);
                const midY = (from.y + to.y) / 2;
                const path = `M ${from.x} ${from.y} C ${from.x} ${midY}, ${to.x} ${midY}, ${to.x} ${to.y - 8}`;
                return (
                  <path
                    key={conn.id}
                    d={path}
                    fill="none"
                    className="stroke-primary"
                    strokeWidth="2"
                    markerEnd="url(#arrowhead)"
                  />
                );
              })}
              {connectingFrom && (() => {
                const fromBlock = blocks.find(b => b.id === connectingFrom);
                if (!fromBlock) return null;
                const from = getBlockCenter(fromBlock, true);
                return (
                  <line
                    x1={from.x}
                    y1={from.y}
                    x2={mousePos.x}
                    y2={mousePos.y}
                    className="stroke-primary"
                    strokeWidth="2"
                    strokeDasharray="5,5"
                  />
                );
              })()}
            </svg>

            {blocks.map(block => (
              <BlockNode
                key={block.id}
                block={block}
                isSelected={selectedBlock === block.id}
                onSelect={setSelectedBlock}
                onDrag={updateBlockPosition}
                onDelete={deleteBlock}
                onStartConnect={startConnect}
                onEndConnect={endConnect}
                onUpdateData={updateBlockData}
              />
            ))}

            {blocks.length === 0 && (
              <div className="absolute inset-0 flex items-center justify-center text-muted-foreground">
                <div className="text-center">
                  <Plus className="w-12 h-12 mx-auto mb-2 opacity-50" />
                  <p>Click a block type above to start building your query</p>
                </div>
              </div>
            )}
          </div>
        </div>

        <aside className="w-full lg:w-80 border-t lg:border-t-0 lg:border-l bg-card flex flex-col">
          <div className="px-3 py-2 border-b text-xs uppercase tracking-wide text-muted-foreground">
            Query Preview
          </div>
          <div className="flex-1 overflow-auto p-3">
            <pre className="text-xs whitespace-pre-wrap bg-secondary/30 border rounded p-3">
              {queryPreview}
            </pre>
          </div>
          <div className="px-3 py-2 border-t flex items-center gap-2 text-xs text-muted-foreground">
            Dialect: <span className="font-medium text-foreground">{dialect === "mongodb" ? "Mongo DSL" : "SQL"}</span>
          </div>
        </aside>
      </div>
    </div>
  );
}
