<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const loading = ref(false)
const showForm = ref(false)
const error = ref('')
const success = ref('')
const form = ref({ title: '', description: '', type: 'video', content: '', is_public: false })
const file = ref<File | null>(null)

const selectedCurso = ref<any | null>(null)
const activeTab = ref<'lecciones' | 'intermedias' | 'examen'>('lecciones')

const lecciones = ref<any[]>([])
const loadingLec = ref(false)
const showLecForm = ref(false)
const lecForm = ref({ title: '', description: '', type: 'video', content: '', orden: 1 })
const lecFile = ref<File | null>(null)

const intermedias = ref<any[]>([])
const loadingInt = ref(false)
const showIntForm = ref(false)
const intForm = ref({
  texto: '',
  tipo: 'multiple_choice',
  orden: 1,
  despues_de_leccion_id: '',
  opciones: [{ texto: '', es_correcta: false }, { texto: '', es_correcta: false }]
})

const misExamenes = ref<any[]>([])

async function load() {
  const res = await api.get('/instructor/capacitaciones')
  capacitaciones.value = res.data || []
}

onMounted(load)

function onFile(e: Event) {
  file.value = (e.target as HTMLInputElement).files?.[0] ?? null
}
function onLecFile(e: Event) {
  lecFile.value = (e.target as HTMLInputElement).files?.[0] ?? null
}

async function guardar() {
  error.value = ''; success.value = ''
  if (!form.value.title) { error.value = 'El titulo es requerido'; return }
  loading.value = true
  try {
    const fd = new FormData()
    fd.append('title', form.value.title)
    fd.append('description', form.value.description)
    fd.append('type', form.value.type)
    fd.append('content', form.value.content)
    fd.append('is_public', String(form.value.is_public))
    if (file.value) fd.append('file', file.value)
    await api.post('/instructor/capacitaciones', fd)
    success.value = 'Curso creado'
    showForm.value = false
    form.value = { title: '', description: '', type: 'video', content: '', is_public: false }
    file.value = null
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loading.value = false
  }
}

async function eliminar(id: string) {
  if (!confirm('Eliminar este curso?')) return
  await api.delete(`/instructor/capacitaciones/${id}`)
  await load()
}

async function togglePublic(id: string) {
  await api.patch(`/instructor/capacitaciones/${id}/toggle-public`)
  await load()
}

async function resetCodigo(id: string) {
  await api.post(`/instructor/capacitaciones/${id}/reset-codigo`)
  await load()
}

async function selectCurso(c: any) {
  if (selectedCurso.value?.id === c.id) { selectedCurso.value = null; return }
  selectedCurso.value = c
  activeTab.value = 'lecciones'
  await Promise.all([loadLecciones(), loadIntermedias(), loadMisExamenes()])
}

async function loadLecciones() {
  if (!selectedCurso.value) return
  loadingLec.value = true
  const res = await api.get(`/instructor/capacitaciones/${selectedCurso.value.id}/lecciones`)
  lecciones.value = res.data || []
  loadingLec.value = false
}

async function guardarLeccion() {
  if (!lecForm.value.title) return
  const fd = new FormData()
  fd.append('title', lecForm.value.title)
  fd.append('description', lecForm.value.description)
  fd.append('type', lecForm.value.type)
  fd.append('content', lecForm.value.content)
  fd.append('orden', String(lecForm.value.orden))
  if (lecFile.value) fd.append('file', lecFile.value)
  await api.post(`/instructor/capacitaciones/${selectedCurso.value.id}/lecciones`, fd)
  showLecForm.value = false
  lecForm.value = { title: '', description: '', type: 'video', content: '', orden: lecciones.value.length + 2 }
  lecFile.value = null
  await loadLecciones()
}

async function eliminarLeccion(leccionId: string) {
  if (!confirm('Eliminar esta leccion?')) return
  await api.delete(`/instructor/capacitaciones/${selectedCurso.value.id}/lecciones/${leccionId}`)
  await loadLecciones()
}

async function moverLeccion(idx: number, dir: -1 | 1) {
  const arr = [...lecciones.value]
  const target = idx + dir
  if (target < 0 || target >= arr.length) return
  ;[arr[idx], arr[target]] = [arr[target], arr[idx]]
  const reorder = arr.map((l, i) => ({ id: l.id, orden: i + 1 }))
  await api.put(`/instructor/capacitaciones/${selectedCurso.value.id}/lecciones/reorder`, reorder)
  await loadLecciones()
}

async function loadIntermedias() {
  if (!selectedCurso.value) return
  loadingInt.value = true
  const res = await api.get(`/instructor/capacitaciones/${selectedCurso.value.id}/intermedias`)
  intermedias.value = res.data || []
  loadingInt.value = false
}

