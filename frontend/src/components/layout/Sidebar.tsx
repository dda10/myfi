"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, LineChart, PieChart, Wallet, Settings, Filter, Zap, GitCompareArrows, Star } from "lucide-react";

const NAV_ITEMS = [
  { icon: <Home size={20} />,      label: "Overview",   href: "/overview" },
  { icon: <Wallet size={20} />,    label: "Portfolio",  href: "/portfolio" },
  { icon: <LineChart size={20} />, label: "Markets",    href: "/markets" },
  { icon: <Filter size={20} />,    label: "Filter",     href: "/filter" },
  { icon: <Zap size={20} />,       label: "Signals",    href: "/signals" },
  { icon: <GitCompareArrows size={20} />, label: "Comparison", href: "/comparison" },
  { icon: <Star size={20} />,             label: "Watchlist",  href: "/watchlist" },
  { icon: <PieChart size={20} />,  label: "Allocation", href: "/allocation" },
];

export function Sidebar() {
  const pathname = usePathname();

  const isActive = (href: string) => {
    const segment = pathname.split("/").filter(Boolean)[0] ?? "overview";
    return `/${segment}` === href;
  };

  return (
    <aside className="w-64 flex-col hidden sm:flex h-screen bg-sidebar-bg border-r border-border-theme fixed left-0 top-0">
      <div className="h-16 flex items-center px-6 border-b border-border-theme">
        <h1 className="text-xl font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
          MyFi Dashboard
        </h1>
      </div>

      <div className="flex-1 py-6 px-4 space-y-2">
        {NAV_ITEMS.map(item => (
          <NavItem
            key={item.href}
            icon={item.icon}
            label={item.label}
            href={item.href}
            active={isActive(item.href)}
          />
        ))}
      </div>

      <div className="p-4 border-t border-border-theme">
        <NavItem
          icon={<Settings size={20} />}
          label="Settings"
          href="/settings"
          active={isActive("/settings")}
        />
      </div>
    </aside>
  );
}

function NavItem({ icon, label, href, active = false }: {
  icon: React.ReactNode;
  label: string;
  href: string;
  active?: boolean;
}) {
  return (
    <Link
      href={href}
      className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 ${
        active
          ? "bg-surface-hover text-foreground"
          : "text-text-muted hover:text-foreground hover:bg-surface"
      }`}
    >
      {icon}
      <span className="font-medium">{label}</span>
    </Link>
  );
}
