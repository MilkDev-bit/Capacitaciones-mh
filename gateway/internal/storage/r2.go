package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"Prueba-Go/gateway/internal/config"

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
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".txt":  "text/plain",
	".csv":  "text/csv",
	".zip":  "application/zip",
	".rar":  "application/vnd.rar",
	".7z":   "application/x-7z-compressed",
}

// BlockedExtensions contiene las extensiones consideradas ejecutables o scripts peligrosos.
var BlockedExtensions = map[string]bool{
	".exe": true, ".bat": true, ".cmd": true, ".sh": true, ".ps1": true,
	".vbs": true, ".js": true, ".msc": true, ".scr": true, ".msi": true,
	".apk": true, ".bin": true, ".elf": true, ".py": true, ".rb": true,
	".php": true, ".cgi": true, ".jar": true, ".com": true, ".inf": true,
	".reg": true, ".htm": true, ".html": true, ".svg": true, ".dll": true,
	".sys": true, ".cpl": true, ".hta": true,
}

// StorageService encapsulates the R2 presign client.
type StorageService struct {
	client    *s3.Client
	bucket    string
	publicURL string
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

// defaultSvc is the package-level singleton, created by Init.
var defaultSvc *StorageService

// Init initialises the package-level R2 client from the gateway config.
func Init(cfg *config.Config) {
	if cfg.R2Bucket == "" || cfg.R2Endpoint == "" || cfg.R2AccessKey == "" || cfg.R2SecretKey == "" {
		slog.Warn("R2: una o más variables de entorno no están configuradas; presign y uploads no funcionarán")
		// Create a no-op service so handlers don't panic
		defaultSvc = &StorageService{publicURL: strings.TrimRight(cfg.R2PublicURL, "/")}
		return
	}
	defaultSvc = New(cfg.R2Bucket, cfg.R2Endpoint, cfg.R2AccessKey, cfg.R2SecretKey, cfg.R2PublicURL)

	// Comprobar que la credencial SIRVE, no solo que esté puesta.
	//
	// Antes bastaba con que las cuatro variables no estuvieran vacías y el log
	// decía "R2 inicializado" aunque la llave fuera de otra cuenta, estuviera
	// revocada o el endpoint llevara el bucket duplicado. El fallo no aparecía
	// hasta que alguien subía un archivo, y le llegaba disfrazado de error de
	// CORS: las respuestas de error de R2 no llevan cabeceras CORS, así que el
	// navegador descarta el motivo real y solo dice "blocked by CORS policy".
	//
	// Va en una goroutine y no bloquea el arranque: el resto de la plataforma
	// funciona sin R2, y un healthcheck que espera a Cloudflare es un arranque
	// que falla por algo que no es suyo.
	go verificarCredenciales(cfg.R2Bucket)
}

// verificarCredenciales hace una lectura mínima del bucket y deja el resultado
// en el log.
func verificarCredenciales(bucket string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// ListObjectsV2 con MaxKeys=1 en lugar de HeadBucket: un token con alcance
	// de objetos —el que recomienda Cloudflare— puede no tener permiso sobre el
	// bucket en sí y fallar un HeadBucket aunque las subidas funcionen.
	_, err := defaultSvc.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		slog.Error("R2: las credenciales NO funcionan; todas las subidas van a fallar",
			"bucket", bucket,
			"error", err,
			"revisa", "R2_ENDPOINT lleva solo la cuenta (https://<cuenta>.r2.cloudflarestorage.com), sin el bucket. "+
				"Si el error dice AccessDenied, el token de Cloudflare es de solo lectura. "+
				"Hay un diagnóstico completo en GET /api/admin/diagnostico/r2.")
		return
	}
	slog.Info("R2 inicializado", "bucket", bucket)
}

// Default returns the package-level storage service.
func Default() *StorageService { return defaultSvc }

// PublicURL devuelve la base pública del bucket, sin barra final.
//
// Se expone para poder comprobar que una URL recibida del cliente apunta a
// nuestro almacenamiento antes de que el servidor la descargue. Sin esa
// comprobación, cualquiera podría hacer que el gateway pida una dirección
// arbitraria —incluidos servicios internos que no son alcanzables desde fuera—.
func (s *StorageService) PublicURL() string {
	if s == nil {
		return ""
	}
	return s.publicURL
}

// Configurado indica si hay credenciales con las que hablar con R2.
//
// `Init` crea un servicio inerte cuando faltan variables, para que los handlers
// no revienten con un puntero nulo. Esto distingue ese caso del real.
func (s *StorageService) Configurado() bool {
	return s != nil && s.client != nil && s.bucket != ""
}

// Bucket devuelve el nombre del bucket configurado. Para diagnóstico.
func (s *StorageService) Bucket() string {
	if s == nil {
		return ""
	}
	return s.bucket
}

// DeleteFile borra un objeto del bucket.
//
// Lo usa la prueba de escritura para no dejar basura. No se expone en ninguna
// ruta de usuario: aquí nada borra archivos subidos, porque una constancia o
// una entrega tienen que seguir estando el día que alguien las reclame.
func (s *StorageService) DeleteFile(ctx context.Context, key string) error {
	if !s.Configurado() {
		return fmt.Errorf("R2 no configurado")
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
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
	if s.client == nil {
		return "", "", fmt.Errorf("R2 no configurado")
	}
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

// UploadFile uploads a reader directly to R2.
func (s *StorageService) UploadFile(ctx context.Context, key, contentType string, body io.Reader, size int64) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("R2 no configurado")
	}
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

// UploadMultipart uploads a multipart file to R2.
func (s *StorageService) UploadMultipart(ctx context.Context, fh *multipart.FileHeader, prefix string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("R2 no configurado")
	}
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
	if BlockedExtensions[ext] {
		return "", fmt.Errorf("no se permiten archivos ejecutables o scripts")
	}
	ct, ok := MimeTypes[ext]
	if !ok {
		ct = "application/octet-stream"
	}
	year := time.Now().Format("2006")
	key := fmt.Sprintf("%s/%s/%s%s", prefix, year, uuid.NewString(), ext)
	return s.UploadFile(ctx, key, ct, f, fh.Size)
}

// --- Package-level wrappers ---

func UploadMultipart(ctx context.Context, fh *multipart.FileHeader, prefix string) (string, error) {
	return defaultSvc.UploadMultipart(ctx, fh, prefix)
}

func GeneratePresignedURL(ctx context.Context, prefix, ext, contentType string, ttl time.Duration) (string, string, error) {
	return defaultSvc.GeneratePresignedURL(ctx, prefix, ext, contentType, ttl)
}
