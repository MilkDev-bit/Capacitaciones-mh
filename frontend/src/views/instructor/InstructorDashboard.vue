<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const capacitaciones = ref<any[]>([])
const examenes = ref<any[]>([])
const estudiantes = ref<any[]>([])
const loading = ref(true)

const firstName = computed(() => auth.user?.name?.split(' ')[0] ?? 'Instructor')

async function loadData() {
  loading.value = true
  try {
    const [cRes, eRes, estRes] = await Promise.all([
      api.get('/instructor/capacitaciones').catch(() => ({ data: [] })),
      api.get('/instructor/examenes').catch(() => ({ data: [] })),
      api.get('/instructor/estudiantes').catch(() => ({ data: [] })),
    ])
    capacitaciones.value = cRes.data || []
    examenes.value = eRes.data || []
    estudiantes.value = estRes.data || []
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

const statCards = computed(() => [
  {
    label: 'Capacitaciones',
    value: capacitaciones.value.length,
    icon: 'book',
    accent: '#ea580c',
    bg: '#fff7ed',
    sub: 'Cursos creados por ti'
  },
  {
    label: 'Exámenes',
    value: examenes.value.length,
    icon: 'exam',
    accent: '#0284c7',
    bg: '#e0f2fe',
    sub: 'Evaluaciones publicadas'
  },
  {
    label: 'Estudiantes',
    value: estudiantes.value.length,
    icon: 'users',
    accent: '#7c3aed',
    bg: '#ede9fe',
    sub: 'Alumnos en tus cursos'
  },
])

const recentCourses = computed(() => capacitaciones.value.slice(0, 5))
</script>

<template>
  <div class="inst-shell">
    <header class="inst-header">
      <div class="inst-welcome">
        <h1>Panel de Instructor</h1>
        <p>Hola, {{ firstName }}. Aquí tienes el resumen de tus capacitaciones y estudiantes.</p>
      </div>
      <button class="btn btn-primary" @click="router.push('/instructor/capacitaciones')">
        <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 4v16m8-8H4"/></svg>
        Mis Capacitaciones
      </button>
    </header>

    <section class="inst-stats">
      <template v-if="loading">
        <div v-for="n in 3" :key="n" class="inst-stat-card skeleton" style="height: 104px; border:none; box-shadow:none"></div>
      </template>
      <template v-else>
        <div v-for="stat in statCards" :key="stat.label" class="inst-stat-card">
          <div class="inst-stat-icon" :style="{ background: stat.bg, color: stat.accent }">
            <svg v-if="stat.icon === 'book'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
            <svg v-else-if="stat.icon === 'exam'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>
            <svg v-else-if="stat.icon === 'users'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
          </div>
          <div class="inst-stat-info">
            <strong>{{ stat.value }}</strong>
            <span>{{ stat.label }}</span>
            <p>{{ stat.sub }}</p>
          </div>
        </div>
      </template>
    </section>

    <section class="inst-section">
      <h2 class="inst-section-title">Acciones rápidas</h2>
      <div class="inst-actions">
        <button class="inst-action-card" @click="router.push('/instructor/capacitaciones')">
          <div class="inst-action-icon" style="background:#fff7ed;color:#ea580c"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg></div>
          <div class="inst-action-info">
            <strong>Mis Capacitaciones</strong>
            <p>Administrar cursos, módulos y lecciones</p>
          </div>
          <span class="inst-action-arrow">→</span>
        </button>

        <button class="inst-action-card" @click="router.push('/instructor/examenes')">
          <div class="inst-action-icon" style="background:#e0f2fe;color:#0284c7"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg></div>
          <div class="inst-action-info">
            <strong>Mis Exámenes</strong>
            <p>Crear evaluaciones y calificar intentos</p>
          </div>
          <span class="inst-action-arrow">→</span>
        </button>

        <button class="inst-action-card" @click="router.push('/instructor/estudiantes')">
          <div class="inst-action-icon" style="background:#ede9fe;color:#7c3aed"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg></div>
          <div class="inst-action-info">
            <strong>Mis Estudiantes</strong>
            <p>Revisar el progreso y asignar capacitaciones</p>
          </div>
          <span class="inst-action-arrow">→</span>
        </button>
      </div>
    </section>

    <section class="inst-section">
      <div class="inst-section-head">
        <h2 class="inst-section-title">Capacitaciones recientes</h2>
        <button v-if="!loading && capacitaciones.length > 0" class="inst-link-btn" @click="router.push('/instructor/capacitaciones')">Ver todas →</button>
      </div>
      <template v-if="loading">
        <div class="inst-recent-list">
          <div v-for="n in 3" :key="n" class="inst-recent-item skeleton" style="height: 64px; border:none; box-shadow:none"></div>
        </div>
      </template>
      <template v-else-if="recentCourses.length === 0">
        <div class="inst-recent-empty">
          <p>No tienes capacitaciones creadas todavía.</p>
          <button class="btn btn-primary btn-sm" @click="router.push('/instructor/capacitaciones')">Ir a Capacitaciones</button>
        </div>
      </template>
      <template v-else>
        <div class="inst-recent-list">
          <div v-for="c in recentCourses" :key="c.id" class="inst-recent-item" @click="router.push('/instructor/capacitaciones')">
            <div class="inst-recent-avatar" :style="{ background: c.color || '#ea580c' }">
              {{ c.title?.charAt(0)?.toUpperCase() || 'C' }}
            </div>
            <div class="inst-recent-info">
              <strong>{{ c.title }}</strong>
              <p>{{ c.description || 'Sin descripción' }}</p>
            </div>
            <span class="badge" :class="c.is_public ? 'badge-success' : 'badge-warning'">
              {{ c.is_public ? 'Público' : 'Privado' }}
            </span>
          </div>
        </div>
      </template>
    </section>
  </div>
</template>

<style scoped>
.inst-shell { display: flex; flex-direction: column; gap: 28px; padding: 28px 40px; max-width: 1400px; margin: 0 auto; width: 100%; }

.inst-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.inst-welcome h1 { font-size: 1.8rem; font-weight: 800; color: var(--dark); letter-spacing: -0.02em; margin: 0; }
.inst-welcome p { color: var(--muted); margin: 4px 0 0 0; font-size: 0.95rem; }

.inst-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.inst-stat-card {
  display: flex; align-items: flex-start; gap: 16px; padding: 22px;
  border-radius: var(--r-lg); background: var(--surface);
  border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
  transition: transform 0.2s, box-shadow 0.2s;
}
.inst-stat-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); }
.inst-stat-icon {
  width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center;
  font-size: 1.4rem; flex-shrink: 0;
}
.inst-stat-info { flex: 1; }
.inst-stat-info strong { display: block; font-size: 1.7rem; font-weight: 800; line-height: 1.1; color: var(--dark); }
.inst-stat-info span { font-size: 0.78rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin-top: 2px; display: block; }
.inst-stat-info p { font-size: 0.75rem; color: var(--subtle); margin-top: 4px; margin-bottom: 0; }

