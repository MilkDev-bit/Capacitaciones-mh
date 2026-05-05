<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import api from '../../api'

const perfil = ref<any>(null)
const stats = ref<any>({})
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const success = ref('')

const form = ref({ name: '', bio: '', phone: '', specialty: '' })
const password = ref({ nueva: '', confirmar: '' })
const showPass = ref(false)

const completion = computed(() => {
  const fields = [form.value.name, form.value.bio, form.value.phone, form.value.specialty]
  return Math.round((fields.filter(v => String(v || '').trim()).length / fields.length) * 100)
})

function initials(name: string) {
  return name ? name.split(' ').slice(0, 2).map(n => n[0]).join('').toUpperCase() : '?'
}

async function load() {
  loading.value = true
  try {
    const res = await api.get('/perfil')
    perfil.value = res.data.user
    stats.value = res.data.stats || {}
    form.value.name = perfil.value.name || ''
    form.value.bio = perfil.value.bio || ''
    form.value.phone = perfil.value.phone || ''
    form.value.specialty = perfil.value.specialty || ''
  } catch (e: any) {
    error.value = e.response?.data?.error || 'No pudimos cargar tu perfil'
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function guardar() {
  error.value = ''; success.value = ''
  if (!form.value.name.trim()) { error.value = 'El nombre es requerido'; return }
  if (showPass.value) {
    if (!password.value.nueva || password.value.nueva.length < 6) { error.value = 'La contraseña debe tener al menos 6 caracteres'; return }
    if (password.value.nueva !== password.value.confirmar) { error.value = 'Las contraseñas no coinciden'; return }
  }
  saving.value = true
  try {
    const payload: any = { name: form.value.name, bio: form.value.bio, phone: form.value.phone, specialty: form.value.specialty }
    if (showPass.value && password.value.nueva) payload.password = password.value.nueva
    await api.put('/perfil', payload)
    success.value = 'Perfil actualizado'
    password.value = { nueva: '', confirmar: '' }
    showPass.value = false
    await load()
    setTimeout(() => success.value = '', 3000)
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="ip-page">
    <!-- Header -->
    <header class="ip-header">
      <div>
        <h1 class="ip-title">Mi Perfil</h1>
        <p class="ip-subtitle">Administra tu información y personaliza tu presencia como instructor.</p>
      </div>
      <button class="btn btn-primary" :disabled="saving || loading" @click="guardar">
        {{ saving ? 'Guardando...' : 'Guardar cambios' }}
      </button>
    </header>

    <!-- Loading skeleton -->
    <div v-if="loading" class="ip-grid">
      <div class="ip-card"><div class="skeleton" style="width:80px;height:80px;border-radius:50%"></div><div class="skeleton" style="height:18px;width:60%;margin-top:14px"></div><div class="skeleton" style="height:12px;width:40%;margin-top:8px"></div></div>
      <div class="ip-card"><div class="skeleton" style="height:18px;width:50%"></div><div class="skeleton" style="height:40px;width:100%;margin-top:12px"></div><div class="skeleton" style="height:40px;width:100%;margin-top:8px"></div></div>
    </div>

    <div v-else class="ip-grid">
      <!-- Sidebar -->
      <aside class="ip-sidebar">
        <div class="ip-card ip-profile-card">
          <div class="ip-avatar">{{ initials(form.name) }}</div>
          <h2 class="ip-name">{{ perfil?.name }}</h2>
          <p class="ip-email">{{ perfil?.email }}</p>
          <span class="ip-role-badge">Instructor</span>

          <div class="ip-completion">
            <div class="ip-completion-header">
              <span>Perfil completo</span>
              <strong>{{ completion }}%</strong>
            </div>
            <div class="ip-progress-bg">
              <div class="ip-progress-fill" :style="`width:${completion}%`"></div>
            </div>
          </div>
        </div>

        <!-- Stats -->
        <div class="ip-card ip-stats-card">
          <h3>Estadísticas</h3>
          <div class="ip-stats-grid">
            <div class="ip-stat">
              <span class="ip-stat-num">{{ stats.cursos_creados ?? 0 }}</span>
              <span class="ip-stat-label">Cursos</span>
            </div>
            <div class="ip-stat">
              <span class="ip-stat-num">{{ stats.estudiantes_total ?? 0 }}</span>
              <span class="ip-stat-label">Estudiantes</span>
            </div>
            <div class="ip-stat">
              <span class="ip-stat-num">{{ stats.examenes_creados ?? 0 }}</span>
              <span class="ip-stat-label">Exámenes</span>
            </div>
          </div>
        </div>
      </aside>

      <!-- Main forms -->
      <main class="ip-main">
        <Transition name="slide-down">
          <div v-if="error" class="alert alert-error">{{ error }}</div>
        </Transition>
        <Transition name="slide-down">
          <div v-if="success" class="alert alert-success">{{ success }}</div>
        </Transition>

        <section class="ip-card">
          <div class="ip-section-head">
            <h2>Información personal</h2>
            <p>Tus datos visibles para los estudiantes.</p>
          </div>
          <div class="ip-form">
            <label class="ip-field">
              <span>Nombre completo *</span>
              <input v-model="form.name" class="field-input" autocomplete="name" />
            </label>
            <label class="ip-field">
              <span>Especialidad</span>
              <input v-model="form.specialty" class="field-input" placeholder="Ej: Seguridad Industrial, Calidad..." />
            </label>
            <label class="ip-field">
              <span>Teléfono</span>
              <input v-model="form.phone" type="tel" class="field-input" placeholder="+52 55 0000 0000" />
            </label>
            <label class="ip-field ip-field-full">
              <span>Biografía</span>
              <textarea v-model="form.bio" rows="4" class="field-input" placeholder="Describe tu experiencia y áreas de conocimiento..." style="resize:vertical"></textarea>
            </label>
          </div>
        </section>

        <section class="ip-card">
          <div class="ip-section-head ip-section-head-row">
            <div>
              <h2>Seguridad</h2>
              <p>Cambia tu contraseña de acceso.</p>
            </div>
            <button class="btn btn-secondary btn-sm" @click="showPass = !showPass">
              {{ showPass ? 'Cancelar' : 'Cambiar contraseña' }}
            </button>
          </div>
          <Transition name="slide-down">
            <div v-if="showPass" class="ip-form ip-pass-form">
              <label class="ip-field">
                <span>Nueva contraseña</span>
                <input v-model="password.nueva" type="password" class="field-input" placeholder="Mínimo 6 caracteres" autocomplete="new-password" />
              </label>
              <label class="ip-field">
                <span>Confirmar</span>
                <input v-model="password.confirmar" type="password" class="field-input" placeholder="Repite la contraseña" autocomplete="new-password" />
              </label>
            </div>
          </Transition>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.ip-page { display: flex; flex-direction: column; gap: 24px; }
.ip-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; padding-bottom: 20px; border-bottom: 1px solid var(--border-light); }
.ip-title { font-size: 1.6rem; font-weight: 700; color: var(--dark); letter-spacing: -0.02em; }
.ip-subtitle { color: var(--muted); font-size: 0.88rem; margin-top: 4px; }
.ip-grid { display: grid; grid-template-columns: 300px minmax(0, 1fr); gap: 24px; align-items: start; }
.ip-sidebar { display: flex; flex-direction: column; gap: 16px; position: sticky; top: calc(var(--topbar-h) + 24px); }
.ip-main { display: flex; flex-direction: column; gap: 16px; min-width: 0; }
.ip-card { padding: 24px; background: var(--surface); border: 1px solid var(--border-light); border-radius: var(--r-lg); box-shadow: var(--shadow-sm); }
.ip-profile-card { display: flex; flex-direction: column; align-items: center; text-align: center; gap: 4px; }
.ip-avatar {
  width: 88px; height: 88px; border-radius: 50%; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, var(--brand), #ef4444); color: #fff; font-size: 1.6rem; font-weight: 800;
  box-shadow: 0 4px 16px rgba(249,115,22,.25); margin-bottom: 8px;
}
.ip-name { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
.ip-email { font-size: 0.82rem; color: var(--muted); overflow-wrap: anywhere; }
.ip-role-badge { margin-top: 8px; padding: 4px 14px; border-radius: 999px; background: var(--brand-light); color: var(--brand-dark); font-size: 0.75rem; font-weight: 700; }
.ip-completion { width: 100%; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border-light); }
.ip-completion-header { display: flex; justify-content: space-between; font-size: 0.78rem; color: var(--muted); margin-bottom: 6px; }
.ip-completion-header strong { color: var(--brand); }
.ip-progress-bg { height: 6px; background: var(--border-light); border-radius: 3px; overflow: hidden; }
.ip-progress-fill { height: 100%; background: linear-gradient(90deg, var(--brand), var(--brand-dark)); border-radius: 3px; transition: width 0.5s var(--ease-apple); }
.ip-stats-card h3 { font-size: 0.82rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .05em; margin-bottom: 14px; }
.ip-stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.ip-stat { text-align: center; padding: 12px 4px; background: var(--bg); border-radius: var(--r); }
.ip-stat-num { display: block; font-size: 1.3rem; font-weight: 800; color: var(--dark); }
.ip-stat-label { font-size: 0.7rem; color: var(--muted); font-weight: 600; }
.ip-section-head { margin-bottom: 20px; }
.ip-section-head h2 { font-size: 1rem; font-weight: 700; color: var(--dark); }
.ip-section-head p { font-size: 0.84rem; color: var(--muted); margin-top: 3px; }
.ip-section-head-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }
.ip-form { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.ip-field { display: flex; flex-direction: column; gap: 6px; }
.ip-field-full { grid-column: 1 / -1; }
.ip-field span { font-size: 0.78rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .04em; }
.ip-pass-form { padding-top: 16px; border-top: 1px solid var(--border-light); }

@media (max-width: 820px) {
  .ip-grid { grid-template-columns: 1fr; }
  .ip-sidebar { position: static; }
}
@media (max-width: 560px) {
  .ip-header { flex-direction: column; align-items: stretch; }
  .ip-form { grid-template-columns: 1fr; }
  .ip-section-head-row { flex-direction: column; }
}
</style>
