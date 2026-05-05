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
    icon: '📚',
    bgClass: 'bg-violet',
  },
  {
    label: 'Cursos completados',
    value: stats.value.completed,
    icon: '✅',
    bgClass: 'bg-emerald',
  },
  {
    label: 'Exámenes',
    value: stats.value.exams,
    icon: '📝',
    bgClass: 'bg-sky',
  },
  {
    label: 'Progreso promedio',
    value: `${stats.value.avgProgress}%`,
    icon: '📈',
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
        <h1>¡Hola, {{ firstName }}! 👋</h1>
        <p>Continúa aprendiendo donde lo dejaste.</p>
      </div>
      <button class="btn btn-primary" @click="router.push('/usuario/capacitaciones')">
        Explorar cursos
      </button>
    </header>

    <!-- Stats -->
    <section v-if="!loading" class="dash-stats">
      <div v-for="stat in statCards" :key="stat.label" :class="['dash-stat-card', stat.bgClass]">
        <div class="dash-stat-icon">{{ stat.icon }}</div>
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
        <div class="dash-empty-icon">🎓</div>
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
          <div class="dash-action-icon" style="background:#e0e7ff;color:#4f46e5">🔍</div>
          <div class="dash-action-info">
            <strong>Explorar cursos</strong>
            <p>Descubre nuevos cursos disponibles</p>
          </div>
        </button>
        <button class="dash-action-card" @click="router.push('/usuario/capacitaciones')">
          <div class="dash-action-icon" style="background:#fef3c7;color:#d97706">🔑</div>
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