.inst-section { display: flex; flex-direction: column; gap: 16px; }
.inst-section-head { display: flex; justify-content: space-between; align-items: center; }
.inst-section-title { font-size: 1.15rem; font-weight: 700; color: var(--dark); margin: 0; }
.inst-link-btn { background: none; border: none; color: var(--brand); font-weight: 600; font-size: 0.88rem; cursor: pointer; }
.inst-link-btn:hover { color: var(--brand-dark); }

.inst-actions { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.inst-action-card {
  display: flex; align-items: center; gap: 16px; padding: 20px;
  background: var(--surface); border: 1px solid var(--border-light); border-radius: var(--r-lg);
  box-shadow: var(--shadow-xs); text-align: left; transition: all 0.2s; cursor: pointer;
}
.inst-action-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-sm); border-color: var(--brand-border); }
.inst-action-icon { width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 1.2rem; flex-shrink: 0; }
.inst-action-info { flex: 1; }
.inst-action-info strong { display: block; font-size: 0.95rem; font-weight: 700; color: var(--dark); }
.inst-action-info p { font-size: 0.8rem; color: var(--muted); margin: 2px 0 0 0; }
.inst-action-arrow { color: var(--muted); font-size: 1.1rem; transition: transform 0.15s; }
.inst-action-card:hover .inst-action-arrow { transform: translateX(3px); color: var(--brand); }

.inst-recent-list {
  background: var(--surface); border-radius: var(--r-lg); border: 1px solid var(--border-light);
  box-shadow: var(--shadow-sm); overflow: hidden;
}
.inst-recent-item {
  display: flex; align-items: center; gap: 14px; padding: 14px 20px;
  border-bottom: 1px solid var(--border-light); transition: background 0.12s; cursor: pointer;
}
.inst-recent-item:last-child { border-bottom: none; }
.inst-recent-item:hover { background: var(--bg); }
.inst-recent-avatar {
  width: 36px; height: 36px; border-radius: 50%; flex-shrink: 0; color: #fff;
  font-size: 0.82rem; font-weight: 700; display: flex; align-items: center; justify-content: center;
}
.inst-recent-info { flex: 1; min-width: 0; }
.inst-recent-info strong { font-size: 0.9rem; font-weight: 700; color: var(--dark); display: block; }
.inst-recent-info p { font-size: 0.78rem; color: var(--muted); margin: 2px 0 0 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.inst-recent-empty {
  background: var(--surface); border-radius: var(--r-lg); border: 1px solid var(--border-light);
  padding: 32px; text-align: center; color: var(--muted); display: flex; flex-direction: column; align-items: center; gap: 12px;
}

@media (max-width: 900px) { .inst-stats { grid-template-columns: 1fr; } .inst-actions { grid-template-columns: 1fr; } .inst-shell { padding: 20px; } }
</style>
