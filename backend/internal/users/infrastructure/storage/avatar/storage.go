package avatar

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
	"github.com/google/uuid"
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
		return nil, fmt.Errorf("storage avatar bucket is required")
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

func (s *Storage) UploadAvatar(ctx context.Context, userID uuid.UUID, contentType string, content []byte) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.objectKey(userID)),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		var noSuchBucket *types.NoSuchBucket
		if !errors.As(err, &noSuchBucket) {
			return "", fmt.Errorf("put avatar object: %w", err)
		}

		if err = s.ensureBucket(ctx); err != nil {
			return "", err
		}

		_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.bucket),
			Key:         aws.String(s.objectKey(userID)),
			Body:        bytes.NewReader(content),
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return "", fmt.Errorf("put avatar object after bucket creation: %w", err)
		}
	}

	return s.ManagedAvatarURL(userID), nil
}

func (s *Storage) DeleteAvatar(ctx context.Context, userID uuid.UUID) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objectKey(userID)),
	})
	if err == nil {
		return nil
	}

	var noSuchBucket *types.NoSuchBucket
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchBucket) || errors.As(err, &noSuchKey) {
		return nil
	}

	return fmt.Errorf("delete avatar object: %w", err)
}

func (s *Storage) ManagedAvatarURL(userID uuid.UUID) string {
	return s.publicBaseURL + "/" + s.bucket + "/" + s.objectKey(userID)
}

func (s *Storage) ensureBucket(ctx context.Context) error {
	_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}

	var bucketAlreadyExists *types.BucketAlreadyExists
	var bucketAlreadyOwnedByYou *types.BucketAlreadyOwnedByYou
	if errors.As(err, &bucketAlreadyExists) || errors.As(err, &bucketAlreadyOwnedByYou) {
		return nil
	}

	return fmt.Errorf("create avatar bucket: %w", err)
}

func (s *Storage) objectKey(userID uuid.UUID) string {
	return "user-" + userID.String() + "/avatar"
}
