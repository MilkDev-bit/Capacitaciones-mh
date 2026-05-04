<script setup lang="ts">

import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const cursosPublicos = ref<any[]>([])
const router = useRouter()
const activeTab = ref<'mis' | 'explorar'>('mis')
const inscribiendose = ref<string | null>(null)

const codigoInput = ref('')
const codigoLoading = ref(false)
const codigoError = ref('')
const codigoSuccess = ref('')

async function loadMis() {
  const res = await api.get('/mis-capacitaciones')
  const cursos = res.data || []
  // Para cada curso, cargar el progreso de lecciones
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
}
async function loadPublicos() {
  const res = await api.get('/cursos-publicos')
  cursosPublicos.value = res.data || []
}
onMounted(() => { loadMis(); loadPublicos() })

async function inscribirse(id: string) {
  inscribiendose.value = id
  try {
    await api.post(`/inscribirse/${id}`)
    await Promise.all([loadMis(), loadPublicos()])
    activeTab.value = 'mis'
  } finally { inscribiendose.value = null }
}

async function unirseConCodigo() {
  const code = codigoInput.value.trim().toUpperCase()
  if (!code) { codigoError.value = 'Ingresa un código'; return }
  codigoError.value = ''; codigoSuccess.value = ''
  codigoLoading.value = true
  try {
    const res = await api.post('/unirse-con-codigo', { codigo: code })
    codigoSuccess.value = `¡Te uniste a "${res.data.title}"!`
    codigoInput.value = ''
    await loadMis()
    setTimeout(() => { codigoSuccess.value = ''; activeTab.value = 'mis' }, 2000)
  } catch (e: any) {
    codigoError.value = e.response?.data?.error || 'Código inválido'
  } finally { codigoLoading.value = false }
}

const thumbClass: Record<string, string> = { video: 'thumb-video', document: 'thumb-document', text: 'thumb-text' }
const typeIcon: Record<string, string> = { video: '🎥', document: '📄', text: '📝' }
const typeLabel: Record<string, string> = { video: 'Video', document: 'Documento', text: 'Texto' }
</script>

