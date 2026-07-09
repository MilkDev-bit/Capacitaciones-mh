<script setup lang="ts">
/**
 * CourseSidebar.vue
 * Sidebar de navegación del alumno: muestra el árbol Módulo → Submódulo → Lección
 * con indicadores de progreso y estado de completado por lección.
 */
import { ref, computed } from 'vue'

interface Leccion {
  id: string
  title: string
  lesson_type: number
  duracion_min: number
  points_reward: number
  completada: boolean
}

interface Submodulo {
  id: string
  title: string
  lecciones: Leccion[]
}

interface Modulo {
  id: string
  title: string
  lecciones: Leccion[]
  submodulos: Submodulo[]
}

interface Tree {
  modulos: Modulo[]
  lecciones: Leccion[]   // lecciones sueltas
}

const props = defineProps<{
  tree: Tree
  selectedId: string | null
  cursoTitle?: string
}>()

const emit = defineEmits<{
  (e: 'select', lec: Leccion): void
}>()

// Colapsado por módulo/submódulo
const collapsed = ref<Set<string>>(new Set())
function toggle(id: string) {
  collapsed.value.has(id) ? collapsed.value.delete(id) : collapsed.value.add(id)
}

// Progreso por sección
function sectionProgress(lecciones: Leccion[]) {
  if (!lecciones || !lecciones.length) return 0
  return Math.round((lecciones.filter(l => l.completada).length / lecciones.length) * 100)
}
function moduleProgress(mod: Modulo) {
  const all: Leccion[] = [...(mod.lecciones ?? [])]
  mod.submodulos?.forEach(s => all.push(...(s.lecciones ?? [])))
  return sectionProgress(all)
}
function moduleTotal(mod: Modulo) {
  return (mod.lecciones?.length ?? 0) + (mod.submodulos ?? []).reduce((s, sub) => s + (sub.lecciones?.length ?? 0), 0)
}

// Total del curso
const allLecciones = computed<Leccion[]>(() => {
  const all: Leccion[] = [...(props.tree?.lecciones ?? [])]
  props.tree?.modulos?.forEach(m => {
    all.push(...(m.lecciones ?? []))
    m.submodulos?.forEach(s => all.push(...(s.lecciones ?? [])))
  })
  return all
})
const totalProgress = computed(() => sectionProgress(allLecciones.value))

// Íconos
const ICONS: Record<number, string> = {
  1: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="20" rx="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/><line x1="2" y1="7" x2="7" y2="7"/><line x1="2" y1="17" x2="7" y2="17"/><line x1="17" y1="17" x2="22" y2="17"/><line x1="17" y1="7" x2="22" y2="7"/></svg>',
  2: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>',
  3: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>',
  4: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
  5: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="9" height="7" rx="1"/><rect x="13" y="3" width="9" height="7" rx="1"/><rect x="2" y="14" width="9" height="7" rx="1"/><rect x="13" y="14" width="9" height="7" rx="1"/></svg>',
  6: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 9l4 4 4-4M5 15l4 4 4-4M17 9l2 2 2-2M17 15l2 2 2-2"/></svg>',
  7: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 7h2M11 7h2M15 7h2M7 11h2M11 11h2M15 11h2M7 15h2M11 15h2M15 15h2"/></svg>',
  8: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18M3 12h12M3 18h8"/></svg>',
  9: '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18M3 12h14M3 18h10"/><path d="M19 9l3 3-3 3M17 12h5"/></svg>'
}
function icon(t: number) { return ICONS[t] ?? ICONS[3] }
function isGame(t: number) { return t >= 5 && t <= 9 }
</script>

