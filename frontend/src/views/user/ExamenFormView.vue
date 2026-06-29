<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../../api'
import iziToast from 'izitoast'

const route = useRoute()
const examen = ref<any>(null)
const respuestas = ref<Record<string, string>>({})
const resultado = ref<any>(null)
const loading = ref(false)
const submitted = ref(false)
const error = ref('')

onMounted(async () => {
  try {
    const res = await api.get(`/examenes/${route.params.id}`)
    examen.value = res.data
    if (res.data.ya_respondido) {
      resultado.value = {
        porcentaje: res.data.porcentaje ?? 0,
        puntaje: res.data.puntaje ?? 0,
        puntaje_max: res.data.puntaje_max ?? 0,
        ya_respondido: true,
      }
      submitted.value = true
    }
  } catch {
    error.value = 'No se pudo cargar el exámen. Verifica que estés autenticado.'
  }
})

async function enviar() {
  const preguntas = examen.value?.preguntas || []
  const sinResponder = preguntas.filter((p: any) => p.tipo !== 'open_text' && !respuestas.value[p.id])
  if (sinResponder.length > 0) {
    iziToast.warning({ title: 'Aviso', message: `Tienes ${sinResponder.length} pregunta(s) sin responder.` })
    return
  }
  loading.value = true
  try {
    const payload = preguntas.map((p: any) => ({
      pregunta_id: String(p.id),
      opcion_id: p.tipo !== 'open_text' ? String(respuestas.value[p.id]) : '',
      respuesta_texto: p.tipo === 'open_text' ? String(respuestas.value[p.id] || '') : '',
    }))
    const res = await api.post(`/examenes/${route.params.id}/submit`, payload)
    resultado.value = res.data
    submitted.value = true
    window.scrollTo({ top: 0, behavior: 'smooth' })
  } finally {
    loading.value = false
  }
}

function getColor(pct: number) {
  if (pct >= 80) return '#10b981'
  if (pct >= 60) return '#f59e0b'
  return '#ef4444'
}

function closeWindow() {
  window.close()
}
</script>

