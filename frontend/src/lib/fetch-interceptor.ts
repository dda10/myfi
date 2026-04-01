"use client";

/**
 * Monkey-patches the global fetch to automatically inject the JWT Authorization header
 * for requests to the backend API. Call this once on app mount.
 */
export function installFetchInterceptor() {
  if (typeof window === "undefined") return;

  const originalFetch = window.fetch;

  window.fetch = async function (input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
    const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : (input as Request).url;

    // Only inject token for requests to our backend API
    const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
    if (url.startsWith(apiUrl) || url.startsWith("http://localhost:8080")) {
      const token = localStorage.getItem("myfi-token");
      if (token) {
        const headers = new Headers(init?.headers);
        if (!headers.has("Authorization")) {
          headers.set("Authorization", `Bearer ${token}`);
        }
        init = { ...init, headers };
      }
    }

    return originalFetch.call(window, input, init);
  };
}
