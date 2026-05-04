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
  <div>
    <div class="ph">
      <div>
        <h1 class="ph-title">Estudiantes</h1>
        <p class="ph-sub">Gestiona qui\u00e9n accede a tus cursos y ex\u00e1menes.</p>
      </div>
      <button class="btn btn-primary" @click="showAssign = !showAssign">
        {{ showAssign ? 'Cancelar' : '+ Asignar contenido' }}
      </button>
    </div>

    <!-- Formulario de asignaci\u00f3n -->
    <div v-if="showAssign" class="form-card">
      <p class="form-card-title">Asignar a usuario</p>
      <div class="form-row">
        <div class="field">
          <label>Usuario *</label>
          <select class="field-input" v-model="assignForm.user_id">
            <option value="">Seleccionar usuario...</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ u.name }} ({{ u.email }})</option>
          </select>
        </div>
        <div class="field">
          <label>Tipo de contenido *</label>
          <select class="field-input" v-model="assignType">
            <option value="capacitacion">Capacitaci\u00f3n</option>
            <option value="examen">Ex\u00e1men</option>
          </select>
        </div>
      </div>
      <div class="form-row">
        <div class="field" v-if="assignType === 'capacitacion'">
          <label>Capacitaci\u00f3n *</label>
          <select class="field-input" v-model="assignForm.capacitacion_id">
            <option value="">Seleccionar capacitaci\u00f3n...</option>
            <option v-for="c in capacitaciones" :key="c.id" :value="c.id">{{ c.title }}</option>
          </select>
        </div>
        <div class="field" v-else>
          <label>Ex\u00e1men *</label>
          <select class="field-input" v-model="assignForm.examen_id">
            <option value="">Seleccionar ex\u00e1men...</option>
            <option v-for="e in examenes" :key="e.id" :value="e.id">{{ e.title }}</option>
          </select>
        </div>
      </div>
      <div v-if="assignError" class="alert alert-error">{{ assignError }}</div>
      <div v-if="assignSuccess" class="alert alert-success">{{ assignSuccess }}</div>
      <button class="btn btn-primary" @click="asignar">Confirmar asignación</button>
    </div>

    <!-- Stats -->
    <div class="stats-grid">
      <div class="stat-box">
        <div class="stat-num">{{ capacitaciones.length }}</div>
        <div class="stat-lbl">Cursos creados</div>
      </div>
      <div class="stat-box">
        <div class="stat-num">{{ capacitaciones.filter(c => c.is_public).length }}</div>
        <div class="stat-lbl">Cursos públicos</div>
      </div>
      <div class="stat-box">
        <div class="stat-num">{{ examenes.length }}</div>
        <div class="stat-lbl">Exámenes creados</div>
      </div>
      <div class="stat-box brand">
        <div class="stat-num">{{ estudiantes.length }}</div>
        <div class="stat-lbl">Estudiantes inscritos</div>
      </div>
    </div>

    <!-- Tabla de estudiantes -->
    <div class="info-tip">
      <span>💡</span>
      <span>Los estudiantes que se inscriban a tus cursos <strong>públicos</strong> aparecerán aquí automáticamente.</span>
    </div>
    <div class="table-card">
      <table v-if="estudiantes.length">
        <thead>
          <tr><th>Nombre</th><th>Correo</th><th>Miembro desde</th></tr>
        </thead>
        <tbody>
          <tr v-for="u in estudiantes" :key="u.id">
            <td>
              <div class="user-cell">
                <div class="ava">{{ u.name.charAt(0).toUpperCase() }}</div>
                <strong>{{ u.name }}</strong>
              </div>
            </td>
            <td class="cell-muted">{{ u.email }}</td>
            <td class="cell-muted">{{ new Date(u.created_at).toLocaleDateString() }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        <div class="empty-icon">🎓</div>
        <h3>Aún no tienes estudiantes inscritos</h3>
        <p>Publica un curso o asigna contenido a usuarios para empezar.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ph { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 12px; margin-bottom: 24px; }
.ph-title { font-size: 1.5rem; font-weight: 800; color: var(--dark); }
.ph-sub { color: var(--muted); font-size: 0.87rem; margin-top: 4px; }
.form-card { background: var(--surface); border-radius: var(--r-lg); padding: 24px; box-shadow: var(--shadow-sm); margin-bottom: 24px; border-top: 4px solid var(--brand); }
.form-card-title { font-size: 1rem; font-weight: 700; color: var(--dark); margin-bottom: 16px; }
.form-row { display: flex; gap: 14px; flex-wrap: wrap; margin-bottom: 14px; }
.field { display: flex; flex-direction: column; gap: 5px; flex: 1; min-width: 200px; }
label { font-size: 0.78rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .04em; }
.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; margin-bottom: 24px; }
.stat-box { background: var(--surface); border-radius: var(--r-lg); padding: 20px; text-align: center; box-shadow: var(--shadow-sm); }
.stat-box.brand { background: var(--brand); }
.stat-num { font-size: 2rem; font-weight: 800; color: var(--dark); }
.stat-box.brand .stat-num { color: #fff; }
.stat-lbl { font-size: 0.78rem; color: var(--muted); margin-top: 4px; }
.stat-box.brand .stat-lbl { color: rgba(255,255,255,.8); }
.info-tip { background: var(--brand-light); border-radius: var(--r); padding: 12px 16px; display: flex; align-items: center; gap: 10px; font-size: 0.87rem; color: var(--brand-darker); margin-bottom: 16px; border-left: 3px solid var(--brand); }
.table-card { background: var(--surface); border-radius: var(--r-lg); box-shadow: var(--shadow-sm); overflow: hidden; }
table { width: 100%; border-collapse: collapse; }
th { background: var(--bg); padding: 11px 16px; text-align: left; font-size: 0.75rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: .05em; border-bottom: 1px solid var(--border); }
td { padding: 12px 16px; border-top: 1px solid var(--border-light); font-size: 0.9rem; }
.user-cell { display: flex; align-items: center; gap: 10px; }
.ava { width: 32px; height: 32px; border-radius: 50%; background: var(--brand-light); color: var(--brand-dark); display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 0.85rem; flex-shrink: 0; }
.cell-muted { color: var(--muted); }
.empty-state { padding: 60px 20px; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.empty-icon { font-size: 3rem; }
.empty-state h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
.empty-state p { color: var(--muted); font-size: 0.9rem; }
@media (max-width: 700px) { .stats-grid { grid-template-columns: repeat(2, 1fr); } .ph { flex-direction: column; } }
</style>
