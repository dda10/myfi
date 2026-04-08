"""Research Report Generator — orchestrates report creation and S3 storage.

Generates periodic research reports:
- Weekly: Factor Snapshot
- Monthly: Sector Deep-Dive

Requirements: 35.1, 35.5, 35.6
"""

from __future__ import annotations

import logging
import time

import pandas as pd

from ezistock_ai.config import Config
from ezistock_ai.infra.s3 import S3Client
from ezistock_ai.research.factor_snapshot import FactorSnapshotGenerator
from ezistock_ai.research.pdf import PDFGenerator
from ezistock_ai.research.sector_deepdive import SectorDeepDiveGenerator

logger = logging.getLogger(__name__)


class ResearchReportGenerator:
    """Orchestrates research report generation and S3 storage."""

    def __init__(self, s3: S3Client, config: Config | None = None) -> None:
        self._s3 = s3
        self._config = config or Config()
        self._factor_gen = FactorSnapshotGenerator()
        self._sector_gen = SectorDeepDiveGenerator()
        self._pdf = PDFGenerator()

    async def generate_weekly_factor_snapshot(
        self,
        signal_space: pd.DataFrame,
        forward_returns: pd.Series,
        vnindex_returns: pd.Series,
    ) -> str:
        """Generate and store weekly factor snapshot. Returns S3 key."""
        report_date = time.strftime("%Y-%m-%d")
        data = self._factor_gen.generate(
            signal_space, forward_returns, vnindex_returns, report_date,
        )

        pdf_bytes = self._pdf.generate_factor_snapshot_pdf(data)
        key = f"factor_snapshot_{report_date}.html"
        self._s3.put_pdf(key, pdf_bytes)

        logger.info("Generated factor snapshot: %s (%d bytes)", key, len(pdf_bytes))
        return key

    async def generate_monthly_sector_deepdive(
        self,
        sector_performance: list[dict],
        signal_space: pd.DataFrame,
    ) -> str:
        """Generate and store monthly sector deep-dive. Returns S3 key."""
        report_date = time.strftime("%Y-%m-%d")
        data = self._sector_gen.generate(
            sector_performance, signal_space, report_date,
        )

        pdf_bytes = self._pdf.generate_sector_deepdive_pdf(data)
        key = f"sector_deepdive_{report_date}.html"
        self._s3.put_pdf(key, pdf_bytes)

        logger.info("Generated sector deep-dive: %s (%d bytes)", key, len(pdf_bytes))
        return key
