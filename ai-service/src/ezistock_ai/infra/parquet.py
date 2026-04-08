"""Parquet read/write using PyArrow and DuckDB for analytical queries.

The Python AI Service reads historical data directly from S3 Parquet files
using PyArrow or DuckDB for in-process analytical queries, avoiding the need
to load bulk time-series data into PostgreSQL.

Requirements: 40.2, 40.3
"""

from __future__ import annotations

import io
import logging
from typing import Optional

import duckdb
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq

from ezistock_ai.infra.s3 import S3Client

logger = logging.getLogger(__name__)


class ParquetStore:
    """Read/write Parquet data via S3 with PyArrow and DuckDB query support."""

    def __init__(self, s3: S3Client) -> None:
        self._s3 = s3

    # ------------------------------------------------------------------
    # Write
    # ------------------------------------------------------------------

    def write_dataframe(self, key: str, df: pd.DataFrame, compression: str = "snappy") -> None:
        """Write a pandas DataFrame as Parquet to S3.

        Args:
            key: S3 key (without the data/ prefix — S3Client adds it).
            df: DataFrame to persist.
            compression: Parquet compression codec (snappy, gzip, zstd).
        """
        table = pa.Table.from_pandas(df)
        buf = io.BytesIO()
        pq.write_table(table, buf, compression=compression)
        self._s3.put_parquet(key, buf.getvalue())
        logger.debug("Wrote parquet %s: %d rows, %d cols", key, len(df), len(df.columns))

    def write_table(self, key: str, table: pa.Table, compression: str = "snappy") -> None:
        """Write a PyArrow Table as Parquet to S3."""
        buf = io.BytesIO()
        pq.write_table(table, buf, compression=compression)
        self._s3.put_parquet(key, buf.getvalue())
        logger.debug("Wrote parquet %s: %d rows", key, table.num_rows)

    # ------------------------------------------------------------------
    # Read
    # ------------------------------------------------------------------

    def read_dataframe(self, key: str, columns: Optional[list[str]] = None) -> pd.DataFrame:
        """Read a Parquet file from S3 into a pandas DataFrame.

        Args:
            key: S3 key (without the data/ prefix).
            columns: Optional column subset to read (reduces I/O).
        """
        data = self._s3.get_parquet(key)
        buf = io.BytesIO(data)
        table = pq.read_table(buf, columns=columns)
        df = table.to_pandas()
        logger.debug("Read parquet %s: %d rows, %d cols", key, len(df), len(df.columns))
        return df

    def read_table(self, key: str, columns: Optional[list[str]] = None) -> pa.Table:
        """Read a Parquet file from S3 into a PyArrow Table."""
        data = self._s3.get_parquet(key)
        buf = io.BytesIO(data)
        return pq.read_table(buf, columns=columns)

    # ------------------------------------------------------------------
    # DuckDB analytical queries
    # ------------------------------------------------------------------

    def query(self, key: str, sql: str) -> pd.DataFrame:
        """Run a DuckDB SQL query against a Parquet file from S3.

        The Parquet data is registered as a temporary table named 'data'
        so the SQL should reference `data`. Example:

            store.query("signals/2024.parquet", "SELECT * FROM data WHERE symbol = 'VNM'")

        Args:
            key: S3 key for the Parquet file.
            sql: DuckDB-compatible SQL query referencing the table as 'data'.
        """
        raw = self._s3.get_parquet(key)
        buf = io.BytesIO(raw)
        table = pq.read_table(buf)

        con = duckdb.connect()
        try:
            con.register("data", table)
            result = con.execute(sql).fetchdf()
            logger.debug("DuckDB query on %s returned %d rows", key, len(result))
            return result
        finally:
            con.close()

    def query_multiple(self, keys: list[str], sql: str) -> pd.DataFrame:
        """Run a DuckDB SQL query across multiple Parquet files.

        All files are unioned into a single table named 'data'.
        """
        tables: list[pa.Table] = []
        for key in keys:
            raw = self._s3.get_parquet(key)
            tables.append(pq.read_table(io.BytesIO(raw)))

        if not tables:
            return pd.DataFrame()

        combined = pa.concat_tables(tables)
        con = duckdb.connect()
        try:
            con.register("data", combined)
            result = con.execute(sql).fetchdf()
            logger.debug("DuckDB multi-query across %d files returned %d rows", len(keys), len(result))
            return result
        finally:
            con.close()
