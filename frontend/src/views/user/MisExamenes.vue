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
    '/usuario/examenes/' + id,
    'examen_' + id,
    'width=1000,height=720,scrollbars=yes,resizable=yes,toolbar=no,menubar=no'
  )
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
        @click="router.push('/usuario/examenes/' + e.id)"
        tabindex="0" @keyup.enter="router.push('/usuario/examenes/' + e.id)"
      >
        <div class="exam-thumb">
          <span class="exam-icon"><svg width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg></span>
        </div>
        <div class="exam-body">
          <span class="exam-badge">Exámen</span>
          <h3>{{ e.title }}</h3>
          <p>{{ e.description || 'Sin descripción' }}</p>
          <div class="exam-meta" v-if="e.preguntas">
            <span>📌 {{ e.preguntas.length }} preguntas</span>
          </div>
          <div class="exam-cta-row">
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
.ph { margin-bottom: 24px; }
.ph-title { font-size: 1.5rem; font-weight: 800; color: var(--dark); }
.ph-sub { color: var(--muted); font-size: 0.9rem; margin-top: 4px; }
.exams-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(268px, 1fr)); gap: 20px; }
.exam-card {
  background: var(--surface); border-radius: var(--r-lg); overflow: hidden;
  box-shadow: var(--shadow-sm); cursor: pointer; transition: transform 0.2s, box-shadow 0.2s;
  display: flex; flex-direction: column;
}
.exam-card:hover { transform: translateY(-4px); box-shadow: var(--shadow-md); }
.exam-thumb {
  height: 130px;
  background: linear-gradient(135deg, #f97316 0%, #dc2626 100%);
  display: flex; align-items: center; justify-content: center;
}
.exam-icon { font-size: 2.8rem; filter: drop-shadow(0 2px 6px rgba(0,0,0,.25)); }
.exam-body { padding: 16px; display: flex; flex-direction: column; gap: 6px; }
.exam-badge {
  font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: .06em;
  color: var(--brand-dark); background: var(--brand-light); padding: 2px 8px; border-radius: 4px;
  display: inline-block; width: fit-content;
}
.exam-body h3 { font-size: 0.97rem; font-weight: 700; color: var(--dark); }
.exam-body p { font-size: 0.83rem; color: var(--muted); }
.exam-meta { font-size: 0.8rem; color: var(--muted); }
.exam-cta-row { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-top: 4px; }
.exam-cta { font-size: 0.83rem; font-weight: 700; color: var(--brand); }
.exam-window-btn {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 0.75rem; font-weight: 600; color: var(--muted);
  background: var(--bg); border: 1px solid var(--border);
  border-radius: var(--r); padding: 4px 9px; cursor: pointer;
  transition: background .15s, color .15s, border-color .15s; white-space: nowrap;
}
.exam-window-btn:hover { background: var(--brand); color: #fff; border-color: var(--brand); }
.empty-state { text-align: center; padding: 60px 20px; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.empty-icon { font-size: 3rem; }
.empty-state h3 { font-size: 1.1rem; font-weight: 700; color: var(--dark); }
.empty-state p { color: var(--muted); max-width: 360px; font-size: 0.9rem; }
@media (max-width: 560px) { .exams-grid { grid-template-columns: 1fr; } }
</style>
