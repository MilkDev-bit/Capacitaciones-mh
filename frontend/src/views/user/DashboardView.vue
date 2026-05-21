<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const capacitaciones = ref<any[]>([])
const examenes = ref<any[]>([])
const loading = ref(true)

const firstName = computed(() => {
  return auth.user?.name?.split(' ')[0] ?? 'Estudiante'
})

const stats = computed(() => {
  const total = capacitaciones.value.length
  const completed = capacitaciones.value.filter(c => {
    const t = c.total_lecciones ?? 0
    const d = c.lecciones_completadas ?? 0
    return t > 0 && d === t
  }).length
  const avgProgress = total > 0
    ? Math.round(
        capacitaciones.value.reduce((sum, c) => {
          const t = c.total_lecciones ?? 0
          const d = c.lecciones_completadas ?? 0
          return sum + (t > 0 ? (d / t) * 100 : 0)
        }, 0) / total
      )
    : 0
  return { total, completed, exams: examenes.value.length, avgProgress }
})

const inProgress = computed(() =>
  capacitaciones.value
    .filter(c => {
      const t = c.total_lecciones ?? 0
      const d = c.lecciones_completadas ?? 0
      return t > 0 && d < t && d > 0
    })
    .slice(0, 3)
)

async function loadData() {
  loading.value = true
  try {
    const [cursosRes, exRes] = await Promise.all([
      api.get('/mis-capacitaciones'),
      api.get('/mis-examenes').catch(() => ({ data: [] })),
    ])
    const cursos = cursosRes.data || []
    const cursosConProgreso = await Promise.all(
      cursos.map(async (c: any) => {
        try {
          const lRes = await api.get(`/capacitaciones/${c.id}/lecciones`)
          const lecciones = lRes.data || []
          c.total_lecciones = lecciones.length
          c.lecciones_completadas = lecciones.filter((l: any) => l.completada).length
        } catch {
          c.total_lecciones = 0
          c.lecciones_completadas = 0
        }
        return c
      })
    )
    capacitaciones.value = cursosConProgreso
    examenes.value = exRes.data || []
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

const statCards = computed(() => [
  {
    label: 'Cursos inscritos',
    value: stats.value.total,
    icon: 'book',
    bgClass: 'bg-violet',
  },
  {
    label: 'Cursos completados',
    value: stats.value.completed,
    icon: 'check',
    bgClass: 'bg-emerald',
  },
  {
    label: 'Exámenes',
    value: stats.value.exams,
    icon: 'clipboard',
    bgClass: 'bg-sky',
  },
  {
    label: 'Progreso promedio',
    value: `${stats.value.avgProgress}%`,
    icon: 'chart',
    bgClass: 'bg-orange',
  },
])

function courseProgress(curso: any) {
  if (!curso.total_lecciones) return 0
  return Math.round((curso.lecciones_completadas / curso.total_lecciones) * 100)
}
</script>

<template>
  <div class="dash-shell">
    <header class="dash-header">
      <div class="dash-welcome">
        <h1>¡Hola, {{ firstName }}!</h1>
        <p>Continúa aprendiendo donde lo dejaste.</p>
      </div>
      <button class="btn btn-primary" @click="router.push('/usuario/capacitaciones')">
        Explorar cursos
      </button>
    </header>

    <!-- Stats -->
    <section v-if="!loading" class="dash-stats">
      <div v-for="stat in statCards" :key="stat.label" :class="['dash-stat-card', stat.bgClass]">
        <div class="dash-stat-icon">
          <svg v-if="stat.icon === 'book'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
          <svg v-else-if="stat.icon === 'check'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          <svg v-else-if="stat.icon === 'clipboard'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>
          <svg v-else-if="stat.icon === 'chart'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/></svg>
        </div>
        <div class="dash-stat-info">
          <strong>{{ stat.value }}</strong>
          <span>{{ stat.label }}</span>
        </div>
      </div>
    </section>

    <!-- En progreso -->
    <section v-if="!loading" class="dash-section">
      <div class="dash-section-head">
        <h2>Continuar aprendiendo</h2>
        <button class="btn-link" @click="router.push('/usuario/capacitaciones')">Ver todos &rarr;</button>
      </div>

      <div v-if="inProgress.length" class="courses-grid">
        <article
          v-for="c in inProgress"
          :key="c.id"
          class="course-card"
          @click="router.push('/usuario/capacitaciones/' + c.id)"
        >
          <div class="course-body">
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description }}</p>
            <div class="progress-wrap">
              <div class="progress-top">
                <span class="progress-label">{{ c.lecciones_completadas || 0 }}/{{ c.total_lecciones || 0 }} completadas</span>
                <span class="progress-pct">{{ courseProgress(c) }}%</span>
              </div>
              <div class="progress-bar-bg">
                <div class="progress-bar-fill" :style="`width:${courseProgress(c)}%`" />
              </div>
            </div>
            <div class="course-cta">Continuar aprendiendo &rarr;</div>
          </div>
        </article>
      </div>
      
      <div v-else class="dash-empty">
        <div class="dash-empty-icon"><svg width="52" height="52" fill="none" stroke="currentColor" stroke-width="1.2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg></div>
        <h3>No tienes cursos en progreso</h3>
        <p>¡Empieza alguno de tus cursos inscritos!</p>
        <button class="btn btn-primary" @click="router.push('/usuario/capacitaciones')">Ir a mis cursos</button>
      </div>
    </section>

    <!-- Acciones rápidas -->
    <section v-if="!loading" class="dash-section">
      <div class="dash-section-head">
        <h2>Acciones rápidas</h2>
      </div>
      <div class="dash-actions">
        <button class="dash-action-card" @click="router.push('/usuario/capacitaciones')">
          <div class="dash-action-icon" style="background:#e0e7ff;color:#4f46e5"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg></div>
          <div class="dash-action-info">
            <strong>Explorar cursos</strong>
            <p>Descubre nuevos cursos disponibles</p>
          </div>
        </button>
        <button class="dash-action-card" @click="router.push('/usuario/capacitaciones')">
          <div class="dash-action-icon" style="background:#fef3c7;color:#d97706"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg></div>
          <div class="dash-action-info">
            <strong>Unirse con código</strong>
            <p>Ingresa el código de tu instructor</p>
          </div>
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.dash-shell { display: flex; flex-direction: column; gap: 32px; }

