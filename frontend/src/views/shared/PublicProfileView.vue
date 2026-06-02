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
  const map: Record<string, string> = { admin: 'role-admin', instructor: 'role-instructor', user: 'role-user' }
  return map[user.value?.role] || 'role-user'
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
  } catch { /* fallback */ }
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
  <div class="pp-page">
    <button class="pp-back" @click="router.back()">
      <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M15 19l-7-7 7-7" stroke-linecap="round" stroke-linejoin="round"/></svg>
      Volver
    </button>

    <div v-if="loading" class="pp-skeleton">
      <div class="skeleton pp-cover-skel"></div>
      <div class="pp-id-skel">
        <div class="skeleton pp-avatar-skel"></div>
        <div style="flex:1;padding:8px 0">
          <div class="skeleton skel-title" style="width:180px;margin-bottom:8px"></div>
          <div class="skeleton skel-text" style="width:100px"></div>
        </div>
      </div>
    </div>

    <div v-else-if="error" class="pp-error">
      <svg width="44" height="44" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/></svg>
      <p>{{ error }}</p>
      <button class="btn btn-secondary btn-sm" @click="router.back()">Volver</button>
    </div>

    <div v-else-if="user" class="pp-content">

      <div class="pp-cover">
        <div :class="['pp-cover-grad', roleClass]"></div>
        <div class="pp-cover-pattern"></div>
      </div>

      <div class="pp-identity">
        <div class="pp-identity-left">
          <div :class="['pp-avatar', roleClass]">
            <img v-if="user.avatar_url" :src="user.avatar_url" class="pp-avatar-img" :alt="user.name" />
            <span v-else>{{ initials(user.name) }}</span>
          </div>
          <div class="pp-id-info">
            <h1 class="pp-name">{{ user.name }}</h1>
            <div class="pp-id-meta">
              <span :class="['pp-role-badge', roleClass]">{{ roleLabel }}</span>
              <span v-if="user.specialty" class="pp-specialty">
                <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/></svg>
                {{ user.specialty }}
              </span>
              <span class="pp-since">
                <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4M8 2v4M3 10h18"/></svg>
                Miembro desde {{ formatDate(user.created_at) }}
              </span>
            </div>
          </div>
        </div>

        <div class="pp-actions">
          <button v-if="!contactVisible" class="btn btn-primary pp-action-btn" @click="contactVisible = true">
            <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
            Contactar
          </button>
          <a v-else :href="`mailto:${user.email}`" class="btn btn-primary pp-action-btn" style="text-decoration:none">
            <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
            Enviar correo
          </a>
        </div>
      </div>

      <div class="pp-body">

        <div class="pp-col-left">

          <div class="pp-card">
            <h3 class="pp-card-title">
              <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
              Presentación
            </h3>
            <p v-if="user.bio" class="pp-bio">{{ user.bio }}</p>
            <p v-else class="pp-bio pp-bio-empty">Este usuario no ha escrito una descripción todavía.</p>
          </div>

          <div class="pp-card">
            <h3 class="pp-card-title">
              <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 8v4l3 3"/></svg>
              Información
            </h3>
            <ul class="pp-info-list">
              <li v-if="user.specialty">
                <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/></svg>
                <span><strong>Especialidad:</strong> {{ user.specialty }}</span>
              </li>
              <li>
                <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4M8 2v4M3 10h18"/></svg>
                <span><strong>Miembro desde:</strong> {{ formatDate(user.created_at) }}</span>
              </li>
              <li>
                <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg>
                <span><strong>Rol:</strong> {{ roleLabel }}</span>
              </li>
            </ul>
          </div>

        </div>

        <div class="pp-col-right">
          <div class="pp-card pp-contact-card">
            <h3 class="pp-card-title">
              <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
              Contacto
            </h3>

            <div v-if="!contactVisible" class="pp-contact-locked">
              <div class="pp-contact-lock-icon">
                <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
              </div>
              <p>Los datos de contacto están protegidos.</p>
              <button class="btn btn-secondary btn-sm pp-unlock-btn" @click="contactVisible = true">
                <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                Ver contacto
              </button>
            </div>

            <div v-else class="pp-contact-revealed">
              <div class="pp-contact-row">
                <span class="pp-contact-lbl">Correo electrónico</span>
                <div class="pp-contact-val-row">
                  <a :href="`mailto:${user.email}`" class="pp-email-link">{{ user.email }}</a>
                  <button class="pp-copy-btn" @click="copyEmail" :title="copied ? 'Copiado' : 'Copiar'">
                    <svg v-if="!copied" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>
                    <svg v-else width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M20 6L9 17l-5-5"/></svg>
                  </button>
                </div>
              </div>
              <a :href="`mailto:${user.email}`" class="btn btn-primary pp-mailto-btn" style="text-decoration:none">
                <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
                Enviar correo
              </a>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
