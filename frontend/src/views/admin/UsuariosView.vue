<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const users = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const examenes = ref<any[]>([])
const asignaciones = ref<any[]>([])

const form = ref({ user_id: '', capacitacion_id: '', examen_id: '' })
const error = ref('')
const success = ref('')
const loading = ref(false)

async function load() {
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
    success.value = 'Asignación realizada'
    form.value = { user_id: '', capacitacion_id: '', examen_id: '' }
    await load()
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

function userName(id: string) { return users.value.find(u => u.id === id)?.name || id }
function capTitle(id: string) { return capacitaciones.value.find(c => c.id === id)?.title || '—' }
function exTitle(id: string) { return examenes.value.find(e => e.id === id)?.title || '—' }
</script>

<template>
  <div>
    <div class="ph">
      <h1 class="ph-title">Usuarios y Asignaciones</h1>
    </div>

    <!-- Asignación -->
    <div class="form-card">
      <p class="form-card-title">Asignar contenido a usuario</p>
      <div class="form-row">
        <div class="field">
          <label>Usuario *</label>
          <select class="field-input" v-model="form.user_id">
            <option value="">— Selecciona usuario —</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ u.name }} ({{ u.email }})</option>
          </select>
        </div>
        <div class="field">
          <label>Capacitación</label>
          <select class="field-input" v-model="form.capacitacion_id">
            <option value="">— Ninguna —</option>
            <option v-for="c in capacitaciones" :key="c.id" :value="c.id">{{ c.title }}</option>
          </select>
        </div>
        <div class="field">
          <label>Exámen</label>
          <select class="field-input" v-model="form.examen_id">
            <option value="">— Ninguno —</option>
            <option v-for="e in examenes" :key="e.id" :value="e.id">{{ e.title }}</option>
          </select>
        </div>
      </div>
      <div v-if="error" class="alert alert-error">{{ error }}</div>
      <div v-if="success" class="alert alert-success">{{ success }}</div>
      <button class="btn btn-primary" :disabled="loading" @click="asignar">
        {{ loading ? 'Asignando…' : 'Asignar' }}
      </button>
    </div>

    <!-- Asignaciones activas -->
    <p class="section-lbl">Asignaciones activas</p>
    <div class="table-card">
      <table v-if="asignaciones.length">
        <thead><tr><th>Usuario</th><th>Capacitación</th><th>Exámen</th><th>Fecha</th><th></th></tr></thead>
        <tbody>
          <tr v-for="a in asignaciones" :key="a.id">
            <td>{{ userName(a.user_id) }}</td>
            <td class="cell-muted">{{ a.capacitacion_id ? capTitle(a.capacitacion_id) : '—' }}</td>
            <td class="cell-muted">{{ a.examen_id ? exTitle(a.examen_id) : '—' }}</td>
            <td class="cell-muted">{{ new Date(a.assigned_at).toLocaleDateString() }}</td>
            <td><button class="btn btn-danger btn-sm" @click="desasignar(a.id)">Quitar</button></td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state"><div class="empty-icon">📋</div><p>No hay asignaciones aún.</p></div>
    </div>

    <!-- Usuarios -->
    <p class="section-lbl" style="margin-top:24px">Usuarios registrados</p>
    <div class="table-card">
      <table v-if="users.length">
        <thead><tr><th>Nombre</th><th>Email</th><th>Rol</th><th>Registro</th></tr></thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td><strong>{{ u.name }}</strong></td>
            <td class="cell-muted">{{ u.email }}</td>
            <td>
              <span :class="['role-badge', u.role]">{{ u.role }}</span>
            </td>
            <td class="cell-muted">{{ new Date(u.created_at).toLocaleDateString() }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.ph { margin-bottom: 24px; }
.ph-title { font-size: 1.5rem; font-weight: 800; color: var(--dark); }
.form-card { background: var(--surface); border-radius: var(--r-lg); padding: 24px; box-shadow: var(--shadow-sm); margin-bottom: 24px; border-top: 4px solid var(--brand); }
.form-card-title { font-size: 1rem; font-weight: 700; color: var(--dark); margin-bottom: 16px; }
.form-row { display: flex; gap: 14px; flex-wrap: wrap; margin-bottom: 14px; }
.field { display: flex; flex-direction: column; gap: 5px; flex: 1; min-width: 180px; }
label { font-size: 0.78rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .04em; }
.section-lbl { font-size: 0.9rem; font-weight: 700; color: var(--dark); margin-bottom: 12px; }
.table-card { background: var(--surface); border-radius: var(--r-lg); box-shadow: var(--shadow-sm); overflow: hidden; margin-bottom: 12px; }
table { width: 100%; border-collapse: collapse; }
th { background: var(--bg); padding: 11px 16px; text-align: left; font-size: 0.75rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .05em; border-bottom: 1px solid var(--border); }
td { padding: 12px 16px; border-top: 1px solid var(--border-light); font-size: 0.9rem; }
.cell-muted { color: var(--muted); }
.role-badge { font-size: 0.73rem; font-weight: 700; padding: 2px 9px; border-radius: 20px; }
.role-badge.admin { background: var(--info-bg); color: var(--info); }
.role-badge.instructor { background: var(--brand-light); color: var(--brand-dark); }
.role-badge.user { background: var(--success-bg); color: var(--success); }
.empty-state { padding: 40px 20px; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 8px; }
.empty-icon { font-size: 2rem; }
.empty-state p { color: var(--muted); }
@media (max-width: 600px) { .form-row { flex-direction: column; } }
</style>
