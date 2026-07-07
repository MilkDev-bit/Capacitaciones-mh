<script setup lang="ts">
/**
 * CourseTreeEditor.vue
 * Editor jerárquico de la estructura de un curso:
 *   Módulo → Submódulo → Lección
 * El instructor puede crear/editar/eliminar/reordenar cada nivel,
 * y configurar minijuegos desde el panel lateral.
 */
import { ref, onMounted, computed } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'
import { uploadToR2 } from '../utils/upload'
import DragDropUpload from './DragDropUpload.vue'
import ContentTypeSelector from './ContentTypeSelector.vue'
import GameConfigEditor from './GameConfigEditor.vue'

const props = defineProps<{ capId: string }>()
const emit = defineEmits<{ (e: 'update'): void }>()

// ── Estado del árbol ───────────────────────────────────────────────────────────
const tree = ref<{ modulos: any[], lecciones: any[] }>({ modulos: [], lecciones: [] })
const loading = ref(false)

async function fetchTree() {
  loading.value = true
  try {
    const res = await api.get(`/instructor/capacitaciones/${props.capId}/tree`)
    tree.value = res.data || { modulos: [], lecciones: [] }
  } catch {
    toast.error('Error al cargar el árbol del curso')
  } finally {
    loading.value = false
  }
}

onMounted(fetchTree)

// ── Panel lateral (drawer) ─────────────────────────────────────────────────────
type PanelMode = 'create-module' | 'edit-module' | 'create-sub' | 'edit-sub' | 'create-lesson' | 'edit-lesson' | null
const panelMode = ref<PanelMode>(null)
const panelCtx  = ref<any>({})       // contexto (moduloId, submoduloId, lección…)
const panelForm = ref<any>({})
const panelFile = ref<File | null>(null)
const panelSaving = ref(false)

function openPanel(mode: PanelMode, ctx: any = {}, form: any = {}) {
  panelMode.value = mode
  panelCtx.value  = ctx
  panelForm.value = { type: '1', lesson_type: '1', duracion_min: 0, points_reward: 100, ...form }
  panelFile.value = null
}
function closePanel() { panelMode.value = null }

const panelTitle = computed(() => ((({
  'create-module':  'Nuevo Módulo',
  'edit-module':    'Editar Módulo',
  'create-sub':     'Nuevo Submódulo',
  'edit-sub':       'Editar Submódulo',
  'create-lesson':  'Nueva Lección',
  'edit-lesson':    'Editar Lección',
} as Record<string, string>)[panelMode.value ?? ''] ?? '')))

const isGameType = computed(() => {
  const t = String(panelForm.value.lesson_type ?? panelForm.value.type ?? '')
  return ['5','6','7','8','9'].includes(t)
})

// ── Módulos ────────────────────────────────────────────────────────────────────
async function saveModule() {
  if (!panelForm.value.title) return toast.error('Título requerido')
  panelSaving.value = true
  try {
    if (panelMode.value === 'create-module') {
      await api.post(`/instructor/capacitaciones/${props.capId}/modulos`, {
        title: panelForm.value.title,
        description: panelForm.value.description,
        orden: tree.value.modulos.length,
      })
      toast.success('Módulo creado')
    } else {
      await api.put(`/instructor/capacitaciones/${props.capId}/modulos/${panelCtx.value.moduloId}`, {
        title: panelForm.value.title,
        description: panelForm.value.description,
      })
      toast.success('Módulo actualizado')
    }
    closePanel(); fetchTree()
  } catch { toast.error('Error al guardar módulo') }
  finally { panelSaving.value = false }
}

async function deleteModule(moduloId: string) {
  if (!await toast.confirm('¿Eliminar módulo y todo su contenido?')) return
  try {
    await api.delete(`/instructor/capacitaciones/${props.capId}/modulos/${moduloId}`)
    toast.success('Módulo eliminado'); fetchTree()
  } catch { toast.error('Error al eliminar') }
}

async function reorderModulos(newOrder: string[]) {
  await api.put(`/instructor/capacitaciones/${props.capId}/modulos/reorder`, { ids: newOrder })
}

