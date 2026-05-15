<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '../../api'

const examenes = ref<any[]>([])
const showForm = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')
const search = ref('')

const form = ref({
  title: '',
  description: '',
  preguntas: [] as Array<{
    texto: string; valor: number; orden: number;
    opciones: Array<{ texto: string; es_correcta: boolean }>
  }>
})

const filtered = computed(() => {
  const term = search.value.toLowerCase().trim()
  if (!term) return examenes.value
  return examenes.value.filter(e => (e.title || '').toLowerCase().includes(term))
})

const totalPuntos = computed(() =>
  form.value.preguntas.reduce((sum, p) => sum + (p.valor || 0), 0)
)

function addPregunta() {
  form.value.preguntas.push({
    texto: '', valor: 1, orden: form.value.preguntas.length,
    opciones: [
      { texto: '', es_correcta: false },
      { texto: '', es_correcta: false },
    ]
  })
}

function removePregunta(i: number) { form.value.preguntas.splice(i, 1) }
function addOpcion(pi: number) {
  const p = form.value.preguntas[pi]
  if (p) p.opciones.push({ texto: '', es_correcta: false })
}
function removeOpcion(pi: number, oi: number) {
  const p = form.value.preguntas[pi]
  if (p && p.opciones.length > 2) p.opciones.splice(oi, 1)
}
function setCorrecta(pi: number, oi: number) {
  const p = form.value.preguntas[pi]
  if (p) p.opciones.forEach((o, idx) => o.es_correcta = idx === oi)
}

