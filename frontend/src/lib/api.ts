const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export function getAuthHeaders(): HeadersInit {
  if (typeof window === "undefined") return {};
  const token = localStorage.getItem("ezistock-token");
  return token ? { Authorization: `Bearer ${token}` } : {};
}

/** Metadata returned alongside API data for staleness/degradation handling. */
export interface ApiMeta {
  stale?: boolean;
  aiUnavailable?: boolean;
  status?: number;
}

/** Result wrapper that includes both data and metadata. */
export interface ApiResult<T> {
  data: T | null;
  meta: ApiMeta;
}

/**
 * Fetch JSON from the Go backend with automatic JWT injection.
 * Returns null on any error (network, 4xx, 5xx).
 * Requirements: 28.3, 29.3
 */
export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T | null> {
  try {
    const res = await fetch(`${API_URL}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...getAuthHeaders(),
        ...init?.headers,
      },
    });
    if (!res.ok) {
      // Auto-redirect to login on 401
      if (res.status === 401 && typeof window !== "undefined") {
        localStorage.removeItem("ezistock-token");
        if (window.location.pathname !== "/login") {
          window.location.href = "/login";
        }
      }
      return null;
    }
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

/**
 * Enhanced fetch that returns both data and metadata (stale flag, AI availability).
 * Use this when the caller needs to display stale-data indicators or AI degradation notices.
 * Requirements: 1.6, 28.3
 */
export async function apiFetchWithMeta<T>(path: string, init?: RequestInit): Promise<ApiResult<T>> {
  try {
    const res = await fetch(`${API_URL}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...getAuthHeaders(),
        ...init?.headers,
      },
    });

    const meta: ApiMeta = { status: res.status };

    if (!res.ok) {
      if (res.status === 401 && typeof window !== "undefined") {
        localStorage.removeItem("ezistock-token");
        if (window.location.pathname !== "/login") {
          window.location.href = "/login";
        }
      }
      // Check for AI service unavailable (503 with flag)
      if (res.status === 503) {
        try {
          const body = await res.json();
          if (body?.ai_service_unavailable) {
            meta.aiUnavailable = true;
          }
        } catch { /* ignore parse errors */ }
      }
      return { data: null, meta };
    }

    const body = await res.json();

    // Detect stale data flag from backend responses
    if (body && typeof body === "object") {
      if ("stale" in body && body.stale) {
        meta.stale = true;
      }
      if ("freshness" in body && body.freshness === "stale") {
        meta.stale = true;
      }
    }

    return { data: body as T, meta };
  } catch {
    return { data: null, meta: { status: 0 } };
  }
}
