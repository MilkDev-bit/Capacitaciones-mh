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
    // dc3_empresa llega anidado desde el backend; se aplana al formulario para
    // poder usar v-model directo en cada campo.
    const emp = props.course.dc3_empresa || {}
    form.value = {
      ...props.course,
      dc3_enabled: props.course.dc3_enabled === true,
      dc3_razon_social: emp.razon_social || '',
      dc3_rfc: emp.rfc || '',
      dc3_nombre_patron: emp.nombre_patron || '',
      dc3_representante_trabajadores: emp.representante_trabajadores || '',
      dc3_area_tematica: emp.area_tematica || '',
      dc3_nombre_capacitador: emp.nombre_capacitador || '',
    }
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
      dc3_enabled: form.value.dc3_enabled === true,
      dc3_empresa: {
        razon_social: form.value.dc3_razon_social || '',
        rfc: form.value.dc3_rfc || '',
        nombre_patron: form.value.dc3_nombre_patron || '',
        representante_trabajadores: form.value.dc3_representante_trabajadores || '',
        area_tematica: form.value.dc3_area_tematica || '',
        nombre_capacitador: form.value.dc3_nombre_capacitador || '',
      },
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
                <p>La constancia se emite sola cuando el alumno termina el curso</p>
              </div>
            </label>
          </div>

          <!--
            Datos de empresa y agentes capacitadores.
            Se capturan una vez por capacitación y se repiten en todas las
            constancias que emita. Sin ellos la emisión queda bloqueada: el
            alumno puede poner su CURP, pero no puede inventar al patrón que
            firma su constancia.
          -->
          <div v-if="form.dc3_enabled" class="dc3-empresa mt-4">
            <div class="dc3-empresa-head">
              <strong>Datos para la constancia</strong>
              <p>Obligatorios para emitir. Se aplican a todos los alumnos de esta capacitación.</p>
            </div>

            <div class="dc3-grid">
              <label class="field">
                <span class="field-label">Razón social <em>*</em></span>
                <input v-model="form.dc3_razon_social" class="field-input"
                       placeholder="Nombre o razón social de la empresa" />
              </label>
              <label class="field">
                <span class="field-label">RFC <em>*</em></span>
                <input v-model="form.dc3_rfc" class="field-input" maxlength="13"
                       placeholder="12 o 13 caracteres" />
              </label>
              <label class="field">
                <span class="field-label">Patrón o representante legal <em>*</em></span>
                <input v-model="form.dc3_nombre_patron" class="field-input" />
              </label>
              <label class="field">
                <span class="field-label">Representante de los trabajadores <em>*</em></span>
                <input v-model="form.dc3_representante_trabajadores" class="field-input" />
              </label>
              <label class="field">
                <span class="field-label">Área temática <em>*</em></span>
                <input v-model="form.dc3_area_tematica" class="field-input"
                       placeholder="Clave del catálogo STPS, ej: 6000" />
              </label>
              <label class="field">
                <span class="field-label">Nombre del capacitador <em>*</em></span>
                <input v-model="form.dc3_nombre_capacitador" class="field-input" />
              </label>
            </div>
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
.dc3-empresa {
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 16px;
  background: var(--bg);
}

.dc3-empresa-head { margin-bottom: 14px; }
.dc3-empresa-head strong { font-size: 0.92rem; color: var(--text); }
.dc3-empresa-head p { margin: 4px 0 0; font-size: 0.8rem; color: var(--muted); }

.dc3-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.dc3-grid .field-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--muted);
  margin-bottom: 5px;
}

.dc3-grid em { color: var(--danger); font-style: normal; }

@media (max-width: 700px) {
  .dc3-grid { grid-template-columns: 1fr; }
}

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
