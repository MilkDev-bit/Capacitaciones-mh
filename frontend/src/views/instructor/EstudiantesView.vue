<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const estudiantes = ref<any[]>([])
const users = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const examenes = ref<any[]>([])

const showAssign = ref(false)
const assignForm = ref({ user_id: '', capacitacion_id: '', examen_id: '' })
const assignType = ref<'capacitacion' | 'examen'>('capacitacion')
const assignError = ref('')
const assignSuccess = ref('')

async function load() {
  const [estRes, capRes, exRes, usrRes] = await Promise.all([
    api.get('/instructor/estudiantes'),
    api.get('/instructor/capacitaciones'),
    api.get('/instructor/examenes'),
    api.get('/instructor/users'),
  ])
  estudiantes.value = estRes.data || []
  capacitaciones.value = capRes.data || []
  examenes.value = exRes.data || []
  users.value = (usrRes.data || []).filter((u: any) => u.role === 'user')
}

onMounted(load)

async function asignar() {
  assignError.value = ''; assignSuccess.value = ''
  if (!assignForm.value.user_id) { assignError.value = 'Selecciona un usuario'; return }
  const body: any = { user_id: assignForm.value.user_id }
  if (assignType.value === 'capacitacion') {
    if (!assignForm.value.capacitacion_id) { assignError.value = 'Selecciona una capacitación'; return }
    body.capacitacion_id = assignForm.value.capacitacion_id
  } else {
    if (!assignForm.value.examen_id) { assignError.value = 'Selecciona un examen'; return }
    body.examen_id = assignForm.value.examen_id
  }
  try {
    await api.post('/instructor/asignar', body)
    assignSuccess.value = 'Asignación realizada exitosamente'
    assignForm.value = { user_id: '', capacitacion_id: '', examen_id: '' }
    await load()
  } catch (e: any) {
    assignError.value = e.response?.data?.error || 'Error al asignar'
  }
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>Estudiantes</h2>
        <p class="subtitle">Gestiona quién accede a tus cursos y exámenes. También puedes asignar contenido directamente.</p>
      </div>
      <button class="btn-primary" @click="showAssign = !showAssign">
        {{ showAssign ? 'Cancelar' : '+ Asignar contenido' }}
      </button>
    </div>

    <!-- Formulario de asignación -->
    <div v-if="showAssign" class="card form-card">
      <h3>Asignar a usuario</h3>
      <div class="form-row">
        <div class="field">
          <label>Usuario *</label>
          <select v-model="assignForm.user_id">
            <option value="">Seleccionar usuario...</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ u.name }} ({{ u.email }})</option>
          </select>
        </div>
        <div class="field">
          <label>Tipo de contenido *</label>
          <select v-model="assignType">
            <option value="capacitacion">Capacitación</option>
            <option value="examen">Examen</option>
          </select>
        </div>
      </div>
      <div class="form-row">
        <div class="field" v-if="assignType === 'capacitacion'">
          <label>Capacitación *</label>
          <select v-model="assignForm.capacitacion_id">
            <option value="">Seleccionar capacitación...</option>
            <option v-for="c in capacitaciones" :key="c.id" :value="c.id">{{ c.title }}</option>
          </select>
        </div>
        <div class="field" v-else>
          <label>Examen *</label>
          <select v-model="assignForm.examen_id">
            <option value="">Seleccionar examen...</option>
            <option v-for="e in examenes" :key="e.id" :value="e.id">{{ e.title }}</option>
          </select>
        </div>
      </div>
      <p v-if="assignError" class="msg error">{{ assignError }}</p>
      <p v-if="assignSuccess" class="msg success">{{ assignSuccess }}</p>
      <button class="btn-primary" @click="asignar">Confirmar asignación</button>
    </div>

    <!-- Info tip -->
    <div class="info-tip">
      <span>💡</span>
      <span>Los estudiantes que se inscriban a tus cursos <strong>públicos</strong> aparecerán aquí automáticamente.</span>
    </div>

    <!-- Tabla de estudiantes -->
    <div class="table-wrap">
      <table v-if="estudiantes.length">
        <thead>
          <tr><th>Nombre</th><th>Correo</th><th>Miembro desde</th></tr>
        </thead>
        <tbody>
          <tr v-for="u in estudiantes" :key="u.id">
            <td>
              <div class="user-cell">
                <div class="avatar">{{ u.name.charAt(0).toUpperCase() }}</div>
                <strong>{{ u.name }}</strong>
              </div>
            </td>
            <td class="email">{{ u.email }}</td>
            <td>{{ new Date(u.created_at).toLocaleDateString() }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty">
        <p>Aún no tienes estudiantes inscritos.</p>
        <p class="hint">Publica un curso o asigna contenido a usuarios para empezar.</p>
      </div>
    </div>

    <!-- Estadísticas rápidas -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-num">{{ capacitaciones.length }}</div>
        <div class="stat-label">Cursos creados</div>
      </div>
      <div class="stat-card">
        <div class="stat-num">{{ capacitaciones.filter(c => c.is_public).length }}</div>
        <div class="stat-label">Cursos públicos</div>
      </div>
      <div class="stat-card">
        <div class="stat-num">{{ examenes.length }}</div>
        <div class="stat-label">Exámenes creados</div>
      </div>
      <div class="stat-card purple">
        <div class="stat-num">{{ estudiantes.length }}</div>
        <div class="stat-label">Estudiantes inscritos</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; gap: 1rem; }
