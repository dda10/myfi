"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";
import { useI18n } from "@/context/I18nContext";
import { LogIn, UserPlus, ShieldCheck } from "lucide-react";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default function LoginPage() {
  const { login, isAuthenticated, disclaimerAcknowledged, acknowledgeDisclaimer } = useAuth();
  const { t } = useI18n();
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [mode, setMode] = useState<"login" | "register">("login");

  // Redirect to dashboard when authenticated and disclaimer acknowledged
  useEffect(() => {
    if (isAuthenticated && disclaimerAcknowledged) {
      router.replace("/dashboard");
    }
  }, [isAuthenticated, disclaimerAcknowledged, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    // Client-side validation
    if (username.length < 3) {
      setError(t("auth.username_min"));
      return;
    }
    if (mode === "register" && password.length < 8) {
      setError(t("auth.password_min"));
      return;
    }

    setLoading(true);
    try {
      if (mode === "register") {
        const res = await fetch(`${API_URL}/api/auth/register`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ username, password }),
        });
        if (!res.ok) {
          const data = await res.json().catch(() => ({}));
          throw new Error(data.error ?? t("auth.register_failed"));
        }
      }
      await login(username, password);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "";
      if (mode === "register") {
        setError(msg || t("auth.register_failed"));
      } else {
        setError(msg || t("auth.login_failed"));
      }
    } finally {
      setLoading(false);
    }
  };

  const handleAcknowledge = async () => {
    await acknowledgeDisclaimer();
  };

  // Show disclaimer step after successful login if not yet acknowledged
  if (isAuthenticated && !disclaimerAcknowledged) {
    return (
      <div className="w-full max-w-md p-8 rounded-2xl bg-surface border border-border-theme shadow-xl">
        <div className="text-center mb-6">
          <div className="mx-auto w-12 h-12 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
            <ShieldCheck size={24} className="text-amber-500" />
          </div>
          <h2 className="text-xl font-bold text-foreground">
            {t("auth.disclaimer_title")}
          </h2>
        </div>

        <div className="p-4 rounded-lg bg-amber-500/5 border border-amber-500/20 mb-6">
          <p className="text-sm text-text-muted leading-relaxed">
            {t("auth.disclaimer_text")}
          </p>
        </div>

        <button
          onClick={handleAcknowledge}
          className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-white font-medium transition-colors"
        >
          <ShieldCheck size={18} />
          {t("auth.disclaimer_accept")}
        </button>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md p-8 rounded-2xl bg-surface border border-border-theme shadow-xl">
      <div className="text-center mb-8">
        <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
          EziStock
        </h1>
        <p className="text-text-muted mt-2">{t("auth.subtitle")}</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="username" className="block text-sm font-medium text-text-muted mb-1">
            {t("auth.username")}
          </label>
          <input
            id="username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="w-full px-4 py-2.5 rounded-lg bg-background border border-border-theme text-foreground focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>
        <div>
          <label htmlFor="password" className="block text-sm font-medium text-text-muted mb-1">
            {t("auth.password")}
          </label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-2.5 rounded-lg bg-background border border-border-theme text-foreground focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>

        {error && <p className="text-red-400 text-sm">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-white font-medium transition-colors disabled:opacity-50"
        >
          {mode === "register" ? <UserPlus size={18} /> : <LogIn size={18} />}
          {loading
            ? t("auth.logging_in")
            : mode === "register"
              ? t("auth.register")
              : t("auth.login")}
        </button>
      </form>

      <p className="text-center text-sm text-text-muted mt-4">
        {mode === "login" ? t("auth.no_account") : t("auth.have_account")}{" "}
        <button
          onClick={() => { setMode(mode === "login" ? "register" : "login"); setError(""); }}
          className="text-blue-400 hover:underline font-medium"
        >
          {mode === "login" ? t("auth.register") : t("auth.login")}
        </button>
      </p>
    </div>
  );
}