<template>
  <div>
    <!-- Page header -->
    <div class="ph">
      <div>
        <h1 class="ph-title">Mis aprendizajes</h1>
        <p class="ph-sub">Todos tus cursos y capacitaciones asignadas</p>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs-bar">
      <button :class="['tab-pill', activeTab === 'mis' ? 'active' : '']" @click="activeTab = 'mis'">
        Mis cursos
        <span class="pill-count">{{ capacitaciones.length }}</span>
      </button>
      <button :class="['tab-pill', activeTab === 'explorar' ? 'active' : '']" @click="activeTab = 'explorar'">
        Explorar
        <span class="pill-count">{{ cursosPublicos.length }}</span>
      </button>
    </div>

    <!-- Mis cursos -->
    <div v-if="activeTab === 'mis'">
      <div v-if="capacitaciones.length" class="courses-grid">
        <div
          v-for="c in capacitaciones" :key="c.id"
          class="course-card"
          @click="router.push('/usuario/capacitaciones/' + c.id)"
          tabindex="0" @keyup.enter="router.push('/usuario/capacitaciones/' + c.id)"
        >
          <div :class="['course-thumb', thumbClass[c.type] || 'thumb-default']">
            <span class="thumb-icon">{{ typeIcon[c.type] || '📚' }}</span>
          </div>
          <div class="course-body">
            <span class="course-type-badge">{{ typeLabel[c.type] || c.type }}</span>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripción' }}</p>
            <div v-if="c.total_lecciones > 0" class="progress-wrap">
              <div class="progress-top">
                <span class="progress-label">{{ c.lecciones_completadas }}/{{ c.total_lecciones }} lecciones</span>
                <span class="progress-pct">{{ Math.round((c.lecciones_completadas/c.total_lecciones)*100) }}%</span>
              </div>
              <div class="progress-bar-bg">
                <div class="progress-bar-fill" :style="`width:${Math.round((c.lecciones_completadas/c.total_lecciones)*100)}%`" />
              </div>
            </div>
            <div class="course-cta">Continuar aprendiendo →</div>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <div class="empty-icon">📚</div>
        <h3>No tienes cursos asignados aún</h3>
        <p>Explora los cursos disponibles o pide a tu instructor un código de acceso.</p>
        <button class="btn btn-primary" @click="activeTab = 'explorar'">Explorar cursos</button>
      </div>
    </div>

    <!-- Explorar -->
    <div v-if="activeTab === 'explorar'">
      <!-- Join by code -->
      <div class="code-banner">
        <div class="code-banner-left">
          <div class="code-key-icon">🔑</div>
          <div>
            <strong>¿Tienes un código de acceso?</strong>
            <p>Ingresa el código de tu instructor para unirte a un curso privado.</p>
          </div>
        </div>
        <div class="code-banner-right">
          <input
            v-model="codigoInput"
            class="code-field"
            placeholder="ABC123"
            maxlength="12"
            @keyup.enter="unirseConCodigo"
          />
          <button class="btn btn-primary" :disabled="codigoLoading" @click="unirseConCodigo">
            {{ codigoLoading ? 'Cargando...' : 'Unirme' }}
          </button>
        </div>
      </div>
      <div v-if="codigoError" class="alert alert-error" style="margin-bottom:16px">{{ codigoError }}</div>
      <div v-if="codigoSuccess" class="alert alert-success" style="margin-bottom:16px">{{ codigoSuccess }}</div>

      <p class="section-label">Cursos disponibles para todos</p>
      <div v-if="cursosPublicos.length" class="courses-grid">
        <div v-for="c in cursosPublicos" :key="c.id" class="course-card public-course">
          <div :class="['course-thumb', thumbClass[c.type] || 'thumb-default']">
            <span class="thumb-icon">{{ typeIcon[c.type] || '📚' }}</span>
            <span v-if="c.inscrito" class="enrolled-ribbon">✓ Inscrito</span>
          </div>
          <div class="course-body">
            <span class="course-type-badge">{{ typeLabel[c.type] || c.type }}</span>
            <h3 class="course-title">{{ c.title }}</h3>
            <p class="course-desc">{{ c.description || 'Sin descripción' }}</p>
            <div class="course-footer-row">
              <span v-if="c.inscrito" class="badge badge-green">✓ Ya inscrito</span>
              <button
                v-else
                class="btn btn-primary btn-sm"
                :disabled="inscribiendose === c.id"
                @click.stop="inscribirse(c.id)"
              >
                {{ inscribiendose === c.id ? 'Inscribiendo...' : '+ Inscribirse gratis' }}
              </button>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-state">
        <div class="empty-icon">🔍</div>
        <h3>No hay cursos públicos disponibles</h3>
        <p>Pide a tu instructor que comparta el enlace o código de su curso.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ph { margin-bottom: 24px; }
.ph-title { font-size: 1.5rem; font-weight: 800; color: var(--dark); }
.ph-sub { color: var(--muted); font-size: 0.9rem; margin-top: 4px; }

/* Tabs */
.tabs-bar { display: flex; gap: 8px; margin-bottom: 24px; }
.tab-pill {
  padding: 9px 20px; border: 2px solid var(--border); border-radius: 24px; background: var(--surface);
  font-size: 0.88rem; font-weight: 600; color: var(--muted); cursor: pointer;
  display: flex; align-items: center; gap: 8px; transition: all 0.18s;
}
.tab-pill:hover { border-color: var(--brand); color: var(--brand); }
.tab-pill.active { border-color: var(--brand); background: var(--brand); color: #fff; }
.tab-pill.active .pill-count { background: rgba(255,255,255,.25); color: #fff; }
.pill-count { background: var(--border-light); color: var(--muted); font-size: 0.75rem; padding: 1px 7px; border-radius: 12px; font-weight: 700; }

/* Course grid */
.courses-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(268px, 1fr)); gap: 20px; }
.course-card {
  background: var(--surface); border-radius: var(--r-lg); overflow: hidden;
  box-shadow: var(--shadow-sm); cursor: pointer; transition: transform 0.2s, box-shadow 0.2s;
  display: flex; flex-direction: column;
}
.course-card:hover { transform: translateY(-4px); box-shadow: var(--shadow-md); }

