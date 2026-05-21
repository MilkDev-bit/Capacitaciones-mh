<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '../../api'
import { toast } from '../../utils/toast'

const estudiantes = ref<any[]>([])
const users = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const examenes = ref<any[]>([])

const showAssign = ref(false)
const selectedUsers = ref<string[]>([])
const selectedCaps = ref<string[]>([])
const selectedExams = ref<string[]>([])
const userSearch = ref('')
const capSearch = ref('')
const examSearch = ref('')
const assigning = ref(false)

const filteredUsers = computed(() =>
  users.value.filter(u =>
    !userSearch.value ||
    u.name.toLowerCase().includes(userSearch.value.toLowerCase()) ||
    u.email.toLowerCase().includes(userSearch.value.toLowerCase())
  )
)
const filteredCaps = computed(() =>
  capacitaciones.value.filter(c =>
    !capSearch.value || c.title.toLowerCase().includes(capSearch.value.toLowerCase())
  )
)
const filteredExams = computed(() =>
  examenes.value.filter(e =>
    !examSearch.value || e.title.toLowerCase().includes(examSearch.value.toLowerCase())
  )
)
const totalTasks = computed(() =>
  selectedUsers.value.length * (selectedCaps.value.length + selectedExams.value.length)
)

function toggleAll(selected: string[], pool: any[], setter: (v: string[]) => void) {
  setter(selected.length === pool.length ? [] : pool.map(i => i.id))
}

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
  if (selectedUsers.value.length === 0) { toast.error('Selecciona al menos un estudiante'); return }
  if (selectedCaps.value.length === 0 && selectedExams.value.length === 0) {
    toast.error('Selecciona al menos una capacitación o examen'); return
  }
  assigning.value = true
  const tasks: Promise<any>[] = []
  for (const uid of selectedUsers.value) {
    for (const cid of selectedCaps.value) tasks.push(api.post('/instructor/asignar', { user_id: uid, capacitacion_id: cid }))
    for (const eid of selectedExams.value) tasks.push(api.post('/instructor/asignar', { user_id: uid, examen_id: eid }))
  }
  const results = await Promise.allSettled(tasks)
  const ok = results.filter(r => r.status === 'fulfilled').length
  const fail = results.filter(r => r.status === 'rejected').length
  if (ok > 0) toast.success(`${ok} asignación${ok !== 1 ? 'es' : ''} realizada${ok !== 1 ? 's' : ''} correctamente`)
  if (fail > 0) toast.warning(`${fail} ya existía${fail !== 1 ? 'n' : ''} o no se pudieron completar`)
  selectedUsers.value = []
  selectedCaps.value = []
  selectedExams.value = []
  assigning.value = false
  await load()
}
</script>

<template>
  <div>
    <div class="ph">
      <div>
        <h1 class="ph-title">Estudiantes</h1>
        <p class="ph-sub">Gestiona quién accede a tus cursos y exámenes.</p>
      </div>
      <button class="btn btn-primary" @click="showAssign = !showAssign">
        {{ showAssign ? 'Cancelar' : '+ Asignar contenido' }}
      </button>
    </div>

    <!-- Formulario de asignación -->
    <div v-if="showAssign" class="form-card">
      <p class="form-card-title">Asignar contenido</p>
      <div class="assign-grid">

        <!-- Estudiantes -->
        <div class="assign-section">
          <div class="assign-sec-header">
            <span class="assign-sec-title">Estudiantes</span>
            <span v-if="selectedUsers.length" class="assign-badge">{{ selectedUsers.length }}</span>
            <button type="button" class="assign-toggle-all" @click="toggleAll(selectedUsers, filteredUsers, v => selectedUsers = v)">
              {{ selectedUsers.length === filteredUsers.length && filteredUsers.length > 0 ? 'Quitar todos' : 'Todos' }}
            </button>
          </div>
          <input v-model="userSearch" class="field-input assign-search" placeholder="Buscar estudiante..." />
          <div class="assign-list">
            <label v-for="u in filteredUsers" :key="u.id" class="assign-item">
              <input type="checkbox" :value="u.id" v-model="selectedUsers" class="assign-check" />
              <div class="assign-ava">{{ u.name.charAt(0).toUpperCase() }}</div>
              <div class="assign-info">
                <span class="assign-name">{{ u.name }}</span>
                <span class="assign-email">{{ u.email }}</span>
              </div>
            </label>
            <div v-if="filteredUsers.length === 0" class="assign-empty">Sin resultados</div>
          </div>
        </div>

        <!-- Capacitaciones -->
        <div class="assign-section">
          <div class="assign-sec-header">
            <span class="assign-sec-title">Capacitaciones</span>
            <span v-if="selectedCaps.length" class="assign-badge">{{ selectedCaps.length }}</span>
            <span class="assign-optional">opcional</span>
            <button type="button" class="assign-toggle-all" @click="toggleAll(selectedCaps, filteredCaps, v => selectedCaps = v)">
              {{ selectedCaps.length === filteredCaps.length && filteredCaps.length > 0 ? 'Quitar todos' : 'Todos' }}
            </button>
          </div>
          <input v-model="capSearch" class="field-input assign-search" placeholder="Buscar capacitación..." />
          <div class="assign-list">
            <label v-for="c in filteredCaps" :key="c.id" class="assign-item">
              <input type="checkbox" :value="c.id" v-model="selectedCaps" class="assign-check" />
              <div class="assign-info">
                <span class="assign-name">{{ c.title }}</span>
                <span class="assign-email">{{ c.type }}</span>
              </div>
            </label>
            <div v-if="filteredCaps.length === 0" class="assign-empty">Sin capacitaciones</div>
          </div>
        </div>

        <!-- Exámenes -->
        <div class="assign-section">
          <div class="assign-sec-header">
            <span class="assign-sec-title">Exámenes</span>
            <span v-if="selectedExams.length" class="assign-badge">{{ selectedExams.length }}</span>
            <span class="assign-optional">opcional</span>
            <button type="button" class="assign-toggle-all" @click="toggleAll(selectedExams, filteredExams, v => selectedExams = v)">
              {{ selectedExams.length === filteredExams.length && filteredExams.length > 0 ? 'Quitar todos' : 'Todos' }}
            </button>
          </div>
          <input v-model="examSearch" class="field-input assign-search" placeholder="Buscar examen..." />
          <div class="assign-list">
            <label v-for="e in filteredExams" :key="e.id" class="assign-item">
              <input type="checkbox" :value="e.id" v-model="selectedExams" class="assign-check" />
              <div class="assign-info">
                <span class="assign-name">{{ e.title }}</span>
              </div>
            </label>
            <div v-if="filteredExams.length === 0" class="assign-empty">Sin exámenes</div>
          </div>
        </div>

      </div>

      <div class="assign-footer">
        <span v-if="totalTasks > 0" class="assign-summary">
          {{ selectedUsers.length }} estudiante{{ selectedUsers.length !== 1 ? 's' : '' }} ×
          {{ selectedCaps.length + selectedExams.length }} contenido{{ (selectedCaps.length + selectedExams.length) !== 1 ? 's' : '' }} =
          <strong>{{ totalTasks }} asignación{{ totalTasks !== 1 ? 'es' : '' }}</strong>
        </span>
        <button class="btn btn-primary" @click="asignar" :disabled="assigning || totalTasks === 0">
          <span v-if="assigning" class="btn-spinner"></span>
          {{ assigning ? 'Asignando...' : 'Confirmar asignación' }}
        </button>
      </div>
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
      <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/></svg>
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
        <div class="empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.2" viewBox="0 0 24 24"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg></div>
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
.form-card { background: var(--surface); border-radius: var(--r-lg); padding: 28px; box-shadow: var(--shadow-sm); margin-bottom: 24px; border: 1px solid var(--border-light); }
.form-card-title { font-size: 1.1rem; font-weight: 800; color: var(--dark); margin-bottom: 20px; }