.dash-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.dash-welcome h1 { font-size: 1.8rem; font-weight: 800; color: var(--dark); letter-spacing: -0.02em; }
.dash-welcome p { color: var(--muted); margin-top: 4px; font-size: 0.95rem; }

.dash-stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.dash-stat-card {
  display: flex; align-items: center; gap: 16px; padding: 20px;
  border-radius: var(--r-lg); background: var(--surface);
  border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
  transition: transform 0.2s;
}
.dash-stat-card:hover { transform: translateY(-2px); }
.dash-stat-icon {
  width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center;
  font-size: 1.4rem; flex-shrink: 0; background: rgba(0,0,0,0.04);
}
.bg-violet .dash-stat-icon { background: #ede9fe; color: #7c3aed; }
.bg-emerald .dash-stat-icon { background: #d1fae5; color: #059669; }
.bg-sky .dash-stat-icon { background: #e0f2fe; color: #0284c7; }
.bg-orange .dash-stat-icon { background: #ffedd5; color: #ea580c; }

.dash-stat-info strong { display: block; font-size: 1.5rem; font-weight: 800; line-height: 1.1; color: var(--dark); }
.dash-stat-info span { font-size: 0.75rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin-top: 4px; display: block; }

.dash-section { display: flex; flex-direction: column; gap: 16px; }
.dash-section-head { display: flex; justify-content: space-between; align-items: flex-end; }
.dash-section-head h2 { font-size: 1.2rem; font-weight: 700; color: var(--dark); }
.btn-link { background: none; border: none; color: var(--brand); font-weight: 600; cursor: pointer; transition: color 0.15s; }
.btn-link:hover { color: var(--brand-dark); }

.dash-empty {
  padding: 48px 24px; text-align: center; background: var(--surface); border-radius: var(--r-lg);
  border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
  display: flex; flex-direction: column; align-items: center; gap: 10px;
}
.dash-empty-icon { font-size: 3rem; }
.dash-empty h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
.dash-empty p { color: var(--muted); margin-bottom: 8px; }

.dash-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.dash-action-card {
  display: flex; align-items: center; gap: 16px; padding: 20px;
  background: var(--surface); border: 1px solid var(--border-light); border-radius: var(--r-lg);
  box-shadow: var(--shadow-xs); text-align: left; transition: all 0.2s; cursor: pointer;
}
.dash-action-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-sm); border-color: var(--brand-border); }
.dash-action-icon { width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 1.2rem; flex-shrink: 0; }
.dash-action-info strong { display: block; font-size: 0.95rem; font-weight: 700; color: var(--dark); }
.dash-action-info p { font-size: 0.8rem; color: var(--muted); margin-top: 2px; }

@media (max-width: 900px) {
  .dash-stats { grid-template-columns: 1fr 1fr; }
}
@media (max-width: 600px) {
  .dash-stats, .dash-actions { grid-template-columns: 1fr; }
}
</style>