<template>
  <nav class="csb-root">
    <!-- Encabezado del curso -->
    <div class="csb-header">
      <p class="csb-course-name">{{ cursoTitle }}</p>
      <div class="csb-progress-bar-wrap">
        <div class="csb-progress-bar">
          <div class="csb-progress-fill" :style="{ width: totalProgress + '%' }" />
        </div>
        <span class="csb-progress-label">{{ totalProgress }}%</span>
      </div>
    </div>

    <div class="csb-tree">
      <!-- ── MÓDULOS ──────────────────────────────────────── -->
      <div v-for="mod in tree.modulos" :key="mod.id" class="csb-module">
        <button class="csb-module-btn" @click="toggle(mod.id)">
          <div class="csb-module-left">
            <span class="csb-module-chevron" :class="{ open: !collapsed.has(mod.id) }">›</span>
            <div class="csb-module-info">
              <span class="csb-module-title">{{ mod.title }}</span>
              <span class="csb-module-meta">{{ moduleTotal(mod) }} lecciones</span>
            </div>
          </div>
          <!-- Mini progress ring -->
          <svg class="csb-ring" width="28" height="28" viewBox="0 0 28 28">
            <circle cx="14" cy="14" r="11" fill="none" stroke="var(--border)" stroke-width="2.5"/>
            <circle
              cx="14" cy="14" r="11" fill="none"
              :stroke="moduleProgress(mod) === 100 ? 'var(--success)' : 'var(--brand)'"
              stroke-width="2.5"
              stroke-linecap="round"
              :stroke-dasharray="69.1"
              :stroke-dashoffset="69.1 * (1 - moduleProgress(mod) / 100)"
              transform="rotate(-90 14 14)"
            />
            <text x="14" y="18" text-anchor="middle" font-size="7" fill="var(--dark)" font-weight="700">
              {{ moduleProgress(mod) }}
            </text>
          </svg>
        </button>

        <div v-show="!collapsed.has(mod.id)" class="csb-module-body">

          <!-- Lecciones directas del módulo -->
          <button
            v-for="lec in mod.lecciones" :key="lec.id"
            class="csb-lesson"
            :class="{ selected: selectedId === lec.id, done: lec.completada, game: isGame(lec.lesson_type) }"
            @click="emit('select', lec)"
          >
            <span class="csb-lesson-icon" v-html="icon(lec.lesson_type)"></span>
            <span class="csb-lesson-title">{{ lec.title }}</span>
            <span v-if="lec.completada" class="csb-check">✓</span>
            <span v-else-if="isGame(lec.lesson_type) && lec.points_reward" class="csb-pts-badge">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
            </span>
          </button>

          <!-- ── SUBMÓDULOS ────────────────────────────────── -->
          <div v-for="sub in mod.submodulos" :key="sub.id" class="csb-sub">
            <button class="csb-sub-btn" @click="toggle(sub.id)">
              <span class="csb-sub-chevron" :class="{ open: !collapsed.has(sub.id) }">›</span>
              <div class="csb-sub-info">
                <span class="csb-sub-title">{{ sub.title }}</span>
                <div class="csb-sub-progress">
                  <div class="csb-sub-bar">
                    <div class="csb-sub-fill" :style="{ width: sectionProgress(sub.lecciones) + '%' }" />
                  </div>
                  <span class="csb-sub-pct">{{ sectionProgress(sub.lecciones) }}%</span>
                </div>
              </div>
            </button>

            <div v-show="!collapsed.has(sub.id)" class="csb-sub-lessons">
              <button
                v-for="lec in sub.lecciones" :key="lec.id"
                class="csb-lesson"
                :class="{ selected: selectedId === lec.id, done: lec.completada, game: isGame(lec.lesson_type) }"
                @click="emit('select', lec)"
              >
                <span class="csb-lesson-icon" v-html="icon(lec.lesson_type)"></span>
                <span class="csb-lesson-title">{{ lec.title }}</span>
                <span v-if="lec.completada" class="csb-check">✓</span>
                <span v-else-if="isGame(lec.lesson_type) && lec.points_reward" class="csb-pts-badge">
                  <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                  {{ lec.points_reward }}
                </span>
              </button>
              <div v-if="!sub.lecciones?.length" class="csb-empty">Submódulo sin lecciones</div>
            </div>
          </div>

          <div v-if="!mod.lecciones?.length && !mod.submodulos?.length" class="csb-empty">Módulo vacío</div>
        </div>
      </div>

      <!-- ── LECCIONES SUELTAS ────────────────────────────── -->
      <button
        v-for="lec in tree.lecciones" :key="lec.id"
        class="csb-lesson csb-lesson-loose"
        :class="{ selected: selectedId === lec.id, done: lec.completada, game: isGame(lec.lesson_type) }"
        @click="emit('select', lec)"
      >
        <span class="csb-lesson-icon" v-html="icon(lec.lesson_type)"></span>
        <span class="csb-lesson-title">{{ lec.title }}</span>
        <span v-if="lec.completada" class="csb-check">✓</span>
        <span v-else-if="isGame(lec.lesson_type) && lec.points_reward" class="csb-pts-badge">
          <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
          {{ lec.points_reward }}
        </span>
      </button>

      <div v-if="!tree.modulos?.length && !tree.lecciones?.length" class="csb-empty csb-empty-root">
        Este curso no tiene contenido aún
      </div>
    </div>
  </nav>
</template>

<style scoped>
.csb-root { display: flex; flex-direction: column; height: 100%; overflow: hidden; }

