package infra

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Storage abstracts S3/MinIO for Parquet files, PDFs, and model artifacts.
//
// Configurable via environment variables:
//   - S3_BUCKET: bucket name (required)
//   - S3_REGION: AWS region (default: ap-southeast-1)
//   - S3_ENDPOINT: custom endpoint for MinIO local dev (optional)
//   - S3_USE_PATH_STYLE: set "true" for MinIO path-style access (optional)
//   - AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY: credentials (uses default chain if unset)
//
// Requirements: 40.2 (S3 Parquet storage), 40.7 (S3 for PDFs/models), 40.8 (env-configurable abstraction)
type Storage struct {
	bucket  string
	client  *s3.Client
	presign *s3.PresignClient
	logger  *slog.Logger
}

// NewStorage creates a Storage client from environment variables.
func NewStorage(logger *slog.Logger) (*Storage, error) {
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET environment variable is required")
	}

	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "ap-southeast-1"
	}

	endpoint := os.Getenv("S3_ENDPOINT")
	usePathStyle := strings.EqualFold(os.Getenv("S3_USE_PATH_STYLE"), "true")

	if logger == nil {
		logger = slog.Default()
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	// For MinIO / local dev: use static credentials from env if endpoint is set.
	if endpoint != "" {
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKey != "" && secretKey != "" {
			opts = append(opts, config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
			))
		}
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = usePathStyle
		})
	}

	client := s3.NewFromConfig(cfg, s3Opts...)
	presign := s3.NewPresignClient(client)

	logger.Info("S3 storage initialized",
		"bucket", bucket,
		"region", region,
		"endpoint", endpoint,
		"pathStyle", usePathStyle,
	)

	return &Storage{
		bucket:  bucket,
		client:  client,
		presign: presign,
		logger:  logger,
	}, nil
}

// PutParquet uploads a Parquet file to S3.
func (s *Storage) PutParquet(ctx context.Context, key string, data []byte) error {
	return s.put(ctx, key, data, "application/vnd.apache.parquet")
}

// GetParquet downloads a Parquet file from S3.
func (s *Storage) GetParquet(ctx context.Context, key string) ([]byte, error) {
	return s.get(ctx, key)
}

// PutPDF uploads a PDF file to S3.
func (s *Storage) PutPDF(ctx context.Context, key string, data []byte) error {
	return s.put(ctx, key, data, "application/pdf")
}

// GetSignedURL generates a pre-signed URL for downloading an object.
func (s *Storage) GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	out, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to presign URL for %s: %w", key, err)
	}
	return out.URL, nil
}

// PutModel uploads a model artifact to S3.
func (s *Storage) PutModel(ctx context.Context, key string, data []byte) error {
	return s.put(ctx, key, data, "application/octet-stream")
}

// GetModel downloads a model artifact from S3.
func (s *Storage) GetModel(ctx context.Context, key string) ([]byte, error) {
	return s.get(ctx, key)
}

// put is the shared upload helper.
func (s *Storage) put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to put object %s: %w", key, err)
	}
	s.logger.Debug("uploaded object", "key", key, "size", len(data), "contentType", contentType)
	return nil
}

// get is the shared download helper.
func (s *Storage) get(ctx context.Context, key string) ([]byte, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", key, err)
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s: %w", key, err)
	}
	s.logger.Debug("downloaded object", "key", key, "size", len(data))
	return data, nil
}