.pp-page { display: flex; flex-direction: column; gap: 0; max-width: 960px; margin: 0 auto; }

.pp-back {
  display: inline-flex; align-items: center; gap: 6px;
  background: none; border: none; cursor: pointer;
  color: var(--muted); font-size: 0.82rem; font-weight: 600;
  padding: 4px 0; margin-bottom: 12px; transition: color 0.15s; align-self: flex-start;
}
.pp-back:hover { color: var(--brand); }

.pp-skeleton { display: flex; flex-direction: column; gap: 0; }
.pp-cover-skel { height: 200px; border-radius: var(--r-xl) var(--r-xl) 0 0; }
.pp-id-skel { display: flex; gap: 16px; padding: 16px 28px; align-items: flex-end; background: var(--surface); border: 1px solid var(--border-light); border-top: none; }
.pp-avatar-skel { width: 92px; height: 92px; border-radius: 50%; margin-top: -46px; flex-shrink: 0; }

.pp-error { display: flex; flex-direction: column; align-items: center; gap: 14px; padding: 60px 24px; text-align: center; color: var(--muted); }
.pp-error p { font-size: 0.92rem; }

.pp-cover {
  position: relative; height: 200px;
  border-radius: var(--r-xl) var(--r-xl) 0 0; overflow: hidden;
}
.pp-cover-grad {
  position: absolute; inset: 0;
  background: linear-gradient(135deg, #1d1f23 0%, #2d1f14 40%, rgba(249,115,22,.5) 100%);
}
.pp-cover-grad.role-instructor {
  background: linear-gradient(135deg, #0f172a 0%, #1e3a5f 40%, rgba(59,130,246,.5) 100%);
}
.pp-cover-grad.role-admin {
  background: linear-gradient(135deg, #1a0533 0%, #2e1065 40%, rgba(168,85,247,.5) 100%);
}
.pp-cover-pattern {
  position: absolute; inset: 0;
  background: radial-gradient(ellipse at 70% 80%, rgba(255,255,255,.06) 0%, transparent 60%),
              radial-gradient(ellipse at 20% 20%, rgba(255,255,255,.04) 0%, transparent 50%);
}

.pp-identity {
  display: flex; align-items: flex-end; justify-content: space-between;
  gap: 16px; padding: 0 28px 20px;
  background: var(--surface);
  border: 1px solid var(--border-light); border-top: none;
  box-shadow: var(--shadow-sm);
  position: relative; z-index: 1;
}
.pp-identity-left { display: flex; align-items: flex-end; gap: 18px; }
.pp-avatar {
  width: 92px; height: 92px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, var(--brand), #ef4444);
  color: #fff; font-size: 2rem; font-weight: 900;
  border: 4px solid var(--surface);
  margin-top: -46px; flex-shrink: 0;
  box-shadow: 0 4px 20px rgba(0,0,0,.2);
}
.pp-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}
.pp-avatar.role-instructor { background: linear-gradient(135deg, #3b82f6, #1d4ed8); }
.pp-avatar.role-admin       { background: linear-gradient(135deg, #a855f7, #7c3aed); }

.pp-id-info { padding-bottom: 4px; }
.pp-name { font-size: 1.45rem; font-weight: 900; color: var(--dark); margin: 0 0 8px; line-height: 1.15; }
.pp-id-meta { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.pp-role-badge {
  padding: 3px 12px; border-radius: 999px;
  font-size: 0.72rem; font-weight: 800; text-transform: uppercase; letter-spacing: .04em;
}
.pp-role-badge.role-user       { background: var(--brand-light); color: var(--brand-dark); }
.pp-role-badge.role-instructor { background: rgba(59,130,246,.12); color: #1d4ed8; border: 1px solid rgba(59,130,246,.25); }
.pp-role-badge.role-admin      { background: rgba(168,85,247,.12); color: #6d28d9; border: 1px solid rgba(168,85,247,.25); }
.pp-specialty, .pp-since {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 0.8rem; color: var(--muted); font-weight: 500;
}

.pp-actions { padding-bottom: 8px; }
.pp-action-btn { display: inline-flex; align-items: center; gap: 7px; }

.pp-body {
  display: grid; grid-template-columns: 1fr 320px;
  gap: 18px; margin-top: 18px;
  align-items: start;
}
.pp-col-left { display: flex; flex-direction: column; gap: 16px; }
.pp-col-right { display: flex; flex-direction: column; gap: 16px; }

.pp-card {
  background: var(--surface); border: 1px solid var(--border-light);
  border-radius: var(--r-lg); padding: 20px; box-shadow: var(--shadow-sm);
}
.pp-card-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 0.95rem; font-weight: 800; color: var(--dark);
  margin: 0 0 14px;
}

.pp-bio { font-size: 0.9rem; color: var(--text); line-height: 1.7; white-space: pre-wrap; margin: 0; }
.pp-bio-empty { color: var(--muted); font-style: italic; }
.pp-info-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 10px; }
.pp-info-list li { display: flex; align-items: center; gap: 9px; font-size: 0.87rem; color: var(--text); }
.pp-info-list li svg { flex-shrink: 0; color: var(--muted); }
.pp-contact-locked { display: flex; flex-direction: column; align-items: center; gap: 12px; padding: 8px 0; text-align: center; }
.pp-contact-lock-icon { width: 48px; height: 48px; border-radius: 50%; background: var(--bg); border: 1px solid var(--border-light); display: flex; align-items: center; justify-content: center; color: var(--muted); }
.pp-contact-locked p { font-size: 0.84rem; color: var(--muted); margin: 0; }
.pp-unlock-btn { display: inline-flex; align-items: center; gap: 7px; }
.pp-contact-revealed { display: flex; flex-direction: column; gap: 14px; }
.pp-contact-row { display: flex; flex-direction: column; gap: 5px; }
.pp-contact-lbl { font-size: 0.74rem; font-weight: 800; color: var(--muted); text-transform: uppercase; letter-spacing: .05em; }
.pp-contact-val-row { display: flex; align-items: center; gap: 8px; }
.pp-email-link { color: var(--dark); font-size: 0.88rem; font-weight: 600; text-decoration: none; word-break: break-all; }
.pp-email-link:hover { color: var(--brand); text-decoration: underline; }
.pp-copy-btn {
  display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 6px;
  border: 1px solid var(--border); background: var(--bg);
  color: var(--muted); cursor: pointer; transition: all 0.15s; flex-shrink: 0;
}
.pp-copy-btn:hover { border-color: var(--brand); color: var(--brand); }
.pp-mailto-btn { display: inline-flex; align-items: center; gap: 7px; }


@media (max-width: 720px) {
  .pp-body { grid-template-columns: 1fr; }
  .pp-cover { height: 140px; }
  .pp-identity { flex-direction: column; align-items: flex-start; padding: 0 16px 16px; }
  .pp-avatar { width: 76px; height: 76px; margin-top: -38px; font-size: 1.7rem; border-width: 3px; }
  .pp-name { font-size: 1.2rem; }
  .pp-actions { width: 100%; }
  .pp-action-btn { width: 100%; justify-content: center; }
}
</style>
