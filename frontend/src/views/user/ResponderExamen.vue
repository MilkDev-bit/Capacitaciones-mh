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
  <div class="page">
    <router-link to="/usuario/examenes" class="back">← Volver a exámenes</router-link>

    <div v-if="!submitted">
      <div v-if="examen">
        <div class="exam-header">
          <h2>{{ examen.title }}</h2>
          <p v-if="examen.description" class="desc">{{ examen.description }}</p>
          <div class="meta">{{ examen.preguntas?.length || 0 }} preguntas</div>
        </div>

        <div
          v-for="(p, i) in examen.preguntas"
          :key="p.id"
          class="pregunta-card"
        >
          <div class="pregunta-header">
            <span class="num">{{ (i as number) + 1 }}</span>
            <span class="texto">{{ p.texto }}</span>
            <span class="pts">{{ p.valor }} pts</span>
          </div>
          <div class="opciones">
            <label
              v-for="o in p.opciones"
              :key="o.id"
              :class="['opcion', { selected: respuestas[p.id] === o.id }]"
            >
              <input type="radio" :name="p.id" :value="o.id" v-model="respuestas[p.id]" hidden />
              <span class="radio-circle"></span>
              {{ o.texto }}
            </label>
          </div>
        </div>

        <button class="btn-submit" :disabled="loading" @click="enviar">
          {{ loading ? 'Enviando...' : 'Enviar examen' }}
        </button>
      </div>
      <div v-else class="loading">Cargando examen...</div>
    </div>

    <div v-else class="resultado">
      <div class="result-icon">{{ resultado.porcentaje >= 60 ? '🎉' : '📖' }}</div>
      <h2>Examen completado</h2>
      <div class="score-circle" :style="{ borderColor: getColor(resultado.porcentaje) }">
        <span :style="{ color: getColor(resultado.porcentaje) }">
          {{ resultado.porcentaje.toFixed(1) }}%
        </span>
      </div>
      <p class="score-detail">
        Obtuviste <strong>{{ resultado.puntaje }}</strong> de <strong>{{ resultado.puntaje_max }}</strong> puntos
      </p>
      <p class="verdict" :style="{ color: getColor(resultado.porcentaje) }">
        {{ resultado.porcentaje >= 80 ? '¡Excelente trabajo!' : resultado.porcentaje >= 60 ? 'Aprobado' : 'No aprobado — repasa el material' }}
      </p>
      <router-link to="/usuario/examenes" class="btn-volver">Volver a mis exámenes</router-link>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; max-width: 780px; }
.back { color: #3b82f6; text-decoration: none; font-size: 0.9rem; font-weight: 600; display: inline-block; margin-bottom: 1.5rem; }
.back:hover { text-decoration: underline; }
.exam-header { background: white; border-radius: 12px; padding: 1.5rem; margin-bottom: 1.5rem; box-shadow: 0 2px 8px rgba(0,0,0,0.07); }
h2 { font-size: 1.4rem; font-weight: 700; color: #1e293b; margin-bottom: 4px; }
.desc { color: #64748b; font-size: 0.9rem; margin-bottom: 8px; }
.meta { font-size: 0.82rem; color: #94a3b8; }
.pregunta-card { background: white; border-radius: 12px; padding: 1.5rem; margin-bottom: 1rem; box-shadow: 0 2px 8px rgba(0,0,0,0.07); }
.pregunta-header { display: flex; align-items: flex-start; gap: 12px; margin-bottom: 1rem; }
.num { background: #eff6ff; color: #3b82f6; font-weight: 700; font-size: 0.8rem; padding: 3px 8px; border-radius: 20px; white-space: nowrap; }
.texto { flex: 1; font-weight: 600; color: #1e293b; line-height: 1.5; }
.pts { font-size: 0.78rem; color: #10b981; font-weight: 700; white-space: nowrap; }
.opciones { display: flex; flex-direction: column; gap: 8px; }
.opcion {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; border: 1.5px solid #e2e8f0; border-radius: 8px;
  cursor: pointer; transition: all 0.15s; font-size: 0.9rem;
}
.opcion:hover { border-color: #93c5fd; background: #eff6ff; }
.opcion.selected { border-color: #3b82f6; background: #eff6ff; }
.radio-circle {
  width: 16px; height: 16px; border-radius: 50%;
  border: 2px solid #cbd5e1; flex-shrink: 0; transition: all 0.15s;
}
.opcion.selected .radio-circle { border-color: #3b82f6; background: #3b82f6; }
.btn-submit {
  margin-top: 1rem; background: #3b82f6; color: white; border: none;
  padding: 13px 32px; border-radius: 10px; font-size: 1rem; font-weight: 700;
  cursor: pointer; width: 100%;
}
.btn-submit:hover:not(:disabled) { background: #2563eb; }
.btn-submit:disabled { opacity: 0.6; cursor: not-allowed; }
.loading { padding: 3rem; text-align: center; color: #94a3b8; }
.resultado { background: white; border-radius: 16px; padding: 3rem 2rem; text-align: center; box-shadow: 0 4px 20px rgba(0,0,0,0.1); margin-top: 1rem; }
.result-icon { font-size: 3rem; margin-bottom: 1rem; }
.score-circle {
  width: 120px; height: 120px; border-radius: 50%; border: 6px solid;
  display: flex; align-items: center; justify-content: center;
  margin: 1.5rem auto;
}
.score-circle span { font-size: 1.6rem; font-weight: 800; }
.score-detail { color: #475569; font-size: 1rem; margin: 0.5rem 0; }
.verdict { font-size: 1.1rem; font-weight: 700; margin: 1rem 0 1.5rem; }
.btn-volver { display: inline-block; background: #3b82f6; color: white; padding: 10px 24px; border-radius: 8px; text-decoration: none; font-weight: 600; }
.btn-volver:hover { background: #2563eb; }
</style>