/* Header */
.csb-header { padding: 16px; border-bottom: 1px solid var(--border); flex-shrink: 0; }
.csb-course-name { font-size: 0.88rem; font-weight: 700; color: var(--dark); margin: 0 0 10px; line-height: 1.3; }
.csb-progress-bar-wrap { display: flex; align-items: center; gap: 10px; }
.csb-progress-bar { flex: 1; height: 5px; background: var(--border); border-radius: 10px; overflow: hidden; }
.csb-progress-fill { height: 100%; background: linear-gradient(90deg, var(--brand), #f97316); border-radius: 10px; transition: width 0.5s ease; }
.csb-progress-label { font-size: 0.75rem; font-weight: 700; color: var(--brand); min-width: 32px; text-align: right; }

/* Tree */
.csb-tree { flex: 1; overflow-y: auto; padding: 8px 6px; display: flex; flex-direction: column; gap: 2px; }

/* Module */
.csb-module { border-radius: var(--r-md); overflow: hidden; }
.csb-module-btn {
  width: 100%; display: flex; align-items: center; justify-content: space-between;
  padding: 10px 12px; gap: 10px;
  background: var(--surface-soft); border: none; cursor: pointer;
  text-align: left; border-radius: var(--r-md); transition: background 0.15s;
}
.csb-module-btn:hover { background: color-mix(in srgb, var(--brand) 8%, var(--surface-soft)); }
.csb-module-left { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
.csb-module-chevron { font-size: 1.1rem; color: var(--muted); transition: transform 0.2s; display: inline-block; }
.csb-module-chevron.open { transform: rotate(90deg); }
.csb-module-info { display: flex; flex-direction: column; min-width: 0; }
.csb-module-title { font-size: 0.88rem; font-weight: 700; color: var(--dark); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.csb-module-meta { font-size: 0.7rem; color: var(--muted); }
.csb-ring { flex-shrink: 0; }
.csb-module-body { padding: 4px 0 4px 14px; display: flex; flex-direction: column; gap: 2px; }

/* Submodule */
.csb-sub { border-radius: var(--r-sm); overflow: hidden; }
.csb-sub-btn {
  width: 100%; display: flex; align-items: center; gap: 8px;
  padding: 8px 10px; background: none; border: none;
  cursor: pointer; text-align: left; border-radius: var(--r-sm); transition: background 0.15s;
}
.csb-sub-btn:hover { background: var(--surface-soft); }
.csb-sub-chevron { font-size: 0.95rem; color: var(--muted); transition: transform 0.2s; display: inline-block; flex-shrink: 0; }
.csb-sub-chevron.open { transform: rotate(90deg); }
.csb-sub-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.csb-sub-title { font-size: 0.82rem; font-weight: 600; color: var(--dark); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.csb-sub-progress { display: flex; align-items: center; gap: 6px; }
.csb-sub-bar { flex: 1; height: 3px; background: var(--border); border-radius: 3px; overflow: hidden; }
.csb-sub-fill { height: 100%; background: var(--brand); border-radius: 3px; transition: width 0.4s; }
.csb-sub-pct { font-size: 0.65rem; color: var(--muted); min-width: 26px; text-align: right; }
.csb-sub-lessons { padding-left: 14px; display: flex; flex-direction: column; gap: 2px; padding-top: 2px; }

/* Lesson */
.csb-lesson {
  display: flex; align-items: center; gap: 8px;
  padding: 7px 10px; border-radius: var(--r-sm);
  background: none; border: none; cursor: pointer;
  text-align: left; width: 100%; transition: all 0.15s;
}
.csb-lesson:hover { background: var(--surface-soft); }
.csb-lesson.selected { background: var(--brand-light); }
.csb-lesson.done .csb-lesson-title { text-decoration: line-through; color: var(--muted); }
.csb-lesson.game { }
.csb-lesson.selected .csb-lesson-title { color: var(--brand); font-weight: 700; }
.csb-lesson-loose { padding-left: 12px; }
.csb-lesson-icon { font-size: 0.95rem; flex-shrink: 0; }
.csb-lesson-title { flex: 1; font-size: 0.83rem; color: var(--dark); font-weight: 500; text-align: left; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.csb-check { color: var(--success); font-size: 0.8rem; font-weight: 700; flex-shrink: 0; }
.csb-pts-badge { font-size: 0.65rem; color: var(--brand); font-weight: 700; background: var(--brand-light); border-radius: 10px; padding: 1px 6px; flex-shrink: 0; white-space: nowrap; }

/* Empties */
.csb-empty { font-size: 0.75rem; color: var(--muted); padding: 8px 12px; text-align: center; }
.csb-empty-root { padding: 30px; }
</style>
