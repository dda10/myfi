"""S3/MinIO client for Parquet files, PDFs, and model artifacts.

Configurable via environment variables for local MinIO vs production AWS S3.

Requirements: 40.2, 40.7, 40.8
"""

from __future__ import annotations

import logging
from typing import Optional

import boto3
from botocore.config import Config as BotoConfig

from ezistock_ai.config import Config

logger = logging.getLogger(__name__)


class S3Client:
    """Thin wrapper around boto3 S3 client for EziStock storage operations.

    Supports both AWS S3 and MinIO (local dev) via endpoint configuration.
    """

    # Key prefixes for organized storage
    PREFIX_PARQUET = "data/"
    PREFIX_PDF = "reports/"
    PREFIX_MODEL = "models/"

    def __init__(self, config: Config) -> None:
        self._bucket = config.s3_bucket
        kwargs: dict = {
            "endpoint_url": config.s3_endpoint,
            "config": BotoConfig(s3={"addressing_style": "path"} if config.s3_use_path_style else {}),
        }
        if config.s3_access_key:
            kwargs["aws_access_key_id"] = config.s3_access_key
            kwargs["aws_secret_access_key"] = config.s3_secret_key

        self._client = boto3.client("s3", **kwargs)
        logger.info("S3 client initialized: bucket=%s endpoint=%s", self._bucket, config.s3_endpoint)

    # ------------------------------------------------------------------
    # Parquet operations
    # ------------------------------------------------------------------

    def put_parquet(self, key: str, data: bytes) -> None:
        """Upload Parquet data to S3."""
        full_key = f"{self.PREFIX_PARQUET}{key}"
        self._client.put_object(
            Bucket=self._bucket,
            Key=full_key,
            Body=data,
            ContentType="application/octet-stream",
        )
        logger.debug("Uploaded parquet: %s (%d bytes)", full_key, len(data))

    def get_parquet(self, key: str) -> bytes:
        """Download Parquet data from S3."""
        full_key = f"{self.PREFIX_PARQUET}{key}"
        resp = self._client.get_object(Bucket=self._bucket, Key=full_key)
        data: bytes = resp["Body"].read()
        logger.debug("Downloaded parquet: %s (%d bytes)", full_key, len(data))
        return data

    # ------------------------------------------------------------------
    # PDF operations
    # ------------------------------------------------------------------

    def put_pdf(self, key: str, data: bytes) -> None:
        """Upload a PDF report to S3."""
        full_key = f"{self.PREFIX_PDF}{key}"
        self._client.put_object(
            Bucket=self._bucket,
            Key=full_key,
            Body=data,
            ContentType="application/pdf",
        )
        logger.debug("Uploaded PDF: %s (%d bytes)", full_key, len(data))

    def get_signed_url(self, key: str, expires_in: int = 3600) -> str:
        """Generate a pre-signed URL for downloading a file."""
        url: str = self._client.generate_presigned_url(
            "get_object",
            Params={"Bucket": self._bucket, "Key": key},
            ExpiresIn=expires_in,
        )
        return url

    # ------------------------------------------------------------------
    # Model artifact operations
    # ------------------------------------------------------------------

    def put_model(self, key: str, data: bytes) -> None:
        """Upload a trained model artifact to S3."""
        full_key = f"{self.PREFIX_MODEL}{key}"
        self._client.put_object(
            Bucket=self._bucket,
            Key=full_key,
            Body=data,
            ContentType="application/octet-stream",
        )
        logger.debug("Uploaded model: %s (%d bytes)", full_key, len(data))

    def get_model(self, key: str) -> bytes:
        """Download a model artifact from S3."""
        full_key = f"{self.PREFIX_MODEL}{key}"
        resp = self._client.get_object(Bucket=self._bucket, Key=full_key)
        data: bytes = resp["Body"].read()
        logger.debug("Downloaded model: %s (%d bytes)", full_key, len(data))
        return data

    # ------------------------------------------------------------------
    # Utility
    # ------------------------------------------------------------------

    def list_keys(self, prefix: str, max_keys: int = 1000) -> list[str]:
        """List object keys under a prefix."""
        resp = self._client.list_objects_v2(
            Bucket=self._bucket,
            Prefix=prefix,
            MaxKeys=max_keys,
        )
        return [obj["Key"] for obj in resp.get("Contents", [])]

    def delete(self, key: str) -> None:
        """Delete an object from S3."""
        self._client.delete_object(Bucket=self._bucket, Key=key)
        logger.debug("Deleted: %s", key)

    def exists(self, key: str) -> bool:
        """Check if an object exists in S3."""
        try:
            self._client.head_object(Bucket=self._bucket, Key=key)
            return True
        except self._client.exceptions.ClientError:
            return False

    def ensure_bucket(self) -> None:
        """Create the bucket if it doesn't exist (useful for local MinIO dev)."""
        try:
            self._client.head_bucket(Bucket=self._bucket)
        except self._client.exceptions.ClientError:
            self._client.create_bucket(Bucket=self._bucket)
            logger.info("Created bucket: %s", self._bucket)
