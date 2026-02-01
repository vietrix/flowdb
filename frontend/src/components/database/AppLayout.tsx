"use client";

import { useMemo } from "react";
import type { ReactNode } from "react";
import { TopBar } from "./TopBar";
import { useAppContext } from "@/lib/app-context";

export function AppLayout({ children }: { children: ReactNode }) {
  const { connections, selectedConnectionId, setSelectedConnectionId, user, loading } =
    useAppContext();

  const databaseOptions = useMemo(
    () => connections.map((c) => ({ id: c.id, name: c.name })),
    [connections]
  );

  const selectedName =
    databaseOptions.find((c) => c.id === selectedConnectionId)?.name ||
    databaseOptions[0]?.name ||
    "default";

  const isConnected = !loading && connections.length > 0;

  const handleDatabaseChange = (name: string) => {
    const match = databaseOptions.find((c) => c.name === name);
    if (match) {
      setSelectedConnectionId(match.id);
    }
  };

  return (
    <div className="h-screen flex flex-col overflow-hidden">
      <TopBar
        selectedDatabase={selectedName}
        onDatabaseChange={handleDatabaseChange}
        databases={databaseOptions.map((c) => c.name)}
        isConnected={isConnected}
        userName={user?.username || "user"}
      />
      {children}
    </div>
  );
}
