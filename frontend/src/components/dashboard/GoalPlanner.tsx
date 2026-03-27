"use client";

import { useEffect, useState, useCallback } from "react";
import { Target, Plus, Pencil, Trash2, AlertTriangle, TrendingUp } from "lucide-react";
import { useI18n } from "@/context/I18nContext";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// --- Types matching backend model/goal_types.go ---

type GoalCategory = "retirement" | "emergency_fund" | "property" | "education" | "custom";

interface FinancialGoal {
  id: number;
  userId: number;
  name: string;
  targetAmount: number;
  targetDate: string;
  associatedAssetTypes?: string[];
  category: GoalCategory;
  createdAt: string;
  updatedAt: string;
}

interface GoalProgress {
  goalId: number;
  currentValue: number;
  targetAmount: number;
  progressPercent: number;
  requiredMonthlyContribution: number;
  monthsRemaining: number;
}

interface GoalFormData {
  name: string;
  targetAmount: string;
  targetDate: string;
  category: GoalCategory;
}

const CATEGORIES: { value: GoalCategory; label: string }[] = [
  { value: "retirement", label: "Retirement" },
  { value: "emergency_fund", label: "Emergency Fund" },
  { value: "property", label: "Property" },
  { value: "education", label: "Education" },
  { value: "custom", label: "Custom" },
];

const EMPTY_FORM: GoalFormData = { name: "", targetAmount: "", targetDate: "", category: "custom" };

// --- Helpers ---

