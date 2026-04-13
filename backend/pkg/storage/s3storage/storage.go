package s3storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Endpoint        string
	PublicBaseURL   string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

type Storage struct {
	publicBaseURL string
	bucket        string
	client        *s3.Client
}

func NewStorage(cfg Config) (*Storage, error) {
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	publicBaseURL := strings.TrimRight(cfg.PublicBaseURL, "/")
	bucket := strings.Trim(strings.TrimSpace(cfg.Bucket), "/")
	accessKeyID := strings.TrimSpace(cfg.AccessKeyID)
	secretAccessKey := strings.TrimSpace(cfg.SecretAccessKey)
	region := strings.TrimSpace(cfg.Region)

	if endpoint == "" {
		return nil, fmt.Errorf("storage endpoint is required")
	}
	if publicBaseURL == "" {
		return nil, fmt.Errorf("storage public base url is required")
	}
	if bucket == "" {
		return nil, fmt.Errorf("storage bucket is required")
	}
	if accessKeyID == "" {
		return nil, fmt.Errorf("storage access key id is required")
	}
	if secretAccessKey == "" {
		return nil, fmt.Errorf("storage secret access key is required")
	}
	if region == "" {
		region = "us-east-1"
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		o.ContinueHeaderThresholdBytes = -1
	})

	return &Storage{
		publicBaseURL: publicBaseURL,
		bucket:        bucket,
		client:        client,
	}, nil
}

func (s *Storage) PutObject(ctx context.Context, key string, contentType string, content []byte) error {
	if err := s.ensureBucket(ctx); err != nil {
		return err
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}

func (s *Storage) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return nil
	}

	var noSuchKey *types.NoSuchKey
	if _, ok := errors.AsType[*types.NoSuchBucket](err); ok || errors.As(err, &noSuchKey) {
		return nil
	}

	return fmt.Errorf("delete object: %w", err)
}

func (s *Storage) ManagedObjectURL(key string) string {
	return s.publicBaseURL + "/" + s.bucket + "/" + key
}

func (s *Storage) ensureBucket(ctx context.Context) error {
	_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}

	var bucketAlreadyOwnedByYou *types.BucketAlreadyOwnedByYou
	if _, ok := errors.AsType[*types.BucketAlreadyExists](err); ok || errors.As(err, &bucketAlreadyOwnedByYou) {
		return nil
	}

	return fmt.Errorf("create bucket: %w", err)
}
