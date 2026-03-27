"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";

export default function LoginPage() {
  const { login, isAuthenticated } = useAuth();
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  // Redirect if already authenticated
  if (isAuthenticated) {
    router.replace("/overview");
    return null;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(username, password);
      router.replace("/overview");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-background">
      <div className="w-full max-w-sm p-8 bg-card-bg border border-border-theme rounded-2xl shadow-xl">
        <h1 className="text-2xl font-bold text-foreground text-center mb-6">
          MyFi Login
        </h1>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div>
            <label htmlFor="username" className="block text-sm font-medium text-text-muted mb-1">
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              autoComplete="username"
              className="w-full px-3 py-2 bg-input-bg border border-border-theme rounded-lg text-foreground placeholder-text-muted focus:outline-none focus:border-accent transition text-sm"
              placeholder="Enter username"
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-text-muted mb-1">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              autoComplete="current-password"
              className="w-full px-3 py-2 bg-input-bg border border-border-theme rounded-lg text-foreground placeholder-text-muted focus:outline-none focus:border-accent transition text-sm"
              placeholder="Enter password"
            />
          </div>

          {error && (
            <p className="text-red-500 text-sm text-center">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-medium rounded-lg transition text-sm"
          >
            {loading ? "Signing in..." : "Sign In"}
          </button>
        </form>

        <p className="text-center text-sm text-text-muted mt-4">
          No account?{" "}
          <button
            onClick={async () => {
              if (!username || !password) { setError("Enter username and password first"); return; }
              setLoading(true); setError("");
              try {
                const res = await fetch("http://localhost:8080/api/auth/register", {
                  method: "POST", headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({ username, password }),
                });
                if (!res.ok) { const d = await res.json(); throw new Error(d.error || "Registration failed"); }
                await login(username, password);
                router.replace("/overview");
              } catch (err) {
                setError(err instanceof Error ? err.message : "Registration failed");
              } finally { setLoading(false); }
            }}
            className="text-indigo-400 hover:underline font-medium"
          >
            Register
          </button>
        </p>
      </div>
    </div>
  );
}
