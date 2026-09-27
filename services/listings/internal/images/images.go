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
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

const MaxBytes int64 = 10 << 20

var allowed = map[string]string{"image/jpeg": "jpg", "image/png": "png", "image/webp": "webp"}

type ObjectStore interface {
	PresignPut(context.Context, string, string, int64) (UploadTarget, error)
	Head(context.Context, string) error
	MarkRegistered(context.Context, string) error
}

// UploadTarget carries headers that must accompany the browser's direct PUT.
// S3 signs the unregistered tag so its lifecycle rule can remove abandoned
// uploads without ever deleting a registered listing image.
type UploadTarget struct {
	URL     string            `json:"upload_url"`
	Headers map[string]string `json:"upload_headers,omitempty"`
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
func (s *S3Store) PresignPut(ctx context.Context, key, contentType string, size int64) (UploadTarget, error) {
	req, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType), ContentLength: aws.Int64(size), Tagging: aws.String("state=unregistered")}, s3.WithPresignExpires(5*time.Minute))
	if err != nil {
		return UploadTarget{}, err
	}
	return UploadTarget{URL: req.URL, Headers: map[string]string{"x-amz-tagging": "state=unregistered"}}, nil
}
func (s *S3Store) Head(ctx context.Context, key string) error {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return err
}

func (s *S3Store) MarkRegistered(ctx context.Context, key string) error {
	_, err := s.client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Tagging: &types.Tagging{TagSet: []types.Tag{
			{Key: aws.String("state"), Value: aws.String("registered")},
		}},
	})
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
