import { ref, onMounted } from 'vue'

export function useTheme() {
  const isDark = ref(false)

  const updateState = () => {
    isDark.value = document.documentElement.classList.contains('dark-theme') || 
      (window.matchMedia('(prefers-color-scheme: dark)').matches && !document.documentElement.classList.contains('light-theme'))
  }

  const toggleTheme = () => {
    const html = document.documentElement
    updateState()
    if (isDark.value) {
      html.classList.remove('dark-theme')
      html.classList.add('light-theme')
      localStorage.setItem('theme', 'light')
      isDark.value = false
    } else {
      html.classList.remove('light-theme')
      html.classList.add('dark-theme')
      localStorage.setItem('theme', 'dark')
      isDark.value = true
    }
  }

  onMounted(() => {
    const saved = localStorage.getItem('theme')
    if (saved === 'dark') {
      document.documentElement.classList.add('dark-theme')
      document.documentElement.classList.remove('light-theme')
    } else if (saved === 'light') {
      document.documentElement.classList.add('light-theme')
      document.documentElement.classList.remove('dark-theme')
    }
    updateState()
  })

  return { isDark, toggleTheme }
}
