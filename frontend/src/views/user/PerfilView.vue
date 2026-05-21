<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import api from '../../api'
import { toast } from '../../utils/toast'

const perfil = ref<any>(null)
const stats = ref<any>({})
const loading = ref(false)
const loadingSave = ref(false)

const form = ref({ name: '', bio: '', phone: '', specialty: '' })
const password = ref({ nueva: '', confirmar: '' })
const showPass = ref(false)
const activeTab = ref<'info' | 'security'>('info')

const roleLabel = computed(() => {
  const labels: Record<string, string> = {
    admin: 'Administrador',
    instructor: 'Instructor',
    user: 'Estudiante',
  }
  return labels[perfil.value?.role] || 'Usuario'
})

const completion = computed(() => {
  const fields = [form.value.name, form.value.bio, form.value.phone]
  const filled = fields.filter((value) => String(value || '').trim()).length
  return Math.round((filled / fields.length) * 100)
})

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
    toast.error(e.response?.data?.error || 'No pudimos cargar tu perfil')
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function guardar() {
  if (!form.value.name.trim()) {
    toast.error('El nombre es requerido')
    return
  }
  if (showPass.value || (activeTab.value === 'security' && (password.value.nueva || password.value.confirmar))) {
    if (!password.value.nueva) {
      toast.error('Ingresa la nueva contraseña')
      return
    }
    if (password.value.nueva.length < 6) {
      toast.error('La contraseña debe tener al menos 6 caracteres')
      return
    }
    if (password.value.nueva !== password.value.confirmar) {
      toast.error('Las contraseñas no coinciden')
      return
    }
  }

  loadingSave.value = true
  try {
    const payload: any = {
      name: form.value.name,
      bio: form.value.bio,
      phone: form.value.phone,
    }
    if (showPass.value && password.value.nueva) payload.password = password.value.nueva

    await api.put('/perfil', payload)
    toast.success('Perfil actualizado correctamente')
    password.value = { nueva: '', confirmar: '' }
    showPass.value = false
    await load()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al guardar')
  } finally {
    loadingSave.value = false
  }
}

function initials(name: string) {
  return name ? name.split(' ').slice(0, 2).map(n => n[0]).join('').toUpperCase() : '?'
}
</script>

