// Package media holds the S3-compatible object store client and the SVG
// sanitizer. The media module (internal/modules/media) uses these to store
// uploads in the self-hosted S3 (Coolify/MinIO) bucket and serve them from the
// public base URL — the API never proxies the bytes.
package media

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Store is the object-storage contract the media module depends on.
type Store interface {
	Put(ctx context.Context, key, contentType string, body []byte) error
	Delete(ctx context.Context, key string) error
}

// S3Config configures the S3-compatible client.
type S3Config struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

// S3Store is an S3-compatible object store (self-hosted MinIO / Cloudflare R2).
type S3Store struct {
	client *s3.Client
	bucket string
}

// NewS3Store builds an S3Store from cfg using static credentials and a custom
// endpoint (path-style for MinIO).
func NewS3Store(cfg S3Config) *S3Store {
	client := s3.New(s3.Options{
		Region:       cfg.Region,
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: cfg.ForcePathStyle,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	})
	return &S3Store{client: client, bucket: cfg.Bucket}
}

// Put stores body under key with the given content type. SVGs are served inline;
// the bytes are already sanitized before this call.
func (s *S3Store) Put(ctx context.Context, key, contentType string, body []byte) error {
	disp := "inline"
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(s.bucket),
		Key:                aws.String(key),
		Body:               bytes.NewReader(body),
		ContentType:        aws.String(contentType),
		ContentDisposition: aws.String(disp),
	})
	if err != nil {
		return fmt.Errorf("s3 put %q: %w", key, err)
	}
	return nil
}

// Delete removes key from the bucket.
func (s *S3Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete %q: %w", key, err)
	}
	return nil
}

// PublicURL joins the public base and key into the served URL.
func PublicURL(base, key string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(key, "/")
}
