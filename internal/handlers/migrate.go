package handlers

import (
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"Prueba-Go/internal/db"
	"Prueba-Go/internal/storage"

	"github.com/gin-gonic/gin"
)

// migrateTarget describes a table column that may hold local /uploads/ paths.
type migrateTarget struct {
	table  string
	column string
}

var migrateTargets = []migrateTarget{
	{"capacitaciones", "thumbnail_url"},
	{"capacitaciones", "file_path"},
	{"lecciones", "file_path"},
	{"users", "avatar_url"},
	{"users", "cover_url"},
}

type migrateResult struct {
	Table   string `json:"table"`
	Column  string `json:"column"`
	ID      string `json:"id"`
	OldPath string `json:"old_path"`
	NewURL  string `json:"new_url,omitempty"`
	Error   string `json:"error,omitempty"`
}

// MigrateLocalToR2 scans every table/column that can hold file paths and
// uploads any /uploads/* local files to R2, then updates the DB record.
// Protected by AdminRequired middleware. Safe to call multiple times
// (records already pointing to R2 URLs won't match the LIKE filter).
func MigrateLocalToR2(c *gin.Context) {
	var results []migrateResult
	migratedCount := 0
	errorCount := 0

	for _, t := range migrateTargets {
		// table and column are hardcoded — not user input, no injection risk.
		query := `SELECT id, ` + t.column + ` FROM ` + t.table +
			` WHERE ` + t.column + ` LIKE '/uploads/%'`

		rows, err := db.DB.Query(query)
		if err != nil {
			slog.Error("MigrateLocalToR2: query", "table", t.table, "column", t.column, "error", err)
			results = append(results, migrateResult{
				Table:  t.table,
				Column: t.column,
				Error:  "query: " + err.Error(),
			})
			errorCount++
			continue
		}

		var pending []migrateResult
		for rows.Next() {
			var id, path string
			if err := rows.Scan(&id, &path); err != nil {
				continue
			}
			pending = append(pending, migrateResult{Table: t.table, Column: t.column, ID: id, OldPath: path})
		}
		rows.Close()

		for _, p := range pending {
			r := uploadLocalFile(c, t.table, t.column, p.ID, p.OldPath)
			results = append(results, r)
			if r.Error != "" {
				errorCount++
			} else {
				migratedCount++
			}
		}
	}

	status := http.StatusOK
	if errorCount > 0 && migratedCount == 0 {
		status = http.StatusInternalServerError
	}
	c.JSON(status, gin.H{
		"migrated": migratedCount,
		"errors":   errorCount,
		"results":  results,
	})
}

func uploadLocalFile(c *gin.Context, table, column, id, localPath string) migrateResult {
	r := migrateResult{Table: table, Column: column, ID: id, OldPath: localPath}

	// localPath = /uploads/thumbnails/uuid.png
	// R2 key   = thumbnails/uuid.png
	trimmed := strings.TrimPrefix(localPath, "/uploads/")
	if trimmed == localPath {
		r.Error = "ruta no empieza con /uploads/"
		return r
	}

	diskPath := filepath.Join(".", "uploads", filepath.FromSlash(trimmed))
	f, err := os.Open(diskPath)
	if err != nil {
		slog.Warn("MigrateLocalToR2: archivo no encontrado", "path", diskPath)
		r.Error = "archivo no encontrado: " + err.Error()
		return r
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		r.Error = "stat: " + err.Error()
		return r
	}

	ext := strings.ToLower(filepath.Ext(localPath))
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		ct = "application/octet-stream"
	}

	// Use forward-slash key for R2 regardless of OS.
	key := strings.ReplaceAll(trimmed, string(filepath.Separator), "/")

	newURL, err := storage.UploadFile(c.Request.Context(), key, ct, f, info.Size())
	if err != nil {
		slog.Error("MigrateLocalToR2: upload R2", "key", key, "error", err)
		r.Error = "upload: " + err.Error()
		return r
	}

	if _, err := db.DB.Exec(
		`UPDATE `+table+` SET `+column+`=$1 WHERE id=$2`,
		newURL, id,
	); err != nil {
		slog.Error("MigrateLocalToR2: update DB", "table", table, "id", id, "error", err)
		r.Error = "update db: " + err.Error()
		return r
	}

	slog.Info("MigrateLocalToR2: migrado", "table", table, "column", column, "id", id, "key", key)
	r.NewURL = newURL
	return r
}
