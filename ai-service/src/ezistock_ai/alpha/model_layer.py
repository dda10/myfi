"""Alpha Mining Engine — Model Layer.

Combines traditional ML models (gradient boosting, random forest) with
LLM-driven factor discovery inspired by AlphaAgent (KDD 2025).

The LLM loop: Idea Agent → Factor Agent → Eval Agent
- Idea Agent: LLM proposes market hypotheses for the VN market
- Factor Agent: LLM constructs pandas factor expressions with regularization
- Eval Agent: Validates via backtesting, feeds results back for refinement

Traditional ML runs in parallel for ensemble diversity.

Requirements: 10.1, 10.2, 10.3, 10.5, 10.6
"""

from __future__ import annotations

import hashlib
import io
import json
import logging
import pickle
import time
from dataclasses import dataclass, field
from typing import Any

import numpy as np
import pandas as pd
from sklearn.ensemble import GradientBoostingRegressor, RandomForestRegressor

from ezistock_ai.alpha.regime_detector import MarketRegime, RegimeDetector, RegimeResult
from ezistock_ai.config import Config
from ezistock_ai.infra.parquet import ParquetStore
from ezistock_ai.llm.router import LLMRouter, TaskType

logger = logging.getLogger(__name__)

_MODEL_KEY_PREFIX = "alpha/models/"
_FACTOR_REGISTRY_KEY = "alpha/factor_registry.json"


# ---------------------------------------------------------------------------
# LLM-Driven Factor Discovery (adapted from AlphaAgent)
# ---------------------------------------------------------------------------


@dataclass
class AlphaFactor:
    """A discovered alpha factor with its expression and metadata."""

    name: str
    hypothesis: str  # Market hypothesis that motivated this factor
    expression: str  # Pandas expression to compute the factor
    category: str    # price, volume, fundamental, etc.
    ic_mean: float = 0.0  # Information coefficient (correlation with forward returns)
    ic_std: float = 0.0
    sharpe: float = 0.0
    decay_rate: float = 0.0  # How fast the factor loses predictive power
    created_at: str = ""
    version: int = 1
    fingerprint: str = ""  # Hash of expression for deduplication


@dataclass
class FactorRegistry:
    """Tracks all discovered factors with deduplication (AlphaAgent regularization)."""

    factors: list[AlphaFactor] = field(default_factory=list)
    expression_hashes: set[str] = field(default_factory=set)

    def add(self, factor: AlphaFactor) -> bool:
        """Add a factor if it's not a duplicate. Returns True if added."""
        fp = hashlib.sha256(factor.expression.encode()).hexdigest()[:16]
        if fp in self.expression_hashes:
            logger.debug("Duplicate factor rejected: %s", factor.name)
            return False
        factor.fingerprint = fp
        self.expression_hashes.add(fp)
        self.factors.append(factor)
        return True

    def top_factors(self, n: int = 20, min_ic: float = 0.02) -> list[AlphaFactor]:
        """Return top N factors by IC, filtering out weak ones."""
        valid = [f for f in self.factors if abs(f.ic_mean) >= min_ic]
        return sorted(valid, key=lambda f: abs(f.ic_mean), reverse=True)[:n]

    def to_dict(self) -> dict:
        return {
            "factors": [
                {
                    "name": f.name, "hypothesis": f.hypothesis,
                    "expression": f.expression, "category": f.category,
                    "ic_mean": f.ic_mean, "ic_std": f.ic_std,
                    "sharpe": f.sharpe, "decay_rate": f.decay_rate,
                    "fingerprint": f.fingerprint, "version": f.version,
                }
                for f in self.factors
            ],
        }

    @classmethod
    def from_dict(cls, data: dict) -> FactorRegistry:
        reg = cls()
        for item in data.get("factors", []):
            factor = AlphaFactor(**{k: v for k, v in item.items() if k != "fingerprint"})
            factor.fingerprint = item.get("fingerprint", "")
            reg.factors.append(factor)
            if factor.fingerprint:
                reg.expression_hashes.add(factor.fingerprint)
        return reg


