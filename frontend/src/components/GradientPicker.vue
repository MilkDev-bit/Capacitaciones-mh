<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const gradients = [
  'linear-gradient(135deg, #f97316 0%, #dc2626 100%)', 
  'linear-gradient(135deg, #3b82f6 0%, #6366f1 100%)', 
  'linear-gradient(135deg, #10b981 0%, #059669 100%)', 
  'linear-gradient(135deg, #0f766e 0%, #2563eb 100%)', 
  'linear-gradient(135deg, #8b5cf6 0%, #6d28d9 100%)', 
  'linear-gradient(135deg, #ec4899 0%, #be185d 100%)', 
  'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)', 
  'linear-gradient(135deg, #14b8a6 0%, #0f766e 100%)', 
  'linear-gradient(135deg, #1e293b 0%, #0f172a 100%)',
]

function selectGradient(g: string) {
  emit('update:modelValue', g)
}
</script>

<template>
  <div class="gp-container">
    <div class="gp-grid">
      <button 
        v-for="(g, i) in gradients" 
        :key="i"
        type="button"
        class="gp-btn"
        :class="{ active: modelValue === g }"
        :style="{ background: g }"
        @click="selectGradient(g)"
        :aria-label="`Seleccionar color de portada ${i + 1}`"
      >
        <div v-if="modelValue === g" class="gp-check">
          <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </div>
      </button>
    </div>
  </div>
</template>

<style scoped>
.gp-container {
  padding: 10px 0;
}
.gp-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.gp-btn {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  position: relative;
  transition: transform 0.2s, box-shadow 0.2s, border-color 0.2s;
  padding: 0;
  outline: none;
}
.gp-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}
.gp-btn:focus-visible {
  outline: 2px solid var(--brand);
  outline-offset: 4px;
}
.gp-btn.active {
  border-color: #fff;
  box-shadow: 0 0 0 3px var(--brand);
  transform: scale(1.05);
}
.gp-check {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  filter: drop-shadow(0 1px 2px rgba(0,0,0,0.5));
}
</style>
