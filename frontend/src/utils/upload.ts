import api from '../api'

/**
 * Sube un archivo directo a Cloudflare R2 con una URL prefirmada.
 *
 * El archivo NO pasa por el gateway: un vídeo de 250 MB atravesándolo ocuparía
 * su memoria y su ancho de banda sin aportar nada. El backend solo emite el
 * permiso de escritura y el navegador escribe contra R2.
 *
 * Devuelve la URL pública y permanente del archivo.
 */

/** Avance de la subida. `porcentaje` es null mientras el total sea desconocido. */
export interface AvanceSubida {
  enviados: number
  total: number
  porcentaje: number | null
}

export async function uploadToR2(
  file: File,
  prefix: string,
  onProgress?: (a: AvanceSubida) => void,
): Promise<string> {
  const ext = '.' + (file.name.split('.').pop()?.toLowerCase() ?? 'bin')
  const { data } = await api.get('/presign', { params: { prefix, ext } })

  // XMLHttpRequest y no fetch, aunque fetch sea lo moderno.
  //
  // `fetch` NO expone el progreso de SUBIDA: no hay forma de saber si un vídeo
  // de 250 MB va por el 2% o por el 99%. Con la barra en "Subiendo…" fija, una
  // subida lenta y una subida colgada se ven exactamente igual, y lo único que
  // puede hacer quien espera es cancelarla y empezar de cero.
  //
  // XHR sí emite `upload.onprogress`. Es la única razón de usarlo aquí.
  return new Promise<string>((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('PUT', data.upload_url, true)
    xhr.setRequestHeader('Content-Type', file.type || 'application/octet-stream')

    if (onProgress) {
      xhr.upload.onprogress = (e) => {
        onProgress({
          enviados: e.loaded,
          total: e.total,
          // lengthComputable es false cuando el navegador no conoce el total.
          // Se devuelve null en vez de 0 para que quien lo pinte pueda
          // distinguir "no empezó" de "no se sabe".
          porcentaje: e.lengthComputable ? Math.round((e.loaded / e.total) * 100) : null,
        })
      }
    }

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(data.final_url as string)
        return
      }
      // 403 tras una subida larga casi siempre es la URL prefirmada caducada,
      // no un problema de permisos. Decirlo evita mandar a nadie a revisar
      // credenciales que están bien.
      if (xhr.status === 403) {
        reject(new Error(
          'La autorización de subida caducó antes de terminar. ' +
          'Suele pasar con archivos grandes en conexiones lentas: vuelve a intentarlo.',
        ))
        return
      }
      reject(new Error(`Error al subir archivo (${xhr.status})`))
    }

    xhr.onerror = () => reject(new Error('Se interrumpió la conexión durante la subida.'))
    xhr.onabort = () => reject(new Error('Subida cancelada.'))

    xhr.send(file)
  })
}

/** Formatea bytes para mostrárselos a una persona. */
export function tamanoLegible(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}
