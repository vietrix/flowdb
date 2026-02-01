"use client";

import { useEffect, useState } from "react";
import { AppShell } from "@/components/database/AppShell";
import { getSettings, updateFlags, updateSecurityMode, updateSettings } from "@/lib/api";
import type { Settings } from "@/lib/types";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [allowlist, setAllowlist] = useState("");

  useEffect(() => {
    const load = async () => {
      try {
        const data = await getSettings();
        setSettings(data);
        setAllowlist((data.ipAllowlist || []).join(", "));
      } catch (err) {
        setError((err as Error).message || "Không tải được settings.");
      }
    };
    load();
  }, []);

  const handleModeChange = async (mode: string) => {
    if (!settings) return;
    setSaving(true);
    try {
      const updated = await updateSecurityMode(mode);
      setSettings(updated);
    } catch (err) {
      setError((err as Error).message || "Cập nhật chế độ thất bại.");
    } finally {
      setSaving(false);
    }
  };

  const handleFlagChange = async (key: string, value: boolean) => {
    if (!settings) return;
    const nextFlags = { ...settings.flags, [key]: value };
    setSettings({ ...settings, flags: nextFlags });
    try {
      await updateFlags({ [key]: value });
    } catch (err) {
      setError((err as Error).message || "Cập nhật flag thất bại.");
    }
  };

  const handleSaveAllowlist = async () => {
    if (!settings) return;
    setSaving(true);
    try {
      const next = await updateSettings({
        ...settings,
        ipAllowlist: allowlist.split(",").map((s) => s.trim()).filter(Boolean),
      });
      setSettings(next);
    } catch (err) {
      setError((err as Error).message || "Lưu allowlist thất bại.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <AppShell>
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-3 py-2 border-b bg-secondary/30">
          <div className="text-sm font-medium">Settings</div>
        </div>
        <div className="flex-1 overflow-auto p-4 space-y-6">
          {error && <div className="text-sm text-destructive">{error}</div>}
          {!settings ? (
            <div className="text-sm text-muted-foreground">Loading...</div>
          ) : (
            <>
              <div className="space-y-2">
                <div className="text-sm font-medium">Security Mode</div>
                <div className="flex items-center gap-2">
                  <select
                    className="dropdown-select"
                    value={settings.securityMode}
                    onChange={(e) => handleModeChange(e.target.value)}
                  >
                    <option value="basic">basic</option>
                    <option value="standard">standard</option>
                    <option value="enterprise">enterprise</option>
                  </select>
                  <span className="text-xs text-muted-foreground">Runtime toggle</span>
                </div>
              </div>

              <div className="space-y-2">
                <div className="text-sm font-medium">Feature Flags</div>
                <div className="grid gap-2">
                  {Object.keys(settings.flags || {}).map((key) => (
                    <div key={key} className="flex items-center justify-between border-b py-2">
                      <span className="text-sm">{key}</span>
                      <Switch
                        checked={Boolean(settings.flags[key])}
                        onCheckedChange={(value) => handleFlagChange(key, value)}
                      />
                    </div>
                  ))}
                </div>
              </div>

              <div className="space-y-2">
                <div className="text-sm font-medium">IP Allowlist</div>
                <Input
                  value={allowlist}
                  onChange={(e) => setAllowlist(e.target.value)}
                  placeholder="10.0.0.0/24, 192.168.1.10/32"
                />
                <Button onClick={handleSaveAllowlist} disabled={saving}>
                  Lưu allowlist
                </Button>
              </div>
            </>
          )}
        </div>
      </div>
    </AppShell>
  );
}
