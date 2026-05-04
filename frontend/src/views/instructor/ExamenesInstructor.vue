<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const examenes = ref<any[]>([])
const capacitaciones = ref<any[]>([])
const showForm = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')

const form = ref({
  title: '',
  description: '',
  capacitacion_id: null as string | null,
  preguntas: [] as Array<{
    texto: string
    tipo: string
    valor: number
    orden: number
    opciones: Array<{ texto: string; es_correcta: boolean }>
  }>
})

function addPregunta() {
  form.value.preguntas.push({
    texto: '', tipo: 'multiple_choice', valor: 1, orden: form.value.preguntas.length + 1,
    opciones: [{ texto: '', es_correcta: false }, { texto: '', es_correcta: false }]
  })
}

function removePregunta(i: number) { form.value.preguntas.splice(i, 1) }

function onTipoChange(pi: number) {
  const p = form.value.preguntas[pi]
  if (!p) return
  if (p.tipo === 'true_false') {
    p.opciones = [{ texto: 'Verdadero', es_correcta: false }, { texto: 'Falso', es_correcta: false }]
  } else if (p.tipo === 'open_text') {
    p.opciones = []
  } else {
    p.opciones = [{ texto: '', es_correcta: false }, { texto: '', es_correcta: false }]
  }
}

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
  const [eRes, cRes] = await Promise.all([
    api.get('/instructor/examenes'),
    api.get('/instructor/capacitaciones')
  ])
  examenes.value = eRes.data || []
  capacitaciones.value = cRes.data || []
}

onMounted(load)

