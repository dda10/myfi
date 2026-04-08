-- 012_research_reports.sql: Research reports with PDF S3 key
-- Requirements: 28.1
CREATE TABLE IF NOT EXISTS research_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(300) NOT NULL,
    report_type VARCHAR(50) NOT NULL,
    -- 'factor_snapshot', 'sector_deepdive', 'market_outlook'
    content JSONB NOT NULL,
    pdf_s3_key VARCHAR(500),
    published_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_research_reports_type ON research_reports(report_type, published_at DESC);