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
const userXP = ref(0)
const selectedLeaderboardCurso = ref<string>('')
const leaderboardList = ref<any[]>([])
const loadingLeaderboard = ref(false)

async function loadLeaderboard(cursoId: string) {
  if (!cursoId) {
    leaderboardList.value = []
    return
  }
  loadingLeaderboard.value = true
  try {
    const res = await api.get(`/capacitaciones/${cursoId}/leaderboard?top=5`)
    leaderboardList.value = res.data?.entries || res.data || []
  } catch {
    leaderboardList.value = []
  } finally {
    loadingLeaderboard.value = false
  }
}

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
    const [cursosRes, exRes, xpRes] = await Promise.all([
      api.get('/mis-capacitaciones'),
      api.get('/mis-examenes').catch(() => ({ data: [] })),
      api.get('/gamificacion/mis-puntos').catch(() => ({ data: { total_points: 0 } })),
    ])
    capacitaciones.value = cursosRes.data || []
    examenes.value = exRes.data || []
    userXP.value = xpRes.data?.total_points || xpRes.data?.points || 0
    if (capacitaciones.value.length > 0 && capacitaciones.value[0]?.id) {
      selectedLeaderboardCurso.value = capacitaciones.value[0].id
      loadLeaderboard(selectedLeaderboardCurso.value)
    }
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
  {
    label: 'Puntos de experiencia',
    value: `${userXP.value} XP`,
    icon: 'star',
    bgClass: 'bg-amber',
  },
])

function courseProgress(curso: any) {
  if (!curso.total_lecciones) return 0
  return Math.round((curso.lecciones_completadas / curso.total_lecciones) * 100)
}
</script>

<template>
  <div class="dash-shell">
    <div class="dash-hero">
      <div class="dash-orb dash-orb--1"></div>
      <div class="dash-orb dash-orb--2"></div>
      <header class="dash-header">
        <div class="dash-welcome">
          <h1>¡Hola, {{ firstName }}!</h1>
          <p>Continúa aprendiendo donde lo dejaste.</p>
        </div>
        <button class="btn btn-primary" @click="router.push('/usuario/capacitaciones')">
          Explorar cursos
        </button>
      </header>
    </div>

    <!-- Stats -->
    <section class="dash-stats">
      <template v-if="loading">
        <div v-for="n in 4" :key="n" class="dash-stat-card skeleton" style="height: 86px; border:none; box-shadow:none"></div>
      </template>
      <template v-else>
        <div v-for="(stat, i) in statCards" :key="stat.label" :class="['dash-stat-card', stat.bgClass]" :style="`--anim-delay: ${i * 80}ms`">
          <div class="dash-stat-icon">
            <svg v-if="stat.icon === 'book'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>
            <svg v-else-if="stat.icon === 'check'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
            <svg v-else-if="stat.icon === 'clipboard'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>
            <svg v-else-if="stat.icon === 'chart'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/></svg>
            <svg v-else-if="stat.icon === 'star'" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"/></svg>
          </div>
          <div class="dash-stat-info">
            <strong>{{ stat.value }}</strong>
            <span>{{ stat.label }}</span>
          </div>
        </div>
      </template>
    </section>

    <!-- En progreso -->
    <section class="dash-section">
      <div class="dash-section-head">
        <h2>Continuar aprendiendo</h2>
        <button v-if="!loading" class="btn-link" @click="router.push('/usuario/capacitaciones')">Ver todos &rarr;</button>
      </div>

      <template v-if="loading">
        <div class="dash-course-list">
          <div v-for="n in 3" :key="n" class="dash-course-item skeleton" style="height: 86px; border:none; box-shadow:none"></div>
        </div>
      </template>

      <template v-else>
        <div v-if="inProgress.length" class="dash-course-list">
          <button
            v-for="(c, i) in inProgress"
            :key="c.id"
            class="dash-course-item"
            :style="`--anim-delay: ${i * 70}ms`"
            @click="router.push('/usuario/capacitaciones/' + c.id)"
          >
            <div class="dash-course-band"></div>
            <div class="dash-course-info">
              <h3>{{ c.title }}</h3>
              <p>{{ c.description }}</p>
              <div class="dash-course-progress">
                <div class="dash-progress-row">
                  <span>{{ c.lecciones_completadas || 0 }}/{{ c.total_lecciones || 0 }} completadas</span>
                  <span class="dash-progress-pct">{{ courseProgress(c) }}%</span>
                </div>
                <div class="progress-bar-bg">
                  <div class="progress-bar-fill" :style="`width:${courseProgress(c)}%`" />
                </div>
              </div>
            </div>
            <span class="dash-course-cta">
              Continuar
              <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </span>
          </button>
        </div>
        
        <div v-else class="dash-empty">
          <h3>No tienes cursos en progreso</h3>
          <p>¡Empieza alguno de tus cursos inscritos!</p>
          <button class="btn btn-primary" @click="router.push('/usuario/capacitaciones')">Ir a mis cursos</button>
        </div>
      </template>
    </section>

    <!-- Tabla de Líderes por Curso -->
    <section v-if="capacitaciones.length" class="dash-section">
      <div class="dash-section-head">
        <h2>Salón de la Fama (Top 5)</h2>
        <select v-model="selectedLeaderboardCurso" @change="loadLeaderboard(selectedLeaderboardCurso)" class="lb-select">
          <option v-for="c in capacitaciones" :key="c.id" :value="c.id">{{ c.title }}</option>
        </select>
      </div>

      <div class="lb-card">
        <div v-if="loadingLeaderboard" class="lb-loading">Cargando tabla de clasificación...</div>
        <div v-else-if="!leaderboardList.length" class="lb-empty">Aún no hay puntajes registrados en este curso. ¡Juega y sé el número 1!</div>
        <div v-else class="lb-list">
          <div v-for="(entry, idx) in leaderboardList" :key="idx" :class="['lb-item', idx === 0 ? 'lb-first' : idx === 1 ? 'lb-second' : idx === 2 ? 'lb-third' : '']">
            <span class="lb-rank">
              <span v-if="idx === 0">🥇</span>
              <span v-else-if="idx === 1">🥈</span>
              <span v-else-if="idx === 2">🥉</span>
              <span v-else>#{{ idx + 1 }}</span>
            </span>
            <div class="lb-user">
              <strong>{{ entry.name || entry.user_name || entry.userName || 'Estudiante Anónimo' }}</strong>
            </div>
            <span class="lb-pts">{{ entry.points || entry.puntaje || entry.totalPoints || 0 }} XP</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Acciones rápidas -->
    <section class="dash-section">
      <div class="dash-section-head">
        <h2>Acciones rápidas</h2>
      </div>
      <template v-if="loading">
        <div class="dash-actions">
          <div v-for="n in 2" :key="n" class="dash-action-card skeleton" style="height: 82px; border:none; box-shadow:none"></div>
        </div>
      </template>
      <template v-else>
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
      </template>
    </section>
  </div>
</template>

<style scoped>
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes floatOrb {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50%       { transform: translate(10px, -14px) scale(1.05); }
}

.dash-shell { display: flex; flex-direction: column; gap: 24px; }

/* ── Hero ───────────────────────────────────────────── */
.dash-hero {
  background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 45%, #292524 100%);
  border-radius: var(--r-xl);
  padding: 28px 28px 58px;
  position: relative; overflow: hidden;
  margin-bottom: -38px;
}
.dash-orb {
  position: absolute; border-radius: 50%; pointer-events: none;
}
.dash-orb--1 {
  width: 340px; height: 340px; top: -110px; right: -90px;
  background: radial-gradient(circle, rgba(249,115,22,0.3) 0%, transparent 70%);
  animation: floatOrb 9s ease-in-out infinite;
}
.dash-orb--2 {
  width: 220px; height: 220px; bottom: -70px; left: 8%;
  background: radial-gradient(circle, rgba(124,58,237,0.22) 0%, transparent 70%);
  animation: floatOrb 11s ease-in-out infinite reverse;
}

