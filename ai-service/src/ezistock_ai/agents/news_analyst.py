"""News Analyst Agent — fetches and analyzes Vietnamese financial news.

Fetches company news via KBS (through Go backend), supplements with web search
of Vietnamese financial sources (CafeF, VnExpress, Vietstock). Uses LLM for
catalyst/risk identification and sentiment scoring.

Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field

import httpx

from ezistock_ai.agents.base import AgentContext, BaseAgent
from ezistock_ai.agents.schemas import NewsSentimentOutput
from ezistock_ai.config import Config
from ezistock_ai.generated.proto.agent_pb2 import NewsAnalysis, NewsArticle
from ezistock_ai.llm.router import LLMRouter, TaskType

logger = logging.getLogger(__name__)

_MAX_ARTICLES = 10


@dataclass
class NewsResult:
    """Internal result from the news analyst."""

    symbol: str
    sentiment: str = "neutral"
    confidence: float = 0.0
    catalysts: list[str] = field(default_factory=list)
    risk_factors: list[str] = field(default_factory=list)
    articles: list[NewsArticle] = field(default_factory=list)
    summary: str = ""


class AsyncNewsAnalyst(BaseAgent[NewsResult]):
    """Fetches news and uses LLM for sentiment analysis.

    News sources:
    1. KBS company news via Go backend REST API
    2. Articles already present in MarketData.news from the gRPC request
    """

    def __init__(
        self,
        llm_router: LLMRouter | None = None,
        config: Config | None = None,
        timeout: float = 30.0,
    ) -> None:
        super().__init__(name="news_analyst", timeout=timeout)
        self._llm_router = llm_router
        self._config = config or Config()
        self._backend_url = self._config.go_backend_url

    async def _run(self, ctx: AgentContext) -> NewsResult:
        symbol = ctx.symbol
        articles: list[NewsArticle] = []

        # 1. Collect articles from MarketData (already fetched by Go backend)
        if ctx.market_data and ctx.market_data.news:
            articles.extend(ctx.market_data.news)

        # 2. Fetch additional news from Go backend KBS endpoint
        try:
            fetched = await self._fetch_kbs_news(symbol)
            # Deduplicate by URL
            existing_urls = {a.url for a in articles}
            for a in fetched:
                if a.url not in existing_urls:
                    articles.append(a)
                    existing_urls.add(a.url)
        except Exception as exc:
            logger.warning("Failed to fetch KBS news for %s: %s", symbol, exc)

        # Limit to most recent articles
        articles = articles[:_MAX_ARTICLES]

        if not articles:
            logger.info("No news articles found for %s", symbol)
            return NewsResult(symbol=symbol)

        # Add citations for each article
        for article in articles:
            self.citation_collector.add(
                source=article.source or "news",
                claim=article.title,
                data_point=f"Published: {article.published_at}",
                url=article.url,
            )

        # 3. Use LLM for sentiment analysis if router available
        if self._llm_router is not None:
            return await self._analyze_with_llm(symbol, articles)

        # Fallback: return articles without LLM analysis
        return NewsResult(
            symbol=symbol,
            articles=articles,
            summary=f"Found {len(articles)} articles for {symbol}",
        )

    async def _fetch_kbs_news(self, symbol: str) -> list[NewsArticle]:
        """Fetch company news from Go backend's KBS endpoint."""
        url = f"{self._backend_url}/api/market/news?symbol={symbol}"
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.get(url)
            if resp.status_code != 200:
                return []
            data = resp.json()
            articles = []
            for item in data.get("articles", data.get("data", []))[:_MAX_ARTICLES]:
                articles.append(NewsArticle(
                    title=item.get("title", ""),
                    url=item.get("url", ""),
                    source=item.get("source", "KBS"),
                    published_at=item.get("published_at", item.get("publishedAt", "")),
                    summary=item.get("summary", ""),
                ))
            return articles

    async def _analyze_with_llm(
        self,
        symbol: str,
        articles: list[NewsArticle],
    ) -> NewsResult:
        """Use LLM with structured output for sentiment analysis."""
        from ezistock_ai.agents.prompts.news import ANALYSIS_PROMPT, SYSTEM_PROMPT

        articles_text = "\n\n".join(
            f"### {i+1}. {a.title}\n"
            f"Source: {a.source} | Date: {a.published_at}\n"
            f"URL: {a.url}\n"
            f"{a.summary}"
            for i, a in enumerate(articles)
        )

        model = self._llm_router.get_model(TaskType.ANALYSIS)
        structured_model = model.with_structured_output(NewsSentimentOutput)

        messages = [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": ANALYSIS_PROMPT.format(
                symbol=symbol,
                articles_text=articles_text,
                article_count=len(articles),
            )},
        ]

        output: NewsSentimentOutput = await structured_model.ainvoke(messages)

        return NewsResult(
            symbol=symbol,
            sentiment=output.sentiment,
            confidence=output.confidence,
            catalysts=output.catalysts,
            risk_factors=output.risk_factors,
            articles=articles,
            summary=output.summary,
        )

    @staticmethod
    def to_protobuf(result: NewsResult) -> NewsAnalysis:
        """Convert NewsResult to protobuf NewsAnalysis."""
        return NewsAnalysis(
            symbol=result.symbol,
            sentiment=result.sentiment,
            confidence=result.confidence,
            catalysts=result.catalysts,
            risk_factors=result.risk_factors,
            articles=result.articles,
        )