<template>
  <div class="gf-page">
    <!-- Barra de marca superior -->
    <div class="gf-topbar">
      <span class="gf-brand">MH Aprende</span>
    </div>

    <div class="gf-body">
      <!-- Error -->
      <div v-if="error" class="gf-error-card">
        <svg width="40" height="40" fill="none" stroke="#ef4444" stroke-width="1.5" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        <p>{{ error }}</p>
        <a href="/login" class="gf-btn-primary">Iniciar sesión</a>
      </div>

      <!-- Cargando -->
      <div v-else-if="!examen && !submitted" class="gf-loading">
        <div class="gf-spinner"></div>
        <p>Cargando exámen…</p>
      </div>

      <!-- Resultado -->
      <div v-else-if="submitted" class="gf-result-card">
        <div class="gf-result-top-bar"></div>
        <div class="gf-result-body">
          <div class="gf-result-icon">
            <svg v-if="resultado.porcentaje >= 60" width="64" height="64" fill="none" stroke="#10b981" stroke-width="1.5" viewBox="0 0 24 24"><path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
            <svg v-else width="64" height="64" fill="none" stroke="#ef4444" stroke-width="1.5" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
          </div>
          <h2 class="gf-result-title">
            {{ resultado.ya_respondido ? 'Ya completaste este exámen' : 'Tu respuesta fue registrada' }}
          </h2>
          <div class="gf-score-ring" :style="{ '--color': getColor(resultado.porcentaje) }">
            <svg viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="44" fill="none" stroke="#e5e7eb" stroke-width="8"/>
              <circle
                cx="50" cy="50" r="44" fill="none"
                :stroke="getColor(resultado.porcentaje)" stroke-width="8"
                stroke-linecap="round"
                :stroke-dasharray="`${2 * Math.PI * 44}`"
                :stroke-dashoffset="`${2 * Math.PI * 44 * (1 - resultado.porcentaje / 100)}`"
                transform="rotate(-90 50 50)"
              />
            </svg>
            <span class="gf-ring-pct" :style="{ color: getColor(resultado.porcentaje) }">
              {{ resultado.porcentaje.toFixed(0) }}%
            </span>
          </div>
          <p class="gf-score-detail">
            <strong>{{ resultado.puntaje }}</strong> de <strong>{{ resultado.puntaje_max }}</strong> puntos
          </p>
          <p class="gf-score-verdict" :style="{ color: getColor(resultado.porcentaje) }">
            {{ resultado.porcentaje >= 80 ? '¡Excelente trabajo!' : resultado.porcentaje >= 60 ? '¡Aprobado!' : 'No aprobado — repasa el material' }}
          </p>
          <button class="gf-btn-primary" @click="closeWindow">Cerrar ventana</button>
        </div>
      </div>

      <!-- Formulario -->
      <template v-else-if="examen">
        <!-- Header card -->
        <div class="gf-header-card">
          <div class="gf-header-accent"></div>
          <div class="gf-header-body">
            <h1 class="gf-exam-title">{{ examen.title }}</h1>
            <p v-if="examen.description" class="gf-exam-desc">{{ examen.description }}</p>
            <div class="gf-exam-meta">
              <span>📌 {{ examen.preguntas?.length || 0 }} pregunta{{ (examen.preguntas?.length || 0) !== 1 ? 's' : '' }}</span>
              <span class="gf-sep">·</span>
              <span>{{ examen.preguntas?.reduce((s: number, p: any) => s + (p.valor || 1), 0) }} puntos en total</span>
            </div>
          </div>
        </div>

        <!-- Preguntas -->
        <div
          v-for="(p, i) in examen.preguntas" :key="p.id"
          class="gf-question-card"
          :class="{ 'gf-answered': !!respuestas[p.id] }"
        >
          <div class="gf-q-title">
            <span class="gf-q-num">{{ (i as number) + 1 }}</span>
            <span class="gf-q-text">{{ p.texto }}</span>
            <span class="gf-q-pts">{{ p.valor }} pt</span>
          </div>

          <!-- Opciones múltiple / verdadero-falso -->
          <div v-if="p.tipo !== 'open_text'" class="gf-options">
            <label
              v-for="o in p.opciones" :key="o.id"
              class="gf-option"
              :class="{ 'gf-option--sel': respuestas[p.id] === o.id }"
            >
              <input type="radio" :name="p.id" :value="o.id" v-model="respuestas[p.id]" />
              <span class="gf-radio"></span>
              <span>{{ o.texto }}</span>
            </label>
          </div>

          <!-- Respuesta abierta -->
          <div v-else class="gf-open">
            <textarea
              v-model="respuestas[p.id]"
              rows="3"
              class="gf-textarea"
              placeholder="Escribe tu respuesta aquí…"
            ></textarea>
          </div>
        </div>

        <!-- Pie del formulario -->
        <div class="gf-footer-card">
          <button class="gf-btn-submit" :disabled="loading" @click="enviar">
            <span v-if="loading" class="gf-spinner gf-spinner--sm"></span>
            {{ loading ? 'Enviando…' : 'Enviar exámen' }}
          </button>
          <p class="gf-footer-note">Nunca envíes contraseñas a través de este formulario.</p>
        </div>
      </template>
    </div>

    <!-- Footer de marca -->
    <div class="gf-bottombar">
      <span>MH Aprende · Plataforma de capacitaciones</span>
    </div>
  </div>
</template>

<style scoped>
/* ── Página base ─────────────────────────────────────────────────────────────── */
.gf-page {
  min-height: 100vh;
  background: #f0f4f8;
  display: flex;
  flex-direction: column;
  font-family: 'Google Sans', 'Segoe UI', sans-serif;
}

/* ── Barra superior ──────────────────────────────────────────────────────────── */
.gf-topbar {
  background: #fff;
  border-bottom: 1px solid #e2e8f0;
  padding: 10px 24px;
  display: flex;
  align-items: center;
  position: sticky;
  top: 0;
  z-index: 10;
  box-shadow: 0 1px 4px rgba(0,0,0,0.07);
}
.gf-brand {
  font-size: 1.1rem;
  font-weight: 700;
  color: #f97316;
  letter-spacing: -0.02em;
}

/* ── Cuerpo ──────────────────────────────────────────────────────────────────── */
.gf-body {
  flex: 1;
  max-width: 680px;
  width: 100%;
  margin: 0 auto;
  padding: 32px 16px 48px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* ── Header card ─────────────────────────────────────────────────────────────── */
.gf-header-card {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.10), 0 4px 12px rgba(0,0,0,0.06);
  border-top: none;
}
.gf-header-accent {
  height: 8px;
  background: linear-gradient(90deg, #f97316, #dc2626);
}
.gf-header-body {
  padding: 24px 28px;
}
.gf-exam-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 8px;
  line-height: 1.3;
}
.gf-exam-desc {
  color: #64748b;
  font-size: 0.95rem;
  line-height: 1.6;
  margin: 0 0 14px;
}
.gf-exam-meta {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 0.82rem;
  color: #64748b;
  font-weight: 600;
}
.gf-sep { color: #cbd5e1; }

/* ── Pregunta card ───────────────────────────────────────────────────────────── */
.gf-question-card {
  background: #fff;
  border-radius: 10px;
  padding: 22px 28px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08), 0 2px 8px rgba(0,0,0,0.04);
  border-left: 4px solid #e2e8f0;
  transition: border-color 0.2s;
}
.gf-question-card.gf-answered {
  border-left-color: #f97316;
}
.gf-q-title {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 16px;
}
.gf-q-num {
  background: #fff7ed;
  color: #c2410c;
  font-size: 0.75rem;
  font-weight: 800;
  padding: 3px 9px;
  border-radius: 20px;
  white-space: nowrap;
  flex-shrink: 0;
  margin-top: 2px;
}
.gf-q-text {
  flex: 1;
  font-weight: 600;
  color: #1e293b;
  font-size: 0.97rem;
  line-height: 1.5;
}
.gf-q-pts {
  font-size: 0.75rem;
  color: #10b981;
  font-weight: 700;
  white-space: nowrap;
  flex-shrink: 0;
  margin-top: 3px;
}