// ── Submódulos ────────────────────────────────────────────────────────────────
async function saveSub() {
  if (!panelForm.value.title) return toast.error('Título requerido')
  panelSaving.value = true
  try {
    if (panelMode.value === 'create-sub') {
      const mod = tree.value.modulos.find((m: any) => m.id === panelCtx.value.moduloId)
      await api.post(`/instructor/capacitaciones/${props.capId}/modulos/${panelCtx.value.moduloId}/submodulos`, {
        title: panelForm.value.title,
        description: panelForm.value.description,
        orden: mod?.submodulos?.length ?? 0,
      })
      toast.success('Submódulo creado')
    } else {
      await api.put(
        `/instructor/capacitaciones/${props.capId}/modulos/${panelCtx.value.moduloId}/submodulos/${panelCtx.value.submoduloId}`,
        { title: panelForm.value.title, description: panelForm.value.description }
      )
      toast.success('Submódulo actualizado')
    }
    closePanel(); fetchTree()
  } catch { toast.error('Error al guardar submódulo') }
  finally { panelSaving.value = false }
}

async function deleteSub(moduloId: string, subId: string) {
  if (!await toast.confirm('¿Eliminar submódulo y sus lecciones?')) return
  try {
    await api.delete(`/instructor/capacitaciones/${props.capId}/modulos/${moduloId}/submodulos/${subId}`)
    toast.success('Submódulo eliminado'); fetchTree()
  } catch { toast.error('Error al eliminar') }
}

// ── Lecciones ─────────────────────────────────────────────────────────────────
function isValidUrl(v: string) {
  try { const u = new URL(v); return u.protocol === 'http:' || u.protocol === 'https:' }
  catch { return false }
}

async function saveLesson() {
  const form = panelForm.value
  if (!form.title) return toast.error('Título requerido')

  const type = String(form.lesson_type ?? form.type ?? '1')
  if (type === '2' && !form.content?.trim()) return toast.error('El contenido de texto es requerido')
  // link es tipo externo — no aplica a enums numéricos
  if (type === '99' && !isValidUrl(form.content || '')) return toast.error('URL inválida')

  const payload: Record<string, any> = {
    title: form.title,
    description: form.description ?? '',
    lesson_type: Number(type),
    content: form.content ?? '',
    duracion_min: Number(form.duracion_min || 0),
    file_path: form.file_path || '',
    modulo_id: panelCtx.value.moduloId ?? '',
    submodulo_id: panelCtx.value.submoduloId ?? '',
    game_config_json: form.game_config_json ?? '',
    points_reward: Number(form.points_reward || 0),
  }

  if (panelFile.value && (type === '1' || type === '3')) {
    const t = toast.loading('Subiendo archivo...')
    try {
      payload.file_path = await uploadToR2(panelFile.value, type === '1' ? 'videos' : 'documents')
    } finally { t.close() }
  }

  panelSaving.value = true
  try {
    if (panelMode.value === 'create-lesson') {
      await api.post(`/instructor/capacitaciones/${props.capId}/lecciones`, payload)
      toast.success('Lección creada')
    } else {
      await api.put(`/instructor/capacitaciones/${props.capId}/lecciones/${panelCtx.value.leccionId}`, payload)
      toast.success('Lección actualizada')
    }
    closePanel(); fetchTree()
  } catch { toast.error('Error al guardar lección') }
  finally { panelSaving.value = false }
}

async function deleteLesson(leccionId: string) {
  if (!await toast.confirm('¿Eliminar lección?')) return
  try {
    await api.delete(`/instructor/capacitaciones/${props.capId}/lecciones/${leccionId}`)
    toast.success('Lección eliminada'); fetchTree()
  } catch { toast.error('Error al eliminar') }
}

// ── Helpers UI ────────────────────────────────────────────────────────────────
const collapsed = ref<Set<string>>(new Set())
function toggleCollapse(id: string) {
  collapsed.value.has(id) ? collapsed.value.delete(id) : collapsed.value.add(id)
}

