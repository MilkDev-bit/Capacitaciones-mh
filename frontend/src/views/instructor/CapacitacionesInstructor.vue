<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
import { toast } from '../../utils/toast'
import CourseWizardModal from '../../components/CourseWizardModal.vue'
import CourseEditorDrawer from '../../components/CourseEditorDrawer.vue'

const courses = ref<any[]>([])
const loading = ref(false)

const showWizard = ref(false)
const selectedCourse = ref<any>(null)
const showDrawer = ref(false)

async function fetchCourses() {
  loading.value = true
  try {
    const res = await api.get('/instructor/capacitaciones')
    courses.value = res.data || []
  } catch (e: any) {
    toast.error('Error al cargar cursos')
  } finally {
    loading.value = false
  }
}

onMounted(fetchCourses)

async function togglePublic(course: any) {
  try {
    await api.patch(`/instructor/capacitaciones/${course.id}/toggle-public`)
    course.is_public = !course.is_public
    toast.success('Visibilidad actualizada')
  } catch (e) {
    toast.error('Error al cambiar visibilidad')
  }
}

async function deleteCourse(id: string) {
  if (!confirm('¿Estás seguro de eliminar este curso? Se perderán todas sus lecciones.')) return
  try {
    await api.delete(`/instructor/capacitaciones/${id}`)
    toast.success('Curso eliminado')
    fetchCourses()
  } catch (e) {
    toast.error('Error al eliminar curso')
  }
}

async function resetCode(id: string) {
  if (!confirm('¿Generar nuevo código de acceso? El anterior dejará de funcionar.')) return
  try {
    const res = await api.post(`/instructor/capacitaciones/${id}/reset-codigo`)
    toast.success('Código actualizado: ' + res.data.codigo_acceso)
    fetchCourses()
  } catch (e) {
    toast.error('Error al resetear código')
  }
}

function openEdit(course: any) {
  selectedCourse.value = course
  showDrawer.value = true
}

function copyCode(code: string) {
  if (!code) return
  navigator.clipboard.writeText(code)
  toast.success('Código copiado')
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <h1 class="page-title">Cursos Creados</h1>
        <p class="page-desc">Administra tus capacitaciones y lecciones</p>
      </div>
      <button class="btn btn-primary" @click="showWizard = true">
        <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 4v16m8-8H4"/></svg>
        Nuevo Curso
      </button>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Cargando cursos...</p>
    </div>
    
    <div v-else-if="courses.length === 0" class="empty-state slide-down-enter-active">
      <div class="empty-icon">
        <svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
      </div>
      <h3>Aún no has creado cursos</h3>
      <p>Comienza creando tu primera capacitación para tus estudiantes.</p>
      <button class="btn btn-primary" @click="showWizard = true">Crear Curso</button>
    </div>

    <div v-else class="courses-grid">
      <div v-for="c in courses" :key="c.id" class="course-card slide-down-enter-active">
        <div class="course-thumb" :style="{ background: c.thumbnail_url ? 'var(--surface-soft)' : (c.color || '#f97316') }">
          <img v-if="c.thumbnail_url" :src="c.thumbnail_url" alt="Cover" class="thumb-img" />
          <div class="thumb-badges">
            <span class="badge" :class="c.is_public ? 'badge-success' : 'badge-warning'">
              {{ c.is_public ? 'Público' : 'Privado' }}
            </span>
            <span class="badge badge-gray" style="background:rgba(255,255,255,0.9);color:#000">{{ c.type }}</span>
          </div>
          <div class="course-actions-overlay">
            <button class="icon-btn" @click="openEdit(c)" title="Editar Curso">
              <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
            </button>
            <button class="icon-btn text-danger" @click="deleteCourse(c.id)" title="Eliminar Curso">
              <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
            </button>
          </div>
        </div>
        <div class="course-body">
          <h3 class="course-title">{{ c.title }}</h3>
          <p class="course-desc">{{ c.description || 'Sin descripción' }}</p>
          
          <div class="course-footer">
            <div class="course-code" v-if="c.codigo_acceso">
              <span class="code-label">Código:</span>
              <strong @click="copyCode(c.codigo_acceso)" title="Copiar código">{{ c.codigo_acceso }}</strong>
              <button class="btn-text small" @click="resetCode(c.id)">Reset</button>
            </div>
            
            <button class="toggle-btn" :class="{ on: c.is_public }" @click="togglePublic(c)" title="Cambiar visibilidad">
              <div class="toggle-track"></div>
              <div class="toggle-thumb"></div>
            </button>
          </div>
        </div>
      </div>
    </div>

    <CourseWizardModal 
      v-if="showWizard" 
      @close="showWizard = false" 
      @created="fetchCourses(); showWizard = false" 
    />
    
    <CourseEditorDrawer 
      :show="showDrawer" 
      :course="selectedCourse" 
      @close="showDrawer = false" 
      @updated="fetchCourses()" 
    />
  </div>
</template>

<style scoped>
.page-container {
  padding: 32px 48px; max-width: 1400px; margin: 0 auto; width: 100%;
}
.page-header {
  display: flex; align-items: flex-end; justify-content: space-between; margin-bottom: 32px;
}
.page-title { font-size: 2rem; font-weight: 800; color: var(--dark); margin: 0 0 8px 0; letter-spacing: -0.02em; }
.page-desc { color: var(--muted); margin: 0; font-size: 1.05rem; }

.loading-state, .empty-state {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 80px 0; text-align: center; color: var(--muted);
}
.empty-icon {
  width: 80px; height: 80px; background: var(--surface); border-radius: 50%;
  display: flex; align-items: center; justify-content: center; color: var(--brand);
  margin-bottom: 24px; box-shadow: var(--shadow-sm);
}
.empty-state h3 { color: var(--dark); font-size: 1.5rem; margin: 0 0 12px 0; }
.empty-state p { margin: 0 0 24px 0; max-width: 400px; }

.courses-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 24px;
}
.course-card {
  background: var(--surface); border-radius: var(--r-lg); border: 1px solid var(--border-light);
  overflow: hidden; display: flex; flex-direction: column; transition: transform 0.2s, box-shadow 0.2s;
}
.course-card:hover {
  transform: translateY(-4px); box-shadow: var(--shadow-md); border-color: var(--brand-light);
}

