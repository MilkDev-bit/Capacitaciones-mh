# CORS del bucket R2

## El síntoma

Subir cualquier archivo desde el navegador falla con:

```
Access to fetch at 'https://capacitacionesmh.<cuenta>.r2.cloudflarestorage.com/...?x-id=PutObject'
from origin 'https://capacitaciones.mhsolucionesempresariales.com'
has been blocked by CORS policy: No 'Access-Control-Allow-Origin' header is present
```

Afecta a **todas** las subidas, no solo a las portadas: avatares, portadas de
curso, vídeos y documentos de lección, adjuntos de mensajes, entregas de
actividad y el logotipo de empresa de la constancia DC-3. Todas pasan por
`uploadToR2()`.

## Por qué pasa

El navegador sube directo a R2 con una URL prefirmada que emite
`GET /api/presign`. El archivo **nunca atraviesa el gateway**: eso es
deliberado, porque un vídeo de 100 MB pasando por el gateway ocuparía su
memoria y su ancho de banda sin aportar nada.

La consecuencia es que quien tiene que autorizar al navegador es R2, no el
gateway. `ALLOWED_ORIGIN` del `.env` configura el CORS **del gateway** y no
tiene ningún efecto sobre el bucket.

Y la petición siempre dispara una comprobación previa (preflight `OPTIONS`):
es un `PUT`, y el método `PUT` nunca se considera simple. Quitar la cabecera
`Content-Type` del `fetch` no lo evitaría.

## El arreglo

No es código. Es la política CORS del bucket, en Cloudflare:

**R2 → el bucket → Settings → CORS Policy → Edit**

```json
[
  {
    "AllowedOrigins": [
      "https://capacitaciones.mhsolucionesempresariales.com",
      "http://localhost:5173"
    ],
    "AllowedMethods": ["PUT", "GET", "HEAD"],
    "AllowedHeaders": ["content-type"],
    "ExposeHeaders": ["ETag"],
    "MaxAgeSeconds": 3600
  }
]
```

Por qué cada parte:

- **AllowedOrigins** — el dominio de producción y el de Vite en desarrollo. R2
  compara la cadena exacta: sin puerto no vale, y `https` y `http` son orígenes
  distintos. Si algún día hay un dominio de pruebas, va aquí también.
- **AllowedMethods** — `PUT` es el que usan las subidas. `GET` y `HEAD` solo
  hacen falta si algo lee el archivo con `fetch`; las etiquetas `<img>` y
  `<video>` no necesitan CORS.
- **AllowedHeaders** — `content-type` es la única cabecera que `uploadToR2()`
  añade, y es justo la que el preflight pregunta.
- **MaxAgeSeconds** — cachea el preflight una hora. Sin esto, cada subida son
  dos viajes a Cloudflare en vez de uno.

Los cambios de CORS en R2 tardan un momento en propagarse. Si después de
guardarlo sigue fallando, recarga sin caché: el navegador recuerda el preflight
fallido.

## Nota aparte sobre el `Content-Type`

La URL prefirmada lleva `X-Amz-SignedHeaders=host`: la cabecera `Content-Type`
**no va firmada**, aunque el backend la calcule en `GeneratePresignedURL`. Eso
significa que el navegador puede subir el archivo declarando el tipo que quiera,
y R2 lo devolverá tal cual al servirlo.

La extensión sí está acotada por el backend —`storage.MimeTypes` y la lista de
prefijos permitidos—, así que la clave del objeto siempre acaba en `.jpeg`,
`.pdf` o lo que toque. Lo que no está acotado es el tipo declarado.

El alcance es limitado: el bucket vive en otro origen que la aplicación, así que
no comparte cookies ni sesión. Pero conviene saberlo antes de servir ese bucket
desde un subdominio propio, donde sí pasaría a compartir origen.
