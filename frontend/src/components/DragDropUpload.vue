<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  accept?: string
  modelValue: File | null
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: File | null): void
}>()

const isDragging = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

function triggerFileInput() {
  if (!props.disabled) fileInput.value?.click()
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    emit('update:modelValue', input.files[0] || null)
  }
}

function handleDragOver(e: DragEvent) {
  e.preventDefault()
  if (!props.disabled) isDragging.value = true
}

function handleDragLeave(e: DragEvent) {
  e.preventDefault()
  isDragging.value = false
}

function handleDrop(e: DragEvent) {
  e.preventDefault()
  isDragging.value = false
  if (props.disabled) return
  
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    const file = e.dataTransfer.files[0]
    if (!file) return
    
    if (props.accept) {
      const allowedTypes = props.accept.split(',').map(t => t.trim().toLowerCase())
      const fileExt = '.' + (file.name.split('.').pop()?.toLowerCase() || '')
      const fileType = file.type.toLowerCase()
      
      const isValid = allowedTypes.some(type => {
        if (type.startsWith('.')) return fileExt === type
        if (type.endsWith('/*')) return fileType.startsWith(type.replace('/*', ''))
        return fileType === type
      })
      
      if (!isValid) {
        alert('Tipo de archivo no permitido.')
        return
      }
    }
    emit('update:modelValue', file)
  }
}

function clearFile(e: Event) {
  e.stopPropagation()
  emit('update:modelValue', null)
  if (fileInput.value) fileInput.value.value = ''
}

const formatSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
</script>

<template>
  <div 
    class="dd-zone" 
    :class="{ 'dd-dragging': isDragging, 'dd-has-file': modelValue, 'dd-disabled': disabled }"
    @dragover="handleDragOver"
    @dragleave="handleDragLeave"
    @drop="handleDrop"
    @click="triggerFileInput"
  >
    <input 
      type="file" 
      ref="fileInput" 
      class="dd-input" 
      :accept="accept" 
      @change="handleFileChange" 
      :disabled="disabled"
    />
    
    <div v-if="!modelValue" class="dd-empty">
      <div class="dd-icon">
        <svg width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5"/></svg>
      </div>
      <div class="dd-text">
        <span class="dd-primary">Haz clic o arrastra un archivo aquí</span>
        <span class="dd-secondary">Formatos soportados: {{ accept || 'Cualquiera' }}</span>
        <span class="dd-secondary">Tamaño máximo: 50MB</span>
      </div>
    </div>
    
    <div v-else class="dd-filled">
      <div class="dd-file-icon">
        <svg v-if="modelValue.type.startsWith('video/')" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14v-4z"/><rect x="3" y="6" width="12" height="12" rx="2"/></svg>
        <svg v-else-if="modelValue.type.startsWith('image/')" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
        <svg v-else width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
      </div>
      <div class="dd-file-info">
        <span class="dd-file-name" :title="modelValue.name">{{ modelValue.name }}</span>
        <span class="dd-file-size">{{ formatSize(modelValue.size) }}</span>
      </div>
      <button type="button" class="dd-remove" @click.stop="clearFile" title="Quitar archivo">
        <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.dd-zone {
  position: relative;
  width: 100%;
  min-height: 120px;
  border: 2px dashed var(--border);
  border-radius: var(--r);
  background: var(--surface-soft);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s ease;
  overflow: hidden;
}

.dd-zone:hover:not(.dd-disabled) {
  border-color: var(--brand);
  background: var(--brand-light);
}

.dd-dragging {
  border-color: var(--brand);
  background: var(--brand-light);
  transform: scale(1.02);
}

.dd-has-file {
  border-style: solid;
  border-color: var(--border);
  background: var(--surface);
}

.dd-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.dd-input {
  display: none;
}

.dd-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  text-align: center;
}

.dd-icon {
  color: var(--brand);
  background: #fff;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(249,115,22,0.15);
}

.dd-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.dd-primary {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--dark);
}

.dd-secondary {
  font-size: 0.75rem;
  color: var(--muted);
}

.dd-filled {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--bg);
  padding: 12px 16px;
  border-radius: var(--r-sm);
  border: 1px solid var(--border-light);
}

.dd-file-icon {
  color: var(--brand);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface);
  padding: 10px;
  border-radius: 10px;
  box-shadow: var(--shadow-xs);
}

.dd-file-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dd-file-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--dark);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dd-file-size {
  font-size: 0.75rem;
  color: var(--muted);
}

.dd-remove {
  background: none;
  border: none;
  color: var(--muted);
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.dd-remove:hover {
  background: var(--danger-bg);
  color: var(--danger);
}
</style>
