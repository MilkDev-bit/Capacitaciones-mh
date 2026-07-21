<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '../../api'
import { toast } from '../../utils/toast'

const entregas = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const loading = ref(true)
const selectedCurso = ref('')
const searchQuery = ref('')

function fileUrl(path: string) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  const base = (import.meta.env.VITE_API_URL || '').replace(/\/api\/?$/, '')
  return `${base}${path}`
}

function formatDate(dateStr: string) {
  if (!dateStr) return 'N/A'
  const d = new Date(dateStr)
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function formatSize(bytes: number) {
  if (!bytes) return '0 KB'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
}

async function load() {
  loading.value = true
  try {
    const [entRes, capRes] = await Promise.all([
      selectedCurso.value
        ? api.get(`/instructor/capacitaciones/${selectedCurso.value}/entregas`)
        : api.get('/instructor/entregas'),
      api.get('/instructor/capacitaciones'),
    ])
    entregas.value = entRes.data?.entregas || entRes.data || []
    if (capRes.data) {
      capacitaciones.value = capRes.data
    }
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al cargar las entregas')
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function onCursoChange() {
  await load()
}

const filteredEntregas = computed(() => {
  if (!searchQuery.value) return entregas.value
  const q = searchQuery.value.toLowerCase()
  return entregas.value.filter(e =>
    (e.user_name && e.user_name.toLowerCase().includes(q)) ||
    (e.user_email && e.user_email.toLowerCase().includes(q)) ||
    (e.curso_title && e.curso_title.toLowerCase().includes(q)) ||
    (e.leccion_title && e.leccion_title.toLowerCase().includes(q)) ||
    (e.file_name && e.file_name.toLowerCase().includes(q))
  )
})

const stats = computed(() => {
  const total = entregas.value.length
  const today = new Date().toDateString()
  const hoyCount = entregas.value.filter(e => e.created_at && new Date(e.created_at).toDateString() === today).length
  const uniqueCursos = new Set(entregas.value.map(e => e.capacitacion_id)).size
  const uniqueUsers = new Set(entregas.value.map(e => e.user_id)).size
  return { total, hoyCount, uniqueCursos, uniqueUsers }
})
</script>

<template>
  <div class="entregas-container">
    <div class="ph">
      <div>
        <h1 class="ph-title">Entregas de Actividades</h1>
        <p class="ph-sub">Revisa y descarga los archivos adjuntos subidos por los estudiantes en tareas y actividades.</p>
      </div>
      <button class="btn btn-secondary" @click="load" :disabled="loading">
        <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" :class="{ 'spin': loading }"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
        Actualizar
      </button>
    </div>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div class="stat-box brand">
        <div class="stat-num">{{ stats.total }}</div>
        <div class="stat-lbl">Entregas totales</div>
      </div>
      <div class="stat-box">
        <div class="stat-num">{{ stats.hoyCount }}</div>
        <div class="stat-lbl">Recibidas hoy</div>
      </div>
      <div class="stat-box">
        <div class="stat-num">{{ stats.uniqueCursos }}</div>
        <div class="stat-lbl">Cursos con entregas</div>
      </div>
      <div class="stat-box">
        <div class="stat-num">{{ stats.uniqueUsers }}</div>
        <div class="stat-lbl">Estudiantes que han entregado</div>
      </div>
    </div>

    <!-- Filters Bar -->
    <div class="filters-bar">
      <div class="filter-group">
        <label class="filter-lbl">Filtrar por curso:</label>
        <select v-model="selectedCurso" @change="onCursoChange" class="field-select">
          <option value="">Todos los cursos</option>
          <option v-for="c in capacitaciones" :key="c.id" :value="c.id">{{ c.title }}</option>
        </select>
      </div>

      <div class="filter-group search-group">
        <label class="filter-lbl">Buscar:</label>
        <div class="search-input-wrapper">
          <svg class="search-icon" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
          <input v-model="searchQuery" class="field-input search-input" placeholder="Nombre, correo, curso o archivo..." />
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="table-card">
      <div v-if="loading" class="loading-state">
        <span class="btn-spinner" style="width:28px;height:28px;border-width:3px"></span>
        <p>Cargando entregas...</p>
      </div>
      <table v-else-if="filteredEntregas.length">
        <thead>
          <tr>
            <th>Estudiante</th>
            <th>Curso / Actividad</th>
            <th>Archivo</th>
            <th>Fecha de entrega</th>
            <th style="text-align: right">Acciones</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in filteredEntregas" :key="e.id">
            <td>
              <div class="user-cell">
                <div class="ava" v-if="!e.avatar_url">{{ (e.user_name || 'U').charAt(0).toUpperCase() }}</div>
                <img v-else :src="e.avatar_url" class="ava-img" alt="Avatar" />
                <div class="user-info">
                  <strong>{{ e.user_name || 'Estudiante' }}</strong>
                  <span class="cell-muted">{{ e.user_email }}</span>
                </div>
              </div>
            </td>
            <td>
              <div class="course-cell">
                <strong class="course-title">{{ e.curso_title || 'Capacitación' }}</strong>
                <span class="lesson-badge">📌 {{ e.leccion_title || 'Actividad' }}</span>
              </div>
            </td>
            <td>
              <div class="file-cell">
                <span class="file-icon">📎</span>
                <div class="file-info">
                  <strong class="file-name" :title="e.file_name">{{ e.file_name || 'Archivo adjunto' }}</strong>
                  <span class="file-size">{{ formatSize(e.file_size) }}</span>
                </div>
              </div>
            </td>
            <td class="cell-muted">
              {{ formatDate(e.created_at) }}
            </td>
            <td>
              <div class="actions-cell">
                <a :href="fileUrl(e.file_path)" target="_blank" rel="noopener noreferrer" class="btn btn-secondary btn-sm" title="Ver en nueva pestaña">
                  <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
                  Ver
                </a>
                <a :href="fileUrl(e.file_path)" :download="e.file_name || 'entrega'" class="btn btn-primary btn-sm" title="Descargar archivo">
                  <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>
                  Descargar
                </a>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        <div class="empty-icon">📂</div>
        <h3>No hay entregas registradas</h3>
        <p v-if="searchQuery || selectedCurso">No se encontraron entregas con los filtros seleccionados.</p>
        <p v-else>Cuando los estudiantes suban archivos en las actividades programadas, aparecerán en esta lista.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.entregas-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.ph {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 12px;
}
.ph-title {
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--dark);
}
.ph-sub {
  color: var(--muted);
  font-size: 0.87rem;
  margin-top: 4px;
}

/* Stats */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.stat-box {
  background: var(--surface);
  border-radius: var(--r-lg);
  padding: 24px;
  text-align: center;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-light);
  transition: transform 0.2s;
}
.stat-box:hover {
  transform: translateY(-2px);
}
.stat-box.brand {
  background: var(--brand);
}
.stat-num {
  font-size: 2rem;
  font-weight: 800;
  color: var(--dark);
}
.stat-box.brand .stat-num {
  color: #fff;
}
.stat-lbl {
  font-size: 0.78rem;
  color: var(--muted);
  margin-top: 4px;
}
.stat-box.brand .stat-lbl {
  color: rgba(255, 255, 255, 0.85);
}

/* Filters */
.filters-bar {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  flex-wrap: wrap;
  gap: 16px;
  background: var(--surface);
  padding: 16px 20px;
  border-radius: var(--r-lg);
  border: 1px solid var(--border-light);
  box-shadow: var(--shadow-sm);
}
.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
  min-width: 220px;
}
.search-group {
  flex: 2;
}
.filter-lbl {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--dark);
}
.field-select {
  padding: 8px 12px;
  border-radius: var(--r);
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--dark);
  font-size: 0.88rem;
}
.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: 12px;
  color: var(--muted);
}
.search-input {
  padding: 8px 12px 8px 36px;
  width: 100%;
  border-radius: var(--r);
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--dark);
  font-size: 0.88rem;
}

