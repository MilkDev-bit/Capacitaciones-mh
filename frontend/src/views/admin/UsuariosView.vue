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
  <div class="page">
    <div class="page-header">
      <h2>Usuarios y Asignaciones</h2>
    </div>

    <div class="card">
      <h3>Asignar contenido a usuario</h3>
      <div class="form-row">
        <div class="field">
          <label>Usuario *</label>
          <select v-model="form.user_id">
            <option value="">— Selecciona usuario —</option>
            <option v-for="u in users" :key="u.id" :value="u.id">
              {{ u.name }} ({{ u.email }})
            </option>
          </select>
        </div>
        <div class="field">
          <label>Capacitación</label>
          <select v-model="form.capacitacion_id">
            <option value="">— Ninguna —</option>
            <option v-for="c in capacitaciones" :key="c.id" :value="c.id">{{ c.title }}</option>
          </select>
        </div>
        <div class="field">
          <label>Examen</label>
          <select v-model="form.examen_id">
            <option value="">— Ninguno —</option>
            <option v-for="e in examenes" :key="e.id" :value="e.id">{{ e.title }}</option>
          </select>
        </div>
      </div>
      <p v-if="error" class="msg error">{{ error }}</p>
      <p v-if="success" class="msg success">{{ success }}</p>
      <button class="btn-primary" :disabled="loading" @click="asignar">
        {{ loading ? 'Asignando...' : 'Asignar' }}
      </button>
    </div>

    <div class="section-title">Asignaciones activas</div>
    <div class="table-wrap">
      <table v-if="asignaciones.length">
        <thead>
          <tr><th>Usuario</th><th>Capacitación</th><th>Examen</th><th>Fecha</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="a in asignaciones" :key="a.id">
            <td>{{ userName(a.user_id) }}</td>
            <td>{{ a.capacitacion_id ? capTitle(a.capacitacion_id) : '—' }}</td>
            <td>{{ a.examen_id ? exTitle(a.examen_id) : '—' }}</td>
            <td>{{ new Date(a.assigned_at).toLocaleDateString() }}</td>
            <td><button class="btn-danger-sm" @click="desasignar(a.id)">Quitar</button></td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">No hay asignaciones aún.</div>
    </div>

    <div class="section-title" style="margin-top:2rem">Usuarios registrados</div>
    <div class="table-wrap">
      <table v-if="users.length">
        <thead><tr><th>Nombre</th><th>Email</th><th>Rol</th><th>Registro</th></tr></thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.name }}</td>
            <td>{{ u.email }}</td>
            <td><span :class="['badge', u.role]">{{ u.role }}</span></td>
            <td>{{ new Date(u.created_at).toLocaleDateString() }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
h2 { font-size: 1.4rem; font-weight: 700; color: #1e293b; }
.btn-primary { background: #3b82f6; color: white; border: none; padding: 9px 18px; border-radius: 8px; cursor: pointer; font-weight: 600; font-size: 0.9rem; margin-top: 0.5rem; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.card { background: white; border-radius: 12px; padding: 1.5rem; box-shadow: 0 2px 8px rgba(0,0,0,0.07); margin-bottom: 1.5rem; }
.card h3 { font-size: 1rem; font-weight: 700; margin-bottom: 1rem; color: #334155; }
.form-row { display: flex; gap: 1rem; flex-wrap: wrap; }
.field { display: flex; flex-direction: column; gap: 4px; flex: 1; min-width: 180px; }
label { font-size: 0.78rem; font-weight: 600; color: #64748b; }
select { padding: 9px 12px; border: 1.5px solid #e2e8f0; border-radius: 8px; font-size: 0.9rem; outline: none; }
select:focus { border-color: #3b82f6; }
.msg { padding: 8px 12px; border-radius: 6px; font-size: 0.85rem; margin: 8px 0; }
.msg.error { background: #fee2e2; color: #b91c1c; }
.msg.success { background: #dcfce7; color: #15803d; }
.section-title { font-weight: 700; color: #475569; font-size: 0.9rem; margin-bottom: 0.75rem; }
.table-wrap { background: white; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.07); overflow: hidden; margin-bottom: 1rem; }
table { width: 100%; border-collapse: collapse; }
th { background: #f8fafc; padding: 12px 16px; text-align: left; font-size: 0.78rem; font-weight: 700; color: #64748b; border-bottom: 1px solid #e2e8f0; }
td { padding: 12px 16px; border-bottom: 1px solid #f1f5f9; font-size: 0.88rem; }
.badge { font-size: 0.75rem; padding: 3px 10px; border-radius: 20px; font-weight: 600; }
.badge.admin { background: #dbeafe; color: #1d4ed8; }
.badge.user { background: #f0fdf4; color: #15803d; }
.btn-danger-sm { background: #fee2e2; color: #b91c1c; border: none; padding: 5px 10px; border-radius: 6px; cursor: pointer; font-size: 0.8rem; font-weight: 600; }
.btn-danger-sm:hover { background: #fecaca; }
.empty { padding: 2rem; text-align: center; color: #94a3b8; font-size: 0.95rem; }
</style>
