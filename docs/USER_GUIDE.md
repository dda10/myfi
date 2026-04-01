# MyFi User Guide

MyFi is a unified finance platform for Vietnamese stock market analysis and portfolio management. It brings together real-time market data, advanced charting, AI-powered insights, and portfolio tracking into a single dashboard.

## Table of Contents

- [Getting Started](#getting-started)
- [Overview Tab](#overview-tab)
- [Markets Tab](#markets-tab)
- [Portfolio Tab](#portfolio-tab)
- [Screener Tab](#screener-tab)
- [Sectors Tab](#sectors-tab)
- [Comparison Tab](#comparison-tab)
- [Goals Tab](#goals-tab)
- [Signals Tab](#signals-tab)
- [Watchlists Tab](#watchlists-tab)
- [Backtest Tab](#backtest-tab)
- [AI Chat](#ai-chat)
- [Settings and Preferences](#settings-and-preferences)
- [Exports and Reports](#exports-and-reports)
- [Offline Mode](#offline-mode)
- [Tips and Best Practices](#tips-and-best-practices)


---

## Getting Started

### Creating an Account

1. Open MyFi in your browser and navigate to the login page.
2. Click **Register**.
3. Fill in your details:
   - Username (minimum 3 characters)
   - Password (minimum 8 characters)
   - Email (optional)
4. Click **Create Account** to register.
5. Log in with your new credentials.

### First Login

After logging in you land on the **Overview** tab. The sidebar on the left lists all available tabs. Click any tab name to navigate.

### Setting Your Preferences

Before diving in, configure your display preferences using the controls in the header bar:
1. Click the **Indicators** panel on the right side of the chart
2. Expand a category (Trend, Momentum, Volatility, Volume) to browse available indicators
3. Click the **+** button next to an indicator to add it to the chart
4. Overlay indicators (SMA, EMA, Bollinger Bands, etc.) appear on the price pane; oscillators (RSI, MACD, etc.) appear in a separate pane below
5. Click an active indicator name to expand its parameter editor — adjust period, multiplier, or other settings
6. Use the eye icon to toggle visibility, or the trash icon to remove an indicator

### Available Indicators (21)

| Category | Indicators |
|----------|-----------|
| Trend | SMA, EMA, VWAP, VWMA, ADX, Aroon, Parabolic SAR, Supertrend |
| Momentum | RSI, MACD, Williams %R, CMO, Stochastic, ROC, Momentum |
| Volatility | Bollinger Bands, Keltner Channel, ATR, Standard Deviation |
| Volume | OBV, Linear Regression |

You can add multiple instances of the same indicator with different parameters (e.g., SMA(20) and SMA(50) at the same time).

### Drawing Tools

Use the drawing toolbar above the chart:

- **Trend Line**: Click and drag to draw a trend line between two price points
- **Horizontal Line**: Click to place a horizontal price level
- **Fibonacci Retracement**: Click and drag to draw Fibonacci levels between a high and low
- **Rectangle**: Click and drag to highlight a price/time zone
- **Clear All**: Remove all drawings from the chart

Drawings are saved in your browser's local storage and persist across page reloads.

---

## Portfolio Management

Navigate to the **Portfolio** tab to manage your holdings and track performance.

### Adding Assets

1. Use the transaction form to record buy/sell transactions
2. Provide the symbol, quantity, unit price, and transaction date
3. Supported asset types: VN stocks, gold, cryptocurrency, savings accounts, term deposits, bonds

### Holdings View

The holdings table shows each position with:
- Current market price (fetched in near real-time)
- Market value and allocation percentage
- Unrealized P&L (profit/loss) in VND and percentage
- Cost basis computed using the weighted average method

### Transactions

View your full transaction history with type (buy, sell, deposit, withdrawal, interest, dividend), date, quantity, and value. Use the date range filter to narrow results.

### Performance Metrics

The Performance section displays:
- **TWR** (Time-Weighted Return): Measures portfolio performance independent of cash flows
- **MWRR** (Money-Weighted Return / XIRR): Reflects the impact of your timing of deposits and withdrawals
- **Equity Curve**: A chart of your portfolio NAV over time
- **Benchmark Comparison**: Compare your returns against VN-Index and VN30

### Risk Metrics

The Risk section provides:
- **Sharpe Ratio**: Risk-adjusted return (higher is better)
- **Max Drawdown**: Largest peak-to-trough decline
- **Portfolio Beta**: Sensitivity to market movements
- **Volatility**: Annualized standard deviation of returns
- **Value at Risk (VaR)**: Estimated maximum loss at a given confidence level

---

## Stock Screener

Navigate to the **Filter** tab to screen stocks across all Vietnamese exchanges (HOSE, HNX, UPCOM).

### Setting Filters

Use the range inputs to set minimum and maximum values for:

| Filter | Description |
|--------|------------|
| P/E | Price-to-Earnings ratio |
| P/B | Price-to-Book ratio |
| EV/EBITDA | Enterprise Value to EBITDA |
| Market Cap | Minimum market capitalization |
| ROE / ROA | Return on Equity / Assets |
| Revenue Growth | Year-over-year revenue growth |
| Profit Growth | Year-over-year profit growth |
| Dividend Yield | Annual dividend yield |
| Debt/Equity | Leverage ratio |

### Sector and Exchange Filters

- Select one or more **ICB sectors** (Technology, Finance, Real Estate, etc.) to narrow results
- Filter by **exchange** (HOSE, HNX, UPCOM)
- Filter by **sector trend** (Uptrend, Downtrend, Sideways)

### Sorting and Pagination

Click any column header to sort results ascending or descending. Results are paginated — use the page controls at the bottom to navigate.

### Saving Presets

1. Configure your desired filters
2. Click **Save Preset** and give it a name
3. Reload saved presets from the preset dropdown to quickly reapply your favorite filter combinations

---

## Stock Comparison

Navigate to the **Comparison** tab to analyze multiple stocks side by side.

### Adding Stocks

Type a stock symbol into the input field and press Enter or click Add. You can compare up to several stocks simultaneously.

### Comparison Modes

Switch between three analysis modes using the tab bar:

1. **Valuation**: Compare P/E, P/B, EV/EBITDA, ROE, dividend yield, and debt/equity across selected stocks with bar charts
2. **Performance**: View normalized price performance over a selected time period (1M, 3M, 6M, 1Y) on a single line chart
3. **Correlation**: See a correlation matrix showing how closely the price movements of selected stocks are related (values from -1 to +1)

---

## Sector Dashboard

The Sector Dashboard provides an overview of all 10 ICB sectors on the Vietnamese market.

### Sector Overview

Each sector card displays:
- Current trend (Uptrend / Downtrend / Sideways) with a visual indicator
- Performance across multiple time horizons: today, 1 week, 1 month, 3 months, 6 months, 1 year
- A mini sparkline chart showing recent price action

### Sector Fundamentals

Click on a sector to view its median fundamental metrics:
- Median P/E, P/B, ROE, ROA, Dividend Yield, Debt/Equity

### ICB Sectors

| Code | Sector |
|------|--------|
| VNIT | Technology |
| VNIND | Industrial |
| VNCONS | Consumer Discretionary |
| VNCOND | Consumer Staples |
| VNHEAL | Healthcare |
| VNENE | Energy |
| VNUTI | Utilities |
| VNREAL | Real Estate |
| VNFIN | Finance |
| VNMAT | Materials |

---

## AI Chat Assistant

The AI chat widget is accessible from any page via the floating chat button in the bottom-right corner.

### How It Works

MyFi uses a multi-agent AI system with five specialized agents:

1. **Price Agent**: Fetches current and historical prices for any asset
2. **Analysis Agent**: Computes technical indicators, identifies patterns, and performs sector-relative analysis
3. **News Agent**: Aggregates and summarizes relevant financial news
4. **Monitor Agent**: Autonomously scans your watchlist for patterns (accumulation, distribution, breakout)
5. **Supervisor Agent**: Orchestrates all agents and synthesizes a final recommendation

### Example Queries

- "Phân tích SSI" — Get a full technical and fundamental analysis of SSI
- "So sánh FPT và VNM" — Compare two stocks
- "Tin tức về ngành ngân hàng" — Get banking sector news
- "Nên mua hay bán HPG?" — Get a buy/sell/hold recommendation with reasoning
- "Danh mục của tôi có rủi ro gì?" — Assess portfolio risk

### Reading AI Responses

Structured responses include clearly labeled sections:
- **Price Data**: Current price, change, volume with color-coded cards
- **Analysis**: Trend assessment, indicator signals, support/resistance levels, confidence score
- **News**: Recent articles with source links and a summary
- **Advice**: Actionable recommendations with position sizing and risk assessment

### Best Practices

- Be specific with stock symbols (e.g., "FPT" not "that tech company")
- Ask follow-up questions to drill deeper into a topic
- The AI considers your portfolio context when making recommendations
- Responses include a confidence score (0–100) — treat lower scores with more caution

---

## Watchlists

Navigate to the **Watchlist** tab to manage your stock watchlists.

### Creating a Watchlist

1. Click **New Watchlist** and enter a name
2. Type a stock symbol and click Add to include it
3. Create multiple watchlists for different strategies or themes

### Managing Symbols

- View real-time quotes (price, change, volume) for all symbols in a watchlist
- Reorder symbols by dragging or using the sort controls
- Remove a symbol by clicking the delete icon

### Price Alerts

Set upper and lower price alert thresholds for any symbol. When the Monitor Agent detects the price crossing your threshold, you receive an in-app notification.

---

## Goal Planning

Set financial targets and track your progress over time.

### Creating a Goal

1. Navigate to the Goals section
2. Click **New Goal** and fill in:
   - Goal name (e.g., "Retirement Fund", "House Down Payment")
   - Target amount in VND
   - Target date
   - Category (retirement, education, emergency, housing, travel, custom)
   - Associated asset types (optional — link to specific portfolio segments)

### Tracking Progress

- View a progress bar showing current NAV vs. target amount
- See the required monthly contribution to reach your goal on time
- The Supervisor Agent can factor your goals into its recommendations

---

## Backtesting

Test trading strategies against historical data before risking real capital.

### Building a Strategy

1. Navigate to the **Signals** tab and select the Backtest section
2. Choose a stock symbol and date range
3. Define **entry conditions** using indicator-based rules:
   - Select a left operand (e.g., RSI value, SMA, price)
   - Choose an operator (crosses above, crosses below, greater than, less than)
   - Select a right operand (another indicator, a constant value, or price)
4. Define **exit conditions** using the same rule builder
5. Set optional **stop-loss** and **take-profit** percentages

### Supported Indicators for Backtesting

All 21 technical indicators are available as condition operands, including SMA, EMA, RSI, MACD, Bollinger Bands, Stochastic, ADX, and more.

### Reading Results

After running a backtest, you see:
- **Total Return**: Overall strategy return percentage
- **Win Rate**: Percentage of profitable trades
- **Max Drawdown**: Largest peak-to-trough decline during the simulation
- **Sharpe Ratio**: Risk-adjusted return
- **Number of Trades** and **Average Holding Period**
- **Equity Curve**: A chart showing how your capital would have grown over time
- **Trade List**: Every individual trade with entry/exit dates, prices, return, and exit reason (signal, stop-loss, take-profit)

---

## Signals

The **Signals** tab provides AI-generated trading and investment signals.

### Signal Types

- **Trading Signals**: Short-term opportunities with entry price, stop-loss, take-profit, and risk/reward ratio
- **Investment Signals**: Long-term opportunities with entry price zone, target price, and suggested holding period

### Signal Details

Each signal includes:
- Confidence score (0–100)
- Factor scores for technical, volume, money flow, and fundamental dimensions
- Detailed reasoning explaining why the signal was generated

### Accuracy Tracking

View historical accuracy metrics to see how past signals performed at 1-day, 7-day, 14-day, and 30-day intervals.

---

## Export & Reports

Export your portfolio data from the Portfolio tab.

### Available Exports

| Export | Format | Description |
|--------|--------|------------|
| Transactions | CSV | Full transaction history with optional date range filter |
| Portfolio Snapshot | CSV | Current holdings with prices, market value, and P&L |
| Portfolio Report | PDF | Summary report with NAV, allocation breakdown, and holdings |
| Tax Report | CSV | Capital gains summary for tax filing with date range filter |

### How to Export

1. Go to the **Portfolio** tab
2. Click the export button for the desired report type
3. The file downloads automatically to your browser's download folder

---

## Offline Mode

MyFi caches data locally so you can still browse when your internet connection drops.

- A **freshness indicator** appears next to data showing how recent it is
- When offline, cached prices and charts remain visible with a stale data warning
- New transactions and chat queries require an active connection
- Data refreshes automatically when connectivity is restored

---

## Keyboard Shortcuts & Tips

- Use the **sidebar** to quickly switch between tabs (Overview, Portfolio, Markets, Filter, Signals, Comparison, Watchlist)
- Toggle **dark/light theme** from the header — charts automatically adapt their color scheme
- Switch **language** (Vietnamese ↔ English) using the language selector in the header — all UI text and number/date formatting updates instantly
- The AI chat widget can be opened from any page without losing your current view
- Bookmark specific tabs (e.g., `/markets`, `/portfolio`) for quick access
- **Theme** — Toggle between dark and light mode. Charts and all UI elements adapt automatically.
- **Language** — Switch between Vietnamese (vi-VN) and English (en-US). All labels, tooltips, and messages update instantly.
- **Currency** — Toggle between VND and USD display. Prices and values are converted using the live USD/VND exchange rate.

All preferences are saved to your browser and persist across sessions.

---

## Overview Tab

The Overview tab is your home dashboard. It provides a snapshot of your financial position and the market at a glance.

### What You See

- **NAV (Net Asset Value)** — The total value of all your assets (stocks, gold, crypto, savings) displayed prominently at the top.
- **Asset Allocation** — A breakdown of your portfolio by asset type, shown as both VND amounts and percentages.
- **Quick Stats** — Key metrics including total P&L, daily change, and number of holdings.
- **Watchlist Preview** — Your default watchlist with live prices and daily change percentages.
- **Gold Prices** — Current Doji/SJC gold buy and sell prices.
- **Crypto Prices** — Live cryptocurrency quotes.
- **Recent Alerts** — Notifications from the AI Monitor Agent about detected market patterns.

### Quick Actions

- Click any stock symbol in the watchlist to jump to its chart in the Markets tab.
- Click the NAV card to navigate to the full Portfolio tab.


---

## Markets Tab

The Markets tab provides detailed market data and an interactive TradingView-style charting experience.

### Viewing a Chart

1. Enter a stock symbol (e.g., `FPT`, `SSI`, `VNM`) in the symbol search bar.
2. The candlestick chart loads with OHLCV (Open, High, Low, Close, Volume) data.
3. A volume histogram is displayed below the price chart, synchronized on the same time axis.
4. Hover over the chart to see a crosshair with price and time tooltips.

### Time Intervals

Select a time interval from the toolbar: **1m**, **5m**, **15m**, **1h**, **1D**, **1W**, **1M**. The chart re-fetches data and recalculates all active indicators when you switch intervals.

### Technical Indicators

MyFi supports 21 technical indicators organized into four categories. Open the **Indicators** panel to browse and add them.

**Trend Indicators (8):**

| Indicator | Default Parameters | Display |
| --- | --- | --- |
| SMA (Simple Moving Average) | Period: 20 | Overlay |
| EMA (Exponential Moving Average) | Period: 12 | Overlay |
| VWAP (Volume Weighted Average Price) | — | Overlay |
| VWMA (Volume Weighted Moving Average) | Period: 20 | Overlay |
| Parabolic SAR | AF Step: 0.02, AF Max: 0.2 | Overlay |
| Supertrend | Period: 10, Multiplier: 3 | Overlay |
| ADX (Average Directional Index) | Period: 14 | Oscillator |
| Aroon | Period: 25 | Oscillator |

**Momentum Indicators (7):**

| Indicator | Default Parameters |
| --- | --- |
| RSI (Relative Strength Index) | Period: 14 |
| MACD | Fast: 12, Slow: 26, Signal: 9 |
| Williams %R | Period: 14 |
| CMO (Chande Momentum Oscillator) | Period: 14 |
| Stochastic Oscillator | %K: 14, %D: 3, Smooth: 3 |
| ROC (Rate of Change) | Period: 12 |
| Momentum | Period: 10 |

**Volatility Indicators (4):**

| Indicator | Default Parameters | Display |
| --- | --- | --- |
| Bollinger Bands | Period: 20, Std Dev: 2 | Overlay |
| Keltner Channel | Period: 20, Multiplier: 1.5 | Overlay |
| ATR (Average True Range) | Period: 14 | Oscillator |
| Standard Deviation | Period: 20 | Oscillator |

**Volume / Statistics Indicators (2):**

| Indicator | Default Parameters |
| --- | --- |
| OBV (On-Balance Volume) | — |
| Linear Regression | Period: 14 |

#### Adding an Indicator

1. Open the **Indicators** panel on the right side of the chart.
2. Expand a category (Trend, Momentum, Volatility, Volume).
3. Click the **+** button next to the indicator name.
4. The indicator appears on the chart immediately.

#### Configuring Parameters

1. In the **Active** section of the Indicators panel, click an indicator name to expand its settings.
2. Adjust parameter values (period, multiplier, etc.) using the number inputs.
3. The chart updates in real time as you change values.

#### Managing Indicators

- **Hide/Show** — Click the eye icon to toggle visibility without removing the indicator.
- **Remove** — Click the trash icon to remove an indicator from the chart.
- **Multiple instances** — You can add the same indicator multiple times with different parameters (e.g., SMA(20) and SMA(50) simultaneously).

### Drawing Tools

The drawing toolbar provides four tools for annotating charts:

| Tool | Description |
| --- | --- |
| **Trend Line** | Draw diagonal lines to mark price trends. Click two points on the chart. |
| **Horizontal Line** | Draw a horizontal line at a specific price level. Click once on the chart. |
| **Fibonacci Retracement** | Draw Fibonacci levels between two price points. Click the high and low. |
| **Rectangle** | Draw a rectangular area to highlight a price/time zone. Click two corners. |

- Select a tool from the toolbar, then click on the chart to draw.
- Click the same tool again to deselect it.
- Click **Clear all drawings** (trash icon) to remove all annotations.
- Drawings are saved in your browser and persist across page reloads.


---

## Portfolio Tab

The Portfolio tab is your central hub for tracking holdings, transactions, performance, and risk.

### Holdings

View all your current asset holdings in a table showing:

- Symbol and asset type
- Quantity held
- Average cost basis
- Current market price
- Market value
- Unrealized P&L (amount and percentage)

Holdings are updated with near real-time prices. Color coding indicates gains (green) and losses (red).

### Recording Transactions

1. Click **Add Transaction** to open the transaction form.
2. Select the transaction type: **Buy**, **Sell**, **Deposit**, **Withdrawal**, **Interest**, or **Dividend**.
3. Enter the asset symbol, quantity, unit price, and date.
4. Click **Save**. The portfolio recalculates NAV and P&L immediately.

Sell transactions are validated — you cannot sell more than your current holding quantity. Cost basis is calculated using the weighted average method.

### Transaction History

View a chronological list of all transactions with filtering options. Each entry shows the asset, type, quantity, price, total value, and date.

### Performance Metrics

The Performance section displays:

- **TWR (Time-Weighted Return)** — Measures portfolio performance independent of cash flows. Useful for comparing against benchmarks.
- **MWRR (Money-Weighted Return / XIRR)** — Measures the actual return earned on your invested capital, accounting for the timing of deposits and withdrawals.
- **NAV Equity Curve** — A chart showing your portfolio value over time.
- **Benchmark Comparison** — Compare your portfolio performance against VN-Index and VN30.

### Risk Metrics

The Risk section provides portfolio-level and per-holding risk analysis:

- **Sharpe Ratio** — Risk-adjusted return (higher is better).
- **Max Drawdown** — The largest peak-to-trough decline in portfolio value.
- **Portfolio Beta** — Sensitivity of your portfolio to market movements.
- **Volatility** — Annualized standard deviation of returns.
- **Value at Risk (VaR)** — Estimated maximum loss at a given confidence level.

### Exporting Data

Use the export buttons at the top of the Portfolio tab:

- **Export Transactions (CSV)** — Download your full transaction history.
- **Export Portfolio (PDF)** — Generate a PDF snapshot of your current holdings and performance.
- **Tax Report (CSV)** — Download a capital gains summary formatted for tax reporting.


---

## Screener Tab

The Screener tab lets you filter the entire Vietnamese stock universe by fundamental metrics, sector, exchange, and trend.

### Available Filters

**Fundamental Metrics (all support min/max ranges):**

| Filter | Description |
| --- | --- |
| P/E Ratio | Price-to-Earnings ratio |
| P/B Ratio | Price-to-Book ratio |
| EV/EBITDA | Enterprise Value to EBITDA |
| Market Cap | Minimum market capitalization |
| ROE | Return on Equity (%) |
| ROA | Return on Assets (%) |
| Revenue Growth | Year-over-year revenue growth (%) |
| Profit Growth | Year-over-year profit growth (%) |
| Dividend Yield | Annual dividend yield (%) |
| Debt-to-Equity | Leverage ratio |

**Classification Filters:**

- **Sectors** — Filter by ICB sectors: Technology (VNIT), Industrial (VNIND), Consumer Staples (VNCONS), Consumer Discretionary (VNCOND), Healthcare (VNHEAL), Energy (VNENE), Utilities (VNUTI), Real Estate (VNREAL), Finance (VNFIN), Materials (VNMAT).
- **Exchanges** — Filter by HOSE, HNX, or UPCOM.
- **Sector Trends** — Filter by sector trend direction: Uptrend, Downtrend, or Sideways.

### Using the Screener

1. Set your desired filter ranges. Leave a field empty to skip that filter.
2. Click **Search** to run the screener. Results appear in a sortable table.
3. Click any column header to sort results by that metric.
4. Use pagination controls to browse through results (10, 20, or 50 per page).
5. Click a stock symbol in the results to view its chart.

### Built-in Presets

Quick-start with common screening strategies:

- **Value Investing** — P/E ≤ 12, P/B ≤ 1.5
- **High Growth** — Revenue Growth ≥ 20%, Profit Growth ≥ 20%
- **High Dividend** — Dividend Yield ≥ 5%
- **Low Debt** — Debt-to-Equity ≤ 0.5

### Saving Custom Presets

1. Configure your filters.
2. Click **Save Preset** and enter a name.
3. Your preset is saved and appears in the preset dropdown for future use.
4. To load a saved preset, select it from the dropdown. To delete, click the trash icon next to it.


---

## Sectors Tab

The Sectors tab provides an ICB sector-level view of the Vietnamese stock market.

### Sector Heatmap

The heatmap displays all 10 ICB sectors with color-coded performance. Green indicates positive performance, red indicates negative. The intensity of the color reflects the magnitude of the change.

### Time Period Selection

Switch between time periods to view sector performance over different horizons: **1D**, **1W**, **1M**, **3M**, **6M**, **1Y**.

### Sector Trend Indicators

Each sector shows a trend indicator (Uptrend, Downtrend, or Sideways) based on price action analysis over the selected period.

### Sector Detail Panel

Click any sector in the heatmap to open its detail panel, which shows:

- **Performance chart** — A mini chart of the sector index over time.
- **Fundamental averages** — Average P/E, P/B, ROE, and other metrics for stocks in the sector.
- **Top stocks** — The leading stocks within that sector by market cap or performance.

Use the Sectors tab to identify which industries are leading or lagging the market, and to contextualize individual stock performance within their sector.

---

## Comparison Tab

The Comparison tab lets you compare up to 10 stocks side-by-side across three dimensions.

### Adding Stocks to Compare

1. Type a stock symbol into the search input and press Enter or click **Add**.
2. Repeat for up to 10 stocks. Added symbols appear as chips that you can remove by clicking the **×** button.

### Comparison Views

Switch between three tabs:

**Valuation:**
- Compares P/E and P/B ratios over time as line charts.
- Quickly spot which stocks are relatively cheap or expensive.

**Performance:**
- Shows normalized return charts starting from a common baseline (100%).
- Visualize how each stock has performed relative to the others over the selected period.

**Correlation:**
- Displays a correlation matrix showing the statistical relationship between each pair of stocks.
- Values range from -1 (perfectly inverse) to +1 (perfectly correlated).
- Use this to assess diversification — stocks with low correlation provide better diversification.

### Tips for Comparison

- Compare stocks within the same sector for peer analysis.
- Compare across sectors to find diversification opportunities.
- Use the performance view to identify relative strength leaders.


---

## Goals Tab

The Goals tab helps you set financial targets and track your progress.

### Creating a Goal

1. Click **Add Goal**.
2. Enter a name (e.g., "Retirement Fund", "House Down Payment").
3. Set the target amount in VND.
4. Set the target date.
5. Optionally associate the goal with specific asset types.
6. Click **Save**.

### Tracking Progress

Each goal displays:

- **Progress bar** — Visual indicator of how close you are to the target.
- **Current value** — Calculated from your portfolio NAV for the associated asset types.
- **Remaining amount** — How much more you need to reach the target.
- **Required monthly contribution** — An estimate of how much you need to save each month to reach the goal on time.

The AI Supervisor Agent can also factor your goals into its recommendations.

---

## Signals Tab

The Signals tab displays AI-generated trading and investment signals.

### Signal Types

- **Trading Signals** — Short-term opportunities with entry price, stop-loss, take-profit, risk/reward ratio, and confidence score (0–100).
- **Investment Signals** — Long-term opportunities with entry price zone, target price, suggested holding period, and fundamental reasoning.

### Signal Details

Click any signal card to expand it and see:

- **Factor scores** — Breakdown of technical, volume, money flow, and fundamental scores.
- **Direction** — Long or short.
- **Reasoning** — The AI's explanation for the signal.
- **Confidence** — A score from 0 to 100 indicating signal strength.

### Signal Accuracy

Switch to the **Accuracy** tab to view historical signal performance:

- Win rate, average return, and total signals tracked.
- Performance breakdown by signal type and time horizon.

### Signal Backtesting

The **Backtest** sub-tab lets you test signal strategies against historical data to validate their effectiveness before acting on them.

---

## Watchlists Tab

The Watchlists tab lets you create and manage multiple named watchlists with price alerts.

### Creating a Watchlist

1. Click **New Watchlist**.
2. Enter a name for the watchlist.
3. Click **Create**.

### Managing Symbols

1. Select a watchlist from the list.
2. Type a symbol in the **Add Symbol** input and press Enter.
3. Symbols appear with live price data, daily change, and volume.
4. Click the **×** button next to a symbol to remove it.

### Price Alerts

1. Click the bell icon next to any symbol in a watchlist.
2. Set an upper and/or lower price threshold.
3. When the price crosses your threshold, you receive an alert notification.

Watchlist data syncs with the AI Monitor Agent, which can incorporate your watched symbols into its autonomous scanning.

---

## Backtest Tab

The Backtest tab lets you define indicator-based trading strategies and simulate them against historical data.

### Defining a Strategy

1. **Select a symbol** — Choose the stock to backtest against.
2. **Set the date range** — Pick start and end dates for the simulation period.
3. **Define entry rules** — Set conditions for when to buy. Each condition compares an indicator value against a threshold or another indicator. For example: "Buy when RSI crosses below 30" or "Buy when SMA(20) crosses above SMA(50)".
4. **Define exit rules** — Set conditions for when to sell, using the same condition format.
5. Click **Run Backtest**.

### Reading Results

After the simulation completes, you see:

- **Total Return** — Overall profit/loss percentage.
- **Win Rate** — Percentage of profitable trades.
- **Max Drawdown** — Largest peak-to-trough decline during the simulation.
- **Sharpe Ratio** — Risk-adjusted return of the strategy.
- **Number of Trades** — Total trades executed.
- **Equity Curve** — A chart showing the strategy's portfolio value over time.
- **Trade List** — Individual trades with entry/exit dates, prices, and P&L.


---

## AI Chat

The AI Chat is a floating widget accessible from any tab. It connects you to a multi-agent AI advisory system powered by five specialized agents.

### Opening the Chat

Click the chat icon in the bottom-right corner of the screen. The chat panel slides open.

### How It Works

When you send a message, the **Supervisor Agent** analyzes your query and dispatches it to the relevant specialist agents:

| Agent | Role |
| --- | --- |
| **Price Agent** | Fetches current and historical prices for any asset |
| **Analysis Agent** | Performs technical analysis (21 indicators), fundamental analysis, and sector-relative analysis |
| **News Agent** | Retrieves and summarizes relevant financial news |
| **Monitor Agent** | Reports on detected market patterns (accumulation, distribution, breakout) |
| **Supervisor Agent** | Orchestrates all agents and synthesizes a final recommendation |

### Structured Responses

The AI chat returns structured data when relevant:

- **Price cards** — Current price, change, volume for requested symbols.
- **Analysis section** — Technical indicator readings, trend assessment, support/resistance levels.
- **News section** — Summarized news articles with source links.
- **Advice section** — Actionable recommendations with reasoning.

Stock symbols mentioned in responses are highlighted and clickable — click them to jump to the chart.

### Example Queries

Here are effective ways to use the AI chat:

**Price inquiries:**
- "What is the current price of FPT?"
- "Show me SSI price history for the last 3 months"

**Technical analysis:**
- "Analyze VNM using RSI and MACD"
- "Is HPG in an uptrend or downtrend?"
- "What are the support and resistance levels for MWG?"

**Fundamental analysis:**
- "Compare the P/E ratios of FPT and VNM"
- "Which bank stocks have the best ROE?"

**Portfolio advice:**
- "Should I buy more FPT at the current price?"
- "What stocks should I consider for my portfolio?"
- "My portfolio is heavy on real estate — how should I diversify?"

**News and events:**
- "What's the latest news about Vingroup?"
- "Any upcoming dividends for my watchlist stocks?"

### Best Practices for AI Chat

1. **Be specific** — Include stock symbols and timeframes. "Analyze FPT for the next week" works better than "What should I buy?"
2. **Ask follow-up questions** — The chat maintains conversation context. You can drill deeper into any response.
3. **Combine perspectives** — Ask for both technical and fundamental views to get a balanced picture.
4. **Verify signals** — Use the AI's analysis as one input alongside your own research. Cross-reference with the Screener and Comparison tools.
5. **Check the reasoning** — The AI provides explanations for its recommendations. Read them to understand the logic, not just the conclusion.

---

## Settings and Preferences

### Theme (Dark / Light)

Click the theme toggle in the header to switch between dark and light mode. The setting is saved in your browser's local storage. Charts, tables, and all UI components adapt to the selected theme.

### Language (Vietnamese / English)

Click the language selector in the header to switch between:

- **vi-VN** — Vietnamese interface with Vietnamese number and date formatting.
- **en-US** — English interface with international formatting.

All labels, button text, tooltips, error messages, and placeholder text update immediately.

### Currency (VND / USD)

Click the currency toggle to switch the display currency. When USD is selected, all VND values are converted using the live USD/VND exchange rate (sourced from CoinGecko USDT/VND pair, with a fallback rate of 25,400).

---

## Exports and Reports

MyFi supports exporting data in CSV and PDF formats from the Portfolio tab.

### Available Exports

| Export | Format | Contents |
| --- | --- | --- |
| Transaction History | CSV | All buy, sell, deposit, withdrawal, interest, and dividend transactions |
| Portfolio Snapshot | PDF | Current holdings, market values, P&L, allocation breakdown |
| Tax Report | CSV | Capital gains summary with cost basis, proceeds, and realized gains/losses |

### How to Export

1. Navigate to the **Portfolio** tab.
2. Click the desired export button in the toolbar.
3. The file downloads to your browser's default download location.

---

## Offline Mode

MyFi supports offline and degraded-mode operation when internet connectivity is lost.

### How It Works

- When the connection drops, MyFi serves cached data for all market prices, portfolio values, and other data.
- A **freshness indicator** appears on the screen showing that data may be stale.
- Cached stock prices have a 15-minute TTL, gold prices 1 hour, and crypto prices 5 minutes.
- When connectivity is restored, data refreshes automatically.

### What Works Offline

- Viewing cached portfolio holdings and NAV.
- Browsing previously loaded charts and screener results.
- Accessing saved watchlists and filter presets.

### What Requires Connectivity

- Fetching live prices and new market data.
- Running the AI chat (requires backend API access).
- Recording new transactions.
- Running backtests and screener queries.

---

## Tips and Best Practices

### For New Investors

1. **Start with the Overview** — Get familiar with the dashboard layout and your NAV.
2. **Set up a watchlist** — Add 5–10 stocks you're interested in and monitor them daily.
3. **Use the Screener** — Start with built-in presets like "Value Investing" or "High Dividend" to discover stocks.
4. **Ask the AI** — Use the chat to ask basic questions like "What does P/E ratio mean?" or "Is this stock overvalued?"

### For Active Traders

1. **Master the chart** — Learn to combine trend indicators (SMA, EMA) with momentum indicators (RSI, MACD) for entry/exit timing.
2. **Use drawing tools** — Mark support/resistance levels and trend lines to track key price levels.
3. **Check Signals daily** — Review AI-generated trading signals for short-term opportunities.
4. **Backtest before trading** — Validate your strategy ideas against historical data before committing real capital.

### For Portfolio Managers

1. **Track performance with TWR** — Use time-weighted return to measure your skill independent of cash flow timing.
2. **Monitor risk metrics** — Keep an eye on Sharpe ratio and max drawdown to ensure your risk is within acceptable bounds.
3. **Use Comparison** — Regularly compare your holdings against peers and check the correlation matrix for diversification.
4. **Set goals** — Use the Goals tab to align your portfolio with specific financial targets.
5. **Export tax reports** — Download capital gains summaries before tax season.

### General Tips

- **Corporate actions** — MyFi automatically tracks dividends and stock splits. Your cost basis adjusts automatically for splits and bonus shares.
- **Multiple watchlists** — Create separate watchlists for different strategies (e.g., "Growth Picks", "Dividend Stocks", "Sector Rotation").
- **Sector context** — Always check the Sectors tab to understand whether a stock's movement is stock-specific or sector-wide.
- **Data freshness** — Look for the freshness indicator to know if you're viewing live or cached data.
