// Utilidad para reCAPTCHA v3
// Carga el script de Google de forma lazy (solo cuando se necesita).
// Requiere la variable de entorno VITE_RECAPTCHA_SITE_KEY en Railway.

const SITE_KEY = import.meta.env.VITE_RECAPTCHA_SITE_KEY as string | undefined

let scriptLoaded = false

function loadScript(): Promise<void> {
  if (scriptLoaded || !SITE_KEY) return Promise.resolve()
  return new Promise((resolve) => {
    const existing = document.querySelector('script[data-recaptcha]')
    if (existing) { scriptLoaded = true; resolve(); return }
    const script = document.createElement('script')
    script.src = `https://www.google.com/recaptcha/api.js?render=${SITE_KEY}`
    script.setAttribute('data-recaptcha', 'true')
    script.onload = () => { scriptLoaded = true; resolve() }
    document.head.appendChild(script)
  })
}

/**
 * Ejecuta reCAPTCHA v3 y devuelve el token para enviarlo al backend.
 * Si VITE_RECAPTCHA_SITE_KEY no está configurado, devuelve '' (útil en dev).
 */
export async function getRecaptchaToken(action: string): Promise<string> {
  if (!SITE_KEY) return ''
  try {
    await loadScript()
    return await new Promise<string>((resolve) => {
      ;(window as any).grecaptcha.ready(() => {
        ;(window as any).grecaptcha.execute(SITE_KEY, { action }).then(resolve)
      })
    })
  } catch {
    return ''
  }
}