/* Table */
.table-card {
  background: var(--surface);
  border-radius: var(--r-lg);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-light);
  overflow-x: auto;
}
table {
  width: 100%;
  border-collapse: collapse;
}
th {
  background: var(--bg);
  padding: 12px 16px;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--border);
}
td {
  padding: 14px 16px;
  border-top: 1px solid var(--border-light);
  font-size: 0.9rem;
  vertical-align: middle;
}
.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}
.ava {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--brand-light);
  color: var(--brand-dark);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.9rem;
  flex-shrink: 0;
}
.ava-img {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}
.user-info {
  display: flex;
  flex-direction: column;
}
.user-info strong {
  color: var(--dark);
  font-size: 0.92rem;
}
.cell-muted {
  color: var(--muted);
  font-size: 0.8rem;
}
.course-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.course-title {
  color: var(--dark);
  font-size: 0.9rem;
}
.lesson-badge {
  display: inline-block;
  font-size: 0.78rem;
  color: var(--brand);
  font-weight: 600;
}
.file-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  max-width: 260px;
}
.file-icon {
  font-size: 1.4rem;
  flex-shrink: 0;
}
.file-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.file-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--dark);
  font-size: 0.88rem;
}
.file-size {
  font-size: 0.75rem;
  color: var(--muted);
}
.actions-cell {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
.loading-state,
.empty-state {
  padding: 60px 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.empty-icon {
  font-size: 3rem;
}
.empty-state h3, .loading-state p {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--dark);
}
.empty-state p {
  color: var(--muted);
  font-size: 0.9rem;
  max-width: 450px;
}
.spin {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media (max-width: 900px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 600px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .actions-cell {
    flex-direction: column;
  }
}
</style>
