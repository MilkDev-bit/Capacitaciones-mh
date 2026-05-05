<script setup lang="ts">
import { computed } from 'vue'

interface Course {
  id: string
  title: string
  description?: string
  type?: string
  total_lecciones?: number
  lecciones_completadas?: number
  inscrito?: boolean
  is_public?: boolean
}

const props = defineProps<{
  course: Course
  mode: 'enrolled' | 'public' | 'instructor'
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'enroll', id: string): void
  (e: 'manage', id: string): void
  (e: 'navigate', id: string): void
}>()

const thumbGradient = computed(() => {
  const map: Record<string, string> = {
    video: 'from-violet-600 to-indigo-700',
    document: 'from-sky-500 to-blue-700',
    text: 'from-emerald-500 to-teal-700',
  }
  return map[props.course.type ?? ''] ?? 'from-orange-500 to-brand-dark'
})

const typeIcon = computed(() => {
  const map: Record<string, string> = { video: '🎥', document: '📄', text: '📝' }
  return map[props.course.type ?? ''] ?? '📚'
})

const typeLabel = computed(() => {
  const map: Record<string, string> = { video: 'Video', document: 'Documento', text: 'Texto' }
  return map[props.course.type ?? ''] ?? (props.course.type ?? 'Curso')
})

const progress = computed(() => {
  const total = props.course.total_lecciones ?? 0
  const done = props.course.lecciones_completadas ?? 0
  return total > 0 ? Math.round((done / total) * 100) : 0
})

const progressColor = computed(() => {
  if (progress.value >= 100) return 'bg-emerald-500'
  if (progress.value >= 50) return 'bg-brand'
  return 'bg-brand'
})
</script>

<template>
  <div
    class="group bg-white rounded-2xl overflow-hidden shadow-sm border border-gray-100 flex flex-col cursor-pointer transition-all duration-200 hover:-translate-y-1 hover:shadow-lg"
    @click="emit('navigate', course.id)"
    @keyup.enter="emit('navigate', course.id)"
    tabindex="0"
    role="article"
  >
    <!-- Thumbnail -->
    <div :class="['bg-gradient-to-br', thumbGradient, 'h-36 flex items-center justify-center relative flex-shrink-0']">
      <span class="text-5xl drop-shadow-md">{{ typeIcon }}</span>
      <!-- Enrolled ribbon -->
      <span
        v-if="mode === 'public' && course.inscrito"
        class="absolute top-2 right-2 bg-black/50 backdrop-blur-sm text-white text-xs font-bold px-3 py-1 rounded-full"
      >
        ✓ Inscrito
      </span>
      <!-- Instructor badge -->
      <span
        v-if="mode === 'instructor'"
        class="absolute top-2 right-2 bg-black/50 backdrop-blur-sm text-white text-xs font-bold px-3 py-1 rounded-full"
      >
        ✏️ Tuyo
      </span>
    </div>

    <!-- Body -->
    <div class="p-4 flex flex-col gap-2 flex-1">
      <!-- Type badge -->
      <span class="text-xs font-bold uppercase tracking-wide text-orange-700 bg-orange-50 px-2 py-0.5 rounded w-fit">
        {{ typeLabel }}
      </span>

      <!-- Title -->
      <h3 class="text-sm font-bold text-gray-900 leading-snug line-clamp-2">{{ course.title }}</h3>

      <!-- Description -->
      <p class="text-xs text-gray-500 leading-relaxed line-clamp-2 flex-1">
        {{ course.description || 'Sin descripción' }}
      </p>

      <!-- Progress bar (enrolled mode) -->
      <div v-if="mode === 'enrolled' && (course.total_lecciones ?? 0) > 0" class="mt-1">
        <div class="flex justify-between items-center mb-1">
          <span class="text-xs text-gray-500">{{ course.lecciones_completadas }}/{{ course.total_lecciones }} lecciones</span>
          <span class="text-xs font-bold text-brand">{{ progress }}%</span>
        </div>
        <div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
          <div
            :class="[progressColor, 'h-full rounded-full transition-all duration-300']"
            :style="`width: ${progress}%`"
          />
        </div>
      </div>

      <!-- CTA -->
      <div class="mt-2">
        <!-- Enrolled: navigate -->
        <button
          v-if="mode === 'enrolled'"
          class="w-full text-sm font-bold text-brand group-hover:text-brand-dark transition-colors flex items-center gap-1"
          @click.stop="emit('navigate', course.id)"
        >
          Continuar aprendiendo
          <svg class="w-4 h-4 transition-transform group-hover:translate-x-0.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M9 5l7 7-7 7"/></svg>
        </button>

        <!-- Public: enroll or enrolled badge -->
        <template v-if="mode === 'public'">
          <span v-if="course.inscrito" class="inline-flex items-center gap-1 text-xs font-bold text-emerald-600 bg-emerald-50 px-3 py-1.5 rounded-full">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>
            Ya inscrito
          </span>
          <button
            v-else
            class="w-full bg-brand hover:bg-brand-dark text-white text-sm font-semibold py-2 rounded-lg transition-colors disabled:opacity-50"
            :disabled="loading"
            @click.stop="emit('enroll', course.id)"
          >
            {{ loading ? 'Inscribiendo...' : '+ Inscribirse gratis' }}
          </button>
        </template>

        <!-- Instructor: manage -->
        <button
          v-if="mode === 'instructor'"
          class="w-full bg-gray-900 hover:bg-gray-700 text-white text-sm font-semibold py-2 rounded-lg transition-colors"
          @click.stop="emit('manage', course.id)"
        >
          Gestionar curso
        </button>
      </div>
    </div>
  </div>
</template>