.dash-header {
  display: flex; align-items: center; justify-content: space-between; gap: 16px;
  flex-wrap: wrap; position: relative; z-index: 1;
}
.dash-welcome h1 {
  font-size: 2rem; font-weight: 900; color: #fff; letter-spacing: -0.03em; line-height: 1.1;
}
.dash-welcome p { color: rgba(255,255,255,0.6); margin-top: 6px; font-size: 0.95rem; }

/* ── Stat cards (glass) ─────────────────────────────── */
.dash-stats { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 14px; position: relative; z-index: 2; }
.dash-stat-card {
  display: flex; align-items: center; gap: 14px; padding: 18px 20px;
  border-radius: var(--r-lg);
  background: rgba(255,255,255,0.9);
  backdrop-filter: blur(18px); -webkit-backdrop-filter: blur(18px);
  border: 1px solid rgba(255,255,255,0.8); border-top: 3px solid transparent;
  box-shadow: 0 8px 32px rgba(0,0,0,0.12), 0 1px 0 rgba(255,255,255,0.95) inset;
  transition: transform 0.22s, box-shadow 0.22s;
  animation: fadeInUp 0.4s ease both;
  animation-delay: var(--anim-delay, 0ms);
}
.dash-stat-card:hover { transform: translateY(-4px); box-shadow: 0 16px 40px rgba(0,0,0,0.14); }
.bg-violet { border-top-color: #7c3aed; }
.bg-emerald { border-top-color: #059669; }
.bg-sky     { border-top-color: #0284c7; }
.bg-orange  { border-top-color: #f97316; }
.dash-stat-icon {
  width: 46px; height: 46px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.bg-violet .dash-stat-icon { background: #ede9fe; color: #7c3aed; }
.bg-emerald .dash-stat-icon { background: #d1fae5; color: #059669; }
.bg-sky .dash-stat-icon     { background: #e0f2fe; color: #0284c7; }
.bg-orange .dash-stat-icon  { background: #ffedd5; color: #ea580c; }
.bg-amber .dash-stat-icon   { background: #fef3c7; color: #d97706; }
.bg-amber { border-top-color: #f59e0b; }
.dash-stat-info { min-width: 0; flex: 1; overflow: hidden; }
.dash-stat-info strong { display: block; font-size: 1.6rem; font-weight: 900; line-height: 1; color: var(--dark); }
.dash-stat-info span {
  font-size: 0.7rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.06em; color: var(--muted); margin-top: 5px; display: block;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}

/* ── Sections ───────────────────────────────────────── */
.dash-section { display: flex; flex-direction: column; gap: 14px; }
.dash-section-head { display: flex; justify-content: space-between; align-items: center; }
.dash-section-head h2 { font-size: 1.05rem; font-weight: 800; color: var(--dark); }
.btn-link {
  background: none; border: none; color: var(--brand);
  font-size: 0.84rem; font-weight: 700; cursor: pointer; transition: color 0.15s;
}
.btn-link:hover { color: var(--brand-dark); }

/* ── Course items ───────────────────────────────────── */
.dash-course-list { display: flex; flex-direction: column; gap: 10px; }
.dash-course-item {
  width: 100%; min-width: 0;
  display: flex; align-items: center; gap: 16px;
  background: var(--surface); border: 1px solid var(--border-light);
  border-radius: var(--r-lg); padding: 16px 20px;
  box-shadow: var(--shadow-sm); cursor: pointer; text-align: left;
  transition: transform 0.18s, box-shadow 0.18s, border-color 0.18s;
  animation: fadeInUp 0.35s ease both;
  animation-delay: var(--anim-delay, 0ms);
}
.dash-course-item:hover {
  transform: translateX(4px); box-shadow: var(--shadow-md);
  border-color: rgba(249,115,22,0.35);
}
.dash-course-band {
  width: 4px; height: 52px; border-radius: 999px; flex-shrink: 0;
  background: linear-gradient(180deg, var(--brand), var(--brand-dark));
}
.dash-course-info { flex: 1; min-width: 0; }
.dash-course-item h3 {
  font-size: 0.95rem; font-weight: 700; color: var(--dark);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 4px;
}
.dash-course-item p {
  font-size: 0.8rem; color: var(--muted); margin-bottom: 8px;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.dash-progress-row {
  display: flex; justify-content: space-between;
  font-size: 0.72rem; color: var(--muted); margin-bottom: 5px;
}
.dash-progress-pct { font-weight: 800; color: var(--brand); }
.dash-course-cta {
  font-size: 0.8rem; font-weight: 700; color: var(--brand); flex-shrink: 0;
  white-space: nowrap; display: flex; align-items: center; gap: 4px;
  padding: 7px 14px; border-radius: 999px;
  background: var(--brand-light); transition: background 0.15s;
}
.dash-course-item:hover .dash-course-cta { background: rgba(249,115,22,0.2); }

/* ── Empty state ────────────────────────────────────── */
.dash-empty {
  padding: 40px 24px; text-align: center; background: var(--surface);
  border-radius: var(--r-lg); border: 1px dashed var(--border);
  display: flex; flex-direction: column; align-items: center; gap: 10px;
}
.dash-empty h3 { font-size: 1rem; font-weight: 700; color: var(--dark); }
.dash-empty p { font-size: 0.88rem; color: var(--muted); }

/* ── Quick actions ──────────────────────────────────── */
.dash-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.dash-action-card {
  display: flex; align-items: center; gap: 16px; padding: 18px 20px;
  background: var(--surface); border: 1px solid var(--border-light);
  border-radius: var(--r-lg); box-shadow: var(--shadow-xs);
  text-align: left; transition: all 0.2s; cursor: pointer;
}
.dash-action-card:hover {
  transform: translateY(-2px); box-shadow: var(--shadow-sm);
  border-color: rgba(249,115,22,0.35);
}
.dash-action-icon {
  width: 44px; height: 44px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.dash-action-info strong { display: block; font-size: 0.92rem; font-weight: 700; color: var(--dark); }
.dash-action-info p { font-size: 0.78rem; color: var(--muted); margin-top: 3px; }

/* ── Responsive ─────────────────────────────────────── */
@media (max-width: 1100px) {
  .dash-stats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .dash-course-item { gap: 12px; }
}
@media (max-width: 640px) {
  .dash-hero { padding: 20px 20px 52px; margin-bottom: -32px; }
  .dash-welcome h1 { font-size: 1.6rem; }
  .dash-stats { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
  .dash-actions { grid-template-columns: 1fr; }
  .dash-course-item { flex-wrap: wrap; }
  .dash-course-cta { width: 100%; justify-content: center; margin-top: 4px; }
}
/* ── Leaderboard Widget ──────────────────────────────── */
.lb-select { padding: 6px 14px; border-radius: var(--r-md); border: 1px solid var(--border); background: var(--surface); color: var(--dark); font-weight: 600; font-size: 0.88rem; outline: none; }
.lb-card { background: var(--surface); border: 1px solid var(--border-light); border-radius: var(--r-lg); padding: 18px 20px; box-shadow: var(--shadow-sm); }
.lb-loading, .lb-empty { text-align: center; color: var(--muted); padding: 20px 0; font-size: 0.9rem; }
.lb-list { display: flex; flex-direction: column; gap: 8px; }
.lb-item { display: flex; align-items: center; justify-content: space-between; padding: 10px 14px; border-radius: var(--r-md); background: var(--surface-soft); border: 1px solid transparent; }
.lb-first { background: linear-gradient(90deg, #fffbeb, #fef3c7); border-color: #fde68a; }
.lb-second { background: linear-gradient(90deg, #f8fafc, #f1f5f9); border-color: #e2e8f0; }
.lb-third { background: linear-gradient(90deg, #fff7ed, #ffedd5); border-color: #fed7aa; }
.lb-rank { font-size: 1.1rem; font-weight: 800; width: 40px; display: inline-block; }
.lb-user strong { font-size: 0.94rem; color: var(--dark); }
.lb-pts { font-weight: 800; color: var(--brand); font-size: 0.9rem; }

@media (max-width: 400px) {
  .dash-stat-card { padding: 12px; gap: 8px; }
  .dash-stat-icon { width: 38px; height: 38px; flex-shrink: 0; }
  .dash-stat-info strong { font-size: 1.2rem; }
  .dash-stat-info span { font-size: 0.65rem; }
}
</style>
