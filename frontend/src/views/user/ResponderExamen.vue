<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../../api'

const route = useRoute()
const examen = ref<any>(null)
const respuestas = ref<Record<string, string>>({})
const resultado = ref<any>(null)
const loading = ref(false)
const submitted = ref(false)

onMounted(async () => {
  const res = await api.get(`/examenes/${route.params.id}`)
  examen.value = res.data
})

async function enviar() {
  const preguntas = examen.value?.preguntas || []
  const sinResponder = preguntas.filter((p: any) => !respuestas.value[p.id])
  if (sinResponder.length > 0) {
    alert(`Tienes ${sinResponder.length} pregunta(s) sin responder.`)
    return
  }
  loading.value = true
  try {
    const payload = preguntas.map((p: any) => ({
      pregunta_id: String(p.id),
      opcion_id: String(respuestas.value[p.id]),
    }))
    const res = await api.post(`/examenes/${route.params.id}/submit`, payload)
    resultado.value = res.data
    submitted.value = true
  } finally {
    loading.value = false
  }
}

function getColor(pct: number) {
  if (pct >= 80) return '#10b981'
  if (pct >= 60) return '#f59e0b'
  return '#ef4444'
}
</script>

<template>
  <div>
    <router-link to="/usuario/examenes" class="back-link">
      <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M19 12H5M12 5l-7 7 7 7"/></svg>
      Volver a exámenes
    </router-link>

    <!-- Examen activo -->
    <div v-if="!submitted">
      <div v-if="examen" class="exam-wrap">
        <!-- Header -->
        <div class="exam-header-card">
          <div class="exam-type-bar"></div>
          <div class="exam-header-body">
            <span class="exam-badge">📝 Exámen</span>
            <h2>{{ examen.title }}</h2>
            <p v-if="examen.description">{{ examen.description }}</p>
            <div class="exam-stats">
              <span>📌 {{ examen.preguntas?.length || 0 }} preguntas</span>
              <span>•</span>
              <span>{{ examen.preguntas?.reduce((s: number, p: any) => s + (p.valor || 1), 0) }} puntos en total</span>
            </div>
          </div>
        </div>

        <!-- Preguntas -->
        <div
          v-for="(p, i) in examen.preguntas" :key="p.id"
          class="question-card"
          :class="{ answered: !!respuestas[p.id] }"
        >
          <div class="q-header">
            <span class="q-num">{{ (i as number) + 1 }}</span>
            <span class="q-text">{{ p.texto }}</span>
            <span class="q-pts">{{ p.valor }} pt{{ p.valor !== 1 ? 's' : '' }}</span>
          </div>
          <div class="q-options">
            <label
              v-for="o in p.opciones" :key="o.id"
              :class="['option', { selected: respuestas[p.id] === o.id }]"
            >
              <input type="radio" :name="p.id" :value="o.id" v-model="respuestas[p.id]" hidden />
              <span class="option-radio"></span>
              <span>{{ o.texto }}</span>
            </label>
          </div>
        </div>

        <button class="btn btn-primary submit-btn" :disabled="loading" @click="enviar">
          <span v-if="loading" class="spinner" style="width:18px;height:18px"></span>
          <span>{{ loading ? 'Enviando…' : 'Enviar exámen' }}</span>
        </button>
      </div>
      <div v-else class="loading">
        <div class="spinner"></div>
        <p>Cargando exámen…</p>
      </div>
    </div>

    <!-- Resultado -->
    <div v-else class="result-card">
      <div class="result-emoji">{{ resultado.porcentaje >= 60 ? '🎉' : '📚' }}</div>
      <h2>Exámen completado</h2>
      <div class="score-ring" :style="{ '--pct': resultado.porcentaje, '--color': getColor(resultado.porcentaje) }">
        <svg viewBox="0 0 100 100" class="ring-svg">
          <circle cx="50" cy="50" r="44" fill="none" stroke="var(--border)" stroke-width="8"/>
          <circle
            cx="50" cy="50" r="44" fill="none"
            :stroke="getColor(resultado.porcentaje)" stroke-width="8"
            stroke-linecap="round"
            :stroke-dasharray="`${2 * Math.PI * 44}`"
            :stroke-dashoffset="`${2 * Math.PI * 44 * (1 - resultado.porcentaje / 100)}`"
            transform="rotate(-90 50 50)"
          />
        </svg>
        <span class="ring-pct" :style="{ color: getColor(resultado.porcentaje) }">{{ resultado.porcentaje.toFixed(0) }}%</span>
      </div>
      <p class="result-detail">
        Obtuviste <strong>{{ resultado.puntaje }}</strong> de <strong>{{ resultado.puntaje_max }}</strong> puntos
      </p>
      <p class="result-verdict" :style="{ color: getColor(resultado.porcentaje) }">
        {{ resultado.porcentaje >= 80 ? '¡Excelente trabajo!' : resultado.porcentaje >= 60 ? '¡Aprobado!' : 'No aprobado — repasa el material' }}
      </p>
      <router-link to="/usuario/examenes" class="btn btn-primary" style="display:inline-flex">Volver a mis exámenes</router-link>
    </div>
  </div>