class IdeaAgent:
    """LLM-driven hypothesis generation for the Vietnamese market.

    Proposes market hypotheses like:
    - "Foreign institutional buying precedes price momentum in VN30 stocks"
    - "Low P/B stocks with improving ROE outperform in bear regimes"
    """

    PROMPT = """\
You are a quantitative researcher specializing in the Vietnamese stock market
(HOSE, HNX, UPCOM). Given the current market regime and existing factor
performance, propose {n} new market hypotheses that could lead to
profitable alpha factors.

Current regime: {regime}
Existing top factors: {existing_factors}
Recent factor performance: {performance_summary}

Requirements:
- Each hypothesis should be specific to Vietnamese market dynamics
- Consider foreign investor behavior, sector rotation, liquidity patterns
- Avoid duplicating existing factors
- Focus on signals that are actionable with daily OHLCV + fundamental data

Return a JSON array of objects with keys:
- hypothesis: string describing the market hypothesis
- category: one of "price", "volume", "fundamental", "money_flow", "technical", "macro"
- rationale: why this might work in the VN market
"""

    def __init__(self, llm_router: LLMRouter) -> None:
        self._llm = llm_router

    async def generate_hypotheses(
        self,
        regime: RegimeResult,
        registry: FactorRegistry,
        n: int = 5,
    ) -> list[dict[str, str]]:
        """Generate N market hypotheses using LLM."""
        model = self._llm.get_model(TaskType.ANALYSIS)

        existing = ", ".join(f.name for f in registry.top_factors(5)) or "None yet"
        perf = ", ".join(
            f"{f.name}: IC={f.ic_mean:.3f}" for f in registry.top_factors(3)
        ) or "No performance data"

        messages = [
            {"role": "user", "content": self.PROMPT.format(
                n=n,
                regime=f"{regime.regime.value} ({regime.description})",
                existing_factors=existing,
                performance_summary=perf,
            )},
        ]

        response = await model.ainvoke(messages)
        try:
            return json.loads(response.content)
        except (json.JSONDecodeError, AttributeError):
            logger.warning("Failed to parse idea agent response")
            return []


class FactorAgent:
    """LLM-driven factor construction from hypotheses.

    Converts market hypotheses into executable pandas expressions
    that can be applied to the signal space DataFrame.
    """

    PROMPT = """\
You are a quantitative developer for the Vietnamese stock market.
Convert the following market hypothesis into a pandas factor expression.

Hypothesis: {hypothesis}
Category: {category}

Available columns in the signal space DataFrame (MultiIndex: date, symbol):
- close, open, high, low, volume (OHLCV)
- For rolling operations, use .groupby(level='symbol').transform(...)

Requirements:
- Expression must be valid pandas/numpy code
- Must produce a single numeric Series
- Use only standard pandas operations (rolling, ewm, shift, pct_change, rank)
- Include the lookback period as a concrete number
- The expression should be a single line or a short function body

Return a JSON object with keys:
- name: short factor name (snake_case)
- expression: the pandas expression string
- description: what the factor captures
"""

    def __init__(self, llm_router: LLMRouter) -> None:
        self._llm = llm_router

    async def construct_factor(self, hypothesis: str, category: str) -> AlphaFactor | None:
        """Convert a hypothesis into an executable factor expression."""
        model = self._llm.get_model(TaskType.ANALYSIS)

        messages = [
            {"role": "user", "content": self.PROMPT.format(
                hypothesis=hypothesis,
                category=category,
            )},
        ]

        response = await model.ainvoke(messages)
        try:
            data = json.loads(response.content)
            return AlphaFactor(
                name=data["name"],
                hypothesis=hypothesis,
                expression=data["expression"],
                category=category,
                created_at=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            )
        except (json.JSONDecodeError, KeyError, AttributeError) as exc:
            logger.warning("Failed to parse factor agent response: %s", exc)
            return None


# ---------------------------------------------------------------------------
# Traditional ML Models + Ensemble
# ---------------------------------------------------------------------------


@dataclass
class ModelArtifact:
    """A trained model with metadata."""

    name: str
    regime: str
    version: str
    feature_importance: dict[str, float] = field(default_factory=dict)
    metrics: dict[str, float] = field(default_factory=dict)
    trained_at: str = ""


