package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/akza/akza-api/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// Client wraps the AWS S3 client with project-specific helpers.
type Client struct {
	s3      *s3.Client
	presign *s3.PresignClient
	bucket  string
	cdnBase string
}

// New creates an S3 client pointed at Yandex Object Storage (or any S3-compatible endpoint).
func New(cfg config.S3Config) (*Client, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{URL: cfg.Endpoint, SigningRegion: cfg.Region}, nil
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("s3 config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &Client{
		s3:      client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
		cdnBase: strings.TrimRight(cfg.CDNBase, "/"),
	}, nil
}

// BuildKey generates a namespaced S3 object key.
// Pattern: {entity}/{entityID}/{uuid}.{ext}
func BuildKey(entity, entityID, filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}
	return fmt.Sprintf("%s/%s/%s%s", entity, entityID, uuid.New().String(), ext)
}

// PublicURL returns the CDN/public URL for a given S3 key.
func (c *Client) PublicURL(key string) string {
	return c.cdnBase + "/" + key
}

// PresignPut generates a presigned PUT URL for direct browser uploads.
func (c *Client) PresignPut(ctx context.Context, key, contentType string) (string, error) {
	req, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}
	return req.URL, nil
}

// DeleteObject removes an object from S3 by key.
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}
