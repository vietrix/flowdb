import { ChevronDown, User, Sun, Moon, LogOut, Settings, Database, ClipboardList, ScrollText, ShieldCheck } from "lucide-react";
import { useTheme } from "@/hooks/useTheme";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { logout } from "@/lib/api";

interface TopBarProps {
  selectedDatabase: string;
  onDatabaseChange: (db: string) => void;
  databases: string[];
  isConnected: boolean;
  userName: string;
}

function Logo() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <ellipse cx="12" cy="6" rx="8" ry="3" stroke="currentColor" strokeWidth="2" />
      <path d="M4 6v6c0 1.657 3.582 3 8 3s8-1.343 8-3V6" stroke="currentColor" strokeWidth="2" />
      <path d="M4 12v6c0 1.657 3.582 3 8 3s8-1.343 8-3v-6" stroke="currentColor" strokeWidth="2" />
    </svg>
  );
}

export function TopBar({ selectedDatabase, onDatabaseChange, databases, isConnected, userName }: TopBarProps) {
  const { theme, toggleTheme } = useTheme();
  const router = useRouter();

  const handleLogout = async () => {
    try {
      await logout();
    } finally {
      router.replace("/login");
    }
  };

  return (
    <header className="app-topbar">
      <div className="flex items-center gap-1.5 font-semibold text-sm">
        <Logo />
        <span>DBAdmin</span>
      </div>
      
      <div className="flex items-center gap-2 ml-4">
        <span className="text-muted-foreground text-sm">Database:</span>
        <select
          className="dropdown-select"
          value={selectedDatabase}
          onChange={(e) => onDatabaseChange(e.target.value)}
          disabled={databases.length === 0}
        >
          {databases.map((db) => (
            <option key={db} value={db}>{db}</option>
          ))}
        </select>
      </div>
      
      <div className="flex-1" />
      
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 text-sm">
          <span className={`status-indicator ${isConnected ? 'online' : 'offline'}`} />
          <span className="text-muted-foreground">
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
        
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="toolbar-button">
              <User size={14} />
              <span>{userName}</span>
              <ChevronDown size={12} />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <DropdownMenuItem asChild>
              <Link href="/connections">
                <Database size={14} className="mr-2" />
                Connections
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href="/approvals">
                <ShieldCheck size={14} className="mr-2" />
                Approvals
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href="/history">
                <ClipboardList size={14} className="mr-2" />
                History
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href="/audit">
                <ScrollText size={14} className="mr-2" />
                Audit
              </Link>
            </DropdownMenuItem>
            <DropdownMenuItem onClick={toggleTheme}>
              {theme === "light" ? (
                <>
                  <Moon size={14} className="mr-2" />
                  Dark Mode
                </>
              ) : (
                <>
                  <Sun size={14} className="mr-2" />
                  Light Mode
                </>
              )}
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
              <Link href="/settings">
                <Settings size={14} className="mr-2" />
                Settings
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="text-destructive" onClick={handleLogout}>
              <LogOut size={14} className="mr-2" />
              Logout
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