/* ── Opciones ────────────────────────────────────────────────────────────────── */
.gf-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.gf-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.93rem;
  color: #374151;
  transition: all 0.15s;
  user-select: none;
}
.gf-option input[type="radio"] { display: none; }
.gf-option:hover {
  border-color: #f97316;
  background: #fff7ed;
}
.gf-option--sel {
  border-color: #f97316;
  background: #fff7ed;
  color: #1e293b;
  font-weight: 600;
}
.gf-radio {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid #cbd5e1;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}
.gf-option--sel .gf-radio {
  border-color: #f97316;
  background: #f97316;
}
.gf-option--sel .gf-radio::after {
  content: '';
  width: 6px;
  height: 6px;
  background: #fff;
  border-radius: 50%;
}

/* ── Respuesta abierta ───────────────────────────────────────────────────────── */
.gf-textarea {
  width: 100%;
  border: none;
  border-bottom: 2px solid #e2e8f0;
  padding: 8px 0;
  font-size: 0.95rem;
  color: #1e293b;
  background: transparent;
  resize: vertical;
  outline: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}
.gf-textarea:focus { border-bottom-color: #f97316; }

/* ── Pie de formulario ───────────────────────────────────────────────────────── */
.gf-footer-card {
  background: #fff;
  border-radius: 10px;
  padding: 20px 28px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.gf-btn-submit {
  background: #f97316;
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 12px 32px;
  font-size: 0.95rem;
  font-weight: 700;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: background 0.15s, transform 0.1s;
}
.gf-btn-submit:hover:not(:disabled) { background: #ea6c0a; transform: translateY(-1px); }
.gf-btn-submit:disabled { opacity: 0.6; cursor: not-allowed; }
.gf-footer-note { color: #94a3b8; font-size: 0.78rem; }

/* ── Resultado ───────────────────────────────────────────────────────────────── */
.gf-result-card {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0,0,0,0.10);
}
.gf-result-top-bar { height: 8px; background: linear-gradient(90deg, #f97316, #dc2626); }
.gf-result-body {
  padding: 40px 32px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}
.gf-result-title {
  font-size: 1.4rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}
.gf-score-ring {
  position: relative;
  width: 130px;
  height: 130px;
}
.gf-score-ring svg { width: 100%; height: 100%; }
.gf-ring-pct {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.7rem;
  font-weight: 800;
}
.gf-score-detail { color: #64748b; font-size: 1rem; margin: 0; }
.gf-score-verdict { font-size: 1.1rem; font-weight: 700; margin: 0; }
.gf-btn-primary {
  background: #f97316;
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 10px 28px;
  font-size: 0.93rem;
  font-weight: 700;
  cursor: pointer;
  text-decoration: none;
  transition: background 0.15s;
  display: inline-block;
  margin-top: 8px;
}
.gf-btn-primary:hover { background: #ea6c0a; }

/* ── Error ───────────────────────────────────────────────────────────────────── */
.gf-error-card {
  background: #fff;
  border-radius: 10px;
  padding: 48px 32px;
  text-align: center;
  box-shadow: 0 1px 6px rgba(0,0,0,0.08);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  color: #64748b;
}

/* ── Cargando ────────────────────────────────────────────────────────────────── */
.gf-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 80px;
  color: #64748b;
}
.gf-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e2e8f0;
  border-top-color: #f97316;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
.gf-spinner--sm { width: 16px; height: 16px; border-width: 2px; }
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Footer ──────────────────────────────────────────────────────────────────── */
.gf-bottombar {
  text-align: center;
  padding: 16px;
  color: #94a3b8;
  font-size: 0.78rem;
  border-top: 1px solid #e2e8f0;
  background: #fff;
}

@media (max-width: 600px) {
  .gf-header-body, .gf-question-card, .gf-footer-card { padding: 18px 18px; }
  .gf-footer-card { flex-direction: column; align-items: flex-start; }
}
</style>
