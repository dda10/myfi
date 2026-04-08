"use client";

import { useState, useEffect } from "react";
import { Bell } from "lucide-react";

interface Notification {
  id: string;
  type: string;
  message: string;
  read: boolean;
  created_at: string;
}

export function NotificationBell() {
  const [open, setOpen] = useState(false);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    // Poll unread count every 30s
    const fetchCount = async () => {
      try {
        const res = await fetch("/api/notifications/unread-count");
        if (res.ok) {
          const data = await res.json();
          setUnreadCount(data.count || 0);
        }
      } catch { /* ignore */ }
    };
    fetchCount();
    const interval = setInterval(fetchCount, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleOpen = async () => {
    setOpen(!open);
    if (!open) {
      try {
        const res = await fetch("/api/notifications?limit=10");
        if (res.ok) {
          const data = await res.json();
          setNotifications(data.notifications || []);
        }
      } catch { /* ignore */ }
    }
  };

  return (
    <div className="relative">
      <button
        onClick={handleOpen}
        className="relative p-2 rounded-lg text-text-muted hover:text-foreground hover:bg-surface transition-colors"
        aria-label={`Notifications${unreadCount > 0 ? ` (${unreadCount} unread)` : ""}`}
      >
        <Bell size={20} />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 w-4 h-4 bg-red-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 top-12 w-80 bg-surface border border-border-theme rounded-xl shadow-xl z-50 overflow-hidden">
          <div className="px-4 py-3 border-b border-border-theme flex items-center justify-between">
            <span className="font-medium text-foreground text-sm">Notifications</span>
            {unreadCount > 0 && (
              <button className="text-xs text-blue-400 hover:text-blue-300">Mark all read</button>
            )}
          </div>
          <div className="max-h-80 overflow-y-auto">
            {notifications.length === 0 ? (
              <div className="px-4 py-8 text-center text-text-muted text-sm">No notifications</div>
            ) : (
              notifications.map(n => (
                <div
                  key={n.id}
                  className={`px-4 py-3 border-b border-border-theme last:border-0 ${
                    !n.read ? "bg-surface-hover" : ""
                  }`}
                >
                  <p className="text-sm text-foreground">{n.message}</p>
                  <p className="text-xs text-text-muted mt-1">{n.created_at}</p>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
