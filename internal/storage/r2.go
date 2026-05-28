package storage

import (
	"context"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var (
	client    *s3.Client
	bucket    string
	publicURL string
)

var mimeTypes = map[string]string{
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

func Init() {
	bucket = os.Getenv("R2_BUCKET")
	publicURL = strings.TrimRight(os.Getenv("R2_PUBLIC_URL"), "/")
	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")

	client = s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		Region:       "auto",
	})

	if bucket == "" || endpoint == "" || accessKey == "" || secretKey == "" {
		log.Println("[storage] ADVERTENCIA: una o más variables R2 no están configuradas")
	} else {
		log.Printf("[storage] R2 inicializado — bucket: %s, endpoint: %s", bucket, endpoint)
	}
}

func UploadFile(ctx context.Context, key, contentType string, body io.Reader, size int64) (string, error) {
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return "", err
	}
	return publicURL + "/" + key, nil
}

func UploadMultipart(ctx context.Context, fh *multipart.FileHeader, prefix string) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	ct, ok := mimeTypes[ext]
	if !ok {
		ct = "application/octet-stream"
	}
	key := prefix + "/" + uuid.NewString() + ext
	return UploadFile(ctx, key, ct, f, fh.Size)
}
