<script setup lang="ts">
import { ref, watch } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'
import { uploadToR2 } from '../utils/upload'
import DragDropUpload from './DragDropUpload.vue'
import GradientPicker from './GradientPicker.vue'
import CourseTreeEditor from './CourseTreeEditor.vue'

const props = defineProps<{
  show: boolean
  course: any
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'updated'): void
}>()

const activeTab = ref('info')
const loading = ref(false)
const form = ref<any>({})
const thumbnailFile = ref<File | null>(null)

watch(() => props.show, (val) => {
  if (val && props.course) {
    form.value = { dc3_enabled: true, ...props.course }
    thumbnailFile.value = null
    activeTab.value = 'info'
  }
})

async function saveInfo() {
  if (!form.value.title) return toast.error('Título requerido')
  loading.value = true
  try {
    const payload: Record<string, any> = {
      title: form.value.title,
      description: form.value.description || '',
      type: form.value.type || 'course',
      is_public: form.value.is_public,
      dc3_enabled: form.value.dc3_enabled !== false,
      welcome_message: form.value.welcome_message || '',
      color: form.value.color || '#f97316',
      thumbnail_url: form.value.thumbnail_url || '',
      precio: Number(form.value.precio) || 0,
    }

    if (thumbnailFile.value) {
      const uploadingToast = toast.loading('Subiendo portada...')
      try {
        payload.thumbnail_url = await uploadToR2(thumbnailFile.value, 'thumbnails')
      } finally {
        uploadingToast.close()
      }
    }

    await api.put(`/instructor/capacitaciones/${form.value.id}`, payload)
    if (payload.thumbnail_url) form.value.thumbnail_url = payload.thumbnail_url
    toast.success('Curso actualizado')
    emit('updated')
  } catch(e:any) {
    toast.error(e?.response?.data?.error || e?.message || 'Error al actualizar curso')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="drawer-overlay" :class="{ open: show }" @mousedown.self="emit('close')">
    <div class="drawer-content" :class="{ open: show }">
      <div class="drawer-header">
        <div>
          <h2 class="drawer-title">Editor de Curso</h2>
          <p class="drawer-sub">{{ form.title }}</p>
        </div>
        <button class="drawer-close" @click="emit('close')">
          <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M6 18L18 6M6 6l12 12"/></svg>
        </button>
      </div>

      <div class="drawer-tabs">
        <button class="tab" :class="{ active: activeTab === 'info' }" @click="activeTab = 'info'">Información</button>
        <button class="tab" :class="{ active: activeTab === 'lessons' }" @click="activeTab = 'lessons'">Lecciones</button>
      </div>

      <div class="drawer-body">
        <div v-if="activeTab === 'info'" class="tab-pane slide-down-enter-active">
          <div class="field">
            <label>Título del curso</label>
            <input v-model="form.title" class="field-input" />
          </div>
          <div class="field mt-4">
            <label>Descripción</label>
            <textarea v-model="form.description" class="field-input" rows="3"></textarea>
          </div>
          <div class="field mt-4">
            <label>Precio Individual (MXN)</label>
            <input type="number" v-model="form.precio" class="field-input" placeholder="Ej: 500.00" min="0" step="0.01" />
          </div>
          <div class="field mt-4">
            <label>Color de Portada</label>
            <GradientPicker v-model="form.color" />
          </div>
          <div class="field mt-4">
            <label>Imagen de portada (Sobrescribe el color)</label>
            <DragDropUpload v-model="thumbnailFile" accept="image/*" />
          </div>
          
          <div class="field mt-6">
            <label class="toggle-wrap">
              <input type="checkbox" v-model="form.is_public" class="toggle-input">
              <div class="toggle-slider"></div>
              <div class="toggle-text">
                <strong>Curso Público</strong>
                <p>Visible para todos los usuarios</p>
              </div>
            </label>
          </div>

          <div class="field mt-4">
            <label class="toggle-wrap">
              <input type="checkbox" v-model="form.dc3_enabled" class="toggle-input">
              <div class="toggle-slider"></div>
              <div class="toggle-text">
                <strong>Habilitar Constancia DC-3</strong>
                <p>Permite a los estudiantes tramitar y obtener su constancia DC-3 en este curso</p>
              </div>
            </label>
          </div>
          
          <div class="mt-6 text-right">
            <button class="btn btn-primary" @click="saveInfo" :disabled="loading">
              {{ loading ? 'Guardando...' : 'Guardar Cambios' }}
            </button>
          </div>
        </div>

        <div v-if="activeTab === 'lessons'" class="tab-pane slide-down-enter-active">
          <CourseTreeEditor v-if="form.id" :capId="form.id" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.drawer-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); backdrop-filter: blur(2px); z-index: 1000; opacity: 0; pointer-events: none; transition: opacity 0.3s; }
.drawer-overlay.open { opacity: 1; pointer-events: auto; }

.drawer-content { position: absolute; right: 0; top: 0; bottom: 0; width: 100%; max-width: 850px; background: var(--surface); box-shadow: -4px 0 24px rgba(0,0,0,0.1); transform: translateX(100%); transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1); display: flex; flex-direction: column; }
.drawer-content.open { transform: translateX(0); }

.drawer-header { padding: 24px; border-bottom: 1px solid var(--border-light); display: flex; align-items: flex-start; justify-content: space-between; }
.drawer-title { font-size: 1.25rem; font-weight: 800; margin: 0; }
.drawer-sub { font-size: 0.85rem; color: var(--muted); margin: 4px 0 0 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 380px; }
.drawer-close { background: var(--surface-soft); border: none; color: var(--muted); cursor: pointer; width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.drawer-close:hover { background: var(--border); color: var(--dark); }

.drawer-tabs { display: flex; border-bottom: 1px solid var(--border-light); background: var(--surface-soft); }
.tab { flex: 1; padding: 14px; background: none; border: none; font-weight: 600; color: var(--muted); cursor: pointer; border-bottom: 2px solid transparent; transition: all 0.2s; }
.tab:hover { color: var(--dark); }
.tab.active { color: var(--brand); border-bottom-color: var(--brand); background: var(--surface); }

.drawer-body { flex: 1; overflow-y: auto; padding: 24px; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: 0.85rem; font-weight: 600; color: var(--dark); }
.mt-4 { margin-top: 16px; }
.mt-6 { margin-top: 24px; }
.text-right { text-align: right; }

.toggle-wrap { display: flex; align-items: flex-start; gap: 12px; cursor: pointer; }
.toggle-input { display: none; }
.toggle-slider { width: 44px; height: 24px; background: var(--border); border-radius: 12px; position: relative; transition: background 0.2s; flex-shrink: 0; margin-top: 2px; }
.toggle-slider::after { content: ''; position: absolute; top: 2px; left: 2px; width: 20px; height: 20px; background: #fff; border-radius: 50%; transition: transform 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.2); }
.toggle-input:checked + .toggle-slider { background: var(--success); }
.toggle-input:checked + .toggle-slider::after { transform: translateX(20px); }
.toggle-text strong { display: block; font-size: 0.95rem; color: var(--dark); }
.toggle-text p { font-size: 0.85rem; color: var(--muted); margin: 0; }
</style>
