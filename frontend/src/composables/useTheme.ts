import { ref } from 'vue'

const isDark = ref(false)

/**
 * Colores de la barra de estado del móvil (el área del notch).
 *
 * Sin un <meta name="theme-color">, Safari en iOS deduce el color muestreando
 * el borde superior de la página. En la tienda eso le hacía tomar el verde de
 * la barra de progreso de lectura (--success, #34c759) y teñir con él toda la
 * barra de estado, dejando ilegibles la hora, la batería y la señal.
 *
 * Declararlo explícitamente quita el margen de interpretación. Y hay que
 * actualizarlo al cambiar de tema: el conmutador de la app no depende de
 * prefers-color-scheme, así que la variante por media query del <head> no
 * bastaría.
 */
const COLOR_BARRA = { claro: '#ffffff', oscuro: '#1d1d1f' }

function sincronizarThemeColor(oscuro: boolean) {
  if (typeof document === 'undefined') return
  let meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = 'theme-color'
    document.head.appendChild(meta)
  }
  meta.content = oscuro ? COLOR_BARRA.oscuro : COLOR_BARRA.claro
}

if (typeof window !== 'undefined') {
  const saved = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  if (saved === 'dark' || (!saved && prefersDark)) {
    document.documentElement.classList.add('dark-theme')
    document.documentElement.classList.remove('light-theme')
    isDark.value = true
  } else {
    document.documentElement.classList.remove('dark-theme')
    document.documentElement.classList.add('light-theme')
    isDark.value = false
  }
  sincronizarThemeColor(isDark.value)
}

export function useTheme() {
  const updateState = () => {
    isDark.value = document.documentElement.classList.contains('dark-theme')
  }

  const toggleTheme = () => {
    const isNowDark = !document.documentElement.classList.contains('dark-theme')
    if (isNowDark) {
      document.documentElement.classList.add('dark-theme')
      document.documentElement.classList.remove('light-theme')
      localStorage.setItem('theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark-theme')
      document.documentElement.classList.add('light-theme')
      localStorage.setItem('theme', 'light')
    }
    updateState()
    sincronizarThemeColor(isNowDark)
  }

  return { isDark, toggleTheme }
}