<template>
  <div class="fp-page">

    <!-- Skeleton carga -->
    <div v-if="loading" class="fp-skeleton">
      <div class="skeleton fp-cover-skel"></div>
      <div class="fp-id-skel">
        <div class="skeleton fp-avatar-skel"></div>
        <div style="flex:1;padding:10px 0">
          <div class="skeleton skel-title" style="width:160px;margin-bottom:8px"></div>
          <div class="skeleton skel-text" style="width:100px"></div>
        </div>
      </div>
    </div>

    <div v-else class="fp-content">

      <!-- Cover photo -->
      <div class="fp-cover">
        <div class="fp-cover-gradient"></div>
        <div class="fp-cover-overlay"></div>
      </div>

      <!-- Barra de identidad: avatar + nombre + acción -->
      <div class="fp-identity-bar">
        <div class="fp-identity-left">
          <div class="fp-avatar">{{ initials(form.name) }}</div>
          <div class="fp-id-info">
            <h1 class="fp-name">{{ perfil?.name }}</h1>
            <div class="fp-id-meta">
              <span class="fp-role-badge">{{ roleLabel }}</span>
              <span v-if="form.specialty" class="fp-specialty">{{ form.specialty }}</span>
              <span class="fp-email-meta">{{ perfil?.email }}</span>
            </div>
          </div>
        </div>
        <button class="btn btn-primary fp-save-btn" :disabled="loadingSave" @click="guardar">
          <span v-if="loadingSave" class="spinner" style="width:14px;height:14px;border-width:2px"></span>
          <svg v-else width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z"/><path d="M17 21v-8H7v8M7 3v5h8"/></svg>
          {{ loadingSave ? 'Guardando...' : 'Guardar cambios' }}
        </button>
      </div>

      <!-- Estadísticas horizontales -->
      <div class="fp-stats-bar" v-if="stats.cursos_inscritos !== undefined">
        <div class="fp-stat">
          <span class="fp-stat-num">{{ stats.cursos_inscritos ?? 0 }}</span>
          <span class="fp-stat-lbl">Cursos inscritos</span>
        </div>
        <div class="fp-stat-divider"></div>
        <div class="fp-stat">
          <span class="fp-stat-num">{{ stats.lecciones_completadas ?? 0 }}</span>
          <span class="fp-stat-lbl">Lecciones completadas</span>
        </div>
        <div class="fp-stat-divider"></div>
        <div class="fp-stat">
          <span class="fp-stat-num fp-stat-pct">{{ stats.total_lecciones ? Math.round((stats.lecciones_completadas / stats.total_lecciones) * 100) : 0 }}%</span>
          <span class="fp-stat-lbl">Progreso total</span>
        </div>
        <div class="fp-stat-divider"></div>
        <div class="fp-stat">
          <span class="fp-stat-num">{{ completion }}%</span>
          <span class="fp-stat-lbl">Perfil completo</span>
        </div>
      </div>

      <!-- Tabs -->
      <div class="fp-tabs">
        <button :class="['fp-tab', activeTab === 'info' ? 'fp-tab-active' : '']" @click="activeTab = 'info'">
          <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          Información
        </button>
        <button :class="['fp-tab', activeTab === 'security' ? 'fp-tab-active' : '']" @click="activeTab = 'security'">
          <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
          Seguridad
        </button>
      </div>

      <!-- Contenido de tabs -->
      <div class="fp-tab-body">

        <!-- Tab: Información -->
        <Transition name="fade" mode="out-in">
          <div v-if="activeTab === 'info'" key="info" class="fp-card">
            <div class="fp-card-head">
              <h2>Datos personales</h2>
              <p>Estos datos son visibles en tu perfil y en los foros de los cursos.</p>
            </div>
            <div class="fp-form">
              <label class="fp-field">
                <span>Nombre completo <em>*</em></span>
                <input v-model="form.name" class="field-input" autocomplete="name" placeholder="Tu nombre completo" />
              </label>
              <label class="fp-field">
                <span>Teléfono</span>
                <input v-model="form.phone" type="tel" class="field-input" placeholder="+52 55 0000 0000" autocomplete="tel" />
              </label>
              <label class="fp-field">
                <span>Especialidad / Área</span>
                <input v-model="form.specialty" class="field-input" placeholder="Ej: Recursos Humanos, Ventas, TI..." />
              </label>
              <label class="fp-field fp-field-full">
                <span>Biografía</span>
                <textarea v-model="form.bio" rows="4" class="field-input fp-textarea"
                  placeholder="Cuenta brevemente tu rol, área o intereses de aprendizaje..." />
              </label>
            </div>
          </div>
        </Transition>

        <!-- Tab: Seguridad -->
        <Transition name="fade" mode="out-in">
          <div v-if="activeTab === 'security'" key="security" class="fp-card">
            <div class="fp-card-head">
              <h2>Cambiar contraseña</h2>
              <p>Elige una contraseña segura de al menos 6 caracteres.</p>
            </div>
            <div class="fp-form fp-form-narrow">
              <label class="fp-field fp-field-full">
                <span>Nueva contraseña</span>
                <input v-model="password.nueva" type="password" placeholder="Mínimo 6 caracteres" class="field-input" autocomplete="new-password" />
              </label>
              <label class="fp-field fp-field-full">
                <span>Confirmar contraseña</span>
                <input v-model="password.confirmar" type="password" placeholder="Repite la nueva contraseña" class="field-input" autocomplete="new-password" />
              </label>
              <div v-if="password.nueva && password.confirmar && password.nueva !== password.confirmar" class="fp-pass-mismatch">
                Las contraseñas no coinciden
              </div>
              <div v-if="password.nueva && password.confirmar && password.nueva === password.confirmar" class="fp-pass-match">
                <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>
                Las contraseñas coinciden
              </div>
            </div>
          </div>
        </Transition>

      </div>
    </div>
  </div>
</template>

<style scoped>
/* ─── Página ────────────────────────────────────────────── */
.fp-page { display: flex; flex-direction: column; min-height: 0; }

/* Skeleton */
.fp-skeleton { display: flex; flex-direction: column; gap: 0; }
.fp-cover-skel { height: 200px; border-radius: var(--r-lg) var(--r-lg) 0 0; }
.fp-id-skel { display: flex; gap: 16px; padding: 16px 24px; align-items: flex-end; }
.fp-avatar-skel { width: 88px; height: 88px; border-radius: 50%; margin-top: -44px; flex-shrink: 0; }