function addOpcion() { intForm.value.opciones.push({ texto: '', es_correcta: false }) }
function removeOpcion(i: number) { intForm.value.opciones.splice(i, 1) }

async function guardarIntermedia() {
  if (!intForm.value.texto) return
  const payload: any = { texto: intForm.value.texto, tipo: intForm.value.tipo, orden: intForm.value.orden }
  if (intForm.value.despues_de_leccion_id) payload.despues_de_leccion_id = intForm.value.despues_de_leccion_id
  if (intForm.value.tipo !== 'open_text') payload.opciones = intForm.value.tipo === 'true_false'
    ? [{ texto: 'Verdadero', es_correcta: false }, { texto: 'Falso', es_correcta: false }]
    : intForm.value.opciones
  await api.post(`/instructor/capacitaciones/${selectedCurso.value.id}/intermedias`, payload)
  showIntForm.value = false
  intForm.value = { texto: '', tipo: 'multiple_choice', orden: 1, despues_de_leccion_id: '', opciones: [{ texto: '', es_correcta: false }, { texto: '', es_correcta: false }] }
  await loadIntermedias()
}

async function eliminarIntermedia(preguntaId: string) {
  if (!confirm('Eliminar esta pregunta?')) return
  await api.delete(`/instructor/capacitaciones/${selectedCurso.value.id}/intermedias/${preguntaId}`)
  await loadIntermedias()
}