const LESSON_TYPE_ICONS: Record<string, string> = {
  '1': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="20" rx="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/><line x1="2" y1="7" x2="7" y2="7"/><line x1="2" y1="17" x2="7" y2="17"/><line x1="17" y1="17" x2="22" y2="17"/><line x1="17" y1="7" x2="22" y2="7"/></svg>',
  '2': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>',
  '3': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>',
  '4': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
  '5': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="9" height="7" rx="1"/><rect x="13" y="3" width="9" height="7" rx="1"/><rect x="2" y="14" width="9" height="7" rx="1"/><rect x="13" y="14" width="9" height="7" rx="1"/></svg>',
  '6': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 9l4 4 4-4M5 15l4 4 4-4M17 9l2 2 2-2M17 15l2 2 2-2"/></svg>',
  '7': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 7h2M11 7h2M15 7h2M7 11h2M11 11h2M15 11h2M7 15h2M11 15h2M15 15h2"/></svg>',
  '8': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18M3 12h12M3 18h8"/></svg>',
  '9': '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18M3 12h14M3 18h10"/><path d="M19 9l3 3-3 3M17 12h5"/></svg>',
}
const LESSON_TYPE_NAMES: Record<string, string> = {
  '1': 'Video', '2': 'Texto', '3': 'PDF', '4': 'Quiz',
  '5': 'Memorama', '6': 'Arrastrar', '7': 'Sopa letras',
  '8': 'Completar', '9': 'Ordenar',
}
function lessonIcon(t: any) { return LESSON_TYPE_ICONS[String(t ?? '1')] ?? LESSON_TYPE_ICONS['3'] }
function lessonTypeName(t: any) { return LESSON_TYPE_NAMES[String(t ?? '1')] ?? 'Lección' }

// Mover módulo arriba/abajo y sincronizar orden
async function moveModule(modulos: any[], i: any, dir: -1 | 1) {
  const idx = Number(i)
  const j = idx + dir
  if (j < 0 || j >= modulos.length) return
  const temp = modulos[idx];
  const target = modulos[j];
  if (temp && target) {
    modulos[idx] = target;
    modulos[j] = temp;
  }
  await reorderModulos(modulos.map((m: any) => m.id)).catch(() => {
    toast.error('Error al reordenar'); fetchTree()
  })
}

async function moveSub(subs: any[], moduloId: string, i: any, dir: -1 | 1) {
  const idx = Number(i)
  const j = idx + dir
  if (j < 0 || j >= subs.length) return
  const temp = subs[idx];
  const target = subs[j];
  if (temp && target) {
    subs[idx] = target;
    subs[j] = temp;
  }
  await api.put(`/instructor/capacitaciones/${props.capId}/modulos/${moduloId}/submodulos/reorder`, {
    ids: subs.map((s: any) => s.id),
  }).catch(() => { toast.error('Error al reordenar'); fetchTree() })
}

async function moveLeccion(lecciones: any[], cursoId: string, i: any, dir: -1 | 1) {
  const idx = Number(i)
  const j = idx + dir
  if (j < 0 || j >= lecciones.length) return
  const temp = lecciones[idx];
  const target = lecciones[j];
  if (temp && target) {
    lecciones[idx] = target;
    lecciones[j] = temp;
  }
  await api.put(`/instructor/capacitaciones/${props.capId}/lecciones/reorder`, {
    ids: lecciones.map((l: any) => l.id),
  }).catch(() => { toast.error('Error al reordenar'); fetchTree() })
}
</script>

