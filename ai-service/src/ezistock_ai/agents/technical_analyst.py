"""Technical Analyst Agent — computes indicators, patterns, and signals.

The quantitative engine underneath is the real differentiator; the LLM is the
storytelling layer.  This module pre-computes ALL indicators from raw OHLCV
data using numpy/pandas, then optionally feeds the structured results to an
LLM for narrative generation.

Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field
from enum import Enum
from typing import Tuple

import numpy as np
import pandas as pd

from ezistock_ai.generated.proto.agent_pb2 import OHLCV, MarketData, TechnicalAnalysis

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Enums & data classes for internal results
# ---------------------------------------------------------------------------


class CompositeSignal(str, Enum):
    STRONGLY_BULLISH = "strongly_bullish"
    BULLISH = "bullish"
    NEUTRAL = "neutral"
    BEARISH = "bearish"
    STRONGLY_BEARISH = "strongly_bearish"


class MoneyFlowClassification(str, Enum):
    STRONG_INFLOW = "strong_inflow"
    MODERATE_INFLOW = "moderate_inflow"
    NEUTRAL = "neutral"
    MODERATE_OUTFLOW = "moderate_outflow"
    STRONG_OUTFLOW = "strong_outflow"


@dataclass
class IndicatorResult:
    """Container for a single computed indicator value + signal direction."""

    name: str
    value: float
    signal: str  # "bullish", "bearish", "neutral"
    detail: str = ""


@dataclass
class SupportResistance:
    level: float
    kind: str  # "support" or "resistance"
    source: str  # "pivot", "volume_profile"


@dataclass
class CandlestickPattern:
    name: str
    direction: str  # "bullish", "bearish", "neutral"
    bar_index: int  # index in the OHLCV series where detected


@dataclass
class MACrossover:
    fast: str
    slow: str
    direction: str  # "golden_cross" or "death_cross"
    bar_index: int


@dataclass
class Divergence:
    indicator: str
    kind: str  # "bullish_divergence" or "bearish_divergence"
    detail: str


@dataclass
class TechnicalResult:
    """Full output of the technical analysis pipeline."""

    symbol: str
    indicators: list[IndicatorResult] = field(default_factory=list)
    support_levels: list[SupportResistance] = field(default_factory=list)
    resistance_levels: list[SupportResistance] = field(default_factory=list)
    patterns: list[CandlestickPattern] = field(default_factory=list)
    crossovers: list[MACrossover] = field(default_factory=list)
    divergences: list[Divergence] = field(default_factory=list)
    smart_money_flow: MoneyFlowClassification = MoneyFlowClassification.NEUTRAL
    composite_signal: CompositeSignal = CompositeSignal.NEUTRAL
    bullish_count: int = 0
    bearish_count: int = 0
    neutral_count: int = 0
    overbought: bool = False
    oversold: bool = False
    ob_os_details: list[str] = field(default_factory=list)


# ---------------------------------------------------------------------------
# Helper: OHLCV list → pandas DataFrame
# ---------------------------------------------------------------------------


def _to_dataframe(ohlcv: list[OHLCV]) -> pd.DataFrame:
    """Convert protobuf OHLCV list to a pandas DataFrame sorted by date ascending."""
    rows = [
        {
            "date": bar.date,
            "open": bar.open,
            "high": bar.high,
            "low": bar.low,
            "close": bar.close,
            "volume": bar.volume,
        }
        for bar in ohlcv
    ]
    df = pd.DataFrame(rows)
    if df.empty:
        return df
    df["date"] = pd.to_datetime(df["date"], errors="coerce")
    df.sort_values("date", inplace=True)
    df.reset_index(drop=True, inplace=True)
    return df


# ---------------------------------------------------------------------------
# Indicator computation functions (pure numpy/pandas, no LLM)
# ---------------------------------------------------------------------------


def _sma(series: pd.Series, period: int) -> pd.Series:
    return series.rolling(window=period, min_periods=period).mean()


def _ema(series: pd.Series, period: int) -> pd.Series:
    return series.ewm(span=period, adjust=False).mean()


def _rsi(close: pd.Series, period: int = 14) -> pd.Series:
    delta = close.diff()
    gain = delta.clip(lower=0)
    loss = -delta.clip(upper=0)
    avg_gain = gain.ewm(alpha=1 / period, min_periods=period).mean()
    avg_loss = loss.ewm(alpha=1 / period, min_periods=period).mean()
    rs = avg_gain / avg_loss.replace(0, np.nan)
    return 100 - (100 / (1 + rs))


def _macd(
    close: pd.Series,
    fast: int = 12,
    slow: int = 26,
    signal: int = 9,
) -> Tuple[pd.Series, pd.Series, pd.Series]:
    """Return (macd_line, signal_line, histogram)."""
    ema_fast = _ema(close, fast)
    ema_slow = _ema(close, slow)
    macd_line = ema_fast - ema_slow
    signal_line = _ema(macd_line, signal)
    histogram = macd_line - signal_line
    return macd_line, signal_line, histogram


def _bollinger_bands(
    close: pd.Series,
    period: int = 20,
    std_dev: float = 2.0,
) -> Tuple[pd.Series, pd.Series, pd.Series]:
    """Return (upper, middle, lower)."""
    middle = _sma(close, period)
    std = close.rolling(window=period, min_periods=period).std()
    upper = middle + std_dev * std
    lower = middle - std_dev * std
    return upper, middle, lower


def _adx(df: pd.DataFrame, period: int = 14) -> pd.Series:
    high, low, close = df["high"], df["low"], df["close"]
    plus_dm = high.diff().clip(lower=0)
    minus_dm = (-low.diff()).clip(lower=0)
    # Zero out when the other is larger
    plus_dm[plus_dm < minus_dm] = 0
    minus_dm[minus_dm < plus_dm] = 0
    tr = _true_range(df)
    atr = tr.ewm(alpha=1 / period, min_periods=period).mean()
    plus_di = 100 * (plus_dm.ewm(alpha=1 / period, min_periods=period).mean() / atr.replace(0, np.nan))
    minus_di = 100 * (minus_dm.ewm(alpha=1 / period, min_periods=period).mean() / atr.replace(0, np.nan))
    dx = 100 * (plus_di - minus_di).abs() / (plus_di + minus_di).replace(0, np.nan)
    adx = dx.ewm(alpha=1 / period, min_periods=period).mean()
    return adx


def _stochastic(
    df: pd.DataFrame,
    k_period: int = 14,
    d_period: int = 3,
    smooth: int = 3,
) -> Tuple[pd.Series, pd.Series]:
    """Return (%K, %D)."""
    low_min = df["low"].rolling(window=k_period, min_periods=k_period).min()
    high_max = df["high"].rolling(window=k_period, min_periods=k_period).max()
    fast_k = 100 * (df["close"] - low_min) / (high_max - low_min).replace(0, np.nan)
    k = fast_k.rolling(window=smooth, min_periods=1).mean()
    d = k.rolling(window=d_period, min_periods=1).mean()
    return k, d


def _true_range(df: pd.DataFrame) -> pd.Series:
    high, low, close = df["high"], df["low"], df["close"]
    prev_close = close.shift(1)
    tr1 = high - low
    tr2 = (high - prev_close).abs()
    tr3 = (low - prev_close).abs()
    return pd.concat([tr1, tr2, tr3], axis=1).max(axis=1)


def _atr(df: pd.DataFrame, period: int = 14) -> pd.Series:
    tr = _true_range(df)
    return tr.ewm(alpha=1 / period, min_periods=period).mean()


def _obv(df: pd.DataFrame) -> pd.Series:
    direction = np.sign(df["close"].diff())
    direction.iloc[0] = 0
    return (direction * df["volume"]).cumsum()


def _mfi(df: pd.DataFrame, period: int = 14) -> pd.Series:
    typical_price = (df["high"] + df["low"] + df["close"]) / 3
    raw_money_flow = typical_price * df["volume"]
    direction = typical_price.diff()
    pos_flow = raw_money_flow.where(direction > 0, 0).rolling(window=period, min_periods=period).sum()
    neg_flow = raw_money_flow.where(direction <= 0, 0).rolling(window=period, min_periods=period).sum()
    mfi = 100 - (100 / (1 + pos_flow / neg_flow.replace(0, np.nan)))
    return mfi


def _aroon(df: pd.DataFrame, period: int = 25) -> Tuple[pd.Series, pd.Series]:
    """Return (aroon_up, aroon_down)."""
    aroon_up = df["high"].rolling(window=period + 1, min_periods=period + 1).apply(
        lambda x: x.argmax() / period * 100, raw=True
    )
    aroon_down = df["low"].rolling(window=period + 1, min_periods=period + 1).apply(
        lambda x: x.argmin() / period * 100, raw=True
    )
    return aroon_up, aroon_down


def _parabolic_sar(
    df: pd.DataFrame,
    af_start: float = 0.02,
    af_max: float = 0.2,
) -> pd.Series:
    """Simplified Parabolic SAR."""
    high, low, close = df["high"].values, df["low"].values, df["close"].values
    n = len(close)
    sar = np.zeros(n)
    af = af_start
    uptrend = True
    ep = high[0]
    sar[0] = low[0]

    for i in range(1, n):
        if uptrend:
            sar[i] = sar[i - 1] + af * (ep - sar[i - 1])
            sar[i] = min(sar[i], low[i - 1], low[max(0, i - 2)])
            if low[i] < sar[i]:
                uptrend = False
                sar[i] = ep
                ep = low[i]
                af = af_start
            else:
                if high[i] > ep:
                    ep = high[i]
                    af = min(af + af_start, af_max)
        else:
            sar[i] = sar[i - 1] + af * (ep - sar[i - 1])
            sar[i] = max(sar[i], high[i - 1], high[max(0, i - 2)])
            if high[i] > sar[i]:
                uptrend = True
                sar[i] = ep
                ep = high[i]
                af = af_start
            else:
                if low[i] < ep:
                    ep = low[i]
                    af = min(af + af_start, af_max)

    return pd.Series(sar, index=df.index)


def _supertrend(df: pd.DataFrame, period: int = 10, multiplier: float = 3.0) -> pd.Series:
    """Return Supertrend values."""
    atr_vals = _atr(df, period)
    hl2 = (df["high"] + df["low"]) / 2
    upper_band = hl2 + multiplier * atr_vals
    lower_band = hl2 - multiplier * atr_vals

    supertrend = pd.Series(np.nan, index=df.index)
    direction = pd.Series(1, index=df.index)  # 1 = up, -1 = down

    for i in range(period, len(df)):
        if pd.isna(upper_band.iloc[i]):
            continue
        if i == period:
            supertrend.iloc[i] = lower_band.iloc[i]
            direction.iloc[i] = 1
            continue

        prev_dir = direction.iloc[i - 1]
        if prev_dir == 1:
            lower_band.iloc[i] = max(lower_band.iloc[i], lower_band.iloc[i - 1]) if not pd.isna(lower_band.iloc[i - 1]) else lower_band.iloc[i]
            if df["close"].iloc[i] < lower_band.iloc[i]:
                direction.iloc[i] = -1
                supertrend.iloc[i] = upper_band.iloc[i]
            else:
                direction.iloc[i] = 1
                supertrend.iloc[i] = lower_band.iloc[i]
        else:
            upper_band.iloc[i] = min(upper_band.iloc[i], upper_band.iloc[i - 1]) if not pd.isna(upper_band.iloc[i - 1]) else upper_band.iloc[i]
            if df["close"].iloc[i] > upper_band.iloc[i]:
                direction.iloc[i] = 1
                supertrend.iloc[i] = lower_band.iloc[i]
            else:
                direction.iloc[i] = -1
                supertrend.iloc[i] = upper_band.iloc[i]

    return supertrend


def _williams_r(df: pd.DataFrame, period: int = 14) -> pd.Series:
    high_max = df["high"].rolling(window=period, min_periods=period).max()
    low_min = df["low"].rolling(window=period, min_periods=period).min()
    return -100 * (high_max - df["close"]) / (high_max - low_min).replace(0, np.nan)


def _vwap(df: pd.DataFrame) -> pd.Series:
    typical_price = (df["high"] + df["low"] + df["close"]) / 3
    cum_tp_vol = (typical_price * df["volume"]).cumsum()
    cum_vol = df["volume"].cumsum()
    return cum_tp_vol / cum_vol.replace(0, np.nan)


def _roc(close: pd.Series, period: int = 12) -> pd.Series:
    return close.pct_change(periods=period) * 100


# ---------------------------------------------------------------------------
# Support / Resistance detection
# ---------------------------------------------------------------------------


def _pivot_support_resistance(df: pd.DataFrame, lookback: int = 5) -> list[SupportResistance]:
    """Detect support/resistance via pivot points (local min/max)."""
    results: list[SupportResistance] = []
    if len(df) < lookback * 2 + 1:
        return results

    high, low = df["high"].values, df["low"].values
    for i in range(lookback, len(df) - lookback):
        # Local high → resistance
        if high[i] == max(high[i - lookback : i + lookback + 1]):
            results.append(SupportResistance(level=float(high[i]), kind="resistance", source="pivot"))
        # Local low → support
        if low[i] == min(low[i - lookback : i + lookback + 1]):
            results.append(SupportResistance(level=float(low[i]), kind="support", source="pivot"))

    return results


def _volume_profile_levels(df: pd.DataFrame, bins: int = 20) -> list[SupportResistance]:
    """Detect support/resistance from volume profile (high-volume price zones)."""
    results: list[SupportResistance] = []
    if df.empty:
        return results

    price_range = df["close"].max() - df["close"].min()
    if price_range <= 0:
        return results

    bin_edges = np.linspace(df["close"].min(), df["close"].max(), bins + 1)
    vol_by_bin = np.zeros(bins)
    for i in range(len(df)):
        idx = np.searchsorted(bin_edges[1:], df["close"].iloc[i])
        idx = min(idx, bins - 1)
        vol_by_bin[idx] += df["volume"].iloc[i]

    # Top 3 volume bins → support/resistance
    threshold = np.percentile(vol_by_bin, 75)
    current_price = df["close"].iloc[-1]
    for i in range(bins):
        if vol_by_bin[i] >= threshold:
            level = (bin_edges[i] + bin_edges[i + 1]) / 2
            kind = "support" if level < current_price else "resistance"
            results.append(SupportResistance(level=float(level), kind=kind, source="volume_profile"))

    return results


# ---------------------------------------------------------------------------
# Candlestick pattern detection
# ---------------------------------------------------------------------------


def _detect_candlestick_patterns(df: pd.DataFrame) -> list[CandlestickPattern]:
    """Detect common candlestick patterns on the last few bars."""
    patterns: list[CandlestickPattern] = []
    if len(df) < 3:
        return patterns

    o, h, l, c = df["open"].values, df["high"].values, df["low"].values, df["close"].values
    n = len(df)

    def body(i: int) -> float:
        return abs(c[i] - o[i])

    def upper_shadow(i: int) -> float:
        return h[i] - max(o[i], c[i])

    def lower_shadow(i: int) -> float:
        return min(o[i], c[i]) - l[i]

    def is_bullish(i: int) -> bool:
        return c[i] > o[i]

    def is_bearish(i: int) -> bool:
        return c[i] < o[i]

    avg_body = np.mean([body(i) for i in range(max(0, n - 20), n)])
    if avg_body == 0:
        return patterns

    # Scan last 10 bars for patterns
    scan_start = max(2, n - 10)

    for i in range(scan_start, n):
        b = body(i)
        us = upper_shadow(i)
        ls = lower_shadow(i)

        # Doji: very small body relative to range
        if b < avg_body * 0.1 and (us + ls) > 0:
            patterns.append(CandlestickPattern(name="doji", direction="neutral", bar_index=i))

        # Hammer: small body at top, long lower shadow (bullish reversal)
        if ls > 2 * b and us < b and is_bullish(i) and b > 0:
            patterns.append(CandlestickPattern(name="hammer", direction="bullish", bar_index=i))

        # Bullish engulfing
        if i >= 1 and is_bearish(i - 1) and is_bullish(i):
            if o[i] <= c[i - 1] and c[i] >= o[i - 1]:
                patterns.append(CandlestickPattern(name="bullish_engulfing", direction="bullish", bar_index=i))

        # Bearish engulfing
        if i >= 1 and is_bullish(i - 1) and is_bearish(i):
            if o[i] >= c[i - 1] and c[i] <= o[i - 1]:
                patterns.append(CandlestickPattern(name="bearish_engulfing", direction="bearish", bar_index=i))

        # Morning star (3-bar bullish reversal)
        if i >= 2 and is_bearish(i - 2) and body(i - 2) > avg_body:
            if body(i - 1) < avg_body * 0.3:  # small middle body
                if is_bullish(i) and c[i] > (o[i - 2] + c[i - 2]) / 2:
                    patterns.append(CandlestickPattern(name="morning_star", direction="bullish", bar_index=i))

        # Evening star (3-bar bearish reversal)
        if i >= 2 and is_bullish(i - 2) and body(i - 2) > avg_body:
            if body(i - 1) < avg_body * 0.3:
                if is_bearish(i) and c[i] < (o[i - 2] + c[i - 2]) / 2:
                    patterns.append(CandlestickPattern(name="evening_star", direction="bearish", bar_index=i))

    # Three white soldiers / three black crows (check last 3 bars)
    if n >= 3:
        last3 = range(n - 3, n)
        if all(is_bullish(i) and body(i) > avg_body * 0.5 for i in last3):
            if c[n - 3] < c[n - 2] < c[n - 1]:
                patterns.append(CandlestickPattern(name="three_white_soldiers", direction="bullish", bar_index=n - 1))
        if all(is_bearish(i) and body(i) > avg_body * 0.5 for i in last3):
            if c[n - 3] > c[n - 2] > c[n - 1]:
                patterns.append(CandlestickPattern(name="three_black_crows", direction="bearish", bar_index=n - 1))

    return patterns


# ---------------------------------------------------------------------------
# MA crossover detection
# ---------------------------------------------------------------------------


def _detect_ma_crossovers(df: pd.DataFrame, lookback: int = 5) -> list[MACrossover]:
    """Detect golden cross / death cross in recent bars."""
    crossovers: list[MACrossover] = []
    if len(df) < 200:
        return crossovers

    close = df["close"]
    ma_pairs = [
        ("SMA50", "SMA200", _sma(close, 50), _sma(close, 200)),
        ("SMA20", "SMA50", _sma(close, 20), _sma(close, 50)),
        ("EMA12", "EMA26", _ema(close, 12), _ema(close, 26)),
    ]

    scan_start = max(0, len(df) - lookback)
    for fast_name, slow_name, fast_ma, slow_ma in ma_pairs:
        for i in range(scan_start + 1, len(df)):
            if pd.isna(fast_ma.iloc[i]) or pd.isna(slow_ma.iloc[i]):
                continue
            if pd.isna(fast_ma.iloc[i - 1]) or pd.isna(slow_ma.iloc[i - 1]):
                continue
            # Golden cross: fast crosses above slow
            if fast_ma.iloc[i - 1] <= slow_ma.iloc[i - 1] and fast_ma.iloc[i] > slow_ma.iloc[i]:
                crossovers.append(MACrossover(fast=fast_name, slow=slow_name, direction="golden_cross", bar_index=i))
            # Death cross: fast crosses below slow
            elif fast_ma.iloc[i - 1] >= slow_ma.iloc[i - 1] and fast_ma.iloc[i] < slow_ma.iloc[i]:
                crossovers.append(MACrossover(fast=fast_name, slow=slow_name, direction="death_cross", bar_index=i))

    return crossovers


# ---------------------------------------------------------------------------
# Overbought / Oversold classification & divergence detection
# ---------------------------------------------------------------------------


def _classify_ob_os(
    rsi_val: float,
    mfi_val: float,
    stoch_k: float,
    williams_r: float,
) -> Tuple[bool, bool, list[str]]:
    """Return (overbought, oversold, detail_strings)."""
    details: list[str] = []
    ob_count = 0
    os_count = 0

    if not np.isnan(rsi_val):
        if rsi_val > 70:
            ob_count += 1
            details.append(f"RSI={rsi_val:.1f} > 70 (overbought)")
        elif rsi_val < 30:
            os_count += 1
            details.append(f"RSI={rsi_val:.1f} < 30 (oversold)")

    if not np.isnan(mfi_val):
        if mfi_val > 80:
            ob_count += 1
            details.append(f"MFI={mfi_val:.1f} > 80 (overbought)")
        elif mfi_val < 20:
            os_count += 1
            details.append(f"MFI={mfi_val:.1f} < 20 (oversold)")

    if not np.isnan(stoch_k):
        if stoch_k > 80:
            ob_count += 1
            details.append(f"Stochastic %K={stoch_k:.1f} > 80 (overbought)")
        elif stoch_k < 20:
            os_count += 1
            details.append(f"Stochastic %K={stoch_k:.1f} < 20 (oversold)")

    if not np.isnan(williams_r):
        if williams_r > -20:
            ob_count += 1
            details.append(f"Williams %R={williams_r:.1f} > -20 (overbought)")
        elif williams_r < -80:
            os_count += 1
            details.append(f"Williams %R={williams_r:.1f} < -80 (oversold)")

    return ob_count >= 2, os_count >= 2, details


def _detect_divergences(
    close: pd.Series,
    rsi: pd.Series,
    macd_line: pd.Series,
    lookback: int = 20,
) -> list[Divergence]:
    """Detect bullish/bearish divergences between price and RSI/MACD."""
    divergences: list[Divergence] = []
    if len(close) < lookback * 2:
        return divergences

    n = len(close)
    recent = slice(n - lookback, n)
    prior = slice(n - lookback * 2, n - lookback)

    # Price vs RSI
    if not rsi.iloc[recent].isna().all() and not rsi.iloc[prior].isna().all():
        price_recent_low = close.iloc[recent].min()
        price_prior_low = close.iloc[prior].min()
        rsi_recent_low = rsi.iloc[recent].min()
        rsi_prior_low = rsi.iloc[prior].min()

        # Bullish divergence: price makes lower low, RSI makes higher low
        if price_recent_low < price_prior_low and rsi_recent_low > rsi_prior_low:
            divergences.append(Divergence(
                indicator="RSI",
                kind="bullish_divergence",
                detail=f"Price lower low ({price_recent_low:.2f} < {price_prior_low:.2f}) but RSI higher low ({rsi_recent_low:.1f} > {rsi_prior_low:.1f})",
            ))

        price_recent_high = close.iloc[recent].max()
        price_prior_high = close.iloc[prior].max()
        rsi_recent_high = rsi.iloc[recent].max()
        rsi_prior_high = rsi.iloc[prior].max()

        # Bearish divergence: price makes higher high, RSI makes lower high
        if price_recent_high > price_prior_high and rsi_recent_high < rsi_prior_high:
            divergences.append(Divergence(
                indicator="RSI",
                kind="bearish_divergence",
                detail=f"Price higher high ({price_recent_high:.2f} > {price_prior_high:.2f}) but RSI lower high ({rsi_recent_high:.1f} < {rsi_prior_high:.1f})",
            ))

    # Price vs MACD
    if not macd_line.iloc[recent].isna().all() and not macd_line.iloc[prior].isna().all():
        price_recent_low = close.iloc[recent].min()
        price_prior_low = close.iloc[prior].min()
        macd_recent_low = macd_line.iloc[recent].min()
        macd_prior_low = macd_line.iloc[prior].min()

        if price_recent_low < price_prior_low and macd_recent_low > macd_prior_low:
            divergences.append(Divergence(
                indicator="MACD",
                kind="bullish_divergence",
                detail=f"Price lower low but MACD higher low",
            ))

        price_recent_high = close.iloc[recent].max()
        price_prior_high = close.iloc[prior].max()
        macd_recent_high = macd_line.iloc[recent].max()
        macd_prior_high = macd_line.iloc[prior].max()

        if price_recent_high > price_prior_high and macd_recent_high < macd_prior_high:
            divergences.append(Divergence(
                indicator="MACD",
                kind="bearish_divergence",
                detail=f"Price higher high but MACD lower high",
            ))

    return divergences


# ---------------------------------------------------------------------------
# Smart Money Flow
# ---------------------------------------------------------------------------


def _compute_smart_money_flow(
    foreign_net_volume: float = 0.0,
    institutional_net_volume: float = 0.0,
    total_volume: float = 1.0,
) -> MoneyFlowClassification:
    """Classify smart money flow from foreign + institutional net volumes.

    In the absence of foreign/institutional breakdown data, this returns NEUTRAL.
    The Go backend provides these values when available from KBS data.
    """
    if total_volume <= 0:
        return MoneyFlowClassification.NEUTRAL

    net_ratio = (foreign_net_volume + institutional_net_volume) / total_volume

    if net_ratio > 0.05:
        return MoneyFlowClassification.STRONG_INFLOW
    elif net_ratio > 0.02:
        return MoneyFlowClassification.MODERATE_INFLOW
    elif net_ratio < -0.05:
        return MoneyFlowClassification.STRONG_OUTFLOW
    elif net_ratio < -0.02:
        return MoneyFlowClassification.MODERATE_OUTFLOW
    return MoneyFlowClassification.NEUTRAL


# ---------------------------------------------------------------------------
# Composite signal aggregation
# ---------------------------------------------------------------------------


def _aggregate_composite_signal(indicators: list[IndicatorResult]) -> Tuple[CompositeSignal, int, int, int]:
    """Count bullish/bearish/neutral signals and return composite."""
    bullish = sum(1 for ind in indicators if ind.signal == "bullish")
    bearish = sum(1 for ind in indicators if ind.signal == "bearish")
    neutral = sum(1 for ind in indicators if ind.signal == "neutral")
    total = bullish + bearish + neutral
    if total == 0:
        return CompositeSignal.NEUTRAL, 0, 0, 0

    bull_ratio = bullish / total
    bear_ratio = bearish / total

    if bull_ratio >= 0.7:
        sig = CompositeSignal.STRONGLY_BULLISH
    elif bull_ratio >= 0.5:
        sig = CompositeSignal.BULLISH
    elif bear_ratio >= 0.7:
        sig = CompositeSignal.STRONGLY_BEARISH
    elif bear_ratio >= 0.5:
        sig = CompositeSignal.BEARISH
    else:
        sig = CompositeSignal.NEUTRAL

    return sig, bullish, bearish, neutral


# ---------------------------------------------------------------------------
# Main analysis pipeline
# ---------------------------------------------------------------------------


class TechnicalAnalystAgent:
    """Computes all technical indicators and patterns from OHLCV data.

    This is the pure-computation engine. It does NOT call an LLM.
    The orchestrator may optionally pass the TechnicalResult to an LLM
    for narrative generation using the prompt templates.
    """

    def analyze(
        self,
        market_data: MarketData,
        foreign_net_volume: float = 0.0,
        institutional_net_volume: float = 0.0,
    ) -> TechnicalResult:
        """Run the full technical analysis pipeline on OHLCV data.

        Args:
            market_data: Protobuf MarketData with symbol and ohlcv list.
            foreign_net_volume: Net foreign buy-sell volume (from KBS data).
            institutional_net_volume: Net institutional buy-sell volume.

        Returns:
            TechnicalResult with all computed indicators, patterns, and signals.
        """
        symbol = market_data.symbol
        df = _to_dataframe(market_data.ohlcv)

        if df.empty or len(df) < 2:
            logger.warning("Insufficient OHLCV data for %s (%d bars)", symbol, len(df))
            return TechnicalResult(symbol=symbol)

        close = df["close"]
        indicators: list[IndicatorResult] = []

        # --- Compute all indicators (Req 5.1) ---
        rsi_series = _rsi(close, 14)
        rsi_val = _safe_last(rsi_series)
        indicators.append(_signal_from_rsi(rsi_val))

        macd_line, macd_signal, macd_hist = _macd(close, 12, 26, 9)
        indicators.append(_signal_from_macd(_safe_last(macd_line), _safe_last(macd_signal), _safe_last(macd_hist)))

        bb_upper, bb_middle, bb_lower = _bollinger_bands(close, 20, 2.0)
        indicators.append(_signal_from_bb(close.iloc[-1], _safe_last(bb_upper), _safe_last(bb_middle), _safe_last(bb_lower)))

        for period in (20, 50, 200):
            sma_val = _safe_last(_sma(close, period))
            indicators.append(_signal_from_ma(f"SMA{period}", close.iloc[-1], sma_val))

        for period in (12, 26):
            ema_val = _safe_last(_ema(close, period))
            indicators.append(_signal_from_ma(f"EMA{period}", close.iloc[-1], ema_val))

        adx_val = _safe_last(_adx(df, 14))
        indicators.append(IndicatorResult(
            name="ADX(14)", value=adx_val,
            signal="bullish" if adx_val > 25 else "neutral",
            detail=f"ADX={adx_val:.1f} — {'strong trend' if adx_val > 25 else 'weak/no trend'}",
        ))

        stoch_k, stoch_d = _stochastic(df, 14, 3, 3)
        stoch_k_val = _safe_last(stoch_k)
        indicators.append(_signal_from_stochastic(stoch_k_val, _safe_last(stoch_d)))

        atr_val = _safe_last(_atr(df, 14))
        indicators.append(IndicatorResult(name="ATR(14)", value=atr_val, signal="neutral", detail=f"ATR={atr_val:.2f}"))

        obv_series = _obv(df)
        obv_val = _safe_last(obv_series)
        obv_trend = "bullish" if len(obv_series) > 5 and obv_series.iloc[-1] > obv_series.iloc[-5] else "bearish" if len(obv_series) > 5 and obv_series.iloc[-1] < obv_series.iloc[-5] else "neutral"
        indicators.append(IndicatorResult(name="OBV", value=obv_val, signal=obv_trend, detail=f"OBV={'rising' if obv_trend == 'bullish' else 'falling' if obv_trend == 'bearish' else 'flat'}"))

        mfi_val = _safe_last(_mfi(df, 14))
        indicators.append(_signal_from_mfi(mfi_val))

        aroon_up, aroon_down = _aroon(df, 25)
        aroon_up_val, aroon_down_val = _safe_last(aroon_up), _safe_last(aroon_down)
        aroon_sig = "bullish" if aroon_up_val > aroon_down_val else "bearish" if aroon_down_val > aroon_up_val else "neutral"
        indicators.append(IndicatorResult(name="Aroon(25)", value=aroon_up_val, signal=aroon_sig, detail=f"Aroon Up={aroon_up_val:.0f}, Down={aroon_down_val:.0f}"))

        psar_series = _parabolic_sar(df, 0.02, 0.2)
        psar_val = _safe_last(psar_series)
        psar_sig = "bullish" if close.iloc[-1] > psar_val else "bearish"
        indicators.append(IndicatorResult(name="Parabolic SAR", value=psar_val, signal=psar_sig, detail=f"SAR={psar_val:.2f}, price {'above' if psar_sig == 'bullish' else 'below'}"))

        st_series = _supertrend(df, 10, 3.0)
        st_val = _safe_last(st_series)
        st_sig = "bullish" if close.iloc[-1] > st_val else "bearish" if not np.isnan(st_val) else "neutral"
        indicators.append(IndicatorResult(name="Supertrend(10,3)", value=st_val, signal=st_sig))

        wr_val = _safe_last(_williams_r(df, 14))
        indicators.append(_signal_from_williams_r(wr_val))

        vwap_val = _safe_last(_vwap(df))
        vwap_sig = "bullish" if close.iloc[-1] > vwap_val else "bearish" if not np.isnan(vwap_val) else "neutral"
        indicators.append(IndicatorResult(name="VWAP", value=vwap_val, signal=vwap_sig, detail=f"VWAP={vwap_val:.2f}"))

        roc_val = _safe_last(_roc(close, 12))
        roc_sig = "bullish" if roc_val > 0 else "bearish" if roc_val < 0 else "neutral"
        indicators.append(IndicatorResult(name="ROC(12)", value=roc_val, signal=roc_sig, detail=f"ROC={roc_val:.2f}%"))

        # --- Support / Resistance (Req 5.2) ---
        sr_levels = _pivot_support_resistance(df) + _volume_profile_levels(df)
        supports = [s for s in sr_levels if s.kind == "support"]
        resistances = [s for s in sr_levels if s.kind == "resistance"]

        # --- Overbought / Oversold (Req 5.3) ---
        overbought, oversold, ob_os_details = _classify_ob_os(rsi_val, mfi_val, stoch_k_val, wr_val)

        # --- Divergences (Req 5.3) ---
        divergences = _detect_divergences(close, rsi_series, macd_line)

        # --- Composite signal (Req 5.4) ---
        composite, bull_cnt, bear_cnt, neut_cnt = _aggregate_composite_signal(indicators)

        # --- Candlestick patterns (Req 5.5) ---
        patterns = _detect_candlestick_patterns(df)

        # --- MA crossovers (Req 5.6) ---
        crossovers = _detect_ma_crossovers(df)

        # --- Smart Money Flow (Req 5.7) ---
        total_vol = float(df["volume"].iloc[-1]) if df["volume"].iloc[-1] > 0 else 1.0
        smart_money = _compute_smart_money_flow(foreign_net_volume, institutional_net_volume, total_vol)

        return TechnicalResult(
            symbol=symbol,
            indicators=indicators,
            support_levels=supports,
            resistance_levels=resistances,
            patterns=patterns,
            crossovers=crossovers,
            divergences=divergences,
            smart_money_flow=smart_money,
            composite_signal=composite,
            bullish_count=bull_cnt,
            bearish_count=bear_cnt,
            neutral_count=neut_cnt,
            overbought=overbought,
            oversold=oversold,
            ob_os_details=ob_os_details,
        )

    def to_protobuf(self, result: TechnicalResult) -> TechnicalAnalysis:
        """Convert internal TechnicalResult to protobuf TechnicalAnalysis."""
        return TechnicalAnalysis(
            symbol=result.symbol,
            composite_signal=result.composite_signal.value,
            indicators={ind.name: ind.value for ind in result.indicators if not np.isnan(ind.value)},
            support_levels=[f"{s.level:.2f} ({s.source})" for s in result.support_levels[:5]],
            resistance_levels=[f"{r.level:.2f} ({r.source})" for r in result.resistance_levels[:5]],
            patterns=[f"{p.name} ({p.direction})" for p in result.patterns],
            smart_money_flow=result.smart_money_flow.value,
            ma_crossovers=[f"{x.fast}/{x.slow} {x.direction}" for x in result.crossovers],
        )


# ---------------------------------------------------------------------------
# Signal classification helpers
# ---------------------------------------------------------------------------


def _safe_last(series: pd.Series) -> float:
    """Return last non-NaN value or NaN."""
    if series.empty:
        return float("nan")
    val = series.iloc[-1]
    return float(val) if not pd.isna(val) else float("nan")


def _signal_from_rsi(val: float) -> IndicatorResult:
    if np.isnan(val):
        return IndicatorResult(name="RSI(14)", value=val, signal="neutral")
    sig = "bearish" if val > 70 else "bullish" if val < 30 else "neutral"
    return IndicatorResult(name="RSI(14)", value=val, signal=sig, detail=f"RSI={val:.1f}")


def _signal_from_macd(line: float, signal: float, hist: float) -> IndicatorResult:
    if np.isnan(line) or np.isnan(signal):
        return IndicatorResult(name="MACD(12,26,9)", value=float("nan"), signal="neutral")
    sig = "bullish" if line > signal else "bearish"
    return IndicatorResult(name="MACD(12,26,9)", value=hist, signal=sig, detail=f"MACD line={line:.4f}, signal={signal:.4f}")


def _signal_from_bb(price: float, upper: float, middle: float, lower: float) -> IndicatorResult:
    if np.isnan(upper):
        return IndicatorResult(name="BB(20,2)", value=float("nan"), signal="neutral")
    if price > upper:
        sig = "bearish"
    elif price < lower:
        sig = "bullish"
    else:
        sig = "neutral"
    return IndicatorResult(name="BB(20,2)", value=middle, signal=sig, detail=f"Price vs BB: upper={upper:.2f}, lower={lower:.2f}")


def _signal_from_ma(name: str, price: float, ma_val: float) -> IndicatorResult:
    if np.isnan(ma_val):
        return IndicatorResult(name=name, value=ma_val, signal="neutral")
    sig = "bullish" if price > ma_val else "bearish"
    return IndicatorResult(name=name, value=ma_val, signal=sig, detail=f"{name}={ma_val:.2f}")


def _signal_from_stochastic(k: float, d: float) -> IndicatorResult:
    if np.isnan(k):
        return IndicatorResult(name="Stochastic(14,3,3)", value=float("nan"), signal="neutral")
    sig = "bearish" if k > 80 else "bullish" if k < 20 else "neutral"
    return IndicatorResult(name="Stochastic(14,3,3)", value=k, signal=sig, detail=f"%K={k:.1f}, %D={d:.1f}")


def _signal_from_mfi(val: float) -> IndicatorResult:
    if np.isnan(val):
        return IndicatorResult(name="MFI(14)", value=val, signal="neutral")
    sig = "bearish" if val > 80 else "bullish" if val < 20 else "neutral"
    return IndicatorResult(name="MFI(14)", value=val, signal=sig, detail=f"MFI={val:.1f}")


def _signal_from_williams_r(val: float) -> IndicatorResult:
    if np.isnan(val):
        return IndicatorResult(name="Williams %R(14)", value=val, signal="neutral")
    sig = "bearish" if val > -20 else "bullish" if val < -80 else "neutral"
    return IndicatorResult(name="Williams %R(14)", value=val, signal=sig, detail=f"Williams %R={val:.1f}")