</template>

<style scoped>
.back-link {
  display: inline-flex; align-items: center; gap: 6px;
  color: var(--muted); font-size: 0.87rem; font-weight: 600;
  margin-bottom: 24px; transition: color 0.15s;
}
.back-link:hover { color: var(--brand); }

.exam-wrap { max-width: 720px; display: flex; flex-direction: column; gap: 16px; }

.exam-header-card { background: var(--surface); border-radius: var(--r-lg); overflow: hidden; box-shadow: var(--shadow-sm); }
.exam-type-bar { height: 5px; background: linear-gradient(90deg, #f97316, #dc2626); }
.exam-header-body { padding: 20px 24px; }
.exam-badge { font-size: 0.75rem; font-weight: 700; text-transform: uppercase; letter-spacing: .06em; color: var(--brand-dark); background: var(--brand-light); padding: 3px 10px; border-radius: 4px; display: inline-block; }
.exam-header-body h2 { font-size: 1.3rem; font-weight: 800; color: var(--dark); margin: 10px 0 6px; }
.exam-header-body p { color: var(--muted); font-size: 0.9rem; }
.exam-stats { display: flex; gap: 10px; align-items: center; margin-top: 10px; font-size: 0.82rem; color: var(--muted); font-weight: 600; }

.question-card {
  background: var(--surface); border-radius: var(--r-lg); padding: 20px 24px;
  box-shadow: var(--shadow-sm); border-left: 4px solid var(--border); transition: border-color 0.2s;
}
.question-card.answered { border-left-color: var(--brand); }
.q-header { display: flex; align-items: flex-start; gap: 12px; margin-bottom: 14px; }
.q-num {
  background: var(--brand-light); color: var(--brand-dark); font-weight: 800; font-size: 0.78rem;
  padding: 3px 9px; border-radius: 20px; white-space: nowrap; flex-shrink: 0;
}
.q-text { flex: 1; font-weight: 600; color: var(--dark); line-height: 1.5; }
.q-pts { font-size: 0.78rem; color: var(--success); font-weight: 700; white-space: nowrap; flex-shrink: 0; }
.q-options { display: flex; flex-direction: column; gap: 8px; }
.option {
  display: flex; align-items: center; gap: 12px;
  padding: 11px 15px; border: 2px solid var(--border); border-radius: var(--r);
  cursor: pointer; font-size: 0.9rem; color: var(--text); transition: all 0.15s;
}
.option:hover { border-color: var(--brand-border); background: var(--brand-light); }
.option.selected { border-color: var(--brand); background: var(--brand-light); color: var(--dark); font-weight: 600; }
.option-radio {
  width: 18px; height: 18px; border-radius: 50%; border: 2px solid var(--border);
  flex-shrink: 0; transition: all 0.15s; display: flex; align-items: center; justify-content: center;
}
.option.selected .option-radio { border-color: var(--brand); background: var(--brand); }
.option.selected .option-radio::after { content: ''; width: 6px; height: 6px; background: #fff; border-radius: 50%; }

.submit-btn { width: 100%; justify-content: center; gap: 8px; padding: 14px; font-size: 1rem; margin-top: 8px; }

.loading { display: flex; flex-direction: column; align-items: center; gap: 16px; padding: 60px; color: var(--muted); }

/* Resultado */
.result-card {
  background: var(--surface); border-radius: var(--r-xl); padding: 48px 32px;
  text-align: center; box-shadow: var(--shadow-lg); max-width: 480px;
  display: flex; flex-direction: column; align-items: center; gap: 14px;
}
.result-emoji { font-size: 3rem; }
.result-card h2 { font-size: 1.5rem; font-weight: 800; color: var(--dark); }
.score-ring { position: relative; width: 130px; height: 130px; margin: 8px auto; }
.ring-svg { width: 100%; height: 100%; }
.ring-pct { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; font-size: 1.6rem; font-weight: 800; }
.result-detail { color: var(--muted); font-size: 1rem; }
.result-verdict { font-size: 1.1rem; font-weight: 700; }

@media (max-width: 600px) { .result-card { padding: 32px 20px; } }
</style>