<template>
  <div class="cte-root">
    <!-- ── Barra de acciones global ───────────────────────────── -->
    <div class="cte-toolbar">
      <h3 class="cte-toolbar-title">Contenido del curso</h3>
      <div class="cte-toolbar-actions">
        <button class="cte-btn cte-btn-outline" @click="openPanel('create-module')">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
          Módulo
        </button>
        <button class="cte-btn cte-btn-outline" @click="openPanel('create-lesson', {})">
          <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
          Lección suelta
        </button>
      </div>
    </div>

    <div v-if="loading" class="cte-loading">
      <span class="cte-spinner" />
      Cargando estructura…
    </div>

    <div v-else class="cte-tree">

      <!-- ── MÓDULOS ─────────────────────────────────────────── -->
      <div
        v-for="(mod, mi) in tree.modulos"
        :key="mod.id"
        class="cte-module"
      >
        <!-- Cabecera del módulo -->
        <div class="cte-module-header">
          <div class="cte-reorder">
            <button class="cte-arr" :disabled="mi === 0" @click="moveModule(tree.modulos, mi, -1)">
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M18 15l-6-6-6 6"/></svg>
            </button>
            <button class="cte-arr" :disabled="mi === tree.modulos.length - 1" @click="moveModule(tree.modulos, mi, 1)">
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9l6 6 6-6"/></svg>
            </button>
          </div>
          <button class="cte-collapse-btn" @click="toggleCollapse(mod.id)">
            <span class="cte-chevron" :class="{ open: !collapsed.has(mod.id) }">›</span>
          </button>
          <div class="cte-module-info">
            <span class="cte-module-label">MÓDULO {{ Number(mi) + 1 }}</span>
            <span class="cte-module-title">{{ mod.title }}</span>
          </div>
          <div class="cte-module-actions">
            <button class="cte-icon-btn" @click="openPanel('create-sub', { moduloId: mod.id })" title="Agregar submódulo">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
              <span>Submódulo</span>
            </button>
            <button class="cte-icon-btn" @click="openPanel('create-lesson', { moduloId: mod.id })" title="Agregar lección">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
              <span>Lección</span>
            </button>
            <button class="cte-icon-btn cte-icon-edit" @click="openPanel('edit-module', { moduloId: mod.id }, { title: mod.title, description: mod.description })" title="Editar">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/></svg>
            </button>
            <button class="cte-icon-btn cte-icon-del" @click="deleteModule(mod.id)" title="Eliminar">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
            </button>
          </div>
        </div>

        <!-- Contenido del módulo (colapsable) -->
        <div v-show="!collapsed.has(mod.id)" class="cte-module-body">

          <!-- ── SUBMÓDULOS ────────────────────────────────── -->
          <div
            v-for="(sub, si) in mod.submodulos"
            :key="sub.id"
            class="cte-sub"
          >
            <div class="cte-sub-header">
              <div class="cte-reorder cte-reorder-sm">
                <button class="cte-arr-sm" :disabled="si === 0" @click="moveSub(mod.submodulos, mod.id, si, -1)">
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M18 15l-6-6-6 6"/></svg>
                </button>
                <button class="cte-arr-sm" :disabled="si === mod.submodulos.length - 1" @click="moveSub(mod.submodulos, mod.id, si, 1)">
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9l6 6 6-6"/></svg>
                </button>
              </div>
              <button class="cte-collapse-btn" @click="toggleCollapse(sub.id)">
                <span class="cte-chevron" :class="{ open: !collapsed.has(sub.id) }">›</span>
              </button>
              <div class="cte-sub-info">
                <span class="cte-sub-label">Submódulo {{ Number(si) + 1 }}</span>
                <span class="cte-sub-title">{{ sub.title }}</span>
              </div>
              <div class="cte-module-actions">
                <button class="cte-icon-btn" @click="openPanel('create-lesson', { moduloId: mod.id, submoduloId: sub.id })" title="Agregar lección">
                  <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
                  <span>Lección</span>
                </button>
                <button class="cte-icon-btn cte-icon-edit" @click="openPanel('edit-sub', { moduloId: mod.id, submoduloId: sub.id }, { title: sub.title, description: sub.description })" title="Editar">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/></svg>
                </button>
                <button class="cte-icon-btn cte-icon-del" @click="deleteSub(mod.id, sub.id)" title="Eliminar">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
                </button>
              </div>
            </div>

            <!-- Lecciones del submódulo -->
            <div v-show="!collapsed.has(sub.id)" class="cte-lecciones">
              <div
                v-for="(lec, li) in sub.lecciones"
                :key="lec.id"
                class="cte-lesson"
              >
                <div class="cte-lesson-reorder">
                  <button class="cte-arr-sm" :disabled="li === 0" @click="moveLeccion(sub.lecciones, capId, li, -1)">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M18 15l-6-6-6 6"/></svg>
                  </button>
                  <button class="cte-arr-sm" :disabled="li === sub.lecciones.length - 1" @click="moveLeccion(sub.lecciones, capId, li, 1)">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9l6 6 6-6"/></svg>
                  </button>
                </div>
                <span class="cte-lesson-icon" v-html="lessonIcon(lec.lesson_type)"></span>
                <div class="cte-lesson-info">
                  <span class="cte-lesson-title">{{ lec.title }}</span>
                  <span class="cte-lesson-meta">{{ lessonTypeName(lec.lesson_type) }} · {{ lec.duracion_min }}min<span v-if="lec.points_reward" class="cte-pts"> · <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1" style="display:inline;vertical-align:middle"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg> {{ lec.points_reward }}pts</span></span>
                </div>
                <div class="cte-lesson-actions">
                  <button class="cte-icon-btn cte-icon-edit" @click="openPanel('edit-lesson', { moduloId: mod.id, submoduloId: sub.id, leccionId: lec.id }, { ...lec, lesson_type: String(lec.lesson_type ?? '1') })" title="Editar">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/></svg>
                  </button>
                  <button class="cte-icon-btn cte-icon-del" @click="deleteLesson(lec.id)" title="Eliminar">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
                  </button>
                </div>
              </div>
              <div v-if="!sub.lecciones?.length" class="cte-empty-sub">No hay lecciones — agrega una arriba ↑</div>
            </div>
          </div>

          <!-- Lecciones directas del módulo (sin submódulo) -->
          <div
            v-for="(lec, li) in mod.lecciones"
            :key="lec.id"
            class="cte-lesson cte-lesson-direct"
          >
            <div class="cte-lesson-reorder">
              <button class="cte-arr-sm" :disabled="li === 0" @click="moveLeccion(mod.lecciones, capId, li, -1)">
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M18 15l-6-6-6 6"/></svg>
              </button>
              <button class="cte-arr-sm" :disabled="li === mod.lecciones.length - 1" @click="moveLeccion(mod.lecciones, capId, li, 1)">
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9l6 6 6-6"/></svg>
              </button>
            </div>
            <span class="cte-lesson-icon" v-html="lessonIcon(lec.lesson_type)"></span>
            <div class="cte-lesson-info">
              <span class="cte-lesson-title">{{ lec.title }}</span>
              <span class="cte-lesson-meta">{{ lessonTypeName(lec.lesson_type) }} · {{ lec.duracion_min }}min<span v-if="lec.points_reward" class="cte-pts"> · <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1" style="display:inline;vertical-align:middle"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg> {{ lec.points_reward }}pts</span></span>
            </div>
            <div class="cte-lesson-actions">
              <button class="cte-icon-btn cte-icon-edit" @click="openPanel('edit-lesson', { moduloId: mod.id, leccionId: lec.id }, { ...lec, lesson_type: String(lec.lesson_type ?? '1') })" title="Editar">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/></svg>
              </button>
              <button class="cte-icon-btn cte-icon-del" @click="deleteLesson(lec.id)" title="Eliminar">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
              </button>
            </div>
          </div>

          <div v-if="!mod.submodulos?.length && !mod.lecciones?.length" class="cte-empty">
            Módulo vacío — usa los botones de arriba para agregar contenido
          </div>
        </div>
      </div>

      <!-- ── LECCIONES SUELTAS (sin módulo) ─────────────── -->
      <div v-if="tree.lecciones?.length" class="cte-module cte-loose">
        <div class="cte-module-header">
          <div class="cte-module-info">
            <span class="cte-module-label">Lecciones sueltas</span>
            <span class="cte-module-title">Sin módulo asignado</span>
          </div>
        </div>
        <div class="cte-module-body">
          <div v-for="(lec, li) in tree.lecciones" :key="lec.id" class="cte-lesson">
            <div class="cte-lesson-reorder">
              <button class="cte-arr-sm" :disabled="li === 0" @click="moveLeccion(tree.lecciones, capId, li, -1)">
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M18 15l-6-6-6 6"/></svg>
              </button>
              <button class="cte-arr-sm" :disabled="li === tree.lecciones.length - 1" @click="moveLeccion(tree.lecciones, capId, li, 1)">
                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9l6 6 6-6"/></svg>
              </button>
            </div>
            <span class="cte-lesson-icon" v-html="lessonIcon(lec.lesson_type)"></span>
            <div class="cte-lesson-info">
              <span class="cte-lesson-title">{{ lec.title }}</span>
              <span class="cte-lesson-meta">{{ lessonTypeName(lec.lesson_type) }} · {{ lec.duracion_min }}min</span>
            </div>
            <div class="cte-lesson-actions">
              <button class="cte-icon-btn cte-icon-edit" @click="openPanel('edit-lesson', { leccionId: lec.id }, { ...lec, lesson_type: String(lec.lesson_type ?? '1') })" title="Editar">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/></svg>
              </button>
              <button class="cte-icon-btn cte-icon-del" @click="deleteLesson(lec.id)" title="Eliminar">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="!tree.modulos?.length && !tree.lecciones?.length" class="cte-empty-root">
        <div class="cte-empty-icon">
          <svg width="32" height="32" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" style="color:var(--brand);margin:0 auto 10px"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
        </div>
        <p>Este curso está vacío</p>
        <p>Agrega un módulo o una lección suelta para comenzar</p>
      </div>
    </div>

    <!-- ════════════════════════════════════════════════════════
         PANEL LATERAL (drawer) — Crear / Editar
         ════════════════════════════════════════════════════════ -->
    <Teleport to="body">
      <Transition name="cte-panel">
        <div v-if="panelMode" class="cte-drawer-overlay" @click.self="closePanel">
          <aside class="cte-drawer">
            <div class="cte-drawer-header">
              <h3 class="cte-drawer-title">{{ panelTitle }}</h3>
              <button class="cte-drawer-close" @click="closePanel">✕</button>
            </div>

            <div class="cte-drawer-body">

              <!-- ── Módulo / Submódulo ── -->
              <template v-if="['create-module','edit-module','create-sub','edit-sub'].includes(panelMode!)">
                <div class="cte-field">
                  <label class="cte-label">Título *</label>
                  <input v-model="panelForm.title" class="cte-input" autofocus :placeholder="panelMode?.includes('module') ? 'Ej: Fundamentos de redes' : 'Ej: Conceptos básicos'" />
                </div>
                <div class="cte-field">
                  <label class="cte-label">Descripción</label>
                  <textarea v-model="panelForm.description" class="cte-input" rows="3" placeholder="Descripción opcional del contenido" />
                </div>
              </template>

              <!-- ── Lección ── -->
              <template v-if="['create-lesson','edit-lesson'].includes(panelMode!)">
                <div class="cte-field">
                  <label class="cte-label">Título *</label>
                  <input v-model="panelForm.title" class="cte-input" autofocus placeholder="Título de la lección" />
                </div>
                <div class="cte-field">
                  <label class="cte-label">Descripción</label>
                  <textarea v-model="panelForm.description" class="cte-input" rows="2" placeholder="Descripción opcional" />
                </div>
                <div class="cte-field">
                  <label class="cte-label">Tipo de lección</label>
                  <ContentTypeSelector v-model="panelForm.lesson_type" />
                </div>

                <!-- Archivo (video/pdf) -->
                <div class="cte-field" v-if="panelForm.lesson_type === '1' || panelForm.lesson_type === '3'">
                  <label class="cte-label">{{ panelForm.lesson_type === '1' ? 'Video' : 'Documento PDF' }}
                    <span v-if="panelForm.file_path" class="cte-existing-file">📎 Archivo actual guardado</span>
                  </label>
                  <DragDropUpload v-model="panelFile" />
                </div>

                <!-- Contenido texto -->
                <div class="cte-field" v-if="panelForm.lesson_type === '2'">
                  <label class="cte-label">Contenido</label>
                  <textarea v-model="panelForm.content" class="cte-input cte-textarea-tall" rows="6" placeholder="Escribe el contenido de la lección…" />
                </div>

                <!-- Duración -->
                <div class="cte-field">
                  <label class="cte-label">Duración estimada (min)</label>
                  <input type="number" v-model.number="panelForm.duracion_min" class="cte-input-sm" min="0" />
                </div>

                <!-- Editor de minijuego -->
                <GameConfigEditor
                  v-if="isGameType"
                  v-model="panelForm.game_config_json"
                  :lesson-type="panelForm.lesson_type"
                  v-model:pointsReward="panelForm.points_reward"
                />
              </template>

            </div>

            <div class="cte-drawer-footer">
              <button class="cte-btn cte-btn-ghost" @click="closePanel">Cancelar</button>
              <button
                class="cte-btn cte-btn-primary"
                :disabled="panelSaving"
                @click="['create-module','edit-module'].includes(panelMode!) ? saveModule() : ['create-sub','edit-sub'].includes(panelMode!) ? saveSub() : saveLesson()"
              >
                <span v-if="panelSaving" class="cte-spinner-sm" />
                {{ panelSaving ? 'Guardando…' : 'Guardar' }}
              </button>
            </div>
          </aside>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.cte-root { display: flex; flex-direction: column; gap: 0; position: relative; }

