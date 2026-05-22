<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'

const examenes = ref<any[]>([])
const router = useRouter()

onMounted(async () => {
  const res = await api.get('/mis-examenes')
  examenes.value = res.data || []
})

function openInWindow(id: string) {
  window.open(
    '/examen/' + id,
    'examen_' + id,
    'width=780,height=900,scrollbars=yes,resizable=yes,toolbar=no,menubar=no,location=no'
  )
}

function scoreColor(pct: number) {
  if (pct >= 80) return '#10b981'
  if (pct >= 60) return '#f59e0b'
  return '#ef4444'
}

function handleCardClick(e: any) {
  if (e.bloqueado || e.ya_respondido) return
  router.push('/usuario/examenes/' + e.id)
}
</script>

<template>
  <div>
    <div class="ph">
      <h1 class="ph-title">Mis exámenes</h1>
      <p class="ph-sub">Completa los exámenes que te han asignado</p>
    </div>

    <div v-if="examenes.length" class="exams-grid">
      <div
        v-for="e in examenes" :key="e.id"
        class="exam-card"
        :class="{ 'exam-card--locked': e.bloqueado, 'exam-card--done': e.ya_respondido }"
        @click="handleCardClick(e)"
        :style="e.bloqueado || e.ya_respondido ? 'cursor:default' : ''"
        tabindex="0"
        @keyup.enter="handleCardClick(e)"
      >
        <!-- Thumb -->
        <div class="exam-thumb" :class="e.bloqueado ? 'thumb--locked' : e.ya_respondido ? 'thumb--done' : ''">
          <!-- Bloqueado -->
          <span v-if="e.bloqueado" class="exam-icon">
            <svg width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
          </span>
          <!-- Ya respondido -->
          <span v-else-if="e.ya_respondido" class="exam-icon">
            <svg width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          </span>
          <!-- Normal -->
          <span v-else class="exam-icon"><svg width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg></span>
          <!-- Badge de puntaje si ya respondido -->
          <div v-if="e.ya_respondido" class="score-badge" :style="{ background: scoreColor(e.porcentaje) }">
            {{ e.porcentaje.toFixed(0) }}%
          </div>
        </div>

        <div class="exam-body">
          <!-- Badge estado -->
          <span v-if="e.bloqueado" class="exam-badge badge--locked">
            <svg width="10" height="10" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
            Bloqueado
          </span>
          <span v-else-if="e.ya_respondido" class="exam-badge badge--done">Completado</span>
          <span v-else class="exam-badge">Exámen</span>

          <h3>{{ e.title }}</h3>
          <p>{{ e.description || 'Sin descripción' }}</p>

          <!-- Bloqueado: aviso -->
          <p v-if="e.bloqueado" class="exam-lock-msg">
            Completa todas las lecciones e intermedias del curso para desbloquear este exámen.
          </p>

          <!-- Completado: puntaje -->
          <div v-else-if="e.ya_respondido" class="exam-score-row">
            <div class="exam-score-bar">
              <div class="exam-score-fill" :style="{ width: e.porcentaje + '%', background: scoreColor(e.porcentaje) }"></div>
            </div>
            <span class="exam-score-label" :style="{ color: scoreColor(e.porcentaje) }">
              {{ e.porcentaje.toFixed(0) }}%
            </span>
          </div>

          <!-- Disponible: botones -->
          <div v-else class="exam-cta-row">
            <span class="exam-cta">Responder exámen →</span>
            <button
              class="exam-window-btn"
              @click.stop="openInWindow(e.id)"
              title="Abrir en ventana independiente"
            >
              <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
              Abrir en ventana
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <div class="empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg></div>
      <h3>No tienes exámenes asignados</h3>
      <p>Cuando tu instructor te asigne un exámen aparecerá aquí.</p>
    </div>
  </div>
</template>

<style scoped>
/* Page header */
.ph { margin-bottom: 28px; }
.ph-title {
  font-size: 1.75rem; font-weight: 900; color: var(--dark);
  letter-spacing: -0.03em; line-height: 1.1;
}
.ph-sub { color: var(--muted); font-size: 0.9rem; margin-top: 6px; }

/* Grid */
.exams-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 18px;
}

