import { DatabaseApp } from "@/components/database/DatabaseApp";
import { AppShell } from "@/components/database/AppShell";

export default function Page() {
  return (
    <AppShell>
      <DatabaseApp />
    </AppShell>
  );
}
