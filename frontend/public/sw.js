// EziStock Service Worker — cache-first for API responses (Req 45.1)
const CACHE_NAME = "ezistock-v1";
const API_CACHE_PATHS = [
  "/api/portfolio/summary",
  "/api/watchlists",
  "/api/market/quote",
  "/api/market/dashboard",
  "/api/health",
  "/api/analyze/",
  "/api/recommendations/history",
  "/api/feedback/performance",
  "/api/market/corporate-actions",
  "/api/portfolio/dividend-history",
];

// Cache duration: 1 hour for API responses
const MAX_AGE_MS = 60 * 60 * 1000;

self.addEventListener("install", (event) => {
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k))
      )
    ).then(() => self.clients.claim())
  );
});

function shouldCacheRequest(url) {
  return API_CACHE_PATHS.some((p) => url.pathname.startsWith(p));
}

self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);

  // Only intercept GET requests to our API paths
  if (event.request.method !== "GET" || !shouldCacheRequest(url)) return;

  event.respondWith(
    fetch(event.request)
      .then((response) => {
        if (response.ok) {
          const clone = response.clone();
          caches.open(CACHE_NAME).then((cache) => cache.put(event.request, clone));
        }
        return response;
      })
      .catch(() =>
        caches.match(event.request).then((cached) => {
          if (cached) return cached;
          return new Response(JSON.stringify({ error: "offline", cached: false }), {
            status: 503,
            headers: { "Content-Type": "application/json" },
          });
        })
      )
  );
});