function authHeaders(): HeadersInit {
  const token = typeof window !== "undefined" ? localStorage.getItem("myfi-token") : null;
  return token ? { Authorization: `Bearer ${token}`, "Content-Type": "application/json" } : { "Content-Type": "application/json" };
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T | null> {
  try {
    const res = await fetch(`${API_URL}${path}`, { headers: authHeaders(), ...init });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

// --- Component ---

export function GoalPlanner() {
  const { formatCurrency, formatDate } = useI18n();
  const [goals, setGoals] = useState<FinancialGoal[]>([]);
  const [progressMap, setProgressMap] = useState<Record<number, GoalProgress>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<GoalFormData>(EMPTY_FORM);

  const loadGoals = useCallback(async () => {
    setLoading(true);
    setError(null);
    const data = await apiFetch<FinancialGoal[]>("/api/goals");
    if (!data) {
      setError("Failed to load goals");
      setLoading(false);
      return;
    }
    setGoals(data);

    // Fetch progress for each goal
    const progressEntries = await Promise.all(
      data.map(async (g) => {
        const p = await apiFetch<GoalProgress>(`/api/goals/${g.id}/progress`);
        return [g.id, p] as const;
      }),
    );
    const map: Record<number, GoalProgress> = {};
    for (const [id, p] of progressEntries) {
      if (p) map[id] = p;
    }
    setProgressMap(map);
    setLoading(false);
  }, []);

  useEffect(() => { loadGoals(); }, [loadGoals]);

  const handleSubmit = async () => {
    const body = {
      name: form.name,
      targetAmount: parseFloat(form.targetAmount) || 0,
      targetDate: new Date(form.targetDate).toISOString(),
      category: form.category,
    };

    if (editingId) {
      await apiFetch(`/api/goals/${editingId}`, { method: "PUT", body: JSON.stringify(body) });
    } else {
      await apiFetch("/api/goals", { method: "POST", body: JSON.stringify(body) });
    }
    setShowForm(false);
    setEditingId(null);
    setForm(EMPTY_FORM);
    loadGoals();
  };

  const handleEdit = (goal: FinancialGoal) => {
    setForm({
      name: goal.name,
      targetAmount: String(goal.targetAmount),
      targetDate: goal.targetDate.slice(0, 10),
      category: goal.category,
    });
    setEditingId(goal.id);
    setShowForm(true);
  };

  const handleDelete = async (id: number) => {
    await apiFetch(`/api/goals/${id}`, { method: "DELETE" });
    loadGoals();
  };

  if (loading) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 animate-pulse">
        <div className="h-6 bg-zinc-800 rounded w-40 mb-4" />
        <div className="space-y-3">
          {[1, 2].map((i) => <div key={i} className="h-24 bg-zinc-800 rounded" />)}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
        <div className="flex items-center gap-2 text-red-400">
          <AlertTriangle size={16} />
          <span className="text-sm">{error}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-zinc-800 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Target size={18} className="text-violet-400" />
          <h2 className="text-lg font-semibold text-white">Goal Planner</h2>
        </div>
        <button
          onClick={() => { setShowForm(!showForm); setEditingId(null); setForm(EMPTY_FORM); }}
          className="flex items-center gap-1 px-3 py-1.5 bg-violet-600 hover:bg-violet-500 text-white text-xs rounded-lg transition-colors"
        >
          <Plus size={14} />
          New Goal
        </button>
      </div>

      {/* Form */}
      {showForm && (
        <div className="p-4 border-b border-zinc-800 bg-zinc-800/30">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <input
              placeholder="Goal name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-violet-500"
            />
            <input
              type="number"
              placeholder="Target amount (VND)"
              value={form.targetAmount}
              onChange={(e) => setForm({ ...form, targetAmount: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-violet-500"
            />
            <input
              type="date"
              value={form.targetDate}
              onChange={(e) => setForm({ ...form, targetDate: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-violet-500"
            />
            <select
              value={form.category}
              onChange={(e) => setForm({ ...form, category: e.target.value as GoalCategory })}
              className="bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-violet-500"
            >
              {CATEGORIES.map((c) => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </select>
          </div>
          <div className="flex gap-2 mt-3">
            <button
              onClick={handleSubmit}
              disabled={!form.name || !form.targetAmount || !form.targetDate}
              className="px-4 py-1.5 bg-violet-600 hover:bg-violet-500 disabled:opacity-40 text-white text-xs rounded-lg transition-colors"
            >
              {editingId ? "Update" : "Create"}
            </button>
            <button
              onClick={() => { setShowForm(false); setEditingId(null); setForm(EMPTY_FORM); }}
              className="px-4 py-1.5 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-xs rounded-lg transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Goals List */}
      <div className="p-4">
        {goals.length === 0 ? (
          <p className="text-zinc-500 text-sm text-center py-8">No goals yet. Create one to get started.</p>
        ) : (
          <div className="space-y-3">
            {goals.map((goal) => {
              const progress = progressMap[goal.id];
              const pct = progress ? Math.min(progress.progressPercent, 100) : 0;
              return (
                <div key={goal.id} className="border border-zinc-800 rounded-lg bg-zinc-800/40 p-4">
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <h3 className="text-white font-medium text-sm">{goal.name}</h3>
                      <span className="text-zinc-500 text-xs capitalize">{goal.category.replace("_", " ")}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <button onClick={() => handleEdit(goal)} className="p-1 text-zinc-500 hover:text-zinc-300">
                        <Pencil size={13} />
                      </button>
                      <button onClick={() => handleDelete(goal.id)} className="p-1 text-zinc-500 hover:text-red-400">
                        <Trash2 size={13} />
                      </button>
                    </div>
                  </div>

                  {/* Progress bar — Req 31.4 */}
                  <div className="w-full bg-zinc-700 rounded-full h-2 mb-2">
                    <div
                      className="h-2 rounded-full bg-violet-500 transition-all"
                      style={{ width: `${pct}%` }}
                    />
                  </div>

                  <div className="grid grid-cols-3 gap-2 text-xs">
                    <div>
                      <span className="text-zinc-500">Progress</span>
                      <p className="text-white font-medium">{pct.toFixed(1)}%</p>
                    </div>
                    <div>
                      <span className="text-zinc-500">Target</span>
                      <p className="text-zinc-300">{formatCurrency(goal.targetAmount)}</p>
                    </div>
                    <div>
                      <span className="text-zinc-500">Deadline</span>
                      <p className="text-zinc-300">{formatDate(goal.targetDate)}</p>
                    </div>
                  </div>

                  {progress && (
                    <div className="mt-2 flex items-center gap-4 text-xs">
                      <div className="flex items-center gap-1 text-zinc-400">
                        <TrendingUp size={12} />
                        <span>Monthly: {formatCurrency(progress.requiredMonthlyContribution)}</span>
                      </div>
                      <span className="text-zinc-500">
                        {progress.monthsRemaining > 0
                          ? `${progress.monthsRemaining} months remaining`
                          : "Target date passed"}
                      </span>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
