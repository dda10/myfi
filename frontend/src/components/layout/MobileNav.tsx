"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, PieChart, Filter, TrendingUp, MessageSquare } from "lucide-react";
import { useI18n } from "@/context/I18nContext";

const MOBILE_NAV_ITEMS = [
  { icon: Home,           labelKey: "nav.dashboard", href: "/dashboard" },
  { icon: PieChart,       labelKey: "nav.portfolio", href: "/portfolio" },
  { icon: Filter,         labelKey: "nav.screener",  href: "/screener" },
  { icon: TrendingUp,     labelKey: "nav.ranking",   href: "/ranking" },
  { icon: MessageSquare,  labelKey: "nav.chat",      href: "/chat" },
];

export function MobileNav() {
  const pathname = usePathname();
  const { t } = useI18n();

  const isActive = (href: string) => {
    if (href === "/dashboard") return pathname === "/dashboard";
    return pathname.startsWith(href);
  };

  return (
    <nav
      className="fixed bottom-0 left-0 right-0 z-40 bg-card-bg border-t border-border-theme md:hidden"
      style={{ height: 56 }}
    >
      <div className="flex items-center justify-around h-full">
        {MOBILE_NAV_ITEMS.map(({ icon: Icon, labelKey, href }) => {
          const active = isActive(href);
          return (
            <Link
              key={href}
              href={href}
              className={`touch-target flex flex-col items-center justify-center gap-1 relative text-[10px] transition-colors ${
                active
                  ? "text-accent font-medium"
                  : "text-text-muted"
              }`}
            >
              {/* Active indicator bar */}
              {active && (
                <span className="absolute top-0 left-1/2 -translate-x-1/2 w-8 h-0.5 rounded-full bg-accent" />
              )}
              <Icon size={22} />
              <span>{t(labelKey)}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
