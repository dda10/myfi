"use client";

import { useAuth } from "@/context/AuthContext";
import { usePathname } from "next/navigation";
import { AppProvider } from "@/context/AppContext";
import { AppShell } from "@/components/layout/AppShell";

/**
 * Renders AppShell for authenticated routes, or bare children for /login.
 * Works together with ProtectedRoute which handles the redirect logic.
 */
export function AuthenticatedLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const pathname = usePathname();

  // Login page renders without AppShell
  if (pathname === "/login") {
    return <>{children}</>;
  }

  // While checking auth, show spinner (ProtectedRoute handles this too)
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="animate-spin rounded-full h-8 w-8 border-2 border-indigo-500 border-t-transparent" />
      </div>
    );
  }

  // Not authenticated and not on login — ProtectedRoute will redirect
  if (!isAuthenticated) {
    return null;
  }

  return (
    <AppProvider>
      <AppShell>{children}</AppShell>
    </AppProvider>
  );
}
