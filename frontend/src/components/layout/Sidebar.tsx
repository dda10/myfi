"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Home, LineChart, TrendingUp, Lightbulb, Filter,
  Wallet, Grid3X3, BarChart3, Globe, MessageSquare,
  FileText, Target, ChevronRight, Zap, Settings,
  Menu, X, User, Landmark,
} from "lucide-react";

interface NavItemConfig {
  icon: React.ComponentType<{ size?: number; className?: string }>;
  label: string;
  href: string;
  badge?: "hot" | "new";
  labelVi?: string;
}

const TOP_ITEMS: NavItemConfig[] = [
  { icon: Home,         label: "Platform",  labelVi: "Platform", href: "/dashboard" },
  { icon: MessageSquare,label: "Agent",     labelVi: "Agent",    href: "/chat" },
  { icon: FileText,     label: "Research",  labelVi: "Research", href: "/research" },
];

const PLATFORM_ITEMS: NavItemConfig[] = [
  { icon: LineChart,    label: "Cổ phiếu",          href: "/stock/VNM",  badge: "hot"  },
  { icon: Home,         label: "Thị trường",          href: "/dashboard"                },
  { icon: TrendingUp,   label: "AI Ranking",          href: "/ranking",    badge: "hot"  },
  { icon: Lightbulb,   label: "Ý tưởng đầu tư",      href: "/ideas"                    },
  { icon: Filter,       label: "Bộ lọc thông minh",   href: "/screener",   badge: "new"  },
  { icon: Wallet,       label: "Danh mục đầu tư",     href: "/portfolio"                },
  { icon: Grid3X3,      label: "Nhiệt đồ",             href: "/heatmap"                  },
  { icon: BarChart3,    label: "Analyst IQ",           href: "/analyst"                  },
  { icon: Globe,        label: "Kinh tế Vĩ mô",       href: "/macro"                    },
  { icon: Landmark,    label: "Quỹ đầu tư",           href: "/funds",    badge: "new"  },
];

const CHAT_HISTORY = [
  "Giải thích tại sao",
  "Phân tích VNM",
];

