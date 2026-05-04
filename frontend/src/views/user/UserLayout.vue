<script setup lang="ts">
import { ref } from 'vue'
import { RouterView, RouterLink } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const sidebarOpen = ref(false)

function initials(name: string) {
  return (name || 'U').split(' ').map((w: string) => w[0]).join('').toUpperCase().slice(0, 2)
}
</script>

<template>
  <div class="shell">
    <!-- Overlay mobile -->
    <div :class="['sidebar-overlay', sidebarOpen ? 'open' : '']" @click="sidebarOpen = false"></div>

    <!-- Sidebar -->
    <aside :class="['sidebar-nav', sidebarOpen ? 'open' : '']">
      <div class="sn-brand">
        <div class="sn-brand-icon">
          <svg width="18" height="18" viewBox="0 0 44 44" fill="none">
            <path d="M10 34L22 12L34 34H10Z" fill="white"/>
          </svg>
        </div>
        <span>MH Aprende</span>
      </div>
      <nav>
        <RouterLink to="/usuario/capacitaciones" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
          Mis cursos
        </RouterLink>
        <RouterLink to="/usuario/examenes" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>
          Mis exámenes
        </RouterLink>
      </nav>
      <div class="sn-footer">
        <button @click="auth.logout()">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>
          Cerrar sesión
        </button>
      </div>
    </aside>

    <!-- Main -->
    <div class="shell-main">
      <header class="topbar">
        <button class="topbar-hamburger" @click="sidebarOpen = true" aria-label="Abrir menú">
          <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6h16M4 12h16M4 18h16"/></svg>
        </button>
        <span class="topbar-title">Mis aprendizajes</span>
        <div class="topbar-right">
          <div class="topbar-user">
            <div class="topbar-avatar">{{ initials(auth.user?.name || '') }}</div>
            <span class="topbar-name">{{ auth.user?.name }}</span>
            <span class="badge badge-green">Estudiante</span>
          </div>
        </div>
      </header>
      <main class="page-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
/* all from main.css globals */
</style>