async function load() {
  loading.value = true
  try {
    const res = await api.get('/admin/examenes')
    examenes.value = res.data || []
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function guardar() {
  error.value = ''; success.value = ''
  if (!form.value.title) { error.value = 'El titulo es requerido'; return }
  if (!form.value.preguntas.length) { error.value = 'Agrega al menos una pregunta'; return }
  for (const p of form.value.preguntas) {
    if (!p.texto) { error.value = 'Todas las preguntas necesitan texto'; return }
    if (p.valor <= 0) { error.value = 'El valor de cada pregunta debe ser mayor a 0'; return }
    if (!p.opciones.some(o => o.es_correcta)) { error.value = 'Cada pregunta debe tener una respuesta correcta'; return }
    if (p.opciones.some(o => !o.texto)) { error.value = 'Todas las opciones deben tener texto'; return }
  }
  loading.value = true
  try {
    await api.post('/admin/examenes', form.value)
    success.value = 'Examen creado exitosamente'
    showForm.value = false
    form.value = { title: '', description: '', preguntas: [] }
    await load()
    setTimeout(() => { success.value = '' }, 3000)
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('¿Eliminar este examen? Esta acción no se puede deshacer.')) return
  await api.delete(`/admin/examenes/${id}`)
  await load()
}
</script>

<template>
  <div class="ae-shell">
    <div class="ae-topbar">
      <div>
        <h1 class="ae-title">Exámenes</h1>
        <p class="ae-sub">{{ examenes.length }} examen{{ examenes.length !== 1 ? 'es' : '' }} creados</p>
      </div>
      <button class="btn btn-primary" @click="showForm = !showForm">
        {{ showForm ? '✕ Cancelar' : '+ Nuevo examen' }}
      </button>
    </div>

    <Transition name="slide-down">
      <div v-if="error" class="alert alert-error">{{ error }}</div>
    </Transition>
    <Transition name="slide-down">
      <div v-if="success" class="alert alert-success">{{ success }}</div>
    </Transition>

    <!-- Form -->
    <Transition name="slide-down">
      <div v-if="showForm" class="ae-form-card">
        <div class="ae-form-header">
          <span>📝</span>
          <div>
            <h2>Nuevo examen</h2>
            <p>Crea un examen con preguntas de opción múltiple</p>
          </div>
        </div>
        <div class="ae-form-body">
          <div class="ae-form-row">
            <div class="ae-field" style="flex:2">
              <label>Título *</label>
              <input class="field-input" v-model="form.title" placeholder="Ej: Examen de inducción" />
            </div>
            <div class="ae-field" style="flex:2">
              <label>Descripción</label>
              <input class="field-input" v-model="form.description" placeholder="Opcional..." />
            </div>
          </div>

          <!-- Preguntas -->
          <div class="ae-preguntas">
            <div class="ae-preguntas-header">
              <span>{{ form.preguntas.length }} pregunta{{ form.preguntas.length !== 1 ? 's' : '' }} · {{ totalPuntos }} pts</span>
              <button class="btn btn-secondary btn-sm" @click="addPregunta">+ Agregar pregunta</button>
            </div>

            <div v-for="(p, pi) in form.preguntas" :key="pi" class="ae-pregunta">
              <div class="ae-pregunta-head">
                <span class="ae-pregunta-num">P{{ pi + 1 }}</span>
                <div style="flex:1;display:flex;gap:10px;flex-wrap:wrap;align-items:flex-start">
                  <textarea class="field-input" v-model="p.texto" rows="2" placeholder="Escribe la pregunta..." style="flex:3;min-width:200px;resize:vertical"></textarea>
                  <div style="flex:1;min-width:80px">
                    <label style="font-size:0.72rem;color:var(--muted);font-weight:700">VALOR</label>
                    <input class="field-input" v-model.number="p.valor" type="number" min="0.5" step="0.5" />
                  </div>
                </div>
                <button class="ae-del-btn" @click="removePregunta(pi)" title="Eliminar pregunta">✕</button>
              </div>
              <div class="ae-opciones">
                <p class="ae-opciones-label">Opciones — selecciona la correcta</p>
                <div v-for="(o, oi) in p.opciones" :key="oi" class="ae-opcion" :class="{ correct: o.es_correcta }">
                  <input type="radio" :name="'ok-' + pi" :checked="o.es_correcta" @change="setCorrecta(pi, oi)" style="accent-color:var(--success);width:16px;height:16px;cursor:pointer;flex-shrink:0" />
                  <input class="field-input" v-model="o.texto" :placeholder="`Opción ${oi + 1}`" style="flex:1" />
                  <button class="ae-del-btn sm" @click="removeOpcion(pi, oi)" :disabled="p.opciones.length <= 2">✕</button>
                </div>
                <button class="ae-add-op" @click="addOpcion(pi)">+ Opción</button>
              </div>
            </div>

            <button v-if="!form.preguntas.length" class="ae-empty-add" @click="addPregunta">
              <span>📋</span>
              <strong>Agrega tu primera pregunta</strong>
              <p>Haz clic aquí para comenzar</p>
            </button>
          </div>

          <div class="ae-form-actions">
            <button class="btn btn-primary" :disabled="loading" @click="guardar">
              {{ loading ? 'Guardando...' : 'Crear examen' }}
            </button>
            <button class="btn btn-secondary" @click="showForm = false">Cancelar</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Search -->
    <div class="ae-search-bar">
      <svg width="17" height="17" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path d="M21 21l-4.35-4.35M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
      </svg>
      <input v-model="search" placeholder="Buscar exámenes..." />
    </div>

    <!-- Grid -->
    <div v-if="filtered.length" class="ae-grid">
      <div v-for="e in filtered" :key="e.id" class="ae-card">
        <div class="ae-card-cover">
          <span>📝</span>
        </div>
        <div class="ae-card-body">
          <h3>{{ e.title }}</h3>
          <p>{{ e.description || 'Sin descripción' }}</p>
          <div class="ae-card-footer">
            <span>{{ new Date(e.created_at).toLocaleDateString('es') }}</span>
            <button class="btn btn-danger btn-sm" @click="eliminar(e.id)">Eliminar</button>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="!loading" class="empty-state">
      <div class="empty-icon">📝</div>
      <h3>{{ search ? 'Sin resultados' : 'No hay exámenes' }}</h3>
      <p>{{ search ? 'Prueba con otro término.' : 'Crea el primer examen de la plataforma.' }}</p>
      <button v-if="!search" class="btn btn-primary" @click="showForm = true">Crear primer examen</button>
    </div>
  </div>
</template>

<style scoped>
.ae-shell { display: flex; flex-direction: column; gap: 20px; }
.ae-topbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.ae-title { font-size: 1.65rem; font-weight: 800; color: var(--dark); letter-spacing: -0.02em; }
.ae-sub { color: var(--muted); font-size: 0.88rem; margin-top: 3px; }

.ae-form-card {
  background: var(--surface); border-radius: var(--r-lg); overflow: hidden;
  border: 1.5px solid var(--border); box-shadow: 0 4px 20px rgba(0,0,0,.06);
}
.ae-form-header {
  display: flex; align-items: center; gap: 16px;
  padding: 22px 28px; background: var(--dark); color: #fff;
}
.ae-form-header span { font-size: 2rem; }
.ae-form-header h2 { font-size: 1.1rem; font-weight: 800; color: #fff; margin: 0; }
.ae-form-header p { font-size: 0.83rem; color: rgba(255,255,255,.65); margin: 2px 0 0; }
.ae-form-body { padding: 24px 28px; }
.ae-form-row { display: flex; gap: 14px; flex-wrap: wrap; margin-bottom: 20px; }
.ae-field { display: flex; flex-direction: column; gap: 6px; min-width: 180px; }
.ae-field label { font-size: 0.82rem; font-weight: 700; color: var(--muted); text-transform: uppercase; letter-spacing: 0.04em; }
.ae-form-actions { display: flex; gap: 10px; margin-top: 20px; }

.ae-preguntas { margin-top: 8px; }
.ae-preguntas-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px;
  font-size: 0.85rem; font-weight: 700; color: var(--muted);
}
.ae-pregunta {
  border: 1.5px solid var(--border); border-radius: var(--r); padding: 18px;
  background: var(--bg); margin-bottom: 14px;
}
.ae-pregunta-head { display: flex; gap: 12px; align-items: flex-start; }
.ae-pregunta-num {
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
  background: var(--brand); color: #fff; font-size: 0.75rem; font-weight: 800;
  display: flex; align-items: center; justify-content: center; margin-top: 6px;
}
.ae-del-btn {
  background: none; border: 1px solid var(--border); border-radius: 6px; cursor: pointer;
  color: var(--subtle); padding: 4px 8px; font-size: 0.78rem; transition: all 0.12s; flex-shrink: 0;
}
.ae-del-btn:hover { color: var(--danger); background: var(--danger-bg); border-color: var(--danger); }
.ae-del-btn.sm { padding: 2px 6px; font-size: 0.72rem; }
.ae-del-btn:disabled { opacity: 0.3; cursor: not-allowed; }

.ae-opciones { margin-top: 14px; padding-left: 44px; }
.ae-opciones-label { font-size: 0.78rem; font-weight: 700; color: var(--muted); margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.04em; }
.ae-opcion {
  display: flex; align-items: center; gap: 10px; margin-bottom: 8px;
  padding: 8px 12px; border-radius: 8px; border: 1px solid var(--border-light); background: var(--surface);
  transition: all 0.15s;
}
.ae-opcion.correct { border-color: var(--success); background: var(--success-bg); }
.ae-add-op {
  background: none; border: 1.5px dashed var(--border); color: var(--muted); padding: 6px 14px;
  border-radius: 8px; cursor: pointer; font-size: 0.82rem; font-weight: 600; margin-top: 4px;
  transition: all 0.15s;
}
.ae-add-op:hover { border-color: var(--brand); color: var(--brand); }

.ae-empty-add {
  display: flex; flex-direction: column; align-items: center; gap: 6px; padding: 36px;
  border: 2px dashed var(--border); border-radius: var(--r); cursor: pointer;
  background: none; color: var(--muted); text-align: center; width: 100%; transition: all 0.15s;
}
.ae-empty-add span { font-size: 2rem; }
.ae-empty-add strong { font-size: 0.95rem; color: var(--dark); }
.ae-empty-add p { font-size: 0.82rem; }
.ae-empty-add:hover { border-color: var(--brand); background: var(--brand-light); }

.ae-search-bar {
  display: flex; align-items: center; gap: 10px; padding: 11px 16px;
  border: 1.5px solid var(--border); border-radius: var(--r); background: var(--surface);
  color: var(--muted); box-shadow: var(--shadow-xs); max-width: 400px;
}
.ae-search-bar input { width: 100%; border: 0; outline: 0; background: transparent; color: var(--dark); font-size: 0.9rem; }
.ae-search-bar:focus-within { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(249,115,22,.12); }

.ae-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 20px; }
.ae-card {
  background: var(--surface); border-radius: var(--r-lg); overflow: hidden;
  border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
  transition: transform 0.22s cubic-bezier(0.34,1.56,0.64,1), box-shadow 0.22s;
}
.ae-card:hover { transform: translateY(-4px); box-shadow: var(--shadow-lg); }
.ae-card-cover {
  height: 100px; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #3b82f6, #6366f1); font-size: 2.2rem;
  filter: drop-shadow(0 2px 6px rgba(0,0,0,.2));
}
.ae-card-body { padding: 16px; display: flex; flex-direction: column; gap: 6px; }
.ae-card-body h3 { font-size: 0.97rem; font-weight: 700; color: var(--dark); }
.ae-card-body p { font-size: 0.82rem; color: var(--muted); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.ae-card-footer { display: flex; align-items: center; justify-content: space-between; padding-top: 10px; border-top: 1px solid var(--border-light); margin-top: 4px; }
.ae-card-footer span { font-size: 0.75rem; color: var(--subtle); }

@media (max-width: 600px) {
  .ae-topbar { flex-direction: column; align-items: stretch; }
  .ae-form-row { flex-direction: column; }
  .ae-grid { grid-template-columns: 1fr; }
  .ae-opciones { padding-left: 0; }
}
</style>
