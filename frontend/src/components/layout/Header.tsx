"use client";
import { useState, useEffect, useRef } from "react";
import { Search, Bell, Settings, Star, Sun, Moon } from "lucide-react";
import { useApp } from "@/context/AppContext";
import { useWatchlist } from "@/context/WatchlistContext";
import { useTheme } from "@/context/ThemeContext";
import { LanguageSelector } from "@/components/layout/LanguageSelector";
import { CurrencyToggle } from "@/components/layout/CurrencyToggle";

interface TickerDetails {
  symbol: string;
  name: string;
  exchange: string;
}

export function Header() {
  const { setActiveSymbol, setActiveTab } = useApp();
  const { isWatched, toggleWatchlist } = useWatchlist();
  const { theme, toggleTheme } = useTheme();
  const [searchInput, setSearchInput] = useState("");
  const [tickers, setTickers] = useState<TickerDetails[]>([]);
  const [filteredTickers, setFilteredTickers] = useState<TickerDetails[]>([]);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetch("http://localhost:8080/api/market/listing", {
      headers: typeof window !== "undefined" && localStorage.getItem("myfi-token")
        ? { Authorization: `Bearer ${localStorage.getItem("myfi-token")}` }
        : {},
    })
      .then(res => res.json())
      .then(json => {
        if (json.data) setTickers(json.data);
      })
      .catch(err => console.error("Failed fetching listing:", err));
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setSearchInput(val);
    if (val.trim()) {
      const filtered = tickers.filter(t => 
        t.symbol.toLowerCase().includes(val.toLowerCase()) || 
        t.name.toLowerCase().includes(val.toLowerCase())
      );
      setFilteredTickers(filtered);
      setIsDropdownOpen(true);
    } else {
      setIsDropdownOpen(false);
    }
  };

  const handleSelectTicker = (symbol: string) => {
    setActiveSymbol(symbol);
    setSearchInput("");
    setIsDropdownOpen(false);
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchInput.trim()) {
      setActiveSymbol(searchInput.trim().toUpperCase());
      setSearchInput("");
      setIsDropdownOpen(false);
    }
  };

  return (
    <header className="h-16 border-b border-border-theme bg-header-bg backdrop-blur-md sticky top-0 z-10 flex items-center justify-between px-8">
      <div ref={dropdownRef} className="relative w-96">
        <form onSubmit={handleSearch} className="flex items-center w-full relative">
          <Search className="absolute left-3 text-text-muted" size={18} />
          <input 
            type="text" 
            value={searchInput}
            onChange={handleInputChange}
            onFocus={() => { if (searchInput.trim() && filteredTickers.length > 0) setIsDropdownOpen(true) }}
            placeholder="Search for symbols (e.g., FPT, VNM)..." 
            className="w-full bg-input-bg border border-border-theme rounded-full py-2 pl-10 pr-4 text-sm text-foreground placeholder-text-muted focus:outline-none focus:border-accent transition"
          />
        </form>

        {/* Autocomplete Dropdown */}
        {isDropdownOpen && filteredTickers.length > 0 && (
          <div className="absolute top-12 left-0 w-full bg-card-bg border border-border-theme rounded-xl shadow-2xl py-2 z-50 max-h-80 overflow-y-auto custom-scrollbar">
            {filteredTickers.map(ticker => (
              <div
                key={ticker.symbol}
                className="px-4 py-3 hover:bg-surface-hover cursor-pointer flex justify-between items-center transition group"
              >
                <div className="flex flex-col flex-1" onClick={() => handleSelectTicker(ticker.symbol)}>
                  <span className="font-bold text-foreground text-sm">{ticker.symbol}</span>
                  <span className="text-xs text-text-muted truncate max-w-[200px]">{ticker.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-[10px] font-medium bg-badge-bg text-badge-text px-2 py-1 rounded">
                    {ticker.exchange}
                  </span>
                  <button
                    onClick={e => { e.stopPropagation(); toggleWatchlist(ticker.symbol); }}
                    className="p-1 rounded transition"
                    title={isWatched(ticker.symbol) ? "Remove from watchlist" : "Add to watchlist"}
                  >
                    <Star
                      size={14}
                      className={isWatched(ticker.symbol)
                        ? "fill-yellow-400 text-yellow-400"
                        : "text-text-muted hover:text-yellow-400 group-hover:text-text-muted"
                      }
                    />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="flex items-center gap-4">
        {/* Currency toggle */}
        <CurrencyToggle />
        {/* Language selector */}
        <LanguageSelector />
        {/* Theme toggle button */}
        <button
          onClick={toggleTheme}
          className="p-2 text-text-muted hover:text-foreground transition rounded-full hover:bg-surface"
          title={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
          aria-label={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
        >
          {theme === "dark" ? <Sun size={20} /> : <Moon size={20} />}
        </button>
        <button className="p-2 text-text-muted hover:text-foreground transition rounded-full hover:bg-surface">
          <Bell size={20} />
        </button>
        <button 
          onClick={() => setActiveTab("Settings")}
          className="p-2 text-text-muted hover:text-foreground transition rounded-full hover:bg-surface"
          title="AI Configuration"
        >
          <Settings size={20} />
        </button>
        <div className="flex items-center gap-2 pl-4 border-l border-border-theme cursor-pointer group">
          <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-indigo-500 to-purple-500 flex items-center justify-center text-white font-medium shadow-lg group-hover:shadow-indigo-500/20 transition">
            A
          </div>
          <span className="text-sm font-medium text-text-muted group-hover:text-foreground transition">Admin</span>
        </div>
      </div>
    </header>
  );
}