/* Thumbnail banner */
.course-thumb {
  height: 140px; display: flex; align-items: center; justify-content: center;
  position: relative; flex-shrink: 0;
}
.thumb-icon { font-size: 2.8rem; filter: drop-shadow(0 2px 6px rgba(0,0,0,.25)); }
.enrolled-ribbon {
  position: absolute; top: 10px; right: 10px;
  background: rgba(0,0,0,.55); color: #fff; font-size: 0.72rem; font-weight: 700;
  padding: 3px 10px; border-radius: 20px; backdrop-filter: blur(4px);
}

/* Course body */
.course-body { padding: 16px; display: flex; flex-direction: column; gap: 6px; flex: 1; }
.course-type-badge {
  font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: .06em;
  color: var(--brand-dark); background: var(--brand-light); padding: 2px 8px; border-radius: 4px;
  display: inline-block; width: fit-content;
}
.course-title { font-size: 0.97rem; font-weight: 700; color: var(--dark); line-height: 1.35; }
.course-desc { font-size: 0.83rem; color: var(--muted); line-height: 1.45; flex: 1; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.course-cta { font-size: 0.83rem; font-weight: 700; color: var(--brand); margin-top: 4px; }
.course-footer-row { display: flex; align-items: center; gap: 10px; margin-top: 4px; }

/* Progress bar */
.progress-wrap { margin-top: 4px; }
.progress-top { display: flex; justify-content: space-between; margin-bottom: 3px; }
.progress-label { font-size: 0.75rem; color: var(--muted); }
.progress-pct { font-size: 0.75rem; font-weight: 700; color: var(--brand); }
.progress-bar-bg { height: 5px; background: var(--border-light); border-radius: 3px; overflow: hidden; }
.progress-bar-fill { height: 100%; background: var(--brand); border-radius: 3px; transition: width 0.3s; }

/* Code banner */
.code-banner {
  display: flex; align-items: center; gap: 20px; flex-wrap: wrap;
  background: var(--surface); border-radius: var(--r-lg); padding: 20px 24px;
  box-shadow: var(--shadow-sm); border-left: 4px solid var(--brand); margin-bottom: 24px;
}
.code-banner-left { display: flex; align-items: center; gap: 14px; flex: 1; min-width: 200px; }
.code-key-icon { font-size: 1.8rem; }
.code-banner-left strong { font-size: 0.95rem; color: var(--dark); display: block; }
.code-banner-left p { font-size: 0.82rem; color: var(--muted); margin-top: 2px; }
.code-banner-right { display: flex; gap: 10px; align-items: center; }
.code-field {
  padding: 9px 14px; border: 2px solid var(--border); border-radius: var(--r);
  font-size: 1.1rem; font-weight: 800; letter-spacing: .15em; text-transform: uppercase;
  width: 130px; outline: none; font-family: 'Courier New', monospace; background: var(--bg);
}
.code-field:focus { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(249,115,22,.12); }

.section-label { font-size: 1rem; font-weight: 700; color: var(--dark); margin-bottom: 16px; }

/* Empty */
.empty-state { text-align: center; padding: 60px 20px; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.empty-icon { font-size: 3rem; }
.empty-state h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
.empty-state p { color: var(--muted); font-size: 0.9rem; max-width: 360px; }

@media (max-width: 560px) {
  .courses-grid { grid-template-columns: 1fr; }
  .code-banner { flex-direction: column; align-items: flex-start; }
  .code-banner-right { width: 100%; }
  .code-field { flex: 1; }
}
</style>
