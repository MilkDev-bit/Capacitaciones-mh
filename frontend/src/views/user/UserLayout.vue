<script setup lang="ts">
import { RouterView, RouterLink } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div class="layout">
    <aside class="sidebar">
      <div class="brand">
        <svg width="28" height="28" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="8" fill="#10b981"/>
          <path d="M10 28L20 12L30 28H10Z" fill="white"/>
        </svg>
        <span>MH Capacitaciones</span>
      </div>
      <nav>
        <RouterLink to="/usuario/capacitaciones">
          <span>📚</span> Mis capacitaciones
        </RouterLink>
        <RouterLink to="/usuario/examenes">
          <span>📝</span> Mis exámenes
        </RouterLink>
      </nav>
      <button class="logout-btn" @click="auth.logout()">Cerrar sesión</button>
    </aside>
    <main class="content">
      <div class="topbar">
        <span class="welcome">Hola, <strong>{{ auth.user?.name }}</strong></span>
        <span class="badge user">Usuario</span>
      </div>
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.layout { display: flex; min-height: 100vh; }
.sidebar {
  width: 230px; background: #064e3b; color: white;
  display: flex; flex-direction: column; padding: 1.5rem 1rem; gap: 0.5rem;
  position: fixed; top: 0; left: 0; height: 100vh;
}
.brand { display: flex; align-items: center; gap: 10px; font-size: 1rem; font-weight: 700; margin-bottom: 1.5rem; color: white; }
nav { display: flex; flex-direction: column; gap: 4px; flex: 1; }
nav a {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; border-radius: 8px;
  color: #6ee7b7; text-decoration: none; font-size: 0.95rem; transition: all 0.15s;
}
nav a:hover, nav a.router-link-active { background: #065f46; color: white; }
.logout-btn {
  margin-top: auto; background: #ef4444; color: white; border: none;
  border-radius: 8px; padding: 10px; cursor: pointer; font-size: 0.9rem; font-weight: 600;
}
.logout-btn:hover { background: #dc2626; }
.content { margin-left: 230px; flex: 1; display: flex; flex-direction: column; }
.topbar { background: white; padding: 1rem 2rem; border-bottom: 1px solid #e2e8f0; display: flex; align-items: center; gap: 1rem; }
.welcome { color: #475569; font-size: 0.95rem; }
.badge { font-size: 0.75rem; padding: 3px 10px; border-radius: 20px; font-weight: 600; }
.badge.user { background: #d1fae5; color: #065f46; }
</style>
