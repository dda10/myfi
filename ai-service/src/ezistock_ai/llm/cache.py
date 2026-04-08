"""LLM response cache — avoids redundant LLM calls for identical inputs.

Caches responses keyed by (model, symbol, data_snapshot_hash) with configurable TTL
(default 1 hour).

Requirements: 48.4
"""

from __future__ import annotations

import hashlib
import json
import logging
import time
from dataclasses import dataclass
from typing import Any, Optional

logger = logging.getLogger(__name__)

_DEFAULT_TTL_SECONDS = 3600  # 1 hour
_MAX_CACHE_SIZE = 1000  # Evict oldest entries beyond this


@dataclass
class CacheEntry:
    """A cached LLM response."""

    key: str
    response: Any
    created_at: float
    ttl: float


class LLMCache:
    """In-memory LLM response cache with TTL-based expiration.

    Keys are derived from (model, input content hash) so identical prompts
    for the same model hit the cache.
    """

    def __init__(self, default_ttl: float = _DEFAULT_TTL_SECONDS, max_size: int = _MAX_CACHE_SIZE) -> None:
        self._default_ttl = default_ttl
        self._max_size = max_size
        self._store: dict[str, CacheEntry] = {}
        self._hits = 0
        self._misses = 0

    def get(self, model: str, inputs: dict[str, Any]) -> Optional[Any]:
        """Look up a cached response. Returns None on miss or expiry."""
        key = self._make_key(model, inputs)
        entry = self._store.get(key)
        if entry is None:
            self._misses += 1
            return None
        if time.time() - entry.created_at > entry.ttl:
            # Expired
            del self._store[key]
            self._misses += 1
            return None
        self._hits += 1
        logger.debug("LLM cache hit: model=%s key=%s", model, key[:16])
        return entry.response

    def put(self, model: str, inputs: dict[str, Any], response: Any, ttl: Optional[float] = None) -> None:
        """Store a response in the cache."""
        key = self._make_key(model, inputs)
        self._store[key] = CacheEntry(
            key=key,
            response=response,
            created_at=time.time(),
            ttl=ttl if ttl is not None else self._default_ttl,
        )
        self._evict_if_needed()

    def invalidate(self, model: str, inputs: dict[str, Any]) -> bool:
        """Remove a specific entry. Returns True if it existed."""
        key = self._make_key(model, inputs)
        return self._store.pop(key, None) is not None

    def clear(self) -> None:
        """Remove all cached entries."""
        self._store.clear()

    @property
    def stats(self) -> dict[str, int]:
        """Return cache hit/miss statistics."""
        return {"hits": self._hits, "misses": self._misses, "size": len(self._store)}

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    @staticmethod
    def _make_key(model: str, inputs: dict[str, Any]) -> str:
        """Deterministic cache key from model + inputs."""
        raw = json.dumps({"model": model, **inputs}, sort_keys=True, default=str)
        return hashlib.sha256(raw.encode()).hexdigest()

    def _evict_if_needed(self) -> None:
        """Evict oldest entries if cache exceeds max size."""
        if len(self._store) <= self._max_size:
            return
        # Sort by created_at, remove oldest
        sorted_keys = sorted(self._store, key=lambda k: self._store[k].created_at)
        to_remove = len(self._store) - self._max_size
        for key in sorted_keys[:to_remove]:
            del self._store[key]
        logger.debug("Evicted %d cache entries", to_remove)
