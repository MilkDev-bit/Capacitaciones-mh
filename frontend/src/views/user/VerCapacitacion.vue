<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import api from '../../api'

const route = useRoute()
const cursoId = route.params.id as string

const curso = ref<any>(null)
const lecciones = ref<any[]>([])
const selectedLeccion = ref<any | null>(null)
const loading = ref(true)

// Preguntas intermedias
const showIntermedias = ref(false)
const preguntas = ref<any[]>([])
const respuestas = ref<Record<string, any>>({})
const resultadoInt = ref<any | null>(null)

// Foro
const foroPosts = ref<any[]>([])
const nuevoPost = ref({ titulo: '', contenido: '' })
const showNuevoPost = ref(false)
const expandedPost = ref<string | null>(null)
const comentariosMap = ref<Record<string, any[]>>({})
const nuevoComentario = ref<Record<string, string>>({})

const progreso = computed(() => {
  if (!lecciones.value.length) return 0
  const completadas = lecciones.value.filter(l => l.completada).length
  return Math.round((completadas / lecciones.value.length) * 100)
})

async function load() {
  loading.value = true
  const [cRes, lRes] = await Promise.all([
    api.get(`/capacitaciones/${cursoId}`),
    api.get(`/capacitaciones/${cursoId}/lecciones`)
  ])
  curso.value = cRes.data
  lecciones.value = lRes.data || []
  if (lecciones.value.length > 0) {
    await selectLeccion(lecciones.value[0])
  }
  loading.value = false
}

onMounted(load)

async function selectLeccion(lec: any) {
  selectedLeccion.value = lec
  showIntermedias.value = false
  resultadoInt.value = null
  respuestas.value = {}
  await loadForo(lec.id)
  await loadPreguntas(lec.id)
}

async function marcarCompleta() {
  if (!selectedLeccion.value || selectedLeccion.value.completada) return
  await api.post(`/lecciones/${selectedLeccion.value.id}/completar`)
  selectedLeccion.value.completada = true
  const idx = lecciones.value.findIndex(l => l.id === selectedLeccion.value.id)
  if (idx >= 0) lecciones.value[idx].completada = true
  // Mostrar preguntas intermedias si hay
  if (preguntas.value.length > 0) {
    showIntermedias.value = true
  }
}

async function loadPreguntas(leccionId: string) {
  const res = await api.get(`/capacitaciones/${cursoId}/intermedias?despues_de_leccion_id=${leccionId}`)
  preguntas.value = res.data || []
}

async function submitIntermedias() {
  const payload = preguntas.value.map(p => {
    const r: any = { pregunta_id: p.id }
    if (p.tipo === 'open_text') {
      r.respuesta_texto = respuestas.value[p.id] || ''
    } else {
      r.opcion_id = respuestas.value[p.id] || ''
    }
    return r
  })
  const res = await api.post(`/capacitaciones/${cursoId}/intermedias/submit`, payload)
  resultadoInt.value = res.data
}

// ── Foro ────────────────────────────────────────────────────────────────────
async function loadForo(leccionId: string) {
  const res = await api.get(`/lecciones/${leccionId}/foro`)
  foroPosts.value = res.data || []
}

async function crearPost() {
  if (!nuevoPost.value.titulo || !nuevoPost.value.contenido) return
  await api.post(`/lecciones/${selectedLeccion.value.id}/foro`, nuevoPost.value)
  nuevoPost.value = { titulo: '', contenido: '' }
  showNuevoPost.value = false
  await loadForo(selectedLeccion.value.id)
}

async function eliminarPost(postId: string) {
  if (!confirm('Eliminar este post?')) return
  await api.delete(`/foro/posts/${postId}`)
  await loadForo(selectedLeccion.value.id)
}

async function togglePost(postId: string) {
  if (expandedPost.value === postId) {
    expandedPost.value = null
    return
  }
  expandedPost.value = postId
  if (!comentariosMap.value[postId]) {
    const res = await api.get(`/foro/posts/${postId}/comentarios`)
    comentariosMap.value[postId] = res.data || []
  }
}

