"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { useRouter } from "next/navigation";
import { getMe, listConnections } from "@/lib/api";
import type { Connection, UserProfile } from "@/lib/types";

interface AppContextValue {
  user: UserProfile | null;
  connections: Connection[];
  selectedConnectionId: string | null;
  setSelectedConnectionId: (id: string | null) => void;
  refreshConnections: () => Promise<void>;
  loading: boolean;
}

const AppContext = createContext<AppContextValue | null>(null);

export function AppProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<UserProfile | null>(null);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [selectedConnectionId, setSelectedConnectionIdState] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const setSelectedConnectionId = (id: string | null) => {
    setSelectedConnectionIdState(id);
    if (typeof window !== "undefined") {
      if (id) {
        localStorage.setItem("flowdb_conn", id);
      } else {
        localStorage.removeItem("flowdb_conn");
      }
    }
  };

  const refreshConnections = useCallback(async () => {
    const list = await listConnections();
    setConnections(list);
    if (!selectedConnectionId) {
      const stored = typeof window !== "undefined" ? localStorage.getItem("flowdb_conn") : null;
      if (stored && list.some((c) => c.id === stored)) {
        setSelectedConnectionIdState(stored);
      } else if (list[0]) {
        setSelectedConnectionIdState(list[0].id);
      }
    }
  }, [selectedConnectionId]);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const me = await getMe();
        if (!mounted) return;
        setUser({
          id: me.id,
          username: me.username,
          isAdmin: me.isAdmin,
          mfaEnabled: me.mfaEnabled,
        });
        await refreshConnections();
      } catch {
        router.replace("/login");
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, [refreshConnections, router]);

  const value = useMemo(
    () => ({
      user,
      connections,
      selectedConnectionId,
      setSelectedConnectionId,
      refreshConnections,
      loading,
    }),
    [user, connections, selectedConnectionId, loading, refreshConnections]
  );

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}

export function useAppContext() {
  const ctx = useContext(AppContext);
  if (!ctx) {
    throw new Error("AppContext not found");
  }
  return ctx;
}
