<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../../api'

const route = useRoute()
const router = useRouter()
const userId = route.params.id as string

const user = ref<any>(null)
const loading = ref(true)
const error = ref('')
const contactVisible = ref(false)
const copied = ref(false)

const roleLabel = computed(() => {
  const map: Record<string, string> = { admin: 'Administrador', instructor: 'Instructor', user: 'Estudiante' }
  return map[user.value?.role] || 'Usuario'
})

const roleClass = computed(() => {
  return { admin: 'role-admin', instructor: 'role-instructor', user: 'role-user' }[user.value?.role] || 'role-user'
})

function initials(name: string) {
  return name ? name.split(' ').slice(0, 2).map((n: string) => n[0]).join('').toUpperCase() : '?'
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('es', { year: 'numeric', month: 'long' })
}

async function copyEmail() {
  if (!user.value?.email) return
  try {
    await navigator.clipboard.writeText(user.value.email)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch { /* fallback — just show the email */ }
}

onMounted(async () => {
  try {
    const res = await api.get(`/usuarios/${userId}/perfil`)
    user.value = res.data.user
  } catch (e: any) {
    error.value = e.response?.data?.error || 'No se pudo cargar el perfil'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="pub-shell">
    <button class="pub-back" @click="router.back()">
      <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M15 19l-7-7 7-7" stroke-linecap="round" stroke-linejoin="round"/></svg>
      Volver
    </button>

    <!-- Skeleton -->
    <div v-if="loading" class="pub-card pub-skeleton-wrap">
      <div class="skeleton pub-skel-avatar"></div>
      <div style="flex:1">
        <div class="skeleton skel-title" style="width:200px;margin-bottom:10px"></div>
        <div class="skeleton skel-text" style="width:120px"></div>
      </div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="pub-error">
      <svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/></svg>
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="router.back()">Volver</button>
    </div>

    <!-- Profile -->
    <div v-else-if="user" class="pub-layout">

      <!-- Hero card -->
      <div class="pub-hero-card">
        <div class="pub-hero-bg"></div>
        <div class="pub-hero-content">
          <div class="pub-avatar">{{ initials(user.name) }}</div>
          <div class="pub-hero-info">
            <h1 class="pub-name">{{ user.name }}</h1>
            <span :class="['pub-role-badge', roleClass]">{{ roleLabel }}</span>
            <p v-if="user.specialty" class="pub-specialty">
              <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/></svg>
              {{ user.specialty }}
            </p>
            <p class="pub-since">
              <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
              Miembro desde {{ formatDate(user.created_at) }}
            </p>
          </div>
        </div>
      </div>

      <div class="pub-grid">

        <!-- Acerca de -->
        <div class="pub-card pub-about-card">
          <div class="pub-card-head">
            <span class="gm-icon gm-icon-person">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            </span>
            <h2>Acerca de</h2>
          </div>
          <p v-if="user.bio" class="pub-bio">{{ user.bio }}</p>
          <p v-else class="pub-bio pub-bio-empty">Este usuario no ha escrito una descripción todavía.</p>
        </div>

        <!-- Contacto -->
        <div class="pub-card pub-contact-card">
          <div class="pub-card-head">
            <span class="gm-icon gm-icon-mail">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
            </span>
            <h2>Contacto</h2>
          </div>

          <div v-if="!contactVisible" class="pub-contact-hidden">
            <p class="pub-contact-hint">Haz clic para ver los datos de contacto de este usuario.</p>
            <button class="btn btn-primary btn-sm pub-show-contact-btn" @click="contactVisible = true">
              <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              Ver contacto
            </button>
          </div>

          <div v-else class="pub-contact-revealed">
            <div class="pub-contact-item">
              <span class="pub-contact-label">Correo electrónico</span>
              <div class="pub-contact-value-row">
                <a :href="`mailto:${user.email}`" class="pub-email-link">{{ user.email }}</a>
                <button class="pub-copy-btn" @click="copyEmail" :title="copied ? 'Copiado' : 'Copiar correo'">
                  <svg v-if="!copied" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
                  <svg v-else width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M20 6L9 17l-5-5"/></svg>
                </button>
              </div>
            </div>
            <a :href="`mailto:${user.email}`" class="btn btn-primary pub-mailto-btn">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
              Enviar correo
            </a>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
.pub-shell {
  max-width: 860px;
  margin: 0 auto;
  padding: 8px 4px 40px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.pub-back {
  display: inline-flex; align-items: center; gap: 6px;
  background: none; border: none; cursor: pointer;
  color: var(--muted); font-size: 0.82rem; font-weight: 600;
  padding: 4px 0; transition: color 0.15s;
}
.pub-back:hover { color: var(--brand); }

/* Hero card */
.pub-hero-card {
  position: relative; overflow: hidden;
  border-radius: 16px;
  border: 1px solid var(--border-light);
  box-shadow: var(--shadow-md);
}
.pub-hero-bg {
  position: absolute; inset: 0;
  background: linear-gradient(135deg, var(--dark) 0%, #374151 60%, rgba(249,115,22,.15) 100%);
}
.pub-hero-content {
  position: relative; z-index: 1;
  display: flex; align-items: flex-start; gap: 24px;
  padding: 36px 36px;
}
.pub-avatar {
  width: 88px; height: 88px; border-radius: 50%; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, var(--brand), #ef4444);
  color: #fff; font-size: 2rem; font-weight: 800;
  box-shadow: 0 4px 20px rgba(249,115,22,.35);
  border: 3px solid rgba(255,255,255,.15);
}
.pub-hero-info { display: flex; flex-direction: column; gap: 8px; }
.pub-name { font-size: 1.6rem; font-weight: 800; color: #fff; margin: 0; line-height: 1.2; letter-spacing: -.02em; }
.pub-role-badge {
  display: inline-flex; align-self: flex-start;
  padding: 4px 14px; border-radius: 999px;
  font-size: 0.75rem; font-weight: 700;
}
.role-user    { background: rgba(249,115,22,.2); color: #fb923c; border: 1px solid rgba(249,115,22,.3); }
.role-instructor { background: rgba(59,130,246,.2); color: #60a5fa; border: 1px solid rgba(59,130,246,.3); }
.role-admin   { background: rgba(168,85,247,.2); color: #c084fc; border: 1px solid rgba(168,85,247,.3); }

.pub-specialty, .pub-since {
  display: flex; align-items: center; gap: 6px;
  color: rgba(255,255,255,.65); font-size: 0.83rem; margin: 0;
}

/* Grid layout */
.pub-grid {
  display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
}

/* Cards */
.pub-card {
  background: rgba(255,255,255,0.72);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(255,255,255,0.55);
  border-radius: 14px;
  box-shadow: 0 4px 24px rgba(0,0,0,.07);
  padding: 22px;
}
.pub-card-head {
  display: flex; align-items: center; gap: 10px; margin-bottom: 16px;
}
.pub-card-head h2 { font-size: 0.95rem; font-weight: 700; color: var(--dark); margin: 0; }

/* Glassmorphism icons */
.gm-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; border-radius: 8px; flex-shrink: 0;
  backdrop-filter: blur(8px);
}
.gm-icon-person { background: rgba(249,115,22,.12); border: 1px solid rgba(249,115,22,.25); color: var(--brand); }
.gm-icon-mail   { background: rgba(59,130,246,.12);  border: 1px solid rgba(59,130,246,.28);  color: #3b82f6; }

/* About */
.pub-bio { font-size: 0.92rem; color: var(--text); line-height: 1.65; white-space: pre-wrap; margin: 0; }
.pub-bio-empty { color: var(--muted); font-style: italic; }

/* Contact */
.pub-contact-hidden { display: flex; flex-direction: column; gap: 12px; }
.pub-contact-hint { font-size: 0.85rem; color: var(--muted); margin: 0; line-height: 1.5; }
.pub-show-contact-btn { display: inline-flex; align-items: center; gap: 7px; align-self: flex-start; }
.pub-contact-revealed { display: flex; flex-direction: column; gap: 16px; }
.pub-contact-item { display: flex; flex-direction: column; gap: 5px; }
.pub-contact-label { font-size: 0.75rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .05em; }
.pub-contact-value-row { display: flex; align-items: center; gap: 8px; }
.pub-email-link {
  color: var(--dark); font-size: 0.9rem; font-weight: 600;
  text-decoration: none; word-break: break-all;
}
.pub-email-link:hover { color: var(--brand); text-decoration: underline; }
.pub-copy-btn {
  display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 6px;
  border: 1px solid var(--border); background: var(--bg);
  color: var(--muted); cursor: pointer; transition: all 0.15s; flex-shrink: 0;
}
.pub-copy-btn:hover { border-color: var(--brand); color: var(--brand); }
.pub-mailto-btn {
  display: inline-flex; align-items: center; gap: 8px;
  text-decoration: none; font-size: 0.88rem; padding: 9px 18px;
  border-radius: var(--r); align-self: flex-start;
}

/* Error state */
.pub-error {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 14px; padding: 60px 24px; text-align: center;
  color: var(--muted);
}

/* Skeleton */
.pub-skeleton-wrap { display: flex; gap: 20px; align-items: center; }
.pub-skel-avatar { width: 72px; height: 72px; border-radius: 50%; flex-shrink: 0; }

@media (max-width: 640px) {
  .pub-grid { grid-template-columns: 1fr; }
  .pub-hero-content { flex-direction: column; padding: 24px 20px; gap: 16px; }
  .pub-name { font-size: 1.3rem; }
}
</style>