/* Card base */
.exam-card {
  background: var(--surface);
  border-radius: var(--r-xl);
  overflow: hidden;
  border: 1px solid var(--border-light);
  box-shadow: var(--shadow-sm);
  cursor: pointer;
  transition: transform 0.22s, box-shadow 0.22s, border-color 0.22s;
  display: flex; flex-direction: column;
}
.exam-card:hover:not(.exam-card--locked):not(.exam-card--done) {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0,0,0,0.1);
  border-color: rgba(249,115,22,0.3);
}
.exam-card--locked { opacity: 0.8; cursor: default; }
.exam-card--done   { cursor: default; }

/* Thumbnail */
.exam-thumb {
  height: 160px;
  background: linear-gradient(135deg, #f97316 0%, #dc2626 100%);
  display: flex; align-items: center; justify-content: center;
  position: relative; overflow: hidden;
}
.exam-thumb::before {
  content: '';
  position: absolute; inset: 0;
  background: radial-gradient(circle at 70% 30%, rgba(255,255,255,0.15), transparent 60%);
}
.thumb--locked { background: linear-gradient(135deg, #94a3b8 0%, #475569 100%); }
.thumb--done   { background: linear-gradient(135deg, #10b981 0%, #059669 100%); }

.score-badge {
  position: absolute; top: 12px; right: 12px;
  color: #fff; font-size: 0.82rem; font-weight: 800;
  padding: 4px 11px; border-radius: 999px;
  box-shadow: 0 3px 10px rgba(0,0,0,0.3);
  backdrop-filter: blur(4px);
}
.exam-icon {
  position: relative; z-index: 1;
  filter: drop-shadow(0 2px 8px rgba(0,0,0,0.3));
  color: rgba(255,255,255,0.95);
}

/* Body */
.exam-body { padding: 18px; display: flex; flex-direction: column; gap: 7px; flex: 1; }
.exam-body h3 { font-size: 1rem; font-weight: 800; color: var(--dark); line-height: 1.3; }
.exam-body > p { font-size: 0.82rem; color: var(--muted); line-height: 1.5; }

/* Badges */
.exam-badge {
  font-size: 0.7rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.07em;
  color: var(--brand-dark); background: var(--brand-light);
  padding: 3px 10px; border-radius: 999px;
  display: inline-flex; align-items: center; gap: 4px; width: fit-content;
}
.badge--locked { background: #f1f5f9; color: #64748b; }
.badge--done   { background: #d1fae5; color: #065f46; }

/* Lock message */
.exam-lock-msg {
  font-size: 0.78rem; color: var(--muted); font-style: italic;
  line-height: 1.5; margin-top: 4px;
}

/* Score bar */
.exam-score-row { display: flex; align-items: center; gap: 10px; margin-top: 6px; }
.exam-score-bar {
  flex: 1; height: 8px;
  background: var(--border-light);
  border-radius: 999px; overflow: hidden;
}
.exam-score-fill {
  height: 100%; border-radius: 999px;
  transition: width 0.5s cubic-bezier(.4,0,.2,1);
  box-shadow: 0 0 8px rgba(16,185,129,0.5);
}
.exam-score-label {
  font-size: 0.85rem; font-weight: 800; min-width: 40px; text-align: right;
}

/* CTA row */
.exam-cta-row {
  display: flex; align-items: center; justify-content: space-between;
  gap: 8px; margin-top: 8px;
}
.exam-cta {
  font-size: 0.88rem; font-weight: 800; color: var(--brand);
  display: flex; align-items: center; gap: 4px;
}
.exam-window-btn {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 0.75rem; font-weight: 600; color: var(--muted);
  background: var(--bg); border: 1px solid var(--border);
  border-radius: 8px; padding: 5px 10px; cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
  white-space: nowrap;
}
.exam-window-btn:hover { background: var(--brand); color: #fff; border-color: var(--brand); }

/* Empty state */
.empty-state {
  text-align: center; padding: 60px 20px;
  display: flex; flex-direction: column; align-items: center; gap: 12px;
  background: var(--surface); border-radius: var(--r-lg);
  border: 1px dashed var(--border); color: var(--muted);
}
.empty-icon { color: var(--border); }
.empty-state h3 { font-size: 1.05rem; font-weight: 700; color: var(--dark); }
.empty-state p { font-size: 0.88rem; max-width: 360px; }

@media (max-width: 560px) { .exams-grid { grid-template-columns: 1fr; } }
</style>
