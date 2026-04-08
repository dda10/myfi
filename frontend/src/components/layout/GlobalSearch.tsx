"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Search, X } from "lucide-react";

/**
 * Global Search (⌘K) — fuzzy search across VN stock symbols.
 * Requirements: 37.1, 37.2, 37.3, 37.4, 37.5, 37.6, 37.7, 37.8
 */

interface SearchResult {
  symbol: string;
  name: string;
  exchange: string;
  price?: number;
  change_percent?: number;
}

interface GlobalSearchProps {
  externalOpen?: boolean;
  onExternalClose?: () => void;
}

export function GlobalSearch({ externalOpen, onExternalClose }: GlobalSearchProps = {}) {
  const [open, setOpen] = useState(false);

  // Sync with external open trigger (mobile search icon)
  useEffect(() => {
    if (externalOpen) setOpen(true);
  }, [externalOpen]);

  const handleClose = () => {
    setOpen(false);
    onExternalClose?.();
  };
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [loading, setLoading] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();

  // ⌘K / Ctrl+K shortcut
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setOpen(prev => !prev);
      }
      if (e.key === "Escape") handleClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  // Focus input when opened
  useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 50);
      setQuery("");
      setResults([]);
      setSelectedIndex(0);
    }
  }, [open]);

  // Search on query change (debounced)
  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      return;
    }

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await fetch(
          `/api/market/search?q=${encodeURIComponent(query)}&limit=8`,
          { signal: AbortSignal.timeout(2000) },
        );
        if (res.ok) {
          const data = await res.json();
          setResults(data.results || []);
        }
      } catch {
        // Ignore timeout/abort
      } finally {
        setLoading(false);
      }
    }, 150);

    return () => clearTimeout(timer);
  }, [query]);

  const handleSelect = useCallback((symbol: string) => {
    handleClose();
    router.push(`/stock/${symbol}`);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [router]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex(i => Math.min(i + 1, results.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex(i => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && results[selectedIndex]) {
      handleSelect(results[selectedIndex].symbol);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50" onClick={handleClose} />

      {/* Search modal */}
      <div className="relative w-full max-w-lg mx-4 bg-surface border border-border-theme rounded-xl shadow-2xl overflow-hidden">
        <div className="flex items-center gap-3 px-4 py-3 border-b border-border-theme">
          <Search size={18} className="text-text-muted" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={e => { setQuery(e.target.value); setSelectedIndex(0); }}
            onKeyDown={handleKeyDown}
            placeholder="Search stocks... (symbol or company name)"
            className="flex-1 bg-transparent text-foreground placeholder:text-text-muted outline-none"
            aria-label="Search stocks"
          />
          <button onClick={handleClose} className="text-text-muted hover:text-foreground" aria-label="Close search">
            <X size={18} />
          </button>
        </div>

        {results.length > 0 && (
          <ul className="max-h-80 overflow-y-auto py-2" role="listbox">
            {results.map((r, i) => (
              <li
                key={r.symbol}
                role="option"
                aria-selected={i === selectedIndex}
                className={`flex items-center justify-between px-4 py-2.5 cursor-pointer transition-colors ${
                  i === selectedIndex ? "bg-surface-hover" : "hover:bg-surface"
                }`}
                onClick={() => handleSelect(r.symbol)}
              >
                <div>
                  <span className="font-medium text-foreground">{r.symbol}</span>
                  <span className="ml-2 text-sm text-text-muted">{r.name}</span>
                  <span className="ml-2 text-xs text-text-muted">{r.exchange}</span>
                </div>
                {r.price != null && (
                  <div className="text-right">
                    <span className="text-sm text-foreground">{r.price.toLocaleString()}</span>
                    {r.change_percent != null && (
                      <span className={`ml-2 text-xs ${r.change_percent >= 0 ? "text-green-400" : "text-red-400"}`}>
                        {r.change_percent >= 0 ? "+" : ""}{r.change_percent.toFixed(2)}%
                      </span>
                    )}
                  </div>
                )}
              </li>
            ))}
          </ul>
        )}

        {query && !loading && results.length === 0 && (
          <div className="px-4 py-6 text-center text-text-muted text-sm">
            No results for &ldquo;{query}&rdquo;
          </div>
        )}

        <div className="px-4 py-2 border-t border-border-theme text-xs text-text-muted flex gap-4">
          <span>↑↓ Navigate</span>
          <span>↵ Select</span>
          <span>Esc Close</span>
        </div>
      </div>
    </div>
  );
}
