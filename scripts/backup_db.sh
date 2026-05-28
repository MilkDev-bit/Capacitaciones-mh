#!/usr/bin/env sh
# backup_db.sh — Realiza un pg_dump y lo sube a Cloudflare R2 usando AWS CLI.
#
# Variables de entorno requeridas:
#   DATABASE_URL          — PostgreSQL connection URL (postgres://user:pass@host/db)
#   R2_BUCKET             — Nombre del bucket R2
#   R2_ENDPOINT           — https://<account-id>.r2.cloudflarestorage.com
#   R2_ACCESS_KEY_ID      — Token S3 de R2
#   R2_SECRET_ACCESS_KEY  — Secreto del token S3 de R2
#
# Dependencias: pg_dump, aws CLI (o compatible: s3cmd, rclone)
#
# Uso en Railway/cron: sh scripts/backup_db.sh
# Cron diario a las 3 AM: 0 3 * * * sh /app/scripts/backup_db.sh

set -e

: "${DATABASE_URL:?backup_db.sh: DATABASE_URL no definida}"
: "${R2_BUCKET:?backup_db.sh: R2_BUCKET no definida}"
: "${R2_ENDPOINT:?backup_db.sh: R2_ENDPOINT no definida}"
: "${R2_ACCESS_KEY_ID:?backup_db.sh: R2_ACCESS_KEY_ID no definida}"
: "${R2_SECRET_ACCESS_KEY:?backup_db.sh: R2_SECRET_ACCESS_KEY no definida}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
FILENAME="backup_${TIMESTAMP}.sql.gz"
TMP_FILE="/tmp/${FILENAME}"

echo "[backup] Iniciando pg_dump → ${FILENAME}"
pg_dump "${DATABASE_URL}" | gzip > "${TMP_FILE}"
echo "[backup] Dump completado ($(du -sh "${TMP_FILE}" | cut -f1))"

# Subir a R2 usando AWS CLI con endpoint personalizado
AWS_ACCESS_KEY_ID="${R2_ACCESS_KEY_ID}" \
AWS_SECRET_ACCESS_KEY="${R2_SECRET_ACCESS_KEY}" \
aws s3 cp "${TMP_FILE}" \
  "s3://${R2_BUCKET}/backups_db/${FILENAME}" \
  --endpoint-url "${R2_ENDPOINT}" \
  --no-progress

echo "[backup] Subido a R2: backups_db/${FILENAME}"

# Limpiar archivo temporal
rm -f "${TMP_FILE}"

# Eliminar backups con más de 30 días en R2
CUTOFF=$(date -d '30 days ago' +%Y-%m-%dT%H:%M:%S 2>/dev/null || date -v-30d +%Y-%m-%dT%H:%M:%S)
echo "[backup] Eliminando backups anteriores a ${CUTOFF}..."

AWS_ACCESS_KEY_ID="${R2_ACCESS_KEY_ID}" \
AWS_SECRET_ACCESS_KEY="${R2_SECRET_ACCESS_KEY}" \
aws s3 ls "s3://${R2_BUCKET}/backups_db/" \
  --endpoint-url "${R2_ENDPOINT}" \
  | awk '{print $4}' \
  | while read -r KEY; do
      FILE_DATE=$(echo "${KEY}" | grep -oP '\d{8}' | head -1)
      if [ -n "${FILE_DATE}" ] && [ "${FILE_DATE}" \< "$(date -d '30 days ago' +%Y%m%d 2>/dev/null || date -v-30d +%Y%m%d)" ]; then
        echo "[backup] Eliminando backup antiguo: ${KEY}"
        AWS_ACCESS_KEY_ID="${R2_ACCESS_KEY_ID}" \
        AWS_SECRET_ACCESS_KEY="${R2_SECRET_ACCESS_KEY}" \
        aws s3 rm "s3://${R2_BUCKET}/backups_db/${KEY}" \
          --endpoint-url "${R2_ENDPOINT}"
      fi
    done

echo "[backup] Proceso completado exitosamente."
