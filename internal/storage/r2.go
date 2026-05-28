package storage

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/google/uuid"
)

// MimeTypes maps file extensions to their MIME content-type.
// Exported so callers (e.g. presign handler) can resolve the correct type.
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

// dangerousMIME holds detected types that indicate potentially malicious content.
var dangerousMIME = map[string]bool{
	"text/html":              true,
	"text/javascript":        true,
	"application/javascript": true,
	"text/xml":               true,
	"application/xml":        true,
	"application/xhtml+xml":  true,
}

// StorageService encapsulates the R2 client and bucket configuration.
// Use Init for the package-level singleton or New to create an isolated instance
// (e.g. in unit tests with a mock S3 client).
type StorageService struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// defaultSvc is the package-level singleton, created by Init.
var defaultSvc *StorageService

// New creates a StorageService from an existing S3 client. Intended for testing.
func New(client *s3.Client, bucket, publicURL string) *StorageService {
	return &StorageService{client: client, bucket: bucket, publicURL: publicURL}
}

func Init() {
	bkt := os.Getenv("R2_BUCKET")
	pubURL := strings.TrimRight(os.Getenv("R2_PUBLIC_URL"), "/")
	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")

	c := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		Region:       "auto",
	})
	defaultSvc = &StorageService{client: c, bucket: bkt, publicURL: pubURL}

	if bkt == "" || endpoint == "" || accessKey == "" || secretKey == "" {
		slog.Warn("una o más variables R2 no están configuradas")
		return
	}
	slog.Info("R2 inicializado", "bucket", bkt, "endpoint", endpoint)
	go configureBucketCORS()
}

func configureBucketCORS() {
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	// Prefer Cloudflare's own API — requires CF_ACCOUNT_ID + CF_API_TOKEN with
	// "Cloudflare R2 Storage: Edit" permission. More reliable than S3 PutBucketCors
	// which requires an "Admin Read & Write" S3 token.
	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")
	if accountID != "" && apiToken != "" {
		if err := configureCORSViaCloudflareAPI(accountID, apiToken, allowedOrigin); err != nil {
			slog.Error("R2: CORS vía Cloudflare API falló", "error", err)
		} else {
			slog.Info("R2: CORS configurado vía Cloudflare API", "origin", allowedOrigin)
		}
		return
	}

	slog.Warn("R2: CF_ACCOUNT_ID y CF_API_TOKEN no configurados — configura CORS manualmente en el dashboard de Cloudflare R2 o agrega esas variables de entorno")
}

// configureCORSViaCloudflareAPI uses Cloudflare's REST API to set the bucket CORS
// policy. The API token must have the 'Cloudflare R2 Storage: Edit' permission.
func configureCORSViaCloudflareAPI(accountID, apiToken, allowedOrigin string) error {
	type corsAllowed struct {
		Origins []string `json:"origins"`
		Methods []string `json:"methods"`
		Headers []string `json:"headers"`
	}
	type corsRule struct {
		Allowed       corsAllowed `json:"allowed"`
		MaxAgeSeconds int         `json:"maxAgeSeconds"`
	}
	type corsPayload struct {
		Rules []corsRule `json:"rules"`
	}

	payload := corsPayload{
		Rules: []corsRule{
			{
				Allowed: corsAllowed{
					Origins: []string{allowedOrigin},
					Methods: []string{"PUT", "GET", "HEAD"},
					Headers: []string{"content-type"},
				},
				MaxAgeSeconds: 86400,
			},
		},
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/r2/buckets/%s/cors",
		accountID, defaultSvc.bucket)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var respBody bytes.Buffer
		respBody.ReadFrom(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, respBody.String())
	}
	return nil
}

// --- StorageService methods ---

func (s *StorageService) UploadFile(ctx context.Context, key, contentType string, body io.Reader, size int64) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
		CacheControl:  aws.String("public, max-age=31536000"),
	})
	if err != nil {
		return "", err
	}
	return s.publicURL + "/" + key, nil
}

func (s *StorageService) UploadMultipart(ctx context.Context, fh *multipart.FileHeader, prefix string) (string, error) {
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
	ct, ok := MimeTypes[ext]
	if !ok {
		ct = "application/octet-stream"
	}
	year := time.Now().Format("2006")
	key := fmt.Sprintf("%s/%s/%s%s", prefix, year, uuid.NewString(), ext)
	return s.UploadFile(ctx, key, ct, f, fh.Size)
}

// GeneratePresignedURL returns a time-limited PUT URL for direct client-to-R2 uploads.
// contentType is locked into the signed URL so R2 rejects uploads with a different MIME type,
// preventing an attacker from substituting the MIME type after the URL is issued.
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

// GeneratePresignedGetURL returns a time-limited GET URL for serving private R2 objects.
func (s *StorageService) GeneratePresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	pc := s3.NewPresignClient(s.client)
	req, err := pc.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// DeleteFile removes an object from R2. Call this when a course, lesson, or profile
// asset is deleted to avoid orphaned files accruing storage costs.
func (s *StorageService) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// ExtractKeyFromURL extracts the R2 object key from a full public URL.
func (s *StorageService) ExtractKeyFromURL(fileURL string) string {
	return strings.TrimPrefix(fileURL, s.publicURL+"/")
}

// --- Package-level wrappers (backward-compatible public API) ---

func UploadFile(ctx context.Context, key, contentType string, body io.Reader, size int64) (string, error) {
	return defaultSvc.UploadFile(ctx, key, contentType, body, size)
}

func UploadMultipart(ctx context.Context, fh *multipart.FileHeader, prefix string) (string, error) {
	return defaultSvc.UploadMultipart(ctx, fh, prefix)
}

// GeneratePresignedURL requires contentType to be locked into the signed URL,
// preventing an attacker from substituting the MIME type on upload.
func GeneratePresignedURL(ctx context.Context, prefix, ext, contentType string, ttl time.Duration) (string, string, error) {
	return defaultSvc.GeneratePresignedURL(ctx, prefix, ext, contentType, ttl)
}

func GeneratePresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	return defaultSvc.GeneratePresignedGetURL(ctx, key, ttl)
}

// DeleteFile removes an R2 object by key.
func DeleteFile(ctx context.Context, key string) error {
	return defaultSvc.DeleteFile(ctx, key)
}

// ExtractKeyFromURL extracts the R2 object key from a full public URL.
func ExtractKeyFromURL(fileURL string) string {
	return defaultSvc.ExtractKeyFromURL(fileURL)
}
