<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '../../api'

const users = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const examenes = ref<any[]>([])
const asignaciones = ref<any[]>([])

const form = ref({ user_id: '', capacitacion_id: '', examen_id: '' })
const error = ref('')
const success = ref('')
const loading = ref(false)
const searchUsers = ref('')
const searchAsig = ref('')
const activeTab = ref<'asignar' | 'asignaciones' | 'usuarios'>('asignar')

const filteredUsers = computed(() => {
  const term = searchUsers.value.toLowerCase().trim()
  if (!term) return users.value
  return users.value.filter(u =>
    (u.name || '').toLowerCase().includes(term) || (u.email || '').toLowerCase().includes(term)
  )
})

const filteredAsig = computed(() => {
  const term = searchAsig.value.toLowerCase().trim()
  if (!term) return asignaciones.value
  return asignaciones.value.filter(a => {
    const name = userName(a.user_id).toLowerCase()
    const cap = capTitle(a.capacitacion_id).toLowerCase()
    return name.includes(term) || cap.includes(term)
  })
})

const roleStats = computed(() => ({
  admin: users.value.filter(u => u.role === 'admin').length,
  instructor: users.value.filter(u => u.role === 'instructor').length,
  user: users.value.filter(u => u.role === 'user').length,
}))

async function load() {
  loading.value = true
  try {
    const [u, c, e, a] = await Promise.all([
      api.get('/admin/users'),
      api.get('/admin/capacitaciones'),
      api.get('/admin/examenes'),
      api.get('/admin/asignaciones'),
    ])
    users.value = u.data || []
    capacitaciones.value = c.data || []
    examenes.value = e.data || []
    asignaciones.value = a.data || []
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function asignar() {
  error.value = ''; success.value = ''
  if (!form.value.user_id) { error.value = 'Selecciona un usuario'; return }
  if (!form.value.capacitacion_id && !form.value.examen_id) {
    error.value = 'Selecciona una capacitación o un examen'; return
  }
  loading.value = true
  try {
    const payload: any = { user_id: form.value.user_id }
    if (form.value.capacitacion_id) payload.capacitacion_id = form.value.capacitacion_id
    if (form.value.examen_id) payload.examen_id = form.value.examen_id
    await api.post('/admin/asignar', payload)
    success.value = 'Asignación realizada correctamente'
    form.value = { user_id: '', capacitacion_id: '', examen_id: '' }
    await load()
    setTimeout(() => { success.value = '' }, 3000)
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al asignar'
  } finally {
    loading.value = false
  }
}

async function desasignar(id: string) {
  if (!confirm('¿Eliminar esta asignación?')) return
  await api.delete(`/admin/asignar/${id}`)
  await load()
}

function userName(id: string) { return users.value.find(u => u.id === id)?.name || 'Desconocido' }
function capTitle(id: string) { return capacitaciones.value.find(c => c.id === id)?.title || '—' }
function exTitle(id: string) { return examenes.value.find(e => e.id === id)?.title || '—' }
function initials(name: string) { return (name || 'U').split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) }
</script>

<template>
  <div class="au-shell">
    <div class="au-topbar">
      <div>
        <h1 class="au-title">Usuarios y Asignaciones</h1>
        <p class="au-sub">{{ users.length }} usuarios · {{ asignaciones.length }} asignaciones activas</p>
      </div>
    </div>

    <Transition name="slide-down">
      <div v-if="error" class="alert alert-error">{{ error }}</div>
    </Transition>
    <Transition name="slide-down">
      <div v-if="success" class="alert alert-success">{{ success }}</div>
    </Transition>

    <!-- Tabs -->
    <div class="tabs-bar">
      <button :class="['tab-pill', activeTab === 'asignar' ? 'active' : '']" @click="activeTab = 'asignar'">
        Asignar contenido
      </button>
      <button :class="['tab-pill', activeTab === 'asignaciones' ? 'active' : '']" @click="activeTab = 'asignaciones'">
        Asignaciones <span class="pill-count">{{ asignaciones.length }}</span>
      </button>
      <button :class="['tab-pill', activeTab === 'usuarios' ? 'active' : '']" @click="activeTab = 'usuarios'">
        Usuarios <span class="pill-count">{{ users.length }}</span>
      </button>
    </div>

    <!-- TAB: Asignar -->
    <div v-if="activeTab === 'asignar'" class="au-tab-content">
      <div class="au-assign-card">
        <div class="au-assign-header">
          <span>🔗</span>
          <div>
            <h2>Asignar contenido a un usuario</h2>
            <p>Selecciona un usuario y el contenido que deseas asignarle</p>
          </div>
        </div>
        <div class="au-assign-body">
          <div class="au-assign-grid">
            <div class="au-field">
              <label>Usuario *</label>
              <select class="field-input" v-model="form.user_id">
                <option value="">— Selecciona usuario —</option>
                <option v-for="u in users" :key="u.id" :value="u.id">{{ u.name }} ({{ u.email }})</option>
              </select>
            </div>
            <div class="au-field">
              <label>Capacitación</label>
              <select class="field-input" v-model="form.capacitacion_id">
                <option value="">— Ninguna —</option>
                <option v-for="c in capacitaciones" :key="c.id" :value="c.id">{{ c.title }}</option>
              </select>
            </div>
            <div class="au-field">
              <label>Examen</label>
              <select class="field-input" v-model="form.examen_id">
                <option value="">— Ninguno —</option>
                <option v-for="e in examenes" :key="e.id" :value="e.id">{{ e.title }}</option>
              </select>
            </div>
          </div>
          <button class="btn btn-primary" :disabled="loading" @click="asignar" style="margin-top:16px">
            {{ loading ? 'Asignando…' : 'Asignar contenido' }}
          </button>
        </div>
      </div>
    </div>

    <!-- TAB: Asignaciones -->
    <div v-if="activeTab === 'asignaciones'" class="au-tab-content">
      <div class="au-search-bar">
        <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z"/></svg>
        <input v-model="searchAsig" placeholder="Buscar en asignaciones..." />
      </div>

      <div v-if="filteredAsig.length" class="au-list">
        <div v-for="a in filteredAsig" :key="a.id" class="au-list-item">
          <div class="au-list-avatar">{{ initials(userName(a.user_id)) }}</div>
          <div class="au-list-info">
            <strong>{{ userName(a.user_id) }}</strong>
            <p>
              <span v-if="a.capacitacion_id" class="badge badge-orange">{{ capTitle(a.capacitacion_id) }}</span>
              <span v-if="a.examen_id" class="badge badge-blue">{{ exTitle(a.examen_id) }}</span>
              <span class="au-list-date">{{ new Date(a.assigned_at).toLocaleDateString('es') }}</span>
            </p>
          </div>
          <button class="btn btn-danger btn-sm" @click="desasignar(a.id)">Quitar</button>
        </div>
      </div>
      <div v-else class="empty-state"><div class="empty-icon">📋</div><h3>Sin asignaciones</h3><p>Asigna contenido a usuarios desde la pestaña "Asignar contenido".</p></div>
    </div>

    <!-- TAB: Usuarios -->
    <div v-if="activeTab === 'usuarios'" class="au-tab-content">
      <!-- Role summary -->
      <div class="au-role-stats">
        <div class="au-role-stat">
          <span class="au-role-dot" style="background:var(--info)"></span>
          <strong>{{ roleStats.admin }}</strong> Admins
        </div>
        <div class="au-role-stat">
          <span class="au-role-dot" style="background:var(--brand)"></span>
          <strong>{{ roleStats.instructor }}</strong> Instructores
        </div>
        <div class="au-role-stat">
          <span class="au-role-dot" style="background:var(--success)"></span>
          <strong>{{ roleStats.user }}</strong> Estudiantes
        </div>
      </div>

      <div class="au-search-bar">
        <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z"/></svg>
        <input v-model="searchUsers" placeholder="Buscar por nombre o email..." />
      </div>

      <div v-if="filteredUsers.length" class="au-users-grid">
        <div v-for="u in filteredUsers" :key="u.id" class="au-user-card">
          <div class="au-user-avatar">{{ initials(u.name) }}</div>
          <div class="au-user-info">
            <strong>{{ u.name }}</strong>
            <p>{{ u.email }}</p>
          </div>
          <span :class="['badge', u.role === 'admin' ? 'badge-blue' : u.role === 'instructor' ? 'badge-orange' : 'badge-green']">
            {{ u.role === 'admin' ? 'Admin' : u.role === 'instructor' ? 'Instructor' : 'Estudiante' }}
          </span>
          <span class="au-user-date">{{ new Date(u.created_at).toLocaleDateString('es') }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.au-shell { display: flex; flex-direction: column; gap: 20px; }
.au-topbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.au-title { font-size: 1.65rem; font-weight: 800; color: var(--dark); letter-spacing: -0.02em; }
.au-sub { color: var(--muted); font-size: 0.88rem; margin-top: 3px; }

.au-tab-content { display: flex; flex-direction: column; gap: 16px; }

.au-assign-card {
  background: var(--surface); border-radius: var(--r-lg); overflow: hidden;
  border: 1.5px solid var(--border); box-shadow: 0 4px 20px rgba(0,0,0,.06);
}
.au-assign-header {
  display: flex; align-items: center; gap: 16px;
  padding: 22px 28px; background: var(--dark); color: #fff;
}
.au-assign-header span { font-size: 2rem; }
.au-assign-header h2 { font-size: 1.1rem; font-weight: 800; color: #fff; margin: 0; }
.au-assign-header p { font-size: 0.83rem; color: rgba(255,255,255,.65); margin: 2px 0 0; }
.au-assign-body { padding: 24px 28px; }
.au-assign-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.au-field { display: flex; flex-direction: column; gap: 6px; }
.au-field label { font-size: 0.82rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: 0.04em; }

.au-search-bar {
  display: flex; align-items: center; gap: 10px; padding: 11px 16px;
  border: 1.5px solid var(--border); border-radius: var(--r); background: var(--surface);
  color: var(--muted); box-shadow: var(--shadow-xs); max-width: 400px;
}
.au-search-bar input { width: 100%; border: 0; outline: 0; background: transparent; color: var(--dark); font-size: 0.9rem; }
.au-search-bar:focus-within { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(249,115,22,.12); }

.au-list {
  background: var(--surface); border-radius: var(--r-lg); border: 1px solid var(--border-light);
  box-shadow: var(--shadow-sm); overflow: hidden;
}
.au-list-item {
  display: flex; align-items: center; gap: 14px; padding: 14px 20px;
  border-bottom: 1px solid var(--border-light); transition: background 0.12s;
}
.au-list-item:last-child { border-bottom: none; }
.au-list-item:hover { background: var(--bg); }
.au-list-avatar {
  width: 36px; height: 36px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, var(--brand), var(--brand-dark)); color: #fff;
  font-size: 0.78rem; font-weight: 700; display: flex; align-items: center; justify-content: center;
}
.au-list-info { flex: 1; min-width: 0; }
.au-list-info strong { font-size: 0.9rem; font-weight: 700; color: var(--dark); display: block; }
.au-list-info p { display: flex; align-items: center; gap: 6px; margin-top: 4px; flex-wrap: wrap; }
.au-list-date { font-size: 0.75rem; color: var(--subtle); }

.au-role-stats { display: flex; gap: 20px; flex-wrap: wrap; }
.au-role-stat { display: flex; align-items: center; gap: 8px; font-size: 0.88rem; color: var(--text); }
.au-role-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.au-role-stat strong { font-weight: 800; }

.au-users-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px; }
.au-user-card {
  display: flex; align-items: center; gap: 14px; padding: 16px 20px;
  background: var(--surface); border: 1px solid var(--border-light); border-radius: var(--r-lg);
  box-shadow: var(--shadow-xs); transition: all 0.15s;
}
.au-user-card:hover { box-shadow: var(--shadow-sm); transform: translateY(-1px); }
.au-user-avatar {
  width: 40px; height: 40px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, #6366f1, #8b5cf6); color: #fff;
  font-size: 0.82rem; font-weight: 700; display: flex; align-items: center; justify-content: center;
}
.au-user-info { flex: 1; min-width: 0; }
.au-user-info strong { font-size: 0.9rem; font-weight: 700; color: var(--dark); display: block; }
.au-user-info p { font-size: 0.8rem; color: var(--muted); margin-top: 1px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.au-user-date { font-size: 0.72rem; color: var(--subtle); white-space: nowrap; }

@media (max-width: 768px) {
  .au-assign-grid { grid-template-columns: 1fr; }
  .au-users-grid { grid-template-columns: 1fr; }
  .au-role-stats { flex-direction: column; gap: 8px; }
}
</style>
