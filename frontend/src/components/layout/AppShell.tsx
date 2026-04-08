"use client";

import { Sidebar } from "@/components/layout/Sidebar";
import { Header } from "@/components/layout/Header";
import { MobileNav } from "@/components/layout/MobileNav";
import { ChatWidget } from "@/components/chat/ChatWidget";
import { WatchlistProvider } from "@/context/WatchlistContext";

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <WatchlistProvider>
      <div className="flex bg-background min-h-screen text-foreground selection:bg-indigo-500/30">
        <Sidebar />
        {/*
          ml-0 on mobile (drawer-based),
          ml-14 on tablet (icon sidebar),
          ml-56 on desktop (full sidebar — matches Sidebar w-56)
        */}
        <div className="flex-1 md:ml-14 lg:ml-56 flex flex-col min-h-screen relative">
          <Header />
          <main className="flex-1 overflow-auto bg-background">
            <div className="p-4 md:p-6 max-w-screen-2xl mx-auto">
              {children}
            </div>
          </main>
        </div>
        <MobileNav />
        <ChatWidget />
      </div>
    </WatchlistProvider>
  );
}
