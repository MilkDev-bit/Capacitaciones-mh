<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'
import DashboardCharts from '../../components/DashboardCharts.vue'

const auth = useAuthStore()
const router = useRouter()

const users = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const examenes = ref<any[]>([])
const asignaciones = ref<any[]>([])
const loading = ref(true)

const firstName = computed(() => auth.user?.name?.split(' ')[0] ?? 'Admin')

const stats = computed(() => ({
  users: users.value.length,
  instructors: users.value.filter(u => u.role === 'instructor').length,
  students: users.value.filter(u => u.role === 'user').length,
  courses: capacitaciones.value.length,
  exams: examenes.value.length,
  assignments: asignaciones.value.length,
}))

const recentAssignments = computed(() =>
  asignaciones.value.slice(0, 5).map(a => ({
    ...a,
    userName: users.value.find(u => u.id === a.user_id)?.name || 'Usuario',
    capTitle: capacitaciones.value.find(c => c.id === a.capacitacion_id)?.title || '',
    exTitle: examenes.value.find(e => e.id === a.examen_id)?.title || '',
  }))
)

async function loadData() {
  loading.value = true
  try {
    const [u, c, e, a] = await Promise.all([
      api.get('/admin/users', { params: { limit: 500 } }),
      api.get('/admin/capacitaciones', { params: { limit: 500 } }),
      api.get('/admin/examenes', { params: { limit: 500 } }),
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

onMounted(loadData)

const statCards = computed(() => [
  { label: 'Usuarios', value: stats.value.users, icon: 'users', accent: '#7c3aed', bg: '#ede9fe', sub: `${stats.value.instructors} instructores · ${stats.value.students} estudiantes` },
  { label: 'Capacitaciones', value: stats.value.courses, icon: 'book', accent: '#ea580c', bg: '#fff7ed', sub: 'Cursos creados' },
  { label: 'Exámenes', value: stats.value.exams, icon: 'clipboard', accent: '#0284c7', bg: '#e0f2fe', sub: 'Evaluaciones activas' },
  { label: 'Asignaciones', value: stats.value.assignments, icon: 'link', accent: '#059669', bg: '#d1fae5', sub: 'Contenido asignado' },
])
</script>

<template>
  <div class="ad-shell">
    <header class="ad-header">
      <div class="ad-welcome">
        <h1>Panel de Administración</h1>
        <p>Hola, {{ firstName }}. Aquí tienes un resumen general de la plataforma.</p>
      </div>
    </header>

    <section class="ad-stats">
      <template v-if="loading">
        <div v-for="n in 4" :key="n" class="ad-stat-card skeleton" style="height: 104px; border:none; box-shadow:none"></div>
      </template>
      <template v-else>
        <div v-for="stat in statCards" :key="stat.label" class="ad-stat-card">
          <div class="ad-stat-icon" :style="{ background: stat.bg, color: stat.accent }">
            <svg v-if="stat.icon === 'users'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
            <svg v-else-if="stat.icon === 'book'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
            <svg v-else-if="stat.icon === 'clipboard'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>
            <svg v-else-if="stat.icon === 'link'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/></svg>
          </div>
          <div class="ad-stat-info">
            <strong>{{ stat.value }}</strong>
            <span>{{ stat.label }}</span>
            <p>{{ stat.sub }}</p>
          </div>
        </div>
      </template>
    </section>

    <template v-if="loading">
      <div class="skeleton" style="height: 340px; margin-top: 28px; margin-bottom: 28px; border-radius: 12px; border:none; box-shadow:none"></div>
    </template>
    <DashboardCharts v-else style="margin-top: 28px; margin-bottom: 28px;" />

    <section class="ad-section">
      <h2 class="ad-section-title">Acciones rápidas</h2>
      <div class="ad-actions">
        <template v-if="loading">
          <div v-for="n in 3" :key="n" class="ad-action-card skeleton" style="height: 84px; border:none; box-shadow:none"></div>
        </template>
        <template v-else>
          <button class="ad-action-card" @click="router.push('/admin/capacitaciones')">
            <div class="ad-action-icon" style="background:#fff7ed;color:#ea580c"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg></div>
            <div class="ad-action-info">
              <strong>Gestionar cursos</strong>
              <p>Crear, editar y eliminar capacitaciones</p>
            </div>
            <span class="ad-action-arrow">→</span>
          </button>
          <button class="ad-action-card" @click="router.push('/admin/examenes')">
            <div class="ad-action-icon" style="background:#e0f2fe;color:#0284c7"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg></div>
            <div class="ad-action-info">
              <strong>Gestionar exámenes</strong>
              <p>Crear evaluaciones con preguntas</p>
            </div>
            <span class="ad-action-arrow">→</span>
          </button>
          <button class="ad-action-card" @click="router.push('/admin/usuarios')">
            <div class="ad-action-icon" style="background:#ede9fe;color:#7c3aed"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg></div>
            <div class="ad-action-info">
              <strong>Usuarios y asignaciones</strong>
              <p>Asignar cursos y exámenes a usuarios</p>
            </div>
            <span class="ad-action-arrow">→</span>
          </button>
        </template>
      </div>
    </section>

    <section v-if="loading || recentAssignments.length" class="ad-section">
      <div class="ad-section-head">
        <h2 class="ad-section-title">Asignaciones recientes</h2>
        <button v-if="!loading" class="ad-link-btn" @click="router.push('/admin/usuarios')">Ver todas →</button>
      </div>
      <template v-if="loading">
        <div class="ad-recent-list">
          <div v-for="n in 3" :key="n" class="ad-recent-item skeleton" style="height: 64px; border:none; box-shadow:none"></div>
        </div>
      </template>
      <template v-else>
        <div class="ad-recent-list">
          <div v-for="a in recentAssignments" :key="a.id" class="ad-recent-item">
            <div class="ad-recent-avatar">{{ a.userName?.charAt(0)?.toUpperCase() || 'U' }}</div>
            <div class="ad-recent-info">
              <strong>{{ a.userName }}</strong>
              <p>{{ a.capTitle || a.exTitle || 'Sin contenido' }} · {{ new Date(a.assigned_at).toLocaleDateString('es') }}</p>
            </div>
            <span class="badge" :class="a.capTitle ? 'badge-orange' : 'badge-blue'">{{ a.capTitle ? 'Curso' : 'Examen' }}</span>
          </div>
        </div>
      </template>
    </section>
  </div>
</template>

<style scoped>
.ad-shell { display: flex; flex-direction: column; gap: 28px; }

.ad-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.ad-welcome h1 { font-size: 1.8rem; font-weight: 800; color: var(--dark); letter-spacing: -0.02em; }
.ad-welcome p { color: var(--muted); margin-top: 4px; font-size: 0.95rem; }

.ad-stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.ad-stat-card {
  display: flex; align-items: flex-start; gap: 16px; padding: 22px;
  border-radius: var(--r-lg); background: var(--surface);
  border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
  transition: transform 0.2s, box-shadow 0.2s;
}
.ad-stat-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); }
.ad-stat-icon {
  width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center;
  font-size: 1.4rem; flex-shrink: 0;
}
.ad-stat-info { flex: 1; }
.ad-stat-info strong { display: block; font-size: 1.7rem; font-weight: 800; line-height: 1.1; color: var(--dark); }
.ad-stat-info span { font-size: 0.78rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin-top: 2px; display: block; }
.ad-stat-info p { font-size: 0.75rem; color: var(--subtle); margin-top: 4px; }

.ad-section { display: flex; flex-direction: column; gap: 16px; }
.ad-section-head { display: flex; justify-content: space-between; align-items: center; }
.ad-section-title { font-size: 1.15rem; font-weight: 700; color: var(--dark); }
.ad-link-btn { background: none; border: none; color: var(--brand); font-weight: 600; font-size: 0.88rem; cursor: pointer; }
.ad-link-btn:hover { color: var(--brand-dark); }

.ad-actions { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.ad-action-card {
  display: flex; align-items: center; gap: 16px; padding: 20px;
  background: var(--surface); border: 1px solid var(--border-light); border-radius: var(--r-lg);
  box-shadow: var(--shadow-xs); text-align: left; transition: all 0.2s; cursor: pointer;
}
.ad-action-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-sm); border-color: var(--brand-border); }
.ad-action-icon { width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 1.2rem; flex-shrink: 0; }
.ad-action-info { flex: 1; }
.ad-action-info strong { display: block; font-size: 0.95rem; font-weight: 700; color: var(--dark); }
.ad-action-info p { font-size: 0.8rem; color: var(--muted); margin-top: 2px; }
.ad-action-arrow { color: var(--muted); font-size: 1.1rem; transition: transform 0.15s; }
.ad-action-card:hover .ad-action-arrow { transform: translateX(3px); color: var(--brand); }

.ad-recent-list {
  background: var(--surface); border-radius: var(--r-lg); border: 1px solid var(--border-light);
  box-shadow: var(--shadow-sm); overflow: hidden;
}
.ad-recent-item {
  display: flex; align-items: center; gap: 14px; padding: 14px 20px;
  border-bottom: 1px solid var(--border-light); transition: background 0.12s;
}
.ad-recent-item:last-child { border-bottom: none; }
.ad-recent-item:hover { background: var(--bg); }
.ad-recent-avatar {
  width: 36px; height: 36px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, var(--brand), var(--brand-dark)); color: #fff;
  font-size: 0.82rem; font-weight: 700; display: flex; align-items: center; justify-content: center;
}
.ad-recent-info { flex: 1; min-width: 0; }
.ad-recent-info strong { font-size: 0.9rem; font-weight: 700; color: var(--dark); display: block; }
.ad-recent-info p { font-size: 0.78rem; color: var(--muted); margin-top: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

@media (max-width: 900px) { .ad-stats { grid-template-columns: 1fr 1fr; } .ad-actions { grid-template-columns: 1fr; } }
@media (max-width: 600px) { .ad-stats { grid-template-columns: 1fr; } }
</style>
