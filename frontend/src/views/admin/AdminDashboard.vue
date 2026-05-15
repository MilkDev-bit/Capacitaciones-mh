<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'

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

onMounted(loadData)

const statCards = computed(() => [
  { label: 'Usuarios', value: stats.value.users, icon: '👥', accent: '#7c3aed', bg: '#ede9fe', sub: `${stats.value.instructors} instructores · ${stats.value.students} estudiantes` },
  { label: 'Capacitaciones', value: stats.value.courses, icon: '📚', accent: '#ea580c', bg: '#fff7ed', sub: 'Cursos creados' },
  { label: 'Exámenes', value: stats.value.exams, icon: '📝', accent: '#0284c7', bg: '#e0f2fe', sub: 'Evaluaciones activas' },
  { label: 'Asignaciones', value: stats.value.assignments, icon: '🔗', accent: '#059669', bg: '#d1fae5', sub: 'Contenido asignado' },
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

    <!-- Stats -->
    <section v-if="!loading" class="ad-stats">
      <div v-for="stat in statCards" :key="stat.label" class="ad-stat-card">
        <div class="ad-stat-icon" :style="{ background: stat.bg, color: stat.accent }">{{ stat.icon }}</div>
        <div class="ad-stat-info">
          <strong>{{ stat.value }}</strong>
          <span>{{ stat.label }}</span>
          <p>{{ stat.sub }}</p>
        </div>
      </div>
    </section>
    <section v-else class="ad-stats">
      <div v-for="n in 4" :key="n" class="ad-stat-card">
        <div class="skeleton" style="width:48px;height:48px;border-radius:12px;flex-shrink:0"></div>
        <div style="flex:1"><div class="skeleton skel-title"></div><div class="skeleton skel-text-sm" style="margin-top:6px"></div></div>
      </div>
    </section>

    <!-- Quick actions -->
    <section v-if="!loading" class="ad-section">
      <h2 class="ad-section-title">Acciones rápidas</h2>
      <div class="ad-actions">
        <button class="ad-action-card" @click="router.push('/admin/capacitaciones')">
          <div class="ad-action-icon" style="background:#fff7ed;color:#ea580c">📚</div>
          <div class="ad-action-info">
            <strong>Gestionar cursos</strong>
            <p>Crear, editar y eliminar capacitaciones</p>
          </div>
          <span class="ad-action-arrow">→</span>
        </button>
        <button class="ad-action-card" @click="router.push('/admin/examenes')">
          <div class="ad-action-icon" style="background:#e0f2fe;color:#0284c7">📝</div>
          <div class="ad-action-info">
            <strong>Gestionar exámenes</strong>
            <p>Crear evaluaciones con preguntas</p>
          </div>
          <span class="ad-action-arrow">→</span>
        </button>
        <button class="ad-action-card" @click="router.push('/admin/usuarios')">
          <div class="ad-action-icon" style="background:#ede9fe;color:#7c3aed">👥</div>
          <div class="ad-action-info">
            <strong>Usuarios y asignaciones</strong>
            <p>Asignar cursos y exámenes a usuarios</p>
          </div>
          <span class="ad-action-arrow">→</span>
        </button>
      </div>
    </section>

    <!-- Recent assignments -->
    <section v-if="!loading && recentAssignments.length" class="ad-section">
      <div class="ad-section-head">
        <h2 class="ad-section-title">Asignaciones recientes</h2>
        <button class="ad-link-btn" @click="router.push('/admin/usuarios')">Ver todas →</button>
      </div>
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
