import api from '../api'

/**
 * Uploads a file directly to Cloudflare R2 using a presigned URL.
 * The backend generates a one-time PUT URL valid for 15 minutes.
 * Returns the permanent public URL of the uploaded file.
 */
export async function uploadToR2(file: File, prefix: string): Promise<string> {
  const ext = '.' + (file.name.split('.').pop()?.toLowerCase() ?? 'bin')
  const { data } = await api.get('/presign', { params: { prefix, ext } })

  await fetch(data.upload_url, {
    method: 'PUT',
    body: file,
    headers: {
      'Content-Type': file.type || 'application/octet-stream',
    },
  })

  return data.final_url as string
}