.course-thumb {
  position: relative; height: 180px; width: 100%; display: flex;
}
.thumb-img {
  width: 100%; height: 100%; object-fit: cover;
}
.thumb-badges {
  position: absolute; top: 12px; left: 12px; display: flex; gap: 8px; z-index: 2;
}
.course-actions-overlay {
  position: absolute; top: 12px; right: 12px; display: flex; gap: 8px; z-index: 2;
  opacity: 0; transition: opacity 0.2s;
}
.course-card:hover .course-actions-overlay { opacity: 1; }

.icon-btn {
  width: 36px; height: 36px; border-radius: 50%; background: rgba(255,255,255,0.9);
  color: var(--dark); border: none; cursor: pointer; display: flex; align-items: center; justify-content: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15); transition: all 0.2s;
}
.icon-btn:hover { background: #fff; transform: scale(1.1); }
.icon-btn.text-danger:hover { color: var(--danger); }

.course-body {
  padding: 20px; display: flex; flex-direction: column; flex: 1;
}
.course-title {
  font-size: 1.15rem; font-weight: 700; color: var(--dark); margin: 0 0 8px 0;
  display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.course-desc {
  font-size: 0.9rem; color: var(--muted); margin: 0 0 20px 0;
  display: -webkit-box; -webkit-line-clamp: 3; line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; flex: 1;
}
.course-footer {
  display: flex; align-items: center; justify-content: space-between;
  padding-top: 16px; border-top: 1px solid var(--border-light);
}
.course-code {
  display: flex; align-items: center; gap: 8px; font-size: 0.85rem; background: var(--surface-soft);
  padding: 6px 12px; border-radius: var(--r-full);
}
.code-label { color: var(--muted); }
.course-code strong { color: var(--brand); cursor: pointer; letter-spacing: 0.05em; }

.toggle-btn {
  position: relative; width: 40px; height: 22px; background: none; border: none; padding: 0; cursor: pointer;
}
.toggle-track {
  position: absolute; inset: 0; background: var(--border); border-radius: 11px; transition: background 0.2s;
}
.toggle-thumb {
  position: absolute; top: 2px; left: 2px; width: 18px; height: 18px; background: #fff; border-radius: 50%;
  transition: transform 0.2s; box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}
.toggle-btn.on .toggle-track { background: var(--success); }
.toggle-btn.on .toggle-thumb { transform: translateX(18px); }

.btn-text.small { font-size: 0.75rem; padding: 2px 6px; }
.spinner {
  width: 40px; height: 40px; border: 3px solid var(--border-light);
  border-top-color: var(--brand); border-radius: 50%; animation: spin 0.8s linear infinite; margin-bottom: 16px;
}
</style>
