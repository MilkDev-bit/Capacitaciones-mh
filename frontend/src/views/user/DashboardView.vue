<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import { useAuthStore } from '../../stores/auth'
import CourseCard from '../../components/CourseCard.vue'

const auth = useAuthStore()
const router = useRouter()

const capacitaciones = ref<any[]>([])
const examenes = ref<any[]>([])
const loading = ref(true)

const firstName = computed(() => {
  return auth.user?.name?.split(' ')[0] ?? 'Estudiante'
})

const stats = computed(() => {
  const total = capacitaciones.value.length
  const completed = capacitaciones.value.filter(c => {
    const t = c.total_lecciones ?? 0
    const d = c.lecciones_completadas ?? 0
    return t > 0 && d === t
  }).length
  const avgProgress = total > 0
    ? Math.round(
        capacitaciones.value.reduce((sum, c) => {
          const t = c.total_lecciones ?? 0
          const d = c.lecciones_completadas ?? 0
          return sum + (t > 0 ? (d / t) * 100 : 0)
        }, 0) / total
      )
    : 0
  return { total, completed, exams: examenes.value.length, avgProgress }
})

const inProgress = computed(() =>
  capacitaciones.value
    .filter(c => {
      const t = c.total_lecciones ?? 0
      const d = c.lecciones_completadas ?? 0
      return t > 0 && d < t && d > 0
    })
    .slice(0, 3)
)

async function loadData() {
  loading.value = true
  try {
    const [cursosRes, exRes] = await Promise.all([
      api.get('/mis-capacitaciones'),
      api.get('/mis-examenes').catch(() => ({ data: [] })),
    ])
    const cursos = cursosRes.data || []
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
    examenes.value = exRes.data || []
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

const statCards = computed(() => [
  {
    label: 'Cursos inscritos',
    value: stats.value.total,
    icon: '📚',
    color: 'bg-violet-50 text-violet-700',
    iconBg: 'bg-violet-100',
  },
  {
    label: 'Cursos completados',
    value: stats.value.completed,
    icon: '✅',
    color: 'bg-emerald-50 text-emerald-700',
    iconBg: 'bg-emerald-100',
  },
  {
    label: 'Exámenes',
    value: stats.value.exams,
    icon: '📝',
    color: 'bg-sky-50 text-sky-700',
    iconBg: 'bg-sky-100',
  },
  {
    label: 'Progreso promedio',
    value: `${stats.value.avgProgress}%`,
    icon: '📈',
    color: 'bg-orange-50 text-orange-700',
    iconBg: 'bg-orange-100',
  },
])
</script>

<template>
  <div class="space-y-8">
    <!-- Welcome header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-gray-900">
          ¡Hola, {{ firstName }}! 👋
        </h1>
        <p class="text-gray-500 mt-1 text-sm">Continúa aprendiendo donde lo dejaste.</p>
      </div>
      <div class="flex gap-3">
        <button
          class="bg-brand hover:bg-brand-dark text-white text-sm font-semibold px-4 py-2 rounded-xl transition-colors shadow-sm"
          @click="router.push('/usuario/capacitaciones')"
        >
          Explorar cursos
        </button>
      </div>
    </div>

    <!-- Stats grid -->
    <div v-if="!loading" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div
        v-for="stat in statCards"
        :key="stat.label"
        :class="['rounded-2xl p-5 flex items-center gap-4', stat.color]"
      >
        <div :class="['w-11 h-11 rounded-xl flex items-center justify-center text-xl flex-shrink-0', stat.iconBg]">
          {{ stat.icon }}
        </div>
        <div>
          <p class="text-2xl font-extrabold leading-none">{{ stat.value }}</p>
          <p class="text-xs font-medium mt-1 opacity-80">{{ stat.label }}</p>
        </div>
      </div>
    </div>
    <!-- Stats skeleton -->
    <div v-else class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="i in 4" :key="i" class="bg-gray-100 animate-pulse rounded-2xl h-20" />
    </div>

    <!-- Continue Learning -->
    <div v-if="!loading">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-bold text-gray-900">Continuar aprendiendo</h2>
        <button
          class="text-sm font-semibold text-brand hover:text-brand-dark transition-colors"
          @click="router.push('/usuario/capacitaciones')"
        >
          Ver todos →
        </button>
      </div>

      <div v-if="inProgress.length" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
        <CourseCard
          v-for="course in inProgress"
          :key="course.id"
          :course="course"
          mode="enrolled"
          @navigate="(id) => router.push('/usuario/capacitaciones/' + id)"
        />
      </div>

      <div v-else class="bg-white rounded-2xl border border-gray-100 shadow-sm p-10 text-center">
        <div class="text-4xl mb-3">🎓</div>
        <p class="font-bold text-gray-800 text-base">No tienes cursos en progreso</p>
        <p class="text-gray-500 text-sm mt-1 mb-5">¡Empieza alguno de tus cursos inscritos!</p>
        <button
          class="bg-brand hover:bg-brand-dark text-white text-sm font-semibold px-5 py-2.5 rounded-xl transition-colors"
          @click="router.push('/usuario/capacitaciones')"
        >
          Ir a mis cursos
        </button>
      </div>
    </div>

    <!-- Quick Actions -->
    <div v-if="!loading">
      <h2 class="text-lg font-bold text-gray-900 mb-4">Acciones rápidas</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <button
          class="flex items-center gap-4 bg-white hover:bg-gray-50 border border-gray-100 rounded-2xl p-5 text-left shadow-sm transition-all hover:shadow-md hover:-translate-y-0.5"
          @click="router.push('/usuario/capacitaciones')"
        >
          <div class="w-11 h-11 bg-violet-100 rounded-xl flex items-center justify-center text-xl flex-shrink-0">🔍</div>
          <div>
            <p class="font-bold text-gray-900 text-sm">Explorar cursos</p>
            <p class="text-xs text-gray-500 mt-0.5">Descubre nuevos cursos disponibles</p>
          </div>
        </button>
        <button
          class="flex items-center gap-4 bg-white hover:bg-gray-50 border border-gray-100 rounded-2xl p-5 text-left shadow-sm transition-all hover:shadow-md hover:-translate-y-0.5"
          @click="router.push('/usuario/capacitaciones')"
        >
          <div class="w-11 h-11 bg-amber-100 rounded-xl flex items-center justify-center text-xl flex-shrink-0">🔑</div>
          <div>
            <p class="font-bold text-gray-900 text-sm">Unirse con código</p>
            <p class="text-xs text-gray-500 mt-0.5">Ingresa el código de tu instructor</p>
          </div>
        </button>
      </div>
    </div>
  </div>
</template>