/* Multi-assign grid */
.assign-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-bottom: 20px; }
.assign-section { background: var(--bg); border-radius: var(--r); border: 1px solid var(--border); overflow: hidden; display: flex; flex-direction: column; }
.assign-sec-header { display: flex; align-items: center; gap: 6px; padding: 10px 14px; background: var(--surface); border-bottom: 1px solid var(--border); flex-wrap: wrap; }
.assign-sec-title { font-size: 0.82rem; font-weight: 700; color: var(--dark); }
.assign-badge { background: var(--brand); color: #fff; font-size: 0.72rem; font-weight: 700; padding: 1px 7px; border-radius: 9999px; }
.assign-optional { font-size: 0.72rem; color: var(--muted); background: var(--bg); border: 1px solid var(--border); padding: 1px 7px; border-radius: 9999px; }
.assign-toggle-all { margin-left: auto; font-size: 0.75rem; font-weight: 600; color: var(--brand); background: none; border: none; cursor: pointer; padding: 0; }
.assign-toggle-all:hover { text-decoration: underline; }
.assign-search { border-radius: 0; border-left: none; border-right: none; border-top: none; font-size: 0.82rem; padding: 8px 14px; }
.assign-list { flex: 1; overflow-y: auto; max-height: 240px; padding: 6px 0; }
.assign-item { display: flex; align-items: center; gap: 10px; padding: 8px 14px; cursor: pointer; transition: background .12s; }
.assign-item:hover { background: var(--surface); }
.assign-item:has(.assign-check:checked) { background: color-mix(in srgb, var(--brand) 8%, transparent); }
.assign-check { accent-color: var(--brand); width: 15px; height: 15px; cursor: pointer; flex-shrink: 0; }
.assign-ava { width: 30px; height: 30px; border-radius: 50%; background: var(--brand-light); color: var(--brand-dark); display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 0.82rem; flex-shrink: 0; }
.assign-info { display: flex; flex-direction: column; min-width: 0; }
.assign-name { font-size: 0.85rem; font-weight: 600; color: var(--dark); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.assign-email { font-size: 0.75rem; color: var(--muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.assign-empty { padding: 20px; text-align: center; color: var(--muted); font-size: 0.82rem; }
.assign-footer { display: flex; align-items: center; justify-content: flex-end; gap: 16px; padding-top: 4px; }
.assign-summary { font-size: 0.85rem; color: var(--muted); }
.assign-summary strong { color: var(--dark); }

.stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
.stat-box { background: var(--surface); border-radius: var(--r-lg); padding: 24px; text-align: center; box-shadow: var(--shadow-sm); border: 1px solid var(--border-light); transition: transform 0.2s; }
.stat-box:hover { transform: translateY(-2px); }
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
@media (max-width: 900px) { .assign-grid { grid-template-columns: 1fr; } }
@media (max-width: 700px) { .stats-grid { grid-template-columns: repeat(2, 1fr); } .ph { flex-direction: column; } }
</style>
