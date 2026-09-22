package images

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const MaxBytes int64 = 10 << 20

var allowed = map[string]string{"image/jpeg": "jpg", "image/png": "png", "image/webp": "webp"}

type ObjectStore interface {
	PresignPut(context.Context, string, string, int64) (string, error)
	Head(context.Context, string) error
}

type S3Store struct {
	bucket  string
	client  *s3.Client
	presign *s3.PresignClient
}

func NewS3Store(ctx context.Context, region, bucket string) (*S3Store, error) {
	if region == "" || bucket == "" {
		return nil, fmt.Errorf("S3_REGION and S3_BUCKET are required")
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3Store{bucket: bucket, client: client, presign: s3.NewPresignClient(client)}, nil
}
func (s *S3Store) PresignPut(ctx context.Context, key, contentType string, size int64) (string, error) {
	req, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType), ContentLength: aws.Int64(size)}, s3.WithPresignExpires(5*time.Minute))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
func (s *S3Store) Head(ctx context.Context, key string) error {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return err
}
func NewKey(listingID uuid.UUID, contentType string) (string, error) {
	ext, ok := allowed[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image content type")
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return path.Join("listings", listingID.String(), id.String()+"."+ext), nil
}
func Validate(contentType string, size int64) error {
	if _, ok := allowed[strings.ToLower(contentType)]; !ok {
		return fmt.Errorf("unsupported image content type")
	}
	if size < 1 || size > MaxBytes {
		return fmt.Errorf("image must be between 1 byte and 10 MB")
	}
	return nil
}
