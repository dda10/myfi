"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";
import { useRouter, usePathname } from "next/navigation";
import { installFetchInterceptor } from "@/lib/fetch-interceptor";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

interface User {
  id: number;
  username: string;
  email?: string;
  disclaimerAcknowledged?: boolean;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  disclaimerAcknowledged: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  acknowledgeDisclaimer: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const DISCLAIMER_STORAGE_KEY = "ezistock-disclaimer-ack";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [disclaimerAcknowledged, setDisclaimerAcknowledged] = useState(false);

  // Install global fetch interceptor to auto-inject JWT token
  useEffect(() => {
    installFetchInterceptor();
  }, []);

  // Load disclaimer acknowledgment from localStorage on mount
  useEffect(() => {
    try {
      const ack = localStorage.getItem(DISCLAIMER_STORAGE_KEY);
      if (ack === "true") {
        setDisclaimerAcknowledged(true);
      }
    } catch {
      // localStorage unavailable
    }
  }, []);

  const checkAuth = useCallback(async () => {
    try {
      const token = typeof window !== "undefined" ? localStorage.getItem("ezistock-token") : null;
      if (!token) { setUser(null); setIsLoading(false); return; }
      const res = await fetch(`${API_URL}/api/auth/me`, {
        headers: { Authorization: `Bearer ${token}` },
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        const userData = data.user ?? data;
        setUser(userData);
        // If backend returns disclaimer status, sync it
        if (userData.disclaimerAcknowledged) {
          setDisclaimerAcknowledged(true);
          try { localStorage.setItem(DISCLAIMER_STORAGE_KEY, "true"); } catch { /* noop */ }
        }
      } else {
        localStorage.removeItem("ezistock-token");
        setUser(null);
      }
    } catch {
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  const login = useCallback(async (username: string, password: string) => {
    const res = await fetch(`${API_URL}/api/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ username, password }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error ?? "Login failed");
    }
    const data = await res.json();
    // Store JWT token for use in API requests
    if (data.token) {
      localStorage.setItem("ezistock-token", data.token);
    }
    if (data.user) {
      setUser(data.user);
      if (data.user.disclaimerAcknowledged) {
        setDisclaimerAcknowledged(true);
        try { localStorage.setItem(DISCLAIMER_STORAGE_KEY, "true"); } catch { /* noop */ }
      }
    } else {
      await checkAuth();
    }
  }, [checkAuth]);

  const logout = useCallback(async () => {
    try {
      const token = typeof window !== "undefined" ? localStorage.getItem("ezistock-token") : null;
      await fetch(`${API_URL}/api/auth/logout`, {
        method: "POST",
        credentials: "include",
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
    } finally {
      localStorage.removeItem("ezistock-token");
      setUser(null);
    }
  }, []);

  /**
   * Acknowledge the investment disclaimer (Req 49.2).
   * Persists to backend and localStorage so the user isn't asked again.
   */
  const acknowledgeDisclaimer = useCallback(async () => {
    try {
      const token = typeof window !== "undefined" ? localStorage.getItem("ezistock-token") : null;
      if (token) {
        // Best-effort call to backend to persist acknowledgment
        await fetch(`${API_URL}/api/auth/disclaimer`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          credentials: "include",
          body: JSON.stringify({ acknowledged: true }),
        }).catch(() => { /* backend may not have this endpoint yet */ });
      }
    } finally {
      setDisclaimerAcknowledged(true);
      try { localStorage.setItem(DISCLAIMER_STORAGE_KEY, "true"); } catch { /* noop */ }
      // Update user object
      setUser((prev) => prev ? { ...prev, disclaimerAcknowledged: true } : prev);
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: user !== null,
        isLoading,
        disclaimerAcknowledged,
        login,
        logout,
        checkAuth,
        acknowledgeDisclaimer,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

/**
 * Wraps children and redirects to /login when no valid session exists.
 * Requirement 29.4: display login screen and prevent access when no valid session.
 */
export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && !isAuthenticated && pathname !== "/login") {
      router.replace("/login");
    }
  }, [isAuthenticated, isLoading, pathname, router]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="animate-spin rounded-full h-8 w-8 border-2 border-indigo-500 border-t-transparent" />
      </div>
    );
  }

  if (!isAuthenticated && pathname !== "/login") {
    return null;
  }

  return <>{children}</>;
}
