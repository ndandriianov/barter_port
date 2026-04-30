package s3storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Endpoint        string
	FilerEndpoint   string
	PublicBaseURL   string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

type Storage struct {
	publicBaseURL string
	bucket        string
	filerEndpoint string
	client        *s3.Client
	httpClient    *http.Client
}

func NewStorage(cfg Config) (*Storage, error) {
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	filerEndpoint := strings.TrimRight(cfg.FilerEndpoint, "/")
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
		awsconfig.WithRetryMaxAttempts(1),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	awsCfg.HTTPClient = &http.Client{
		Timeout: 15 * time.Second,
	}
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		o.ContinueHeaderThresholdBytes = -1
	})

	storage := &Storage{
		publicBaseURL: publicBaseURL,
		bucket:        bucket,
		filerEndpoint: defaultFilerEndpoint(endpoint, filerEndpoint),
		client:        client,
		httpClient:    httpClient,
	}

	initCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := storage.ensureBucket(initCtx); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) PutObject(ctx context.Context, key string, contentType string, content []byte) error {
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

func (s *Storage) ReplaceObject(ctx context.Context, key string, contentType string, content []byte) error {
	if err := s.DeleteObject(ctx, key); err != nil {
		return err
	}

	if err := s.PutObject(ctx, key, contentType, content); err != nil {
		return err
	}

	return nil
}

func (s *Storage) CopyObject(ctx context.Context, sourceKey string, destinationKey string) error {
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(s.bucket + "/" + sourceKey),
		Key:        aws.String(destinationKey),
	})
	if err != nil {
		return fmt.Errorf("copy object: %w", err)
	}

	return nil
}

func (s *Storage) CopyManagedObject(ctx context.Context, rawURL string, destinationKey string) error {
	sourceKey, ok := s.ObjectKeyFromManagedURL(rawURL)
	if !ok {
		return fmt.Errorf("managed object url is invalid")
	}

	return s.CopyObject(ctx, sourceKey, destinationKey)
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

func (s *Storage) DeleteManagedObject(ctx context.Context, rawURL string) error {
	key, ok := s.ObjectKeyFromManagedURL(rawURL)
	if !ok {
		return nil
	}

	return s.DeleteObject(ctx, key)
}

func (s *Storage) ManagedObjectURL(key string) string {
	return s.publicBaseURL + "/" + s.bucket + "/" + key
}

func (s *Storage) ObjectKeyFromManagedURL(rawURL string) (string, bool) {
	if strings.TrimSpace(rawURL) == "" {
		return "", false
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}

	managedBaseURL, err := url.Parse(s.publicBaseURL)
	if err != nil {
		return "", false
	}

	if !sameOrigin(managedBaseURL, parsedURL) {
		return "", false
	}

	managedBasePath := strings.Trim(managedBaseURL.EscapedPath(), "/")
	managedObjectPrefix := s.bucket + "/"
	if managedBasePath != "" {
		managedObjectPrefix = managedBasePath + "/" + managedObjectPrefix
	}

	path := strings.Trim(parsedURL.EscapedPath(), "/")
	if !strings.HasPrefix(path, managedObjectPrefix) {
		return "", false
	}

	key := strings.TrimPrefix(path, managedObjectPrefix)
	if key == "" {
		return "", false
	}

	return key, true
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
	if shouldFallbackToFiler(err) {
		if filerErr := s.ensureBucketViaFiler(ctx); filerErr == nil {
			return nil
		} else {
			return fmt.Errorf("create bucket via s3: %w; create bucket via filer: %v", err, filerErr)
		}
	}

	return fmt.Errorf("create bucket: %w", err)
}

func sameOrigin(a *url.URL, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}

func shouldFallbackToFiler(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "SignatureDoesNotMatch")
}

func (s *Storage) ensureBucketViaFiler(ctx context.Context) error {
	if s.filerEndpoint == "" {
		return fmt.Errorf("filer endpoint is not configured")
	}

	bucketURL := s.filerEndpoint + "/buckets/" + url.PathEscape(s.bucket) + "/"

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, bucketURL, nil)
	if err != nil {
		return fmt.Errorf("build filer get bucket request: %w", err)
	}
	getResp, err := s.httpClient.Do(getReq)
	if err != nil {
		return fmt.Errorf("check filer bucket: %w", err)
	}
	io.Copy(io.Discard, getResp.Body)
	getResp.Body.Close()
	if getResp.StatusCode == http.StatusOK {
		return nil
	}
	if getResp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("check filer bucket: unexpected status %d", getResp.StatusCode)
	}

	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, bucketURL, nil)
	if err != nil {
		return fmt.Errorf("build filer create bucket request: %w", err)
	}
	postResp, err := s.httpClient.Do(postReq)
	if err != nil {
		return fmt.Errorf("create filer bucket: %w", err)
	}
	io.Copy(io.Discard, postResp.Body)
	postResp.Body.Close()
	if postResp.StatusCode >= 200 && postResp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("create filer bucket: unexpected status %d", postResp.StatusCode)
}

func defaultFilerEndpoint(s3Endpoint string, configured string) string {
	if configured != "" {
		return configured
	}

	parsed, err := url.Parse(s3Endpoint)
	if err != nil {
		return ""
	}

	host := parsed.Hostname()
	if host == "" {
		return ""
	}

	port := parsed.Port()
	switch port {
	case "8333":
		parsed.Host = host + ":8888"
	default:
		return ""
	}

	parsed.Path = ""
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/")
}
