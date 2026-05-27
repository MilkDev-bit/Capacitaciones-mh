<script setup lang="ts">
const props = defineProps<{
  modelValue: string;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

const types = [
  {
    id: "video",
    name: "Video",
    desc: "Clases grabadas, MP4",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><polygon points="10 8 16 12 10 16 10 8" fill="currentColor" stroke="none"/></svg>',
  },
  {
    id: "document",
    name: "Documento",
    desc: "PDF, Word, Guías",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>',
  },
  {
    id: "text",
    name: "Texto",
    desc: "Lecturas, Artículos",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>',
  },
  {
    id: "link",
    name: "Enlace",
    desc: "YouTube, Vimeo, URL externa",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M10 13a5 5 0 007.07 0l2.83-2.83a5 5 0 10-7.07-7.07L11.7 4.22"/><path d="M14 11a5 5 0 01-7.07 0L4.1 8.17a5 5 0 017.07-7.07L12.3 2.22"/></svg>',
  },
];
</script>

<template>
  <div class="cts-grid">
    <label v-for="t in types" :key="t.id" class="cts-card" :class="{ active: modelValue === t.id }">
      <input
        type="radio"
        :value="t.id"
        :checked="modelValue === t.id"
        @change="emit('update:modelValue', t.id)"
        class="cts-radio"
      />
      <div class="cts-icon" v-html="t.icon"></div>
      <div class="cts-info">
        <span class="cts-name">{{ t.name }}</span>
        <span class="cts-desc">{{ t.desc }}</span>
      </div>
      <div class="cts-check" v-if="modelValue === t.id">
        <svg width="16" height="16" fill="none" stroke="#fff" stroke-width="3" viewBox="0 0 24 24">
          <path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </div>
    </label>
  </div>
</template>

<style scoped>
.cts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 16px;
}
.cts-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 20px 14px;
  background: var(--surface);
  border: 2px solid var(--border);
  border-radius: var(--r-lg);
  cursor: pointer;
  transition: all 0.2s ease;
}
.cts-card:hover {
  border-color: var(--brand-border);
  background: var(--brand-light);
}
.cts-card.active {
  border-color: var(--brand);
  background: var(--brand-light);
  box-shadow: 0 4px 14px rgba(249, 115, 22, 0.1);
}
.cts-radio {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}
.cts-icon {
  color: var(--muted);
  margin-bottom: 12px;
  transition: color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cts-card.active .cts-icon {
  color: var(--brand);
}
.cts-name {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--dark);
  margin-bottom: 4px;
}
.cts-desc {
  font-size: 0.8rem;
  color: var(--muted);
  line-height: 1.3;
}
.cts-check {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 24px;
  height: 24px;
  background: var(--brand);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 6px rgba(249, 115, 22, 0.4);
}
</style>
