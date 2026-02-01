"use client";

import { useMemo, useState } from "react";
import { AppShell } from "@/components/database/AppShell";
import { useAppContext } from "@/lib/app-context";
import { createConnection, deleteConnection, testConnection, updateConnection } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";

type FormState = {
  id?: string;
  name: string;
  type: string;
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
};

const emptyForm: FormState = {
  name: "",
  type: "postgres",
  host: "",
  port: 5432,
  database: "",
  username: "",
  password: "",
};

export default function ConnectionsPage() {
  const { connections, refreshConnections, setSelectedConnectionId } = useAppContext();
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [error, setError] = useState<string | null>(null);

  const isEdit = useMemo(() => Boolean(form.id), [form.id]);

  const handleSave = async () => {
    setError(null);
    try {
      if (isEdit && form.id) {
        await updateConnection(form.id, {
          name: form.name,
          type: form.type,
          host: form.host,
          port: form.port,
          database: form.database,
          username: form.username,
          password: form.password || undefined,
        });
      } else {
        await createConnection({
          name: form.name,
          type: form.type,
          host: form.host,
          port: form.port,
          database: form.database,
          username: form.username,
          password: form.password,
        });
      }
      await refreshConnections();
      setOpen(false);
      setForm(emptyForm);
    } catch (err) {
      setError((err as Error).message || "Lưu kết nối thất bại.");
    }
  };

  const handleEdit = (id: string) => {
    const conn = connections.find((c) => c.id === id);
    if (!conn) return;
    setForm({
      id: conn.id,
      name: conn.name,
      type: conn.type,
      host: conn.host,
      port: conn.port,
      database: conn.database,
      username: conn.username,
      password: "",
    });
    setOpen(true);
  };

  const handleDelete = async (id: string) => {
    await deleteConnection(id);
    await refreshConnections();
  };

  const handleTest = async (id: string) => {
    await testConnection(id);
  };

  return (
    <AppShell>
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-3 py-2 border-b bg-secondary/30">
          <div className="text-sm font-medium">Connections</div>
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
              <button className="toolbar-button">New Connection</button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>{isEdit ? "Edit Connection" : "New Connection"}</DialogTitle>
              </DialogHeader>
              <div className="space-y-2">
                <Input
                  placeholder="Name"
                  value={form.name}
                  onChange={(e) => setForm((s) => ({ ...s, name: e.target.value }))}
                />
                <Input
                  placeholder="Type (postgres/mongodb)"
                  value={form.type}
                  onChange={(e) => setForm((s) => ({ ...s, type: e.target.value }))}
                />
                <Input
                  placeholder="Host"
                  value={form.host}
                  onChange={(e) => setForm((s) => ({ ...s, host: e.target.value }))}
                />
                <Input
                  placeholder="Port"
                  value={form.port}
                  onChange={(e) => setForm((s) => ({ ...s, port: Number(e.target.value) }))}
                />
                <Input
                  placeholder="Database"
                  value={form.database}
                  onChange={(e) => setForm((s) => ({ ...s, database: e.target.value }))}
                />
                <Input
                  placeholder="Username"
                  value={form.username}
                  onChange={(e) => setForm((s) => ({ ...s, username: e.target.value }))}
                />
                <Input
                  type="password"
                  placeholder="Password"
                  value={form.password}
                  onChange={(e) => setForm((s) => ({ ...s, password: e.target.value }))}
                />
                {error && <div className="text-sm text-destructive">{error}</div>}
              </div>
              <DialogFooter>
                <Button onClick={handleSave}>{isEdit ? "Update" : "Create"}</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
        <div className="flex-1 overflow-auto">
          <table className="data-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Host</th>
                <th>Database</th>
                <th>Owner</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {connections.map((conn) => (
                <tr key={conn.id}>
                  <td>{conn.name}</td>
                  <td>{conn.type}</td>
                  <td>{conn.host}:{conn.port}</td>
                  <td>{conn.database}</td>
                  <td>{String(conn.tags?.owner || "-")}</td>
                  <td className="space-x-1">
                    <button className="toolbar-button px-2" onClick={() => setSelectedConnectionId(conn.id)}>
                      Use
                    </button>
                    <button className="toolbar-button px-2" onClick={() => handleEdit(conn.id)}>
                      Edit
                    </button>
                    <button className="toolbar-button px-2" onClick={() => handleTest(conn.id)}>
                      Test
                    </button>
                    <button className="toolbar-button px-2" onClick={() => handleDelete(conn.id)}>
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
              {connections.length === 0 && (
                <tr>
                  <td colSpan={6} className="text-center text-muted-foreground">
                    No connections
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </AppShell>
  );
}