async function loadMisExamenes() {
  const res = await api.get('/instructor/examenes')
  misExamenes.value = res.data || []
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="max-w-5xl mx-auto">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold text-gray-800">Mis Cursos</h1>
        <button @click="showForm = !showForm" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition text-sm font-medium">
          + Nuevo Curso
        </button>
      </div>

      <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">{{ error }}</div>
      <div v-if="success" class="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded-lg text-sm">{{ success }}</div>

      <div v-if="showForm" class="bg-white rounded-xl shadow-sm border border-gray-200 p-5 mb-6">
        <h2 class="font-semibold text-gray-700 mb-4">Nuevo Curso</h2>
        <div class="grid grid-cols-2 gap-4">
          <div class="col-span-2">
            <label class="block text-sm font-medium text-gray-600 mb-1">Titulo *</label>
            <input v-model="form.title" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          </div>
          <div class="col-span-2">
            <label class="block text-sm font-medium text-gray-600 mb-1">Descripcion</label>
            <textarea v-model="form.description" rows="2" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-600 mb-1">Tipo</label>
            <select v-model="form.type" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
              <option value="video">Video</option>
              <option value="document">Documento</option>
              <option value="text">Texto</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-600 mb-1">Archivo</label>
            <input type="file" @change="onFile" class="text-sm text-gray-600" />
          </div>
          <div v-if="form.type === 'text'" class="col-span-2">
            <textarea v-model="form.content" placeholder="Contenido..." rows="4" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" />
          </div>
          <div class="flex items-center gap-2">
            <input type="checkbox" v-model="form.is_public" id="pub" class="rounded" />
            <label for="pub" class="text-sm text-gray-600">Curso publico</label>
          </div>
        </div>
        <div class="flex gap-2 mt-4">
          <button @click="guardar" :disabled="loading" class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-blue-700 disabled:opacity-50">
            {{ loading ? 'Guardando...' : 'Guardar Curso' }}
          </button>
          <button @click="showForm = false" class="px-4 py-2 rounded-lg text-sm border border-gray-300 hover:bg-gray-50">Cancelar</button>
        </div>
      </div>

      <div class="space-y-4">
        <div v-for="c in capacitaciones" :key="c.id" class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div class="p-4 flex items-start justify-between">
            <div class="flex-1 min-w-0 cursor-pointer" @click="selectCurso(c)">
              <div class="flex items-center gap-2 flex-wrap">
                <h2 class="font-semibold text-gray-800 truncate">{{ c.title }}</h2>
                <span :class="c.is_public ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'" class="text-xs px-2 py-0.5 rounded-full">
                  {{ c.is_public ? 'Publico' : 'Privado' }}
                </span>
                <span class="text-xs bg-blue-50 text-blue-600 px-2 py-0.5 rounded-full capitalize">{{ c.type }}</span>
              </div>
              <p class="text-sm text-gray-500 mt-1 truncate">{{ c.description }}</p>
              <p class="text-xs text-gray-400 mt-1">Codigo: <span class="font-mono font-semibold text-blue-600">{{ c.codigo }}</span></p>
            </div>
            <div class="flex gap-1 ml-3 flex-shrink-0">
              <button @click.stop="togglePublic(c.id)" title="Cambiar visibilidad" class="p-1.5 rounded-lg text-gray-500 hover:bg-gray-100 text-xs">{{ c.is_public ? 'Ocultar' : 'Publicar' }}</button>
              <button @click.stop="resetCodigo(c.id)" title="Regenerar codigo" class="p-1.5 rounded-lg text-gray-500 hover:bg-gray-100 text-xs">Nuevo cod.</button>
              <button @click.stop="eliminar(c.id)" class="p-1.5 rounded-lg text-red-500 hover:bg-red-50 text-xs">Eliminar</button>
              <button @click="selectCurso(c)" class="p-1.5 rounded-lg text-blue-500 hover:bg-blue-50 text-xs font-bold">{{ selectedCurso?.id === c.id ? 'Cerrar' : 'Editar' }}</button>
            </div>
          </div>

          <div v-if="selectedCurso?.id === c.id" class="border-t border-gray-100 bg-gray-50">
            <div class="flex border-b border-gray-200">
              <button @click="activeTab = 'lecciones'" :class="activeTab === 'lecciones' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-500 hover:text-gray-700'" class="px-4 py-2.5 text-sm font-medium">Lecciones</button>
              <button @click="activeTab = 'intermedias'" :class="activeTab === 'intermedias' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-500 hover:text-gray-700'" class="px-4 py-2.5 text-sm font-medium">Preguntas Intermedias</button>
              <button @click="activeTab = 'examen'" :class="activeTab === 'examen' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-500 hover:text-gray-700'" class="px-4 py-2.5 text-sm font-medium">Examen</button>
            </div>

            <div v-if="activeTab === 'lecciones'" class="p-4">
              <div class="flex justify-between items-center mb-3">
                <span class="text-sm font-medium text-gray-700">{{ lecciones.length }} leccion(es)</span>
                <button @click="showLecForm = !showLecForm; lecForm.orden = lecciones.length + 1" class="text-xs bg-blue-600 text-white px-3 py-1.5 rounded-lg hover:bg-blue-700">+ Agregar</button>
              </div>
              <div v-if="showLecForm" class="bg-white rounded-lg border border-gray-200 p-4 mb-4">
                <div class="grid grid-cols-2 gap-3">
                  <div class="col-span-2">
                    <input v-model="lecForm.title" placeholder="Titulo de la leccion *" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                  </div>
                  <div class="col-span-2">
                    <input v-model="lecForm.description" placeholder="Descripcion" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" />
                  </div>
                  <div>
                    <select v-model="lecForm.type" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
                      <option value="video">Video</option>
                      <option value="document">Documento</option>
                      <option value="text">Texto</option>
                    </select>
                  </div>
                  <div>
                    <input type="number" v-model="lecForm.orden" placeholder="Orden" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" />
                  </div>
                  <div v-if="lecForm.type !== 'text'" class="col-span-2">
                    <input type="file" @change="onLecFile" class="text-sm text-gray-600" />
                  </div>
                  <div v-if="lecForm.type === 'text'" class="col-span-2">
                    <textarea v-model="lecForm.content" placeholder="Contenido..." rows="4" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" />
                  </div>
                </div>
                <div class="flex gap-2 mt-3">
                  <button @click="guardarLeccion" class="bg-blue-600 text-white px-3 py-1.5 rounded-lg text-sm hover:bg-blue-700">Guardar Leccion</button>
                  <button @click="showLecForm = false" class="px-3 py-1.5 rounded-lg text-sm border border-gray-300 hover:bg-gray-50">Cancelar</button>
                </div>
              </div>
              <div v-if="loadingLec" class="text-center text-sm text-gray-400 py-4">Cargando...</div>
              <div v-else-if="lecciones.length === 0" class="text-center text-sm text-gray-400 py-4">Sin lecciones aun.</div>
              <div v-else class="space-y-2">
                <div v-for="(lec, idx) in lecciones" :key="lec.id" class="flex items-center gap-3 bg-white rounded-lg border border-gray-200 px-3 py-2.5">
                  <span class="text-xs font-bold text-gray-400 w-5 text-center">{{ idx + 1 }}</span>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-gray-800 truncate">{{ lec.title }}</p>
                    <p class="text-xs text-gray-500 capitalize">{{ lec.type }}</p>
                  </div>
                  <div class="flex gap-1">
                    <button @click="moverLeccion(idx, -1)" :disabled="idx === 0" class="px-1.5 py-0.5 text-gray-400 hover:text-gray-600 disabled:opacity-30 text-xs">Subir</button>
                    <button @click="moverLeccion(idx, 1)" :disabled="idx === lecciones.length - 1" class="px-1.5 py-0.5 text-gray-400 hover:text-gray-600 disabled:opacity-30 text-xs">Bajar</button>
                    <button @click="eliminarLeccion(lec.id)" class="px-1.5 py-0.5 text-red-400 hover:text-red-600 text-xs">Eliminar</button>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="activeTab === 'intermedias'" class="p-4">
              <div class="flex justify-between items-center mb-3">
                <span class="text-sm font-medium text-gray-700">{{ intermedias.length }} pregunta(s)</span>
                <button @click="showIntForm = !showIntForm" class="text-xs bg-blue-600 text-white px-3 py-1.5 rounded-lg hover:bg-blue-700">+ Agregar</button>
              </div>
              <div v-if="showIntForm" class="bg-white rounded-lg border border-gray-200 p-4 mb-4">
                <div class="space-y-3">
                  <textarea v-model="intForm.texto" placeholder="Texto de la pregunta *" rows="2" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">Tipo</label>
                      <select v-model="intForm.tipo" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
                        <option value="multiple_choice">Opcion multiple</option>
                        <option value="true_false">Verdadero/Falso</option>
                        <option value="open_text">Respuesta abierta</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">Mostrar despues de</label>
                      <select v-model="intForm.despues_de_leccion_id" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm">
                        <option value="">Al inicio</option>
                        <option v-for="lec in lecciones" :key="lec.id" :value="lec.id">{{ lec.title }}</option>
                      </select>
                    </div>
                  </div>
                  <div v-if="intForm.tipo === 'multiple_choice'" class="space-y-2">
                    <div class="flex justify-between items-center">
                      <span class="text-xs font-medium text-gray-600">Opciones</span>
                      <button @click="addOpcion" class="text-xs text-blue-600 hover:underline">+ Opcion</button>
                    </div>
                    <div v-for="(op, i) in intForm.opciones" :key="i" class="flex items-center gap-2">
                      <input v-model="op.texto" :placeholder="`Opcion ${i+1}`" class="flex-1 border border-gray-300 rounded px-2 py-1 text-sm" />
                      <label class="flex items-center gap-1 text-xs text-gray-600 whitespace-nowrap">
                        <input type="radio" :name="`int-c`" :checked="op.es_correcta" @change="intForm.opciones.forEach((o,j) => o.es_correcta = j===i)" />
                        Correcta
                      </label>
                      <button v-if="intForm.opciones.length > 2" @click="removeOpcion(i)" class="text-red-400 text-xs">X</button>
                    </div>
                  </div>
                  <div v-if="intForm.tipo === 'true_false'" class="text-xs text-gray-500">Se generaran opciones Verdadero y Falso automaticamente.</div>
                </div>
                <div class="flex gap-2 mt-3">
                  <button @click="guardarIntermedia" class="bg-blue-600 text-white px-3 py-1.5 rounded-lg text-sm hover:bg-blue-700">Guardar</button>
                  <button @click="showIntForm = false" class="px-3 py-1.5 rounded-lg text-sm border border-gray-300 hover:bg-gray-50">Cancelar</button>
                </div>
              </div>
              <div v-if="loadingInt" class="text-center text-sm text-gray-400 py-4">Cargando...</div>
              <div v-else-if="intermedias.length === 0" class="text-center text-sm text-gray-400 py-4">Sin preguntas intermedias.</div>
              <div v-else class="space-y-2">
                <div v-for="preg in intermedias" :key="preg.id" class="flex items-start gap-3 bg-white rounded-lg border border-gray-200 px-3 py-2.5">
                  <div class="flex-1">
                    <p class="text-sm text-gray-800">{{ preg.texto }}</p>
                    <p class="text-xs text-gray-500 mt-0.5 capitalize">{{ preg.tipo }}</p>
                  </div>
                  <button @click="eliminarIntermedia(preg.id)" class="text-red-400 hover:text-red-600 text-xs">Eliminar</button>
                </div>
              </div>
            </div>

            <div v-if="activeTab === 'examen'" class="p-4">
              <p class="text-sm text-gray-600 mb-3">Examenes enlazados a este curso. Para enlazar, selecciona el curso al crear el examen.</p>
              <div v-for="ex in misExamenes.filter(e => e.capacitacion_id === c.id)" :key="ex.id" class="flex items-center gap-3 bg-white rounded-lg border border-green-200 px-3 py-2.5 mb-2">
                <span class="text-green-600 text-xs font-medium">Enlazado</span>
                <span class="text-sm text-gray-800">{{ ex.title }}</span>
              </div>
              <p v-if="!misExamenes.find(e => e.capacitacion_id === c.id)" class="text-sm text-gray-400">Ningun examen enlazado. Crea uno seleccionando este curso.</p>
            </div>
          </div>
        </div>

        <div v-if="capacitaciones.length === 0" class="text-center py-12 text-gray-400">
          <p class="text-lg mb-2">Sin cursos</p>
          <p class="text-sm">Crea tu primer curso con el boton de arriba</p>
        </div>
      </div>
    </div>
  </div>
</template>
