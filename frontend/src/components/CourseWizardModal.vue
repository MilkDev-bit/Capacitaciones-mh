<script setup lang="ts">
import { ref } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'
import { uploadToR2 } from '../utils/upload'
import DragDropUpload from './DragDropUpload.vue'
import GradientPicker from './GradientPicker.vue'

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', id: string): void
}>()

const step = ref(1)
const loading = ref(false)

const form = ref({
  title: '',
  description: '',
  type: 'course',
  precio: 0,
  is_public: false,
  dc3_enabled: true,
  welcome_message: '',
  color: 'linear-gradient(135deg, #f97316 0%, #dc2626 100%)',
})

const thumbnailFile = ref<File | null>(null)

function nextStep() {
  if (step.value === 1 && !form.value.title) {
    toast.error('El título es requerido')
    return
  }
  step.value++
}

function prevStep() {
  step.value--
}

async function guardar() {
  if (!form.value.title) { toast.error('El titulo es requerido'); return }
  loading.value = true
  const loadingToast = thumbnailFile.value
    ? toast.loading('Subiendo portada...')
    : null
  try {
    const payload: Record<string, any> = {
      title: form.value.title,
      description: form.value.description,
      type: 'course',
      is_public: form.value.is_public,
      dc3_enabled: form.value.dc3_enabled === true,
      welcome_message: form.value.welcome_message,
      color: form.value.color,
      precio: Number(form.value.precio) || 0,
    }
    if (thumbnailFile.value) {
      payload.thumbnail_url = await uploadToR2(thumbnailFile.value, 'thumbnails')
    }

    const res = await api.post('/instructor/capacitaciones', payload)
    toast.success('Curso creado')
    emit('created', res.data.id)
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al guardar')
  } finally {
    loading.value = false
    loadingToast?.close()
  }
}
</script>