/* ─── Cover photo ──────────────────────────────────────── */
.fp-cover {
  position: relative;
  height: 200px;
  border-radius: var(--r-xl) var(--r-xl) 0 0;
  overflow: hidden;
  flex-shrink: 0;
}
.fp-cover-gradient {
  position: absolute; inset: 0;
  background: linear-gradient(135deg,
    #1d1f23 0%,
    #2d1f14 30%,
    rgba(249,115,22,.6) 70%,
    #c2410c 100%);
}
.fp-cover-overlay {
  position: absolute; inset: 0;
  background: radial-gradient(ellipse at 20% 80%, rgba(249,115,22,.3) 0%, transparent 60%),
              radial-gradient(ellipse at 80% 20%, rgba(239,68,68,.2) 0%, transparent 50%);
}

/* ─── Barra de identidad ────────────────────────────────── */
.fp-identity-bar {
  display: flex; align-items: flex-end; justify-content: space-between;
  gap: 16px; padding: 0 24px 18px;
  background: var(--surface);
  border: 1px solid var(--border-light);
  border-top: none;
  box-shadow: var(--shadow-sm);
}
.fp-identity-left { display: flex; align-items: flex-end; gap: 16px; }
.fp-avatar {
  width: 88px; height: 88px; border-radius: 50%;
  border: 4px solid var(--surface);
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, var(--brand), #ef4444);
  color: #fff; font-size: 2rem; font-weight: 900;
  margin-top: -44px; flex-shrink: 0;
  box-shadow: 0 4px 20px rgba(0,0,0,.2);
}
.fp-id-info { padding-bottom: 4px; }
.fp-name { font-size: 1.35rem; font-weight: 900; color: var(--dark); line-height: 1.2; margin-bottom: 6px; }
.fp-id-meta { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.fp-role-badge {
  padding: 3px 12px; border-radius: 999px;
  background: var(--brand-light); color: var(--brand-dark);
  font-size: 0.72rem; font-weight: 800; text-transform: uppercase; letter-spacing: .04em;
}
.fp-specialty { font-size: 0.82rem; color: var(--muted); font-weight: 600; }
.fp-email-meta { font-size: 0.8rem; color: var(--muted); }
.fp-save-btn { display: flex; align-items: center; gap: 7px; flex-shrink: 0; margin-bottom: 4px; }

/* ─── Stats bar ─────────────────────────────────────────── */
.fp-stats-bar {
  display: flex; align-items: center;
  background: var(--surface); border: 1px solid var(--border-light); border-top: none;
  padding: 14px 28px; gap: 0; flex-wrap: wrap;
}
.fp-stat { display: flex; flex-direction: column; align-items: center; gap: 3px; flex: 1; min-width: 80px; padding: 6px 12px; }
.fp-stat-num { font-size: 1.25rem; font-weight: 900; color: var(--dark); line-height: 1; }
.fp-stat-pct { color: var(--brand); }
.fp-stat-lbl { font-size: 0.72rem; color: var(--muted); font-weight: 600; text-align: center; }
.fp-stat-divider { width: 1px; align-self: stretch; background: var(--border-light); }

/* ─── Tabs ──────────────────────────────────────────────── */
.fp-tabs {
  display: flex; gap: 4px;
  background: var(--surface); border: 1px solid var(--border-light); border-top: none;
  padding: 0 20px;
  border-bottom: 2px solid var(--border-light);
}
.fp-tab {
  display: flex; align-items: center; gap: 7px;
  padding: 14px 18px; border: none; background: transparent;
  color: var(--muted); font-size: 0.88rem; font-weight: 600;
  cursor: pointer; border-bottom: 3px solid transparent;
  margin-bottom: -2px; transition: all 0.15s;
}
.fp-tab:hover { color: var(--dark); background: var(--bg); }
.fp-tab-active { color: var(--brand) !important; border-bottom-color: var(--brand) !important; }

/* ─── Alerta ────────────────────────────────────────────── */
.fp-alert { margin: 16px 0 0; }

/* ─── Tab body ──────────────────────────────────────────── */
.fp-tab-body { margin-top: 20px; }
.fp-card {
  background: var(--surface); border: 1px solid var(--border-light);
  border-radius: var(--r-lg); padding: 24px; box-shadow: var(--shadow-sm);
}
.fp-card-head { margin-bottom: 22px; }
.fp-card-head h2 { font-size: 1.05rem; font-weight: 900; color: var(--dark); margin-bottom: 4px; }
.fp-card-head p { font-size: 0.83rem; color: var(--muted); line-height: 1.5; }

/* ─── Formulario ────────────────────────────────────────── */
.fp-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0,1fr));
  gap: 16px;
}
.fp-form-narrow { grid-template-columns: 1fr; max-width: 480px; }
.fp-field { display: flex; flex-direction: column; gap: 7px; }
.fp-field-full { grid-column: 1 / -1; }
.fp-field span {
  font-size: 0.75rem; font-weight: 800; color: var(--muted);
  text-transform: uppercase; letter-spacing: .05em;
}
.fp-field em { color: var(--brand); font-style: normal; }
.fp-textarea { min-height: 110px; resize: vertical; }

/* Password feedback */
.fp-pass-mismatch { grid-column: 1/-1; font-size: 0.83rem; color: #dc2626; font-weight: 600; display: flex; align-items: center; gap: 6px; }
.fp-pass-match { grid-column: 1/-1; font-size: 0.83rem; color: #16a34a; font-weight: 600; display: flex; align-items: center; gap: 6px; }

/* ─── Responsive ────────────────────────────────────────── */
@media (max-width: 680px) {
  .fp-cover { height: 140px; border-radius: var(--r-lg) var(--r-lg) 0 0; }
  .fp-identity-bar { flex-direction: column; align-items: flex-start; padding: 0 16px 16px; }
  .fp-avatar { width: 72px; height: 72px; margin-top: -36px; font-size: 1.6rem; }
  .fp-name { font-size: 1.15rem; }
  .fp-save-btn { width: 100%; justify-content: center; }
  .fp-stats-bar { padding: 10px 16px; }
  .fp-stat { min-width: 60px; }
  .fp-form { grid-template-columns: 1fr; }
}
</style>
