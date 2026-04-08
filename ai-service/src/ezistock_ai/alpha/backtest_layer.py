"""Alpha Mining Engine — Backtest Layer.

Walk-forward analysis with regime-aware validation, transaction cost
simulation, signal stability analysis, and alpha decay detection.

Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field

import numpy as np
import pandas as pd

from ezistock_ai.alpha.regime_detector import MarketRegime

logger = logging.getLogger(__name__)

# Vietnamese market transaction costs (Req 11.4)
_COMMISSION_RATE = 0.002  # 0.15-0.25% per trade, use 0.2% average
_SLIPPAGE_BPS = 5  # 5 basis points slippage estimate
_TAX_SELL = 0.001  # 0.1% sell tax


@dataclass
class BacktestConfig:
    """Configuration for walk-forward backtesting."""

    train_window_days: int = 504  # ~2 years
    test_window_days: int = 63   # ~3 months
    step_days: int = 21          # ~1 month step
    commission_rate: float = _COMMISSION_RATE
    slippage_bps: float = _SLIPPAGE_BPS
    sell_tax: float = _TAX_SELL
    top_n_stocks: int = 20       # Long top N stocks
    rebalance_frequency: int = 5 # Rebalance every N trading days
    alpha_decay_threshold: float = 0.01  # IC below this = decayed


@dataclass
class BacktestMetrics:
    """Performance metrics for a backtest period (Req 11.5)."""

    cumulative_return: float = 0.0
    annualized_return: float = 0.0
    sharpe_ratio: float = 0.0
    max_drawdown: float = 0.0
    win_rate: float = 0.0
    profit_factor: float = 0.0
    information_ratio: float = 0.0  # vs VN-Index benchmark
    total_trades: int = 0
    total_cost: float = 0.0


@dataclass
class RegimeMetrics:
    """Per-regime backtest metrics (Req 11.2)."""

    regime: str
    metrics: BacktestMetrics
    n_periods: int = 0


@dataclass
class AlphaDecayResult:
    """Alpha decay detection output (Req 11.6)."""

    factor_name: str
    ic_series: list[float] = field(default_factory=list)
    ic_current: float = 0.0
    ic_peak: float = 0.0
    decay_detected: bool = False
    decay_rate: float = 0.0  # IC loss per month


@dataclass
class BacktestResult:
    """Full backtest output (Req 11.7)."""

    overall: BacktestMetrics = field(default_factory=BacktestMetrics)
    by_regime: list[RegimeMetrics] = field(default_factory=list)
    monthly_returns: list[float] = field(default_factory=list)
    yearly_returns: dict[int, float] = field(default_factory=dict)
    drawdown_series: list[float] = field(default_factory=list)
    alpha_decay: list[AlphaDecayResult] = field(default_factory=list)
    signal_stability: dict[str, float] = field(default_factory=dict)


class BacktestLayer:
    """Walk-forward backtesting with regime-aware validation."""

    def __init__(self, config: BacktestConfig | None = None) -> None:
        self._config = config or BacktestConfig()

    def run_walkforward(
        self,
        predictions: pd.DataFrame,
        forward_returns: pd.Series,
        regimes: pd.Series | None = None,
        benchmark_returns: pd.Series | None = None,
    ) -> BacktestResult:
        """Run walk-forward backtest (Req 11.1).

        Args:
            predictions: Model predictions (MultiIndex: date, symbol).
            forward_returns: Actual forward returns aligned with predictions.
            regimes: Optional Series of MarketRegime per date.
            benchmark_returns: VN-Index daily returns for information ratio.
        """
        result = BacktestResult()

        dates = predictions.index.get_level_values("date").unique().sort_values()
        if len(dates) < self._config.test_window_days:
            logger.warning("Insufficient dates for walk-forward: %d", len(dates))
            return result

        # Walk-forward windows
        all_portfolio_returns = []
        all_benchmark_returns = []

        step = self._config.step_days
        test_len = self._config.test_window_days

        for start_idx in range(0, len(dates) - test_len, step):
            test_dates = dates[start_idx:start_idx + test_len]

            # For each test date, pick top N stocks by prediction score
            period_returns = self._simulate_period(
                predictions, forward_returns, test_dates,
            )
            all_portfolio_returns.extend(period_returns)

            if benchmark_returns is not None:
                bench = benchmark_returns.reindex(test_dates).fillna(0).tolist()
                all_benchmark_returns.extend(bench)

        if not all_portfolio_returns:
            return result

        # Compute overall metrics (Req 11.5)
        ret_series = pd.Series(all_portfolio_returns)
        result.overall = self._compute_metrics(
            ret_series,
            pd.Series(all_benchmark_returns) if all_benchmark_returns else None,
        )

        # Monthly/yearly returns (Req 11.7)
        result.monthly_returns = self._aggregate_monthly(ret_series)
        result.drawdown_series = self._compute_drawdown_series(ret_series)

        # Regime-aware validation (Req 11.2)
        if regimes is not None:
            result.by_regime = self._compute_regime_metrics(
                predictions, forward_returns, regimes, benchmark_returns,
            )

        # Signal stability (Req 11.3)
        result.signal_stability = self._compute_signal_stability(predictions)

        # Alpha decay detection (Req 11.6)
        result.alpha_decay = self._detect_alpha_decay(predictions, forward_returns)

        return result

    # ------------------------------------------------------------------
    # Period simulation
    # ------------------------------------------------------------------

    def _simulate_period(
        self,
        predictions: pd.DataFrame,
        forward_returns: pd.Series,
        test_dates: pd.DatetimeIndex,
    ) -> list[float]:
        """Simulate a single test period with transaction costs."""
        period_returns = []
        cfg = self._config

        for date in test_dates:
            if date not in predictions.index.get_level_values("date"):
                continue

            day_preds = predictions.xs(date, level="date")
            if day_preds.empty:
                continue

            # Average across model columns, pick top N
            if isinstance(day_preds, pd.DataFrame):
                scores = day_preds.mean(axis=1)
            else:
                scores = day_preds

            top_symbols = scores.nlargest(cfg.top_n_stocks).index

            # Get actual returns for top stocks
            day_returns = forward_returns.xs(date, level="date") if date in forward_returns.index.get_level_values("date") else pd.Series(dtype=float)
            selected_returns = day_returns.reindex(top_symbols).fillna(0)

            # Equal-weight portfolio return minus costs
            if len(selected_returns) > 0:
                gross_return = selected_returns.mean()
                cost = cfg.commission_rate + cfg.slippage_bps / 10000 + cfg.sell_tax / 2
                net_return = gross_return - cost / cfg.rebalance_frequency
                period_returns.append(net_return)

        return period_returns

    # ------------------------------------------------------------------
    # Metrics computation
    # ------------------------------------------------------------------

    def _compute_metrics(
        self,
        returns: pd.Series,
        benchmark: pd.Series | None = None,
    ) -> BacktestMetrics:
        """Compute standard backtest metrics (Req 11.5)."""
        if returns.empty:
            return BacktestMetrics()

        cum_ret = (1 + returns).prod() - 1
        n_years = len(returns) / 252
        ann_ret = (1 + cum_ret) ** (1 / max(n_years, 0.01)) - 1 if cum_ret > -1 else -1.0

        vol = returns.std() * np.sqrt(252)
        sharpe = ann_ret / vol if vol > 0 else 0.0

        # Max drawdown
        cum = (1 + returns).cumprod()
        peak = cum.cummax()
        dd = (cum - peak) / peak
        max_dd = float(dd.min())

        # Win rate and profit factor
        wins = returns[returns > 0]
        losses = returns[returns < 0]
        win_rate = len(wins) / len(returns) if len(returns) > 0 else 0.0
        profit_factor = wins.sum() / abs(losses.sum()) if len(losses) > 0 and losses.sum() != 0 else 0.0

        # Information ratio vs benchmark
        ir = 0.0
        if benchmark is not None and len(benchmark) == len(returns):
            excess = returns.values - benchmark.values
            ir = float(np.mean(excess) / (np.std(excess) + 1e-10) * np.sqrt(252))

        return BacktestMetrics(
            cumulative_return=float(cum_ret),
            annualized_return=float(ann_ret),
            sharpe_ratio=float(sharpe),
            max_drawdown=float(max_dd),
            win_rate=float(win_rate),
            profit_factor=float(profit_factor),
            information_ratio=ir,
            total_trades=len(returns),
        )

    def _compute_regime_metrics(
        self,
        predictions: pd.DataFrame,
        forward_returns: pd.Series,
        regimes: pd.Series,
        benchmark: pd.Series | None,
    ) -> list[RegimeMetrics]:
        """Compute metrics separately per regime (Req 11.2)."""
        results = []
        for regime_val in regimes.unique():
            regime_dates = regimes[regimes == regime_val].index
            mask = predictions.index.get_level_values("date").isin(regime_dates)
            if mask.sum() == 0:
                continue

            period_returns = self._simulate_period(
                predictions[mask], forward_returns, regime_dates,
            )
            if period_returns:
                metrics = self._compute_metrics(pd.Series(period_returns), None)
                results.append(RegimeMetrics(
                    regime=str(regime_val),
                    metrics=metrics,
                    n_periods=len(period_returns),
                ))
        return results

    # ------------------------------------------------------------------
    # Signal stability & alpha decay
    # ------------------------------------------------------------------

    def _compute_signal_stability(self, predictions: pd.DataFrame) -> dict[str, float]:
        """Measure consistency of signal rankings across time (Req 11.3)."""
        stability = {}
        dates = predictions.index.get_level_values("date").unique()

        if len(dates) < 10 or not isinstance(predictions, pd.DataFrame):
            return stability

        for col in predictions.columns:
            # Rank correlation between consecutive periods
            rank_corrs = []
            for i in range(1, min(len(dates), 20)):
                try:
                    prev = predictions.xs(dates[i-1], level="date")[col].rank()
                    curr = predictions.xs(dates[i], level="date")[col].rank()
                    common = prev.index.intersection(curr.index)
                    if len(common) > 10:
                        corr = prev[common].corr(curr[common])
                        if not np.isnan(corr):
                            rank_corrs.append(corr)
                except (KeyError, ValueError):
                    continue

            if rank_corrs:
                stability[col] = float(np.mean(rank_corrs))

        return stability

    def _detect_alpha_decay(
        self,
        predictions: pd.DataFrame,
        forward_returns: pd.Series,
    ) -> list[AlphaDecayResult]:
        """Detect alpha decay by monitoring IC over time (Req 11.6)."""
        results = []
        dates = predictions.index.get_level_values("date").unique().sort_values()

        if len(dates) < 40 or not isinstance(predictions, pd.DataFrame):
            return results

        # Compute rolling IC (20-day windows) for each model
        window = 20
        for col in predictions.columns:
            ic_series = []
            for i in range(window, len(dates)):
                window_dates = dates[i-window:i]
                try:
                    pred_window = predictions.loc[predictions.index.get_level_values("date").isin(window_dates), col]
                    ret_window = forward_returns.reindex(pred_window.index)
                    mask = pred_window.notna() & ret_window.notna()
                    if mask.sum() > 10:
                        ic = float(pred_window[mask].corr(ret_window[mask]))
                        ic_series.append(ic if not np.isnan(ic) else 0.0)
                except Exception:
                    ic_series.append(0.0)

            if len(ic_series) < 2:
                continue

            ic_current = ic_series[-1] if ic_series else 0.0
            ic_peak = max(abs(x) for x in ic_series) if ic_series else 0.0

            # Decay rate: slope of IC over last 10 windows
            recent = ic_series[-10:] if len(ic_series) >= 10 else ic_series
            decay_rate = 0.0
            if len(recent) >= 2:
                x = np.arange(len(recent))
                slope = np.polyfit(x, recent, 1)[0]
                decay_rate = float(slope)

            decayed = abs(ic_current) < self._config.alpha_decay_threshold and ic_peak > self._config.alpha_decay_threshold * 2

            results.append(AlphaDecayResult(
                factor_name=col,
                ic_series=ic_series[-20:],
                ic_current=ic_current,
                ic_peak=ic_peak,
                decay_detected=decayed,
                decay_rate=decay_rate,
            ))

        return results

    # ------------------------------------------------------------------
    # Helpers
    # ------------------------------------------------------------------

    @staticmethod
    def _aggregate_monthly(returns: pd.Series) -> list[float]:
        """Aggregate daily returns into monthly returns."""
        if returns.empty:
            return []
        # Simple chunking by ~21 trading days
        monthly = []
        for i in range(0, len(returns), 21):
            chunk = returns.iloc[i:i+21]
            monthly.append(float((1 + chunk).prod() - 1))
        return monthly

    @staticmethod
    def _compute_drawdown_series(returns: pd.Series) -> list[float]:
        cum = (1 + returns).cumprod()
        peak = cum.cummax()
        dd = (cum - peak) / peak.replace(0, 1)
        return dd.tolist()