<template>
  <div class="modal-overlay" @mousedown.self="emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h2 class="modal-title">Crear nuevo curso</h2>
        <button class="modal-close" @click="emit('close')">
          <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>
        </button>
      </div>

      <div class="stepper">
        <div class="step" :class="{ active: step >= 1 }"><div class="step-num">1</div><div class="step-text">Información</div></div>
        <div class="step-line" :class="{ active: step >= 2 }"></div>
        <div class="step" :class="{ active: step >= 2 }"><div class="step-num">2</div><div class="step-text">Diseño</div></div>
        <div class="step-line" :class="{ active: step >= 3 }"></div>
        <div class="step" :class="{ active: step >= 3 }"><div class="step-num">3</div><div class="step-text">Publicar</div></div>
      </div>

      <div class="modal-body">
        <div v-if="step === 1" class="step-pane slide-down-enter-active">
          <div class="field">
            <label>Título del curso <span class="req">*</span></label>
            <input class="field-input" v-model="form.title" placeholder="Ej: Introducción a Vue 3" autofocus />
            <span class="hint">{{ form.title.length }}/100 caracteres</span>
          </div>
          <div class="field mt-4">
            <label>Descripción</label>
            <textarea class="field-input" v-model="form.description" rows="4" placeholder="¿Qué aprenderán los estudiantes?"></textarea>
          </div>
          <div class="field mt-4">
            <label>Precio Individual (MXN)</label>
            <input type="number" class="field-input" v-model="form.precio" placeholder="Ej: 500.00 (0 para gratis)" min="0" step="0.01" />
          </div>
        </div>

        <div v-if="step === 2" class="step-pane slide-down-enter-active">
          <div class="field">
            <label>Color de portada (si no subes imagen)</label>
            <GradientPicker v-model="form.color" />
          </div>
          <div class="field mt-4">
            <label>Imagen de portada (Opcional)</label>
            <DragDropUpload v-model="thumbnailFile" accept="image/*" />
          </div>
          <div class="field mt-4">
            <label>Mensaje de bienvenida</label>
            <textarea class="field-input" v-model="form.welcome_message" rows="2" placeholder="Mensaje para los nuevos estudiantes"></textarea>
          </div>
        </div>

        <div v-if="step === 3" class="step-pane text-center slide-down-enter-active">
          <div class="preview-card" :style="{ background: thumbnailFile ? 'var(--surface-soft)' : form.color }">
            <div v-if="thumbnailFile" class="preview-img-wrap">
              <span class="preview-text">Imagen seleccionada</span>
            </div>
            <div class="preview-overlay">
              <h3>{{ form.title || 'Sin título' }}</h3>
              <span class="badge badge-gray" style="background:rgba(255,255,255,0.2);color:#fff">Curso Completo</span>
            </div>
          </div>
          <div class="field mt-6 text-left">
            <label class="toggle-wrap">
              <input type="checkbox" v-model="form.is_public" class="toggle-input">
              <div class="toggle-slider"></div>
              <div class="toggle-text">
                <strong>Curso Público</strong>
                <p>Cualquiera podrá ver e inscribirse a este curso.</p>
              </div>
            </label>
          </div>
          <div class="field mt-4 text-left">
            <label class="toggle-wrap">
              <input type="checkbox" v-model="form.dc3_enabled" class="toggle-input">
              <div class="toggle-slider"></div>
              <div class="toggle-text">
                <strong>Habilitar Constancia DC-3</strong>
                <p>Permite a los estudiantes tramitar y obtener su constancia DC-3 en este curso.</p>
              </div>
            </label>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button v-if="step > 1" class="btn btn-secondary" @click="prevStep" :disabled="loading">Atrás</button>
        <div class="spacer"></div>
        <button v-if="step < 3" class="btn btn-primary" @click="nextStep">Siguiente</button>
        <button v-else class="btn btn-primary" @click="guardar" :disabled="loading">
          <span v-if="loading" class="spinner-small"></span>
          {{ loading ? 'Publicando...' : 'Publicar curso' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center; z-index: 1000; padding: 20px;
}
.modal-content {
  background: var(--surface); width: 100%; max-width: 600px;
  border-radius: var(--r-xl); box-shadow: var(--shadow-lg);
  display: flex; flex-direction: column; max-height: 90vh;
}
.modal-header {
  padding: 20px 24px; border-bottom: 1px solid var(--border-light);
  display: flex; align-items: center; justify-content: space-between;
}
.modal-title { font-size: 1.25rem; font-weight: 800; margin: 0; }
.modal-close {
  background: none; border: none; color: var(--muted); cursor: pointer;
  padding: 4px; border-radius: 50%; transition: background 0.2s;
}
.modal-close:hover { background: var(--surface-soft); color: var(--dark); }

.stepper {
  display: flex; align-items: center; justify-content: space-between;
  padding: 24px 32px 12px;
}
.step {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  color: var(--muted); position: relative; z-index: 2;
}
.step.active { color: var(--brand); }
.step-num {
  width: 32px; height: 32px; border-radius: 50%; background: var(--surface);
  border: 2px solid var(--border); display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 0.9rem; transition: all 0.3s;
}
.step.active .step-num {
  border-color: var(--brand); background: var(--brand); color: #fff;
}
.step-text { font-size: 0.75rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }
.step-line {
  flex: 1; height: 2px; background: var(--border); margin: 0 12px;
  transform: translateY(-10px); transition: background 0.3s;
}
.step-line.active { background: var(--brand); }

.modal-body {
  padding: 24px 32px; overflow-y: auto; flex: 1;
}
.field { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: 0.85rem; font-weight: 600; color: var(--dark); }
.req { color: var(--danger); }
.hint { font-size: 0.75rem; color: var(--muted); align-self: flex-end; }
.mt-4 { margin-top: 16px; }
.mt-6 { margin-top: 24px; }
.text-center { text-align: center; }
.text-left { text-align: left; }

.preview-card {
  width: 100%; height: 200px; border-radius: var(--r-lg); overflow: hidden;
  position: relative; display: flex; align-items: flex-end; padding: 20px;
  box-shadow: var(--shadow-sm);
}
.preview-img-wrap {
  position: absolute; inset: 0; background: var(--surface-soft);
  display: flex; align-items: center; justify-content: center;
}
.preview-text { font-weight: 700; color: var(--muted); }
.preview-overlay {
  position: relative; z-index: 1; color: #fff; text-align: left;
  text-shadow: 0 2px 4px rgba(0,0,0,0.5);
}
.preview-overlay h3 { color: #fff; font-size: 1.5rem; margin: 0 0 8px; }

.toggle-wrap {
  display: flex; align-items: flex-start; gap: 12px; cursor: pointer;
}
.toggle-input { display: none; }
.toggle-slider {
  width: 44px; height: 24px; background: var(--border); border-radius: 12px;
  position: relative; transition: background 0.2s; flex-shrink: 0; margin-top: 2px;
}
.toggle-slider::after {
  content: ''; position: absolute; top: 2px; left: 2px;
  width: 20px; height: 20px; background: #fff; border-radius: 50%;
  transition: transform 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}
.toggle-input:checked + .toggle-slider { background: var(--success); }
.toggle-input:checked + .toggle-slider::after { transform: translateX(20px); }
.toggle-text strong { display: block; font-size: 0.95rem; color: var(--dark); }
.toggle-text p { font-size: 0.85rem; color: var(--muted); margin: 0; }

.modal-footer {
  padding: 20px 24px; border-top: 1px solid var(--border-light);
  display: flex; align-items: center;
}
.spacer { flex: 1; }
.spinner-small {
  width: 16px; height: 16px; border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%; animation: spin 0.8s linear infinite;
  display: inline-block;
}
</style>
