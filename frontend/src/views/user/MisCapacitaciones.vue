<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import CourseCard from '../../components/CourseCard.vue'

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
  if (!code) { codigoError.value = 'Ingresa un codigo'; return }
  codigoError.value = ''; codigoSuccess.value = ''
  codigoLoading.value = true
  try {
    const res = await api.post('/unirse-con-codigo', { codigo: code })
    codigoSuccess.value = `Te uniste a "${res.data.title}"!`
    codigoInput.value = ''
    await loadMis()
    setTimeout(() => { codigoSuccess.value = ''; activeTab.value = 'mis' }, 2000)
  } catch (e: any) {
    codigoError.value = e.response?.data?.error || 'Codigo invalido'
  } finally { codigoLoading.value = false }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Page header -->
    <div>
      <h1 class="text-2xl font-extrabold text-gray-900">Mis aprendizajes</h1>
      <p class="text-gray-500 text-sm mt-1">Todos tus cursos y capacitaciones asignadas</p>
    </div>

    <!-- Tabs -->
    <div class="flex gap-2 flex-wrap">
      <button
        :class="[
          'flex items-center gap-2 px-5 py-2 rounded-full border-2 text-sm font-semibold transition-all',
          activeTab === 'mis'
            ? 'border-orange-500 bg-orange-500 text-white'
            : 'border-gray-200 bg-white text-gray-500 hover:border-orange-400 hover:text-orange-500'
        ]"
        @click="activeTab = 'mis'"
      >
        Mis cursos
        <span
          :class="[
            'text-xs font-bold px-2 py-0.5 rounded-full',
            activeTab === 'mis' ? 'bg-white/25 text-white' : 'bg-gray-100 text-gray-500'
          ]"
        >
          {{ capacitaciones.length }}
        </span>
      </button>
      <button
        :class="[
          'flex items-center gap-2 px-5 py-2 rounded-full border-2 text-sm font-semibold transition-all',
          activeTab === 'explorar'
            ? 'border-orange-500 bg-orange-500 text-white'
            : 'border-gray-200 bg-white text-gray-500 hover:border-orange-400 hover:text-orange-500'
        ]"
        @click="activeTab = 'explorar'"
      >
        Explorar
        <span
          :class="[
            'text-xs font-bold px-2 py-0.5 rounded-full',
            activeTab === 'explorar' ? 'bg-white/25 text-white' : 'bg-gray-100 text-gray-500'
          ]"
        >
          {{ cursosPublicos.length }}
        </span>
      </button>
    </div>

    <!-- Mis cursos -->
    <div v-if="activeTab === 'mis'">
      <div v-if="capacitaciones.length" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
        <CourseCard
          v-for="c in capacitaciones"
          :key="c.id"
          :course="c"
          mode="enrolled"
          @navigate="(id) => router.push('/usuario/capacitaciones/' + id)"
        />
      </div>
      <div v-else class="flex flex-col items-center text-center py-16 gap-3 bg-white rounded-2xl border border-gray-100 shadow-sm">
        <div class="text-5xl">&#128218;</div>
        <h3 class="text-base font-bold text-gray-800">No tienes cursos asignados aun</h3>
        <p class="text-gray-500 text-sm max-w-sm">Explora los cursos disponibles o pide a tu instructor un codigo de acceso.</p>
        <button
          class="mt-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-semibold px-5 py-2.5 rounded-xl transition-colors"
          @click="activeTab = 'explorar'"
        >
          Explorar cursos
        </button>
      </div>
    </div>

    <!-- Explorar -->
    <div v-if="activeTab === 'explorar'" class="space-y-6">
      <!-- Join by code banner -->
      <div class="bg-white border border-gray-100 shadow-sm rounded-2xl p-5 border-l-4 border-l-orange-500 flex flex-col sm:flex-row sm:items-center gap-4">
        <div class="flex items-center gap-4 flex-1 min-w-0">
          <div class="w-11 h-11 bg-amber-100 rounded-xl flex items-center justify-center text-xl flex-shrink-0">&#128273;</div>
          <div>
            <p class="font-bold text-gray-900 text-sm">Tienes un codigo de acceso?</p>
            <p class="text-gray-500 text-xs mt-0.5">Ingresa el codigo de tu instructor para unirte a un curso privado.</p>
          </div>
        </div>
        <div class="flex gap-2 sm:flex-shrink-0">
          <input
            v-model="codigoInput"
            class="px-4 py-2.5 border-2 border-gray-200 rounded-xl text-sm font-bold uppercase tracking-widest w-32 focus:outline-none focus:border-orange-400 focus:ring-2 focus:ring-orange-100 bg-gray-50 font-mono"
            placeholder="ABC123"
            maxlength="12"
            @keyup.enter="unirseConCodigo"
          />
          <button
            class="bg-orange-500 hover:bg-orange-600 disabled:opacity-50 text-white text-sm font-semibold px-4 py-2.5 rounded-xl transition-colors whitespace-nowrap"
            :disabled="codigoLoading"
            @click="unirseConCodigo"
          >
            {{ codigoLoading ? 'Cargando...' : 'Unirme' }}
          </button>
        </div>
      </div>

      <div v-if="codigoError" class="bg-red-50 border border-red-200 text-red-700 text-sm rounded-xl px-4 py-3">{{ codigoError }}</div>
      <div v-if="codigoSuccess" class="bg-emerald-50 border border-emerald-200 text-emerald-700 text-sm rounded-xl px-4 py-3">{{ codigoSuccess }}</div>

      <div>
        <p class="text-base font-bold text-gray-900 mb-4">Cursos disponibles para todos</p>
        <div v-if="cursosPublicos.length" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
          <CourseCard
            v-for="c in cursosPublicos"
            :key="c.id"
            :course="c"
            mode="public"
            :loading="inscribiendose === c.id"
            @enroll="(id) => inscribirse(id)"
            @navigate="(id) => router.push('/usuario/capacitaciones/' + id)"
          />
        </div>
        <div v-else class="flex flex-col items-center text-center py-16 gap-3 bg-white rounded-2xl border border-gray-100 shadow-sm">
          <div class="text-5xl">&#128269;</div>
          <h3 class="text-base font-bold text-gray-800">No hay cursos publicos disponibles</h3>
          <p class="text-gray-500 text-sm max-w-sm">Pide a tu instructor que comparta el enlace o codigo de su curso.</p>
        </div>
      </div>
    </div>
  </div>
</template>
