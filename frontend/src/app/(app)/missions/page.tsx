"use client";

import { useState, useEffect, useCallback } from "react";
import { Target, Plus, Pause, Play, Trash2, Clock, Zap, Calendar, Newspaper, X } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface Mission {
  id: string;
  name: string;
  status: "active" | "paused" | "completed";
  triggerType: "price_threshold" | "schedule" | "event" | "news";
  triggerParams: Record<string, string>;
  targetSymbols: string[];
  actionType: string;
  notificationPrefs: { email: boolean; inApp: boolean };
  lastTrigger?: string;
  nextRun?: string;
}

const STATUS_STYLE: Record<string, string> = {
  active: "bg-green-500/20 text-green-400",
  paused: "bg-yellow-500/20 text-yellow-400",
  completed: "bg-zinc-500/20 text-zinc-400",
};

const TRIGGER_ICON: Record<string, typeof Zap> = {
  price_threshold: Zap,
  schedule: Calendar,
  event: Clock,
  news: Newspaper,
};

export default function MissionsPage() {
  const { t, formatDate, formatTime } = useI18n();
  const [missions, setMissions] = useState<Mission[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const fetchMissions = useCallback(() => {
    setLoading(true);
    apiFetch<Mission[]>("/api/missions")
      .then((res) => { if (res) setMissions(res); })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => { fetchMissions(); }, [fetchMissions]);

  const handlePause = async (id: string) => {
    await apiFetch(`/api/missions/${id}/pause`, { method: "PUT" });
    fetchMissions();
  };

  const handleResume = async (id: string) => {
    await apiFetch(`/api/missions/${id}/resume`, { method: "PUT" });
    fetchMissions();
  };

  const handleDelete = async (id: string) => {
    await apiFetch(`/api/missions/${id}`, { method: "DELETE" });
    fetchMissions();
  };

  const filtered = statusFilter === "all" ? missions : missions.filter((m) => m.status === statusFilter);

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-foreground">{t("missions.title")}</h1>
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="bg-card-bg border border-border-theme rounded-xl p-5 h-24 animate-pulse">
              <div className="h-4 w-1/2 bg-surface rounded mb-2" />
              <div className="h-3 w-1/3 bg-surface rounded" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">{t("missions.title")}</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 text-sm bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg transition"
        >
          <Plus size={16} />
          {t("missions.create")}
        </button>
      </div>

      {/* Status Filter */}
      <div className="flex gap-2">
        {["all", "active", "paused", "completed"].map((s) => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`text-xs px-3 py-1.5 rounded-full transition ${
              statusFilter === s
                ? "bg-indigo-600 text-white"
                : "bg-surface text-text-muted hover:bg-surface/80"
            }`}
          >
            {s === "all" ? t("common.all") : t(`missions.${s}`)}
          </button>
        ))}
      </div>

      {/* Mission List */}
      {filtered.length === 0 ? (
        <div className="p-8 rounded-xl bg-surface border border-border-theme text-center text-text-muted">
          {t("common.no_data")}
        </div>
      ) : (
        <div className="space-y-3">
          {filtered.map((mission) => {
            const TriggerIcon = TRIGGER_ICON[mission.triggerType] ?? Zap;
            return (
              <div key={mission.id} className="bg-card-bg border border-border-theme rounded-xl p-5">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-lg bg-indigo-500/10 flex items-center justify-center">
                      <TriggerIcon size={18} className="text-indigo-400" />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold text-foreground">{mission.name}</h3>
                      <div className="flex items-center gap-2 mt-0.5">
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${STATUS_STYLE[mission.status]}`}>
                          {t(`missions.${mission.status}`)}
                        </span>
                        <span className="text-xs text-text-muted">
                          {t("missions.trigger")}: {mission.triggerType.replace("_", " ")}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-1.5">
                    {mission.status === "active" && (
                      <button onClick={() => handlePause(mission.id)} className="p-1.5 rounded-lg hover:bg-surface text-yellow-400 transition" title={t("missions.pause")}>
                        <Pause size={16} />
                      </button>
                    )}
                    {mission.status === "paused" && (
                      <button onClick={() => handleResume(mission.id)} className="p-1.5 rounded-lg hover:bg-surface text-green-400 transition" title={t("missions.resume")}>
                        <Play size={16} />
                      </button>
                    )}
                    <button onClick={() => handleDelete(mission.id)} className="p-1.5 rounded-lg hover:bg-surface text-red-400 transition" title={t("btn.delete")}>
                      <Trash2 size={16} />
                    </button>
                  </div>
                </div>
                <div className="flex items-center gap-4 mt-3 text-xs text-text-muted">
                  <span>{t("table.symbol")}: {mission.targetSymbols.join(", ")}</span>
                  {mission.lastTrigger && <span>Last: {formatDate(mission.lastTrigger)} {formatTime(mission.lastTrigger)}</span>}
                  {mission.nextRun && <span>Next: {formatDate(mission.nextRun)} {formatTime(mission.nextRun)}</span>}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Create Mission Modal */}
      {showCreate && <CreateMissionModal onClose={() => { setShowCreate(false); fetchMissions(); }} />}
    </div>
  );
}

function CreateMissionModal({ onClose }: { onClose: () => void }) {
  const { t } = useI18n();
  const [name, setName] = useState("");
  const [triggerType, setTriggerType] = useState("price_threshold");
  const [symbols, setSymbols] = useState("");
  const [actionType, setActionType] = useState("alert");
  const [emailNotif, setEmailNotif] = useState(true);
  const [inAppNotif, setInAppNotif] = useState(true);
  const [saving, setSaving] = useState(false);

  const handleSubmit = async () => {
    if (!name.trim() || !symbols.trim()) return;
    setSaving(true);
    await apiFetch("/api/missions", {
      method: "POST",
      body: JSON.stringify({
        name: name.trim(),
        triggerType,
        triggerParams: {},
        targetSymbols: symbols.split(",").map((s) => s.trim().toUpperCase()).filter(Boolean),
        actionType,
        notificationPrefs: { email: emailNotif, inApp: inAppNotif },
      }),
    });
    setSaving(false);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-card-bg border border-border-theme rounded-2xl w-full max-w-md p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-bold text-foreground">{t("missions.create")}</h2>
          <button onClick={onClose} className="text-text-muted hover:text-foreground transition"><X size={20} /></button>
        </div>

        <div>
          <label className="text-xs text-text-muted font-medium block mb-1">{t("table.name")}</label>
          <input value={name} onChange={(e) => setName(e.target.value)} className="w-full bg-surface border border-border-theme rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:border-indigo-500" />
        </div>

        <div>
          <label className="text-xs text-text-muted font-medium block mb-1">{t("missions.trigger")}</label>
          <select value={triggerType} onChange={(e) => setTriggerType(e.target.value)} className="w-full bg-surface border border-border-theme rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:border-indigo-500">
            <option value="price_threshold">Price Threshold</option>
            <option value="schedule">Schedule</option>
            <option value="event">Event</option>
            <option value="news">News</option>
          </select>
        </div>

        <div>
          <label className="text-xs text-text-muted font-medium block mb-1">{t("table.symbol")} (comma-separated)</label>
          <input value={symbols} onChange={(e) => setSymbols(e.target.value)} placeholder="FPT, VNM, VCB" className="w-full bg-surface border border-border-theme rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:border-indigo-500" />
        </div>

        <div>
          <label className="text-xs text-text-muted font-medium block mb-1">{t("missions.action")}</label>
          <select value={actionType} onChange={(e) => setActionType(e.target.value)} className="w-full bg-surface border border-border-theme rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:border-indigo-500">
            <option value="alert">Alert</option>
            <option value="report">Report</option>
            <option value="agent_analysis">Agent Analysis</option>
          </select>
        </div>

        <div className="flex gap-4">
          <label className="flex items-center gap-2 text-sm text-foreground">
            <input type="checkbox" checked={emailNotif} onChange={(e) => setEmailNotif(e.target.checked)} className="rounded" />
            Email
          </label>
          <label className="flex items-center gap-2 text-sm text-foreground">
            <input type="checkbox" checked={inAppNotif} onChange={(e) => setInAppNotif(e.target.checked)} className="rounded" />
            In-app
          </label>
        </div>

        <div className="flex gap-2 pt-2">
          <button onClick={onClose} className="flex-1 text-sm bg-surface text-text-muted hover:bg-surface/80 py-2 rounded-lg transition">
            {t("btn.cancel")}
          </button>
          <button onClick={handleSubmit} disabled={saving || !name.trim() || !symbols.trim()} className="flex-1 text-sm bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white py-2 rounded-lg transition">
            {saving ? t("common.loading") : t("btn.save")}
          </button>
        </div>
      </div>
    </div>
  );
}
