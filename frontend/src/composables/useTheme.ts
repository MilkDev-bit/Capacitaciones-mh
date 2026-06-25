import { ref } from 'vue'

const isDark = ref(false)

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
  }

  return { isDark, toggleTheme }
}
