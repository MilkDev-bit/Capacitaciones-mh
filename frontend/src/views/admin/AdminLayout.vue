<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterView, RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useTheme } from '../../composables/useTheme'
import NotificationBell from '../../components/NotificationBell.vue'
import { getAvatarUrl } from '../../utils/avatars'

const auth = useAuthStore()
const router = useRouter()
const { isDark, toggleTheme } = useTheme()
const sidebarOpen = ref(false)
const profileOpen = ref(false)

const isUUID = (str: string) => /^[0-9a-fA-F-]{32,36}$/.test(str) || str.length > 25
const breadcrumbs = computed(() => {
  const parts = router.currentRoute.value.path.split('/').filter(Boolean)
  return parts.map((p, idx) => {
    let name = p.charAt(0).toUpperCase() + p.slice(1)
    if (isUUID(p)) {
      name = 'Detalle'
    }
    return {
      name,
      path: '/' + parts.slice(0, idx + 1).join('/')
    }
  })
})

// Theme is managed by useTheme

function initials(name: string) {
  return (name || 'A').split(' ').map((w: string) => w[0]).join('').toUpperCase().slice(0, 2)
}
</script>

<template>
  <div class="shell">
    <div :class="['sidebar-overlay', sidebarOpen ? 'open' : '']" @click="sidebarOpen = false"></div>

    <aside :class="['sidebar-nav', sidebarOpen ? 'open' : '']">
      <div class="sn-brand">
        <div class="sn-brand-icon" style="background: transparent;">
          <img src="../../assets/logo-capacitaciones.png" alt="Logo" style="width: 24px; height: 24px; object-fit: contain;" />
        </div>
        <span>MH Aprende</span>
      </div>
      <nav class="sn-links">
        <RouterLink to="/admin/dashboard" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/></svg>
          Panel Admin
        </RouterLink>
        <RouterLink to="/admin/capacitaciones" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
          Capacitaciones
        </RouterLink>
        <RouterLink to="/admin/licencias" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1-0 011-1h2a1 1 0 011 1v5m-4 0h4"/></svg>
          Empresas / Licencias
        </RouterLink>
        <RouterLink to="/admin/usuarios" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"/></svg>
          Usuarios
        </RouterLink>
      </nav>
      <div class="sn-footer">
        <button @click="auth.logout()">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>
          Cerrar sesión
        </button>
      </div>
    </aside>

    <!-- Main Column -->
    <div class="main-column">
      <header class="topbar">
        <button class="menu-toggle" @click="sidebarOpen = !sidebarOpen">
          <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6h16M4 12h16M4 18h16"/></svg>
        </button>
        <div class="topbar-left">
          <template v-for="(crumb, idx) in breadcrumbs" :key="crumb.path">
            <span v-if="idx > 0" style="color:var(--muted); margin: 0 6px;">/</span>
            <RouterLink v-if="idx < breadcrumbs.length - 1" :to="crumb.path" class="breadcrumb-link">{{ crumb.name }}</RouterLink>
            <span v-else class="breadcrumb-current" style="color:var(--dark); font-weight:600;">{{ crumb.name }}</span>
          </template>
        </div>
        <div class="topbar-right">
          <!-- Theme Toggle -->
          <button class="icon-btn" @click="toggleTheme" style="border:none; background:transparent; width: 36px; height: 36px" data-tooltip="Cambiar tema">
            <svg v-if="!isDark" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" /></svg>
            <svg v-else width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
          </button>
          
          <!-- Notification Bell -->
          <NotificationBell />

          <div v-if="profileOpen" class="pd-overlay" @click="profileOpen = false" />
          <div class="topbar-user" @click.stop="profileOpen = !profileOpen">
            <div class="topbar-avatar">
              <img :src="getAvatarUrl(auth.user?.avatar_url, auth.user?.id || auth.user?.name)" class="avatar-img" alt="avatar" />
            </div>
            <span class="topbar-name">{{ (auth.user?.name || '').slice(0, 20) }}</span>
            <svg class="topbar-chevron" :class="{ open: profileOpen }" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M19 9l-7 7-7-7"/></svg>
            <Transition name="slide-down">
              <div v-if="profileOpen" class="profile-dropdown" @click.stop>
                <div class="pd-header">
                  <div class="pd-avatar">
                    <img :src="getAvatarUrl(auth.user?.avatar_url, auth.user?.id || auth.user?.name)" class="avatar-img" alt="avatar" />
                  </div>
                  <div class="pd-info">
                    <strong>{{ auth.user?.name }}</strong>
                    <span>{{ auth.user?.email }}</span>
                  </div>
                </div>
                <div class="pd-divider" />
                <RouterLink to="/admin/usuarios" class="pd-item" @click="profileOpen = false">
                  <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
                  Gestión de usuarios
                </RouterLink>
                <div class="pd-divider" />
                <button class="pd-item pd-item-danger" @click="auth.logout()">
                  <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>
                  Cerrar sesión
                </button>
              </div>
            </Transition>
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
.breadcrumb-link {
  color: var(--muted);
  text-decoration: none;
  transition: color 0.15s;
}
.breadcrumb-link:hover {
  color: var(--brand);
}
</style>
