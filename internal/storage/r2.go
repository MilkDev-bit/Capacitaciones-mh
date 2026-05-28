package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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

// dangerousMIME holds detected types that indicate potentially malicious content.
var dangerousMIME = map[string]bool{
	"text/html":              true,
	"text/javascript":        true,
	"application/javascript": true,
	"text/xml":               true,
	"application/xml":        true,
	"application/xhtml+xml":  true,
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
		slog.Warn("una o más variables R2 no están configuradas")
		return
	}
	slog.Info("R2 inicializado", "bucket", bucket, "endpoint", endpoint)
	go configureBucketCORS()
}

func configureBucketCORS() {
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := client.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: aws.String(bucket),
		CORSConfiguration: &s3types.CORSConfiguration{
			CORSRules: []s3types.CORSRule{
				{
					AllowedOrigins: []string{allowedOrigin},
					AllowedMethods: []string{"PUT", "GET", "HEAD"},
					AllowedHeaders: []string{"content-type", "cache-control"},
					MaxAgeSeconds:  aws.Int32(86400),
				},
			},
		},
	})
	if err != nil {
		slog.Warn("R2: no se pudo configurar CORS", "error", err)
	} else {
		slog.Info("R2: CORS configurado", "origin", allowedOrigin)
	}
}

func UploadFile(ctx context.Context, key, contentType string, body io.Reader, size int64) (string, error) {
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
		CacheControl:  aws.String("public, max-age=31536000"),
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

	// Detect content type from first 512 bytes to reject dangerous files.
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if dangerousMIME[http.DetectContentType(buf[:n])] {
		return "", fmt.Errorf("tipo de archivo no permitido")
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	ct, ok := mimeTypes[ext]
	if !ok {
		ct = "application/octet-stream"
	}
	year := time.Now().Format("2006")
	key := fmt.Sprintf("%s/%s/%s%s", prefix, year, uuid.NewString(), ext)
	return UploadFile(ctx, key, ct, f, fh.Size)
}

// GeneratePresignedURL returns a time-limited PUT URL for direct client-to-R2 uploads
// and the permanent public URL the file will have once uploaded.
func GeneratePresignedURL(ctx context.Context, prefix, ext string, ttl time.Duration) (uploadURL, finalURL string, err error) {
	year := time.Now().Format("2006")
	key := fmt.Sprintf("%s/%s/%s%s", prefix, year, uuid.NewString(), ext)
	pc := s3.NewPresignClient(client)
	req, presignErr := pc.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if presignErr != nil {
		return "", "", presignErr
	}
	return req.URL, publicURL + "/" + key, nil
}

// GeneratePresignedGetURL returns a time-limited GET URL for serving private R2 objects.
func GeneratePresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	pc := s3.NewPresignClient(client)
	req, err := pc.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// ExtractKeyFromURL extracts the R2 object key from a full public URL.
func ExtractKeyFromURL(fileURL string) string {
	return strings.TrimPrefix(fileURL, publicURL+"/")
}