class ModelLayer:
    """Combines ML models with LLM-driven factor discovery.

    Training pipeline:
    1. Load signal space from Data Layer
    2. Run LLM factor discovery loop (Idea → Factor → Eval)
    3. Train gradient boosting + random forest on signal space + LLM factors
    4. Output ranked signal importance per regime
    """

    def __init__(
        self,
        parquet_store: ParquetStore,
        llm_router: LLMRouter | None = None,
        regime_detector: RegimeDetector | None = None,
    ) -> None:
        self._store = parquet_store
        self._llm_router = llm_router
        self._regime_detector = regime_detector or RegimeDetector()
        self._registry = FactorRegistry()
        self._models: dict[str, Any] = {}  # regime → trained model

        # LLM agents (only if router available)
        self._idea_agent = IdeaAgent(llm_router) if llm_router else None
        self._factor_agent = FactorAgent(llm_router) if llm_router else None

    async def train(
        self,
        signal_space: pd.DataFrame,
        forward_returns: pd.Series,
        regime: RegimeResult,
        n_hypotheses: int = 5,
    ) -> list[ModelArtifact]:
        """Train models on signal space with optional LLM factor augmentation.

        Args:
            signal_space: DataFrame from DataLayer (MultiIndex: date, symbol).
            forward_returns: Series of N-day forward returns aligned with signal_space.
            regime: Current market regime for regime-aware training.
            n_hypotheses: Number of LLM hypotheses to generate per training cycle.

        Returns:
            List of trained ModelArtifacts.
        """
        artifacts = []

        # Step 1: LLM factor discovery (if available)
        if self._idea_agent and self._factor_agent:
            await self._discover_factors(signal_space, regime, n_hypotheses)

        # Step 2: Augment signal space with discovered factors
        augmented = self._apply_discovered_factors(signal_space)

        # Step 3: Prepare training data
        X, y = self._prepare_training_data(augmented, forward_returns)
        if X.empty or len(y) < 100:
            logger.warning("Insufficient training data: %d samples", len(y))
            return artifacts

        # Step 4: Train gradient boosting (Req 10.1)
        gb_artifact = self._train_gradient_boosting(X, y, regime)
        artifacts.append(gb_artifact)

        # Step 5: Train random forest (Req 10.1)
        rf_artifact = self._train_random_forest(X, y, regime)
        artifacts.append(rf_artifact)

        # Step 6: Save models to S3
        for artifact in artifacts:
            self._save_model(artifact)

        # Step 7: Save factor registry
        self._save_registry()

        logger.info(
            "Training complete: %d models, %d factors, regime=%s",
            len(artifacts), len(self._registry.factors), regime.regime.value,
        )
        return artifacts

    def predict(self, signal_space: pd.DataFrame) -> pd.DataFrame:
        """Generate predictions from all trained models.

        Returns DataFrame with columns for each model's predictions.
        """
        augmented = self._apply_discovered_factors(signal_space)
        X = augmented.select_dtypes(include=[np.number]).dropna(axis=1, how="all")
        X = X.fillna(0)

        predictions = {}
        for name, model in self._models.items():
            try:
                cols = [c for c in model.feature_names_in_ if c in X.columns]
                if cols:
                    predictions[name] = model.predict(X[cols])
                else:
                    predictions[name] = np.zeros(len(X))
            except Exception as exc:
                logger.warning("Prediction failed for %s: %s", name, exc)
                predictions[name] = np.zeros(len(X))

        return pd.DataFrame(predictions, index=X.index)

    def get_feature_importance(self, regime: str = "") -> dict[str, float]:
        """Return ranked signal importance scores (Req 10.6)."""
        importance: dict[str, float] = {}
        for name, model in self._models.items():
            if regime and regime not in name:
                continue
            if hasattr(model, "feature_importances_"):
                for feat, imp in zip(model.feature_names_in_, model.feature_importances_):
                    importance[feat] = importance.get(feat, 0) + imp

        # Normalize
        total = sum(importance.values()) or 1.0
        return {k: v / total for k, v in sorted(importance.items(), key=lambda x: -x[1])}

    # ------------------------------------------------------------------
    # LLM Factor Discovery Loop
    # ------------------------------------------------------------------

    async def _discover_factors(
        self,
        signal_space: pd.DataFrame,
        regime: RegimeResult,
        n: int,
    ) -> None:
        """Run the Idea → Factor → (register) loop."""
        if not self._idea_agent or not self._factor_agent:
            return

        hypotheses = await self._idea_agent.generate_hypotheses(regime, self._registry, n)
        logger.info("LLM generated %d hypotheses", len(hypotheses))

        for hyp in hypotheses:
            factor = await self._factor_agent.construct_factor(
                hyp.get("hypothesis", ""),
                hyp.get("category", "price"),
            )
            if factor:
                added = self._registry.add(factor)
                if added:
                    logger.info("New factor discovered: %s", factor.name)

    def _apply_discovered_factors(self, signal_space: pd.DataFrame) -> pd.DataFrame:
        """Apply LLM-discovered factor expressions to the signal space."""
        result = signal_space.copy()
        for factor in self._registry.factors:
            if factor.expression and f"llm_{factor.name}" not in result.columns:
                try:
                    # Execute the pandas expression safely
                    val = eval(factor.expression, {"df": result, "np": np, "pd": pd})  # noqa: S307
                    if isinstance(val, pd.Series):
                        result[f"llm_{factor.name}"] = val
                except Exception as exc:
                    logger.debug("Factor %s expression failed: %s", factor.name, exc)
        return result

    # ------------------------------------------------------------------
    # ML Training
    # ------------------------------------------------------------------

    def _prepare_training_data(
        self,
        signal_space: pd.DataFrame,
        forward_returns: pd.Series,
    ) -> tuple[pd.DataFrame, pd.Series]:
        """Align features and target, drop NaN rows."""
        X = signal_space.select_dtypes(include=[np.number])
        y = forward_returns.reindex(X.index)

        mask = X.notna().all(axis=1) & y.notna()
        return X[mask], y[mask]

    def _train_gradient_boosting(
        self,
        X: pd.DataFrame,
        y: pd.Series,
        regime: RegimeResult,
    ) -> ModelArtifact:
        model = GradientBoostingRegressor(
            n_estimators=200, max_depth=5, learning_rate=0.05,
            subsample=0.8, random_state=42,
        )
        model.fit(X, y)

        name = f"gb_{regime.regime.value}"
        self._models[name] = model

        importance = dict(zip(X.columns, model.feature_importances_))
        top_features = dict(sorted(importance.items(), key=lambda x: -x[1])[:10])

        return ModelArtifact(
            name=name,
            regime=regime.regime.value,
            version=time.strftime("%Y%m%d"),
            feature_importance=top_features,
            metrics={"train_r2": float(model.score(X, y))},
            trained_at=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        )

    def _train_random_forest(
        self,
        X: pd.DataFrame,
        y: pd.Series,
        regime: RegimeResult,
    ) -> ModelArtifact:
        model = RandomForestRegressor(
            n_estimators=200, max_depth=8, min_samples_leaf=20,
            random_state=42, n_jobs=-1,
        )
        model.fit(X, y)

        name = f"rf_{regime.regime.value}"
        self._models[name] = model

        importance = dict(zip(X.columns, model.feature_importances_))
        top_features = dict(sorted(importance.items(), key=lambda x: -x[1])[:10])

        return ModelArtifact(
            name=name,
            regime=regime.regime.value,
            version=time.strftime("%Y%m%d"),
            feature_importance=top_features,
            metrics={"train_r2": float(model.score(X, y))},
            trained_at=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        )

    # ------------------------------------------------------------------
    # Persistence
    # ------------------------------------------------------------------

    def _save_model(self, artifact: ModelArtifact) -> None:
        key = f"{_MODEL_KEY_PREFIX}{artifact.name}_v{artifact.version}.pkl"
        model = self._models.get(artifact.name)
        if model:
            buf = io.BytesIO()
            pickle.dump(model, buf)
            self._store._s3.put_model(key, buf.getvalue())
            logger.info("Saved model: %s", key)

    def _save_registry(self) -> None:
        data = json.dumps(self._registry.to_dict(), indent=2)
        self._store._s3.put_parquet(_FACTOR_REGISTRY_KEY, data.encode())

    def load_registry(self) -> None:
        try:
            raw = self._store._s3.get_parquet(_FACTOR_REGISTRY_KEY)
            self._registry = FactorRegistry.from_dict(json.loads(raw))
            logger.info("Loaded factor registry: %d factors", len(self._registry.factors))
        except Exception:
            logger.debug("No existing factor registry found")