.page-header h2 { font-size: 1.4rem; font-weight: 700; color: #1e293b; margin: 0; }
.subtitle { color: #64748b; font-size: 0.85rem; margin-top: 4px; }
.btn-primary { background: #7c3aed; color: white; border: none; padding: 9px 18px; border-radius: 8px; cursor: pointer; font-weight: 600; font-size: 0.9rem; white-space: nowrap; }
.btn-primary:hover { background: #6d28d9; }
.card { background: white; border-radius: 12px; padding: 1.5rem; box-shadow: 0 2px 8px rgba(0,0,0,0.07); margin-bottom: 1.5rem; }
.form-card h3 { font-size: 1rem; font-weight: 700; margin-bottom: 1rem; color: #334155; }
.form-row { display: flex; gap: 1rem; margin-bottom: 1rem; flex-wrap: wrap; }
.field { display: flex; flex-direction: column; gap: 4px; flex: 1; min-width: 200px; }
label { font-size: 0.78rem; font-weight: 600; color: #64748b; }
select, input { padding: 9px 12px; border: 1.5px solid #e2e8f0; border-radius: 8px; font-size: 0.9rem; outline: none; font-family: inherit; }
select:focus, input:focus { border-color: #7c3aed; }
.msg { font-size: 0.85rem; margin-top: 8px; padding: 8px 12px; border-radius: 6px; }
.msg.error { background: #fee2e2; color: #dc2626; }
.msg.success { background: #d1fae5; color: #065f46; }
.info-tip {
  background: #ede9fe; border-radius: 10px; padding: 12px 16px;
  display: flex; align-items: center; gap: 10px; font-size: 0.87rem; color: #4c1d95;
  margin-bottom: 1.5rem;
}
.table-wrap { background: white; border-radius: 12px; box-shadow: 0 1px 6px rgba(0,0,0,0.08); overflow: hidden; margin-bottom: 1.5rem; }
table { width: 100%; border-collapse: collapse; }
th { background: #f8fafc; padding: 10px 16px; text-align: left; font-size: 0.8rem; font-weight: 700; color: #64748b; text-transform: uppercase; letter-spacing: 0.05em; }
td { padding: 12px 16px; border-top: 1px solid #f1f5f9; font-size: 0.9rem; vertical-align: middle; }
.user-cell { display: flex; align-items: center; gap: 10px; }
.avatar {
  width: 32px; height: 32px; border-radius: 50%; background: #ede9fe; color: #7c3aed;
  display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 0.85rem;
}
.email { color: #64748b; }
.empty { padding: 3rem; text-align: center; color: #64748b; }
.hint { font-size: 0.85rem; margin-top: 6px; }
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; }
.stat-card {
  background: white; border-radius: 12px; padding: 1.2rem 1.5rem;
  box-shadow: 0 1px 6px rgba(0,0,0,0.08); text-align: center;
}
.stat-card.purple { background: #7c3aed; color: white; }
.stat-num { font-size: 2rem; font-weight: 800; color: #1e293b; }
.stat-card.purple .stat-num { color: white; }
.stat-label { font-size: 0.8rem; color: #64748b; margin-top: 4px; }
.stat-card.purple .stat-label { color: #ddd6fe; }
</style>