async function crearComentario(postId: string) {
  const texto = nuevoComentario.value[postId]
  if (!texto?.trim()) return
  await api.post(`/foro/posts/${postId}/comentarios`, { contenido: texto })
  nuevoComentario.value[postId] = ''
  const res = await api.get(`/foro/posts/${postId}/comentarios`)
  comentariosMap.value[postId] = res.data || []
}

function fileUrl(path: string) {
  // path ya viene con /uploads/... desde el backend
  return path ? `${import.meta.env.VITE_API_URL || ''}${path}` : ''
}

function getEmbedUrl(url: string): string {
  if (!url) return ''
  // YouTube
  const yt = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/)([^&?\s]+)/)
  if (yt) return `https://www.youtube.com/embed/${yt[1]}?rel=0`
  // Vimeo
  const vim = url.match(/vimeo\.com\/(\d+)/)
  if (vim) return `https://player.vimeo.com/video/${vim[1]}`
  // Otro (iframe generico)
  return url
}

function typeLabel(t: string) {
  const map: Record<string, string> = { video: 'Video', document: 'PDF / Documento', text: 'Lectura', link: 'Enlace / Video' }
  return map[t] || t
}

function typeIcon(t: string) {
  const map: Record<string, string> = { video: 'V', document: 'D', text: 'L', link: 'YT' }
  return map[t] || '?'
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <div v-if="loading" class="flex items-center justify-center h-screen text-gray-400">Cargando...</div>
    <div v-else class="flex h-screen overflow-hidden">

      <!-- Sidebar lecciones -->
      <aside class="w-72 bg-white border-r border-gray-200 flex flex-col flex-shrink-0 overflow-y-auto">
        <div class="p-4 border-b border-gray-100">
          <h2 class="font-bold text-gray-800 text-sm leading-tight">{{ curso?.title }}</h2>
          <div class="mt-2">
            <div class="flex justify-between text-xs text-gray-500 mb-1">
              <span>Progreso</span>
              <span>{{ progreso }}%</span>
            </div>
            <div class="h-1.5 bg-gray-200 rounded-full overflow-hidden">
              <div class="h-full bg-blue-600 rounded-full transition-all" :style="`width:${progreso}%`" />
            </div>
          </div>
        </div>
        <nav class="flex-1 p-2">
          <button v-for="(lec, idx) in lecciones" :key="lec.id"
            @click="selectLeccion(lec)"
            :class="[
              'w-full text-left px-3 py-2.5 rounded-lg mb-1 transition flex items-start gap-2',
              selectedLeccion?.id === lec.id ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-50 text-gray-700'
            ]">
            <span class="mt-0.5 flex-shrink-0">
              <span v-if="lec.completada" class="text-green-500 text-xs">&#10003;</span>
              <span v-else class="text-gray-300 text-xs">{{ idx + 1 }}</span>
            </span>
            <div class="min-w-0">
              <p class="text-sm font-medium truncate">{{ lec.title }}</p>
              <p class="text-xs text-gray-400">{{ typeLabel(lec.type) }}<span v-if="lec.duracion_min" class="ml-1">· {{ lec.duracion_min }} min</span></p>
            </div>
          </button>
          <div v-if="lecciones.length === 0" class="text-xs text-gray-400 text-center py-4">Sin lecciones</div>
        </nav>
      </aside>

      <!-- Contenido principal -->
      <main class="flex-1 overflow-y-auto">
        <div class="max-w-3xl mx-auto p-6">
          <div v-if="!selectedLeccion" class="text-center py-20 text-gray-400">
            <p>Selecciona una leccion para comenzar</p>
          </div>

          <div v-else>
            <!-- Header leccion -->
            <div class="flex items-start justify-between mb-4">
              <div>
                <h1 class="text-xl font-bold text-gray-800">{{ selectedLeccion.title }}</h1>
                <p class="text-sm text-gray-500 mt-1">{{ selectedLeccion.description }}</p>
              </div>
              <button v-if="!selectedLeccion.completada"
                @click="marcarCompleta"
                class="ml-4 flex-shrink-0 bg-green-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-green-700 transition">
                Marcar completada
              </button>
              <span v-else class="ml-4 flex-shrink-0 bg-green-100 text-green-700 px-3 py-1.5 rounded-lg text-sm font-medium">
                &#10003; Completada
              </span>
            </div>

            <!-- Reproductor / contenido -->
            <div class="bg-white rounded-xl border border-gray-200 overflow-hidden mb-6">
              <!-- Video subido -->
              <div v-if="selectedLeccion.type === 'video'" class="aspect-video bg-black">
                <video v-if="selectedLeccion.file_path" :src="fileUrl(selectedLeccion.file_path)" controls class="w-full h-full" />
                <div v-else class="flex items-center justify-center h-full text-gray-500 text-sm">Sin video disponible</div>
              </div>

              <!-- PDF / Documento embebido -->
              <div v-else-if="selectedLeccion.type === 'document'">
                <div v-if="selectedLeccion.file_path">
                  <iframe :src="fileUrl(selectedLeccion.file_path)" class="w-full border-0" style="height:75vh" />
                  <div class="px-4 py-2 border-t border-gray-100 flex justify-end">
                    <a :href="fileUrl(selectedLeccion.file_path)" target="_blank"
                      class="text-xs text-blue-600 hover:underline">Abrir en nueva pestana</a>
                  </div>
                </div>
                <p v-else class="p-6 text-gray-400 text-sm">Sin documento adjunto</p>
              </div>

              <!-- Texto / lectura -->
              <div v-else-if="selectedLeccion.type === 'text'" class="p-6">
                <div class="prose prose-sm max-w-none text-gray-700 leading-relaxed" style="white-space: pre-wrap">{{ selectedLeccion.content }}</div>
              </div>

              <!-- Enlace externo: YouTube, Vimeo, otro -->
              <div v-else-if="selectedLeccion.type === 'link'">
                <div v-if="selectedLeccion.content">
                  <div class="aspect-video">
                    <iframe :src="getEmbedUrl(selectedLeccion.content)"
                      class="w-full h-full border-0"
                      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                      allowfullscreen />
                  </div>
                  <div class="px-4 py-2 border-t border-gray-100 flex justify-end">
                    <a :href="selectedLeccion.content" target="_blank" rel="noopener"
                      class="text-xs text-blue-600 hover:underline">Abrir enlace original</a>
                  </div>
                </div>
                <p v-else class="p-6 text-gray-400 text-sm">Sin enlace configurado</p>
              </div>
            </div>

            <!-- Preguntas intermedias -->
            <div v-if="showIntermedias && preguntas.length > 0" class="bg-yellow-50 border border-yellow-200 rounded-xl p-5 mb-6">
              <h3 class="font-semibold text-yellow-800 mb-4">Preguntas de la leccion</h3>
              <div v-if="resultadoInt" class="text-center py-4">
                <p class="text-lg font-bold text-gray-800">{{ resultadoInt.puntaje.toFixed(1) }} / {{ resultadoInt.puntaje_max.toFixed(1) }}</p>
                <p class="text-sm text-gray-500">{{ resultadoInt.porcentaje?.toFixed(0) }}% correcto</p>
                <button @click="showIntermedias = false" class="mt-3 text-sm text-blue-600 hover:underline">Continuar</button>
              </div>
              <div v-else class="space-y-4">
                <div v-for="p in preguntas" :key="p.id">
                  <p class="text-sm font-medium text-gray-800 mb-2">{{ p.texto }}</p>
                  <div v-if="p.tipo === 'open_text'">
                    <textarea v-model="respuestas[p.id]" rows="2" placeholder="Tu respuesta..." class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" />
                  </div>
                  <div v-else class="space-y-1.5">
                    <label v-for="op in p.opciones" :key="op.id" class="flex items-center gap-2 cursor-pointer">
                      <input type="radio" :name="p.id" :value="op.id" v-model="respuestas[p.id]" />
                      <span class="text-sm text-gray-700">{{ op.texto }}</span>
                    </label>
                  </div>
                </div>
                <button @click="submitIntermedias" class="bg-yellow-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-yellow-700">
                  Enviar respuestas
                </button>
              </div>
            </div>

            <!-- Foro -->
            <div class="bg-white rounded-xl border border-gray-200 p-5">
              <div class="flex items-center justify-between mb-4">
                <h3 class="font-semibold text-gray-700">Foro de la leccion</h3>
                <button @click="showNuevoPost = !showNuevoPost" class="text-xs bg-blue-600 text-white px-3 py-1.5 rounded-lg hover:bg-blue-700">
                  + Nuevo post
                </button>
              </div>

              <div v-if="showNuevoPost" class="border border-gray-200 rounded-lg p-4 mb-4 bg-gray-50">
                <input v-model="nuevoPost.titulo" placeholder="Titulo del post" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm mb-2 focus:outline-none focus:ring-2 focus:ring-blue-500" />
                <textarea v-model="nuevoPost.contenido" placeholder="Escribe tu pregunta o comentario..." rows="3" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                <div class="flex gap-2 mt-2">
                  <button @click="crearPost" class="bg-blue-600 text-white px-3 py-1.5 rounded-lg text-sm hover:bg-blue-700">Publicar</button>
                  <button @click="showNuevoPost = false" class="px-3 py-1.5 rounded-lg text-sm border border-gray-300">Cancelar</button>
                </div>
              </div>

              <div v-if="foroPosts.length === 0" class="text-sm text-gray-400 text-center py-4">Sin posts aun. Se el primero en preguntar.</div>
              <div class="space-y-3">
                <div v-for="post in foroPosts" :key="post.id" class="border border-gray-200 rounded-lg overflow-hidden">
                  <div class="px-4 py-3 flex items-start justify-between cursor-pointer hover:bg-gray-50" @click="togglePost(post.id)">
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-gray-800">{{ post.titulo }}</p>
                      <p class="text-xs text-gray-500 mt-0.5">{{ post.user_name }} - {{ new Date(post.created_at).toLocaleDateString('es') }}</p>
                    </div>
                    <div class="flex gap-1 ml-2">
                      <button @click.stop="eliminarPost(post.id)" class="text-xs text-red-400 hover:text-red-600 px-1">X</button>
                      <span class="text-xs text-gray-400">{{ expandedPost === post.id ? 'Cerrar' : 'Ver' }}</span>
                    </div>
                  </div>
                  <div v-if="expandedPost === post.id" class="border-t border-gray-100 px-4 py-3 bg-gray-50">
                    <p class="text-sm text-gray-700 mb-4" style="white-space: pre-wrap">{{ post.contenido }}</p>
                    <div class="space-y-2 mb-3">
                      <div v-for="com in (comentariosMap[post.id] || [])" :key="com.id" class="bg-white rounded-lg px-3 py-2 border border-gray-100">
                        <p class="text-sm text-gray-700">{{ com.contenido }}</p>
                        <p class="text-xs text-gray-400 mt-0.5">{{ com.user_name }}</p>
                      </div>
                      <p v-if="!(comentariosMap[post.id]?.length)" class="text-xs text-gray-400">Sin comentarios aun.</p>
                    </div>
                    <div class="flex gap-2">
                      <input v-model="nuevoComentario[post.id]" placeholder="Agregar comentario..." class="flex-1 border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" @keydown.enter="crearComentario(post.id)" />
                      <button @click="crearComentario(post.id)" class="bg-blue-600 text-white px-3 py-1.5 rounded-lg text-sm hover:bg-blue-700">Enviar</button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>
