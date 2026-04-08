-- Article-level sentiment analysis results from LLM processing.
CREATE TABLE IF NOT EXISTS article_sentiments (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    source VARCHAR(50) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    analyzed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sentiment VARCHAR(20) NOT NULL CHECK (sentiment IN ('positive', 'negative', 'neutral')),
    confidence_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    summary TEXT NOT NULL DEFAULT '',
    key_topics JSONB NOT NULL DEFAULT '[]',
    impact_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    raw_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_article_sentiments_symbol ON article_sentiments (symbol);
CREATE INDEX IF NOT EXISTS idx_article_sentiments_analyzed_at ON article_sentiments (analyzed_at);
CREATE INDEX IF NOT EXISTS idx_article_sentiments_symbol_analyzed ON article_sentiments (symbol, analyzed_at);
CREATE INDEX IF NOT EXISTS idx_article_sentiments_sentiment ON article_sentiments (sentiment);