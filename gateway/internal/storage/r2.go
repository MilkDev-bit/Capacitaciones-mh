package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// MimeTypes maps file extensions to their MIME content-type.
var MimeTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".gif":  "image/gif",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

// StorageService encapsulates the R2 presign client.
type StorageService struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// New creates a StorageService.
func New(bucket, endpoint, accessKey, secretKey, publicURL string) *StorageService {
	c := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		Region:       "auto",
	})
	return &StorageService{
		client:    c,
		bucket:    bucket,
		publicURL: strings.TrimRight(publicURL, "/"),
	}
}

// GeneratePresignedURL returns a time-limited PUT URL for direct client-to-R2 uploads.
func (s *StorageService) GeneratePresignedURL(ctx context.Context, prefix, ext, contentType string, ttl time.Duration) (uploadURL, finalURL string, err error) {
	year := time.Now().Format("2006")
	key := fmt.Sprintf("%s/%s/%s%s", prefix, year, uuid.NewString(), ext)
	pc := s3.NewPresignClient(s.client)
	req, presignErr := pc.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(ttl))
	if presignErr != nil {
		return "", "", presignErr
	}
	return req.URL, s.publicURL + "/" + key, nil
}