export function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const isActive = (href: string) => {
    if (href === "/dashboard") return pathname === "/dashboard";
    if (href.startsWith("/stock/")) return pathname.startsWith("/stock/");
    return pathname.startsWith(href);
  };

  const sidebarContent = (onNavClick?: () => void) => (
    <div className="flex flex-col h-full">
      {/* Logo + collapse toggle */}
      <div className="h-12 flex items-center justify-between px-4 border-b border-border-theme flex-shrink-0">
        {!collapsed && (
          <Link href="/dashboard" onClick={onNavClick} className="flex items-center gap-2">
            <div className="w-6 h-6 rounded bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center flex-shrink-0">
              <Zap size={12} className="text-white" />
            </div>
            <span className="text-sm font-bold gradient-text tracking-tight">MyFi</span>
          </Link>
        )}
        {collapsed && (
          <Link href="/dashboard" onClick={onNavClick} className="mx-auto">
            <div className="w-6 h-6 rounded bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center">
              <Zap size={12} className="text-white" />
            </div>
          </Link>
        )}
        <button
          onClick={() => setCollapsed(v => !v)}
          className="p-1 rounded hover:bg-surface transition text-text-muted hidden lg:block"
          aria-label="Toggle sidebar"
        >
          <Menu size={15} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto no-scrollbar">
        {/* Top-level nav */}
        <div className="px-2 pt-3 pb-1 space-y-0.5">
          {TOP_ITEMS.map(item => (
            <NavRow key={item.href} item={item} active={isActive(item.href)} collapsed={collapsed} onClick={onNavClick} />
          ))}
        </div>

        {/* Divider + Nền tảng label */}
        <div className="mx-3 my-2 border-t border-border-theme" />
        {!collapsed && (
          <p className="px-4 pb-1 text-[10px] font-semibold text-text-muted uppercase tracking-widest">
            Nền tảng
          </p>
        )}
        <div className="px-2 space-y-0.5">
          {PLATFORM_ITEMS.map(item => (
            <NavRow key={item.href} item={item} active={isActive(item.href)} collapsed={collapsed} onClick={onNavClick} />
          ))}
        </div>

        {/* Chat History */}
        {!collapsed && (
          <>
            <div className="mx-3 my-2 border-t border-border-theme" />
            <p className="px-4 pb-1 text-[10px] font-semibold text-text-muted uppercase tracking-widest">
              Đoạn chat
            </p>
            <div className="px-2 space-y-0.5">
              {CHAT_HISTORY.map((item, i) => (
                <Link
                  key={i}
                  href="/chat"
                  onClick={onNavClick}
                  className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs text-text-muted hover:text-foreground hover:bg-surface transition truncate"
                >
                  <MessageSquare size={12} className="flex-shrink-0 opacity-50" />
                  <span className="truncate">{item}</span>
                </Link>
              ))}
            </div>
          </>
        )}
      </div>

      {/* Bottom: Upgrade CTA + User panel */}
      <div className="flex-shrink-0 border-t border-border-theme">
        {!collapsed && (
          <div className="p-3">
            <div className="rounded-xl bg-gradient-to-br from-indigo-600/20 to-purple-600/20 border border-indigo-500/20 p-3">
              <p className="text-xs font-semibold text-white mb-0.5">Nâng cấp ngay</p>
              <p className="text-[10px] text-text-muted mb-2">Mở khoá gói cao cấp</p>
              <button className="w-full bg-gradient-to-r from-indigo-500 to-purple-600 text-white text-xs font-semibold py-1.5 rounded-lg hover:opacity-90 transition flex items-center justify-center gap-1">
                <Zap size={11} />
                Advanced
              </button>
            </div>
          </div>
        )}
        <div className={`flex items-center gap-2 p-3 cursor-pointer hover:bg-surface transition ${collapsed ? "justify-center" : ""}`}>
          <div className="w-7 h-7 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center flex-shrink-0">
            <User size={13} className="text-white" />
          </div>
          {!collapsed && (
            <div className="flex-1 min-w-0">
              <p className="text-xs font-semibold text-foreground truncate">da da</p>
              <p className="text-[10px] text-text-muted">STARTER</p>
            </div>
          )}
          {!collapsed && (
            <Link href="/settings" onClick={onNavClick}>
              <Settings size={13} className="text-text-muted hover:text-foreground transition" />
            </Link>
          )}
        </div>
      </div>
    </div>
  );

  return (
    <>
      {/* Mobile hamburger */}
      <button
        onClick={() => setDrawerOpen(true)}
        className="fixed top-3 left-3 z-50 p-2 rounded-lg bg-surface hover:bg-surface-hover text-foreground md:hidden touch-target"
        aria-label="Open navigation menu"
      >
        <Menu size={20} />
      </button>

      {/* Mobile overlay */}
      {drawerOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/60 md:hidden backdrop-blur-sm"
          onClick={() => setDrawerOpen(false)}
        />
      )}

      {/* Mobile drawer */}
      <aside
        className={`fixed inset-y-0 left-0 z-50 w-56 bg-sidebar-bg border-r border-border-theme transform transition-transform duration-200 ease-in-out md:hidden ${
          drawerOpen ? "translate-x-0 animate-slide-in" : "-translate-x-full"
        }`}
      >
        <div className="absolute top-3 right-3">
          <button
            onClick={() => setDrawerOpen(false)}
            className="p-1.5 rounded-lg text-text-muted hover:text-foreground hover:bg-surface transition"
          >
            <X size={16} />
          </button>
        </div>
        {sidebarContent(() => setDrawerOpen(false))}
      </aside>

      {/* Desktop full sidebar */}
      <aside
        className={`hidden lg:flex flex-col h-screen bg-sidebar-bg border-r border-border-theme fixed left-0 top-0 z-30 transition-all duration-200 ${
          collapsed ? "w-14" : "w-56"
        }`}
      >
        {sidebarContent()}
      </aside>

      {/* Tablet icon-only sidebar */}
      <aside className="hidden md:flex lg:hidden w-14 flex-col h-screen bg-sidebar-bg border-r border-border-theme fixed left-0 top-0 z-30">
        <div className="h-12 flex items-center justify-center border-b border-border-theme">
          <Link href="/dashboard">
            <div className="w-6 h-6 rounded bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center">
              <Zap size={12} className="text-white" />
            </div>
          </Link>
        </div>
        <nav className="flex-1 py-2 px-1.5 space-y-0.5 overflow-y-auto no-scrollbar">
          {[...TOP_ITEMS, ...PLATFORM_ITEMS].map(item => (
            <Link
              key={item.href}
              href={item.href}
              title={item.label}
              className={`flex items-center justify-center p-2 rounded-lg transition-all duration-150 ${
                isActive(item.href)
                  ? "bg-surface-hover text-foreground"
                  : "text-text-muted hover:text-foreground hover:bg-surface"
              }`}
            >
              <item.icon size={17} />
            </Link>
          ))}
        </nav>
        <div className="p-1.5 border-t border-border-theme">
          <Link href="/settings" title="Settings" className="flex items-center justify-center p-2 rounded-lg text-text-muted hover:text-foreground hover:bg-surface transition">
            <Settings size={17} />
          </Link>
        </div>
      </aside>
    </>
  );
}

// ----- Nav Row -----

function NavRow({ item, active, collapsed, onClick }: {
  item: NavItemConfig;
  active: boolean;
  collapsed: boolean;
  onClick?: () => void;
}) {
  return (
    <Link
      href={item.href}
      onClick={onClick}
      title={collapsed ? item.label : undefined}
      className={`flex items-center gap-2.5 px-2.5 py-1.5 rounded-lg transition-all duration-150 text-xs group ${
        active
          ? "bg-surface-hover text-foreground font-medium"
          : "text-text-muted hover:text-foreground hover:bg-surface"
      } ${collapsed ? "justify-center" : ""}`}
    >
      <item.icon size={16} className="flex-shrink-0" />
      {!collapsed && (
        <>
          <span className="flex-1 truncate">{item.label}</span>
          {item.badge === "hot" && <span className="badge-hot">HOT</span>}
          {item.badge === "new" && <span className="badge-new">NEW</span>}
          {active && !item.badge && <ChevronRight size={12} className="opacity-30" />}
        </>
      )}
    </Link>
  );
}
