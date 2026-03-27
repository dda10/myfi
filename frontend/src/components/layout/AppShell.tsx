"use client";

import { Sidebar } from "@/components/layout/Sidebar";
import { Header } from "@/components/layout/Header";
import { ChatWidget } from "@/components/chat/ChatWidget";
import { WatchlistProvider } from "@/context/WatchlistContext";

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <WatchlistProvider>
      <div className="flex bg-background min-h-screen text-foreground font-sans selection:bg-indigo-500/30">
        <Sidebar />
        <div className="flex-1 sm:ml-64 flex flex-col min-h-screen relative border-border-theme border-l">
          <Header />
          <main className="flex-1 p-8 bg-background">
            <div className="max-w-7xl mx-auto">
              {children}
            </div>
          </main>
        </div>
        <ChatWidget />
      </div>
    </WatchlistProvider>
  );
}