async function guardar() {
  error.value = ''; success.value = ''
  if (!form.value.title) { error.value = 'El titulo es requerido'; return }
  if (!form.value.preguntas.length) { error.value = 'Agrega al menos una pregunta'; return }
  for (const p of form.value.preguntas) {
    if (!p.texto) { error.value = 'Todas las preguntas necesitan texto'; return }
    if (p.valor <= 0) { error.value = 'El valor debe ser mayor a 0'; return }
    if (p.tipo !== 'open_text' && !p.opciones.some(o => o.es_correcta)) {
      error.value = 'Cada pregunta debe tener una respuesta correcta'; return
    }
    if (p.tipo === 'multiple_choice' && p.opciones.some(o => !o.texto)) {
      error.value = 'Todas las opciones deben tener texto'; return
    }
  }
  loading.value = true
  try {
    const payload: any = {
      title: form.value.title,
      description: form.value.description,
      preguntas: form.value.preguntas.map(p => ({
        texto: p.texto,
        tipo: p.tipo,
        valor: p.valor,
        orden: p.orden,
        opciones: p.tipo !== 'open_text' ? p.opciones : []
      }))
    }
    if (form.value.capacitacion_id) payload.capacitacion_id = form.value.capacitacion_id
    await api.post('/instructor/examenes', payload)
    success.value = 'Examen creado exitosamente'
    showForm.value = false
    form.value = { title: '', description: '', capacitacion_id: null, preguntas: [] }
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('Eliminar este examen?')) return
  await api.delete(`/instructor/examenes/${id}`)
  await load()
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="max-w-4xl mx-auto">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold text-gray-800">Mis Examenes</h1>
        <button @click="showForm = !showForm" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition text-sm font-medium">
          + Nuevo Examen
        </button>
      </div>

      <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">{{ error }}</div>
      <div v-if="success" class="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded-lg text-sm">{{ success }}</div>

      <div v-if="showForm" class="bg-white rounded-xl shadow-sm border border-gray-200 p-5 mb-6">
        <h2 class="font-semibold text-gray-700 mb-4">Crear Examen</h2>
        <div class="space-y-3 mb-4">
          <input v-model="form.title" placeholder="Titulo del examen *" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <textarea v-model="form.description" placeholder="Descripcion (opcional)" rows="2" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <div>
            <label class="block text-sm font-medium text-gray-600 mb-1">Enlazar a curso (opcional)</label>
            <select v-model="form.capacitacion_id" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
              <option :value="null">Sin curso enlazado</option>
              <option v-for="cap in capacitaciones" :key="cap.id" :value="cap.id">{{ cap.title }}</option>
            </select>
          </div>
        </div>

        <div class="space-y-4 mb-4">
          <div v-for="(p, pi) in form.preguntas" :key="pi" class="border border-gray-200 rounded-lg p-4 bg-gray-50">
            <div class="flex items-center gap-2 mb-3">
              <span class="text-xs font-bold text-gray-500">Pregunta {{ pi + 1 }}</span>
              <button @click="removePregunta(pi)" class="ml-auto text-xs text-red-500 hover:text-red-700">Eliminar</button>
            </div>
            <div class="space-y-2">
              <textarea v-model="p.texto" placeholder="Texto de la pregunta *" rows="2"
                class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
              <div class="grid grid-cols-3 gap-2">
                <div>
                  <label class="block text-xs text-gray-500 mb-1">Tipo</label>
                  <select v-model="p.tipo" @change="onTipoChange(pi)" class="w-full border border-gray-300 rounded-lg px-2 py-1.5 text-sm">
                    <option value="multiple_choice">Opcion multiple</option>
                    <option value="true_false">Verdadero/Falso</option>
                    <option value="open_text">Respuesta abierta</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs text-gray-500 mb-1">Valor (pts)</label>
                  <input type="number" v-model="p.valor" min="0.1" step="0.5" class="w-full border border-gray-300 rounded-lg px-2 py-1.5 text-sm" />
                </div>
                <div>
                  <label class="block text-xs text-gray-500 mb-1">Orden</label>
                  <input type="number" v-model="p.orden" min="1" class="w-full border border-gray-300 rounded-lg px-2 py-1.5 text-sm" />
                </div>
              </div>

              <div v-if="p.tipo === 'open_text'" class="text-xs text-gray-500 italic px-1">
                Pregunta de respuesta abierta. No requiere opciones.
              </div>
              <div v-if="p.tipo === 'true_false'" class="text-xs text-gray-500 italic px-1">
                Opciones Verdadero/Falso. Marca la correcta:
                <div class="flex gap-4 mt-1">
                  <label class="flex items-center gap-1">
                    <input type="radio" :name="`tf-${pi}`" @change="setCorrecta(pi, 0)" :checked="p.opciones[0]?.es_correcta" />
                    Verdadero
                  </label>
                  <label class="flex items-center gap-1">
                    <input type="radio" :name="`tf-${pi}`" @change="setCorrecta(pi, 1)" :checked="p.opciones[1]?.es_correcta" />
                    Falso
                  </label>
                </div>
              </div>

              <div v-if="p.tipo === 'multiple_choice'" class="space-y-1.5">
                <div class="flex justify-between items-center">
                  <span class="text-xs font-medium text-gray-600">Opciones</span>
                  <button @click="addOpcion(pi)" class="text-xs text-blue-600 hover:underline">+ Opcion</button>
                </div>
                <div v-for="(op, oi) in p.opciones" :key="oi" class="flex items-center gap-2">
                  <input v-model="op.texto" :placeholder="`Opcion ${oi + 1}`" class="flex-1 border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" />
                  <label class="flex items-center gap-1 text-xs text-gray-600 whitespace-nowrap">
                    <input type="radio" :name="`correcta-${pi}`" @change="setCorrecta(pi, oi)" :checked="op.es_correcta" />
                    Correcta
                  </label>
                  <button v-if="p.opciones.length > 2" @click="removeOpcion(pi, oi)" class="text-red-400 hover:text-red-600 text-xs">X</button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex gap-2 flex-wrap">
          <button @click="addPregunta" class="border border-blue-300 text-blue-600 px-3 py-1.5 rounded-lg text-sm hover:bg-blue-50">+ Pregunta</button>
          <button @click="guardar" :disabled="loading" class="bg-blue-600 text-white px-4 py-1.5 rounded-lg text-sm hover:bg-blue-700 disabled:opacity-50">
            {{ loading ? 'Guardando...' : 'Guardar Examen' }}
          </button>
          <button @click="showForm = false" class="px-3 py-1.5 rounded-lg text-sm border border-gray-300 hover:bg-gray-50">Cancelar</button>
        </div>
      </div>

      <div class="space-y-3">
        <div v-for="ex in examenes" :key="ex.id" class="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
          <div class="flex items-start justify-between">
            <div>
              <h2 class="font-semibold text-gray-800">{{ ex.title }}</h2>
              <p class="text-sm text-gray-500 mt-0.5">{{ ex.description }}</p>
              <div v-if="ex.capacitacion_nombre" class="mt-1">
                <span class="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">Enlazado: {{ ex.capacitacion_nombre }}</span>
              </div>
            </div>
            <button @click="eliminar(ex.id)" class="text-sm text-red-500 hover:text-red-700 hover:bg-red-50 px-2 py-1 rounded">Eliminar</button>
          </div>
        </div>
        <div v-if="examenes.length === 0" class="text-center py-12 text-gray-400">
          <p>Sin examenes aun. Crea tu primer examen.</p>
        </div>
      </div>
    </div>
  </div>
</template>
