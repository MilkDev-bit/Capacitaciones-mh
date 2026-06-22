<script setup lang="ts">
import { ref } from 'vue'
import { RouterView, RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const sidebarOpen = ref(false)
const profileOpen = ref(false)

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
          <img src="/logo-capacitaciones.png" alt="Logo" style="width: 24px; height: 24px; object-fit: contain;" />
        </div>
        <span>MH Admin</span>
      </div>
      <nav>
        <RouterLink to="/admin/dashboard" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/></svg>
          Dashboard
        </RouterLink>
        <RouterLink to="/admin/capacitaciones" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
          Capacitaciones
        </RouterLink>
        <RouterLink to="/admin/examenes" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>
          Exámenes
        </RouterLink>
        <RouterLink to="/admin/usuarios" @click="sidebarOpen = false">
          <svg class="nav-icon" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
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

    <div class="shell-main">
      <header class="topbar">
        <button class="topbar-hamburger" @click="sidebarOpen = true">
          <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6h16M4 12h16M4 18h16"/></svg>
        </button>
        <span class="topbar-title">Panel Administrador</span>
        <div class="topbar-right">
          <div v-if="profileOpen" class="pd-overlay" @click="profileOpen = false" />
          <div class="topbar-user" @click.stop="profileOpen = !profileOpen">
            <div class="topbar-avatar">
              <img v-if="auth.user?.avatar_url" :src="auth.user.avatar_url" class="avatar-img" alt="avatar" />
              <template v-else>{{ initials(auth.user?.name || '') }}</template>
            </div>
            <span class="topbar-name">{{ (auth.user?.name || '').slice(0, 20) }}</span>
            <svg class="topbar-chevron" :class="{ open: profileOpen }" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M19 9l-7 7-7-7"/></svg>
            <Transition name="slide-down">
              <div v-if="profileOpen" class="profile-dropdown" @click.stop>
                <div class="pd-header">
                  <div class="pd-avatar">
                    <img v-if="auth.user?.avatar_url" :src="auth.user.avatar_url" class="avatar-img" alt="avatar" />
                    <template v-else>{{ initials(auth.user?.name || '') }}</template>
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

<style scoped></style>