/* Toolbar */
.cte-toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.cte-toolbar-title { font-size: 1.05rem; font-weight: 700; color: var(--dark); margin: 0; }
.cte-toolbar-actions { display: flex; gap: 8px; }

/* Tree */
.cte-tree { display: flex; flex-direction: column; gap: 8px; }

/* Module */
.cte-module {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  overflow: hidden;
  transition: box-shadow 0.2s;
}
.cte-module:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.06); }
.cte-module.cte-loose { border-style: dashed; opacity: 0.9; }

.cte-module-header {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 16px; background: var(--surface-soft);
  border-bottom: 1px solid var(--border);
}
.cte-module-info { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.cte-module-label { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: var(--brand); }
.cte-module-title { font-size: 0.98rem; font-weight: 700; color: var(--dark); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cte-module-actions { display: flex; gap: 4px; align-items: center; }
.cte-module-body { padding: 10px 12px; display: flex; flex-direction: column; gap: 6px; }

/* Sub */
.cte-sub { background: color-mix(in srgb, var(--brand) 3%, var(--surface)); border: 1px solid var(--border-light); border-radius: var(--r-md); overflow: hidden; margin-left: 24px; }
.cte-sub-header { display: flex; align-items: center; gap: 8px; padding: 9px 12px; border-bottom: 1px solid var(--border-light); }
.cte-sub-info { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.cte-sub-label { font-size: 0.65rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; color: var(--muted); }
.cte-sub-title { font-size: 0.9rem; font-weight: 600; color: var(--dark); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* Lessons */
.cte-lecciones { padding: 6px 8px; display: flex; flex-direction: column; gap: 4px; }
.cte-lesson {
  display: flex; align-items: center; gap: 8px;
  padding: 9px 12px; border-radius: var(--r-sm);
  background: var(--surface); border: 1px solid var(--border-light);
  transition: background 0.15s;
}
.cte-lesson.cte-lesson-direct { margin-left: 0; }
.cte-lesson:hover { background: var(--surface-soft); }
.cte-lesson-icon { font-size: 1.1rem; flex-shrink: 0; }
.cte-lesson-info { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.cte-lesson-title { font-size: 0.88rem; font-weight: 600; color: var(--dark); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cte-lesson-meta { font-size: 0.72rem; color: var(--muted); }
.cte-pts { color: var(--brand); font-weight: 600; }
.cte-lesson-actions { display: flex; gap: 4px; flex-shrink: 0; }
.cte-lesson-reorder { display: flex; flex-direction: column; gap: 2px; flex-shrink: 0; }

/* Reorder buttons */
.cte-reorder { display: flex; flex-direction: column; gap: 2px; flex-shrink: 0; }
.cte-arr, .cte-arr-sm {
  background: none; border: 1px solid var(--border); border-radius: 3px;
  cursor: pointer; font-size: 0.6rem; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s; color: var(--muted);
}
.cte-arr { width: 20px; height: 20px; }
.cte-arr-sm { width: 16px; height: 16px; font-size: 0.55rem; }
.cte-arr:hover:not(:disabled), .cte-arr-sm:hover:not(:disabled) { background: var(--brand); color: white; border-color: var(--brand); }
.cte-arr:disabled, .cte-arr-sm:disabled { opacity: 0.25; cursor: not-allowed; }

/* Collapse */
.cte-collapse-btn { background: none; border: none; cursor: pointer; padding: 2px; color: var(--muted); flex-shrink: 0; }
.cte-chevron { display: inline-block; font-size: 1.1rem; transition: transform 0.2s; transform: rotate(0deg); }
.cte-chevron.open { transform: rotate(90deg); }

/* Icon buttons */
.cte-icon-btn {
  display: flex; align-items: center; gap: 4px;
  background: none; border: 1px solid var(--border); border-radius: var(--r-sm);
  padding: 4px 8px; font-size: 0.75rem; cursor: pointer; color: var(--muted);
  transition: all 0.15s; white-space: nowrap;
}
.cte-icon-btn:hover { background: var(--surface-soft); color: var(--dark); border-color: var(--border-dark); }
.cte-icon-edit:hover { border-color: var(--brand); color: var(--brand); background: var(--brand-light); }
.cte-icon-del:hover { border-color: var(--danger); color: var(--danger); background: var(--danger-bg); }

/* Empties */
.cte-empty { padding: 16px; text-align: center; font-size: 0.82rem; color: var(--muted); }
.cte-empty-sub { padding: 10px 12px; font-size: 0.78rem; color: var(--muted); text-align: center; }
.cte-empty-root { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 60px 20px; color: var(--muted); text-align: center; }
.cte-empty-icon { font-size: 3rem; }

/* Loading */
.cte-loading { display: flex; align-items: center; gap: 10px; padding: 30px; justify-content: center; color: var(--muted); }
.cte-spinner { width: 20px; height: 20px; border: 2px solid var(--border); border-top-color: var(--brand); border-radius: 50%; animation: spin 0.8s linear infinite; }
.cte-spinner-sm { width: 14px; height: 14px; border: 2px solid rgba(255,255,255,0.3); border-top-color: white; border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Buttons */
.cte-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: var(--r-md); font-size: 0.88rem;
  font-weight: 600; cursor: pointer; transition: all 0.15s; border: none;
}
.cte-btn-primary { background: var(--brand); color: white; }
.cte-btn-primary:hover:not(:disabled) { background: var(--brand-dark); }
.cte-btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.cte-btn-outline { background: transparent; border: 1.5px solid var(--border); color: var(--dark); }
.cte-btn-outline:hover { border-color: var(--brand); color: var(--brand); background: var(--brand-light); }
.cte-btn-ghost { background: transparent; color: var(--muted); border: 1px solid var(--border); }
.cte-btn-ghost:hover { color: var(--dark); }

/* Drawer */
.cte-drawer-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.45);
  backdrop-filter: blur(4px); z-index: 1000;
  display: flex; justify-content: flex-end;
}
.cte-drawer {
  width: min(600px, 95vw);
  background: var(--surface);
  display: flex; flex-direction: column;
  height: 100%; overflow: hidden;
  box-shadow: -8px 0 40px rgba(0,0,0,0.18);
}
.cte-drawer-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px; border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.cte-drawer-title { font-size: 1.15rem; font-weight: 700; color: var(--dark); margin: 0; }
.cte-drawer-close { background: none; border: none; font-size: 1.2rem; cursor: pointer; color: var(--muted); border-radius: var(--r-sm); padding: 4px 8px; transition: all 0.15s; }
.cte-drawer-close:hover { background: var(--surface-soft); color: var(--dark); }
.cte-drawer-body { flex: 1; overflow-y: auto; padding: 24px; display: flex; flex-direction: column; gap: 20px; }
.cte-drawer-footer { padding: 16px 24px; border-top: 1px solid var(--border); display: flex; gap: 10px; justify-content: flex-end; flex-shrink: 0; }

/* Form fields inside drawer */
.cte-field { display: flex; flex-direction: column; gap: 6px; }
.cte-label { font-size: 0.85rem; font-weight: 700; color: var(--dark); }
.cte-input {
  width: 100%; padding: 10px 12px;
  background: var(--surface-soft); border: 1.5px solid var(--border);
  border-radius: var(--r-md); font-size: 0.9rem; color: var(--dark);
  transition: border-color 0.2s; resize: vertical;
}
.cte-input:focus { outline: none; border-color: var(--brand); }
.cte-input-sm { width: 120px; padding: 8px 10px; background: var(--surface-soft); border: 1.5px solid var(--border); border-radius: var(--r-md); font-size: 0.9rem; color: var(--dark); }
.cte-textarea-tall { min-height: 140px; }
.cte-existing-file { font-size: 0.75rem; color: var(--brand); font-weight: 400; margin-left: 8px; }

/* Panel transition */
.cte-panel-enter-active, .cte-panel-leave-active { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
.cte-panel-enter-from, .cte-panel-leave-to { opacity: 0; }
.cte-panel-enter-from .cte-drawer, .cte-panel-leave-to .cte-drawer { transform: translateX(100%); }
</style>
