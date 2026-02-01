"use client";

import { AppProvider } from "@/lib/app-context";
import { AppLayout } from "./AppLayout";
import type { ReactNode } from "react";

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <AppProvider>
      <AppLayout>{children}</AppLayout>
    </AppProvider>
  );
}
