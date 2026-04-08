"use client";
import { useState, useEffect, useRef } from "react";
import { Search, ClipboardList, LayoutDashboard, ChevronDown } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useApp } from "@/context/AppContext";
import { NotificationBell } from "@/components/layout/NotificationBell";
import { ConnectionStatus } from "@/components/layout/ConnectionStatus";
import { GlobalSearch } from "@/components/layout/GlobalSearch";

interface TickerDetails {
  symbol: string;
  name: string;
  exchange: string;
}

export function Header() {
  const router = useRouter();
  const { setActiveSymbol } = useApp();
  const [globalSearchOpen, setGlobalSearchOpen] = useState(false);
  const [tickers, setTickers] = useState<TickerDetails[]>([]);
  const [searchInput, setSearchInput] = useState("");
  const [filteredTickers, setFilteredTickers] = useState<TickerDetails[]>([]);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const token = typeof window !== "undefined" ? localStorage.getItem("ezistock-token") : null;
    fetch("http://localhost:8080/api/market/listing", {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
      .then(r => r.json())
      .then(j => { if (j.data) setTickers(j.data); })
      .catch(() => {});
  }, []);

  useEffect(() => {
    const handleOutsideClick = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handleOutsideClick);
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setSearchInput(val);
    if (val.trim()) {
      setFilteredTickers(
        tickers.filter(t =>
          t.symbol.toLowerCase().includes(val.toLowerCase()) ||
          t.name.toLowerCase().includes(val.toLowerCase())
        ).slice(0, 8)
      );
      setIsDropdownOpen(true);
    } else {
      setIsDropdownOpen(false);
    }
  };

  const handleSelectTicker = (symbol: string) => {
    setActiveSymbol(symbol);
    router.push(`/stock/${symbol}`);
    setSearchInput("");
    setIsDropdownOpen(false);
  };

  return (
    <header className="h-12 border-b border-border-theme bg-header-bg backdrop-blur-md sticky top-0 z-20 flex items-center gap-3 px-4">
      {/* Left: spacer on mobile for hamburger */}
      <div className="w-10 md:hidden" />

      {/* Center-left: Nhiệm vụ + Biểu đồ buttons */}
      <div className="flex items-center gap-1">
        <Link
          href="/missions"
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium text-text-muted hover:text-foreground hover:bg-surface transition"
        >
          <ClipboardList size={14} />
          <span className="hidden sm:inline">Nhiệm vụ</span>
        </Link>
        <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium text-text-muted hover:text-foreground hover:bg-surface transition">
          <LayoutDashboard size={14} />
          <span className="hidden sm:inline">Biểu đồ</span>
          <span className="badge-new hidden sm:inline-block">NEW</span>
        </button>
      </div>

      {/* Spacer */}
      <div className="flex-1" />

      {/* Right: notification + search */}
      <div className="flex items-center gap-2">
        <ConnectionStatus />
        <NotificationBell />

        {/* Mobile: just icon */}
        <button
          onClick={() => setGlobalSearchOpen(true)}
          className="md:hidden p-2 rounded-lg text-text-muted hover:text-foreground hover:bg-surface transition"
          aria-label="Search"
        >
          <Search size={16} />
        </button>

        {/* Desktop: inline search */}
        <div ref={dropdownRef} className="relative hidden md:block">
          <div className="flex items-center gap-2 bg-surface border border-border-theme rounded-lg px-3 py-1.5 cursor-text w-56 hover:border-accent/50 transition"
            onClick={() => setGlobalSearchOpen(true)}
          >
            <Search size={13} className="text-text-muted flex-shrink-0" />
            <span className="text-xs text-text-muted flex-1">Tìm kiếm cổ phiếu...</span>
            <kbd className="text-[10px] text-text-muted bg-badge-bg px-1.5 py-0.5 rounded font-mono">⌘K</kbd>
          </div>
        </div>
      </div>

      <GlobalSearch externalOpen={globalSearchOpen} onExternalClose={() => setGlobalSearchOpen(false)} />
    </header>
  );
}
