<script setup lang="ts">
const props = defineProps<{ modelValue: string }>();
const emit = defineEmits<{ (e: "update:modelValue", value: string): void }>();

// lesson_type numérico tal como lo define el proto LessonType enum
// LESSON_TYPE_UNSPECIFIED=0, VIDEO=1, TEXT=2, PDF=3, QUIZ=4,
// GAME_MEMORY=5, GAME_DRAGDROP=6, GAME_WORDSEARCH=7, GAME_FILLBLANK=8, GAME_ORDER=9
const types = [
  {
    id: "1", name: "Video", desc: "Clases grabadas, MP4",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><polygon points="10 8 16 12 10 16 10 8" fill="currentColor" stroke="none"/></svg>',
    group: "Contenido",
  },
  {
    id: "2", name: "Texto", desc: "Lecturas, Artículos",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>',
    group: "Contenido",
  },
  {
    id: "3", name: "PDF", desc: "Documentos, Guías",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>',
    group: "Contenido",
  },
  {
    id: "4", name: "Quiz", desc: "Preguntas de opción múltiple",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 015.83 1c0 2-3 3-3 3M12 17h.01"/></svg>',
    group: "Contenido",
  },
  {
    id: "5", name: "Memorama", desc: "Encontrar pares de tarjetas",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><rect x="2" y="3" width="9" height="7" rx="1"/><rect x="13" y="3" width="9" height="7" rx="1"/><rect x="2" y="14" width="9" height="7" rx="1"/><rect x="13" y="14" width="9" height="7" rx="1"/></svg>',
    group: "Minijuego",
    color: "#f59e0b",
  },
  {
    id: "6", name: "Arrastrar", desc: "Clasificar elementos en categorías",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M5 9l4 4 4-4M5 15l4 4 4-4M17 9l2 2 2-2M17 15l2 2 2-2"/></svg>',
    group: "Minijuego",
    color: "#10b981",
  },
  {
    id: "7", name: "Sopa letras", desc: "Encontrar palabras ocultas",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 7h2M11 7h2M15 7h2M7 11h2M11 11h2M15 11h2M7 15h2M11 15h2M15 15h2"/></svg>',
    group: "Minijuego",
    color: "#3b82f6",
  },
  {
    id: "8", name: "Completar", desc: "Rellenar espacios en blanco",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M3 6h18M3 12h12M3 18h8"/></svg>',
    group: "Minijuego",
    color: "#ec4899",
  },
  {
    id: "9", name: "Ordenar", desc: "Secuenciar pasos correctamente",
    icon: '<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path d="M3 6h18M3 12h14M3 18h10"/><path d="M19 9l3 3-3 3M17 12h5"/></svg>',
    group: "Minijuego",
    color: "#f97316",
  },
];

const groups = ["Contenido", "Minijuego"];
function typesByGroup(g: string) { return types.filter(t => t.group === g); }
</script>

<template>
  <div class="cts-root">
    <div v-for="group in groups" :key="group" class="cts-group">
      <div class="cts-group-label">
        <span v-if="group === 'Minijuego'" class="cts-badge glass-badge-purple">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="6" width="20" height="12" rx="6"/><path d="M6 12h4m-2-2v4"/><circle cx="17" cy="11" r="1" fill="currentColor"/><circle cx="15" cy="13" r="1" fill="currentColor"/></svg>
          Minijuegos (Gamificación)
        </span>
        <span v-else class="cts-badge glass-badge-blue">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1-0-5H20"/></svg>
          Contenido estático
        </span>
      </div>
      <div class="cts-grid">
        <label
          v-for="t in typesByGroup(group)"
          :key="t.id"
          class="cts-card"
          :class="{ active: modelValue === t.id, game: t.group === 'Minijuego' }"
          :style="t.color && modelValue === t.id ? `--accent:${t.color}` : ''"
        >
          <input
            type="radio"
            :value="t.id"
            :checked="modelValue === t.id"
            @change="emit('update:modelValue', t.id)"
            class="cts-radio"
          />
          <div class="cts-icon" v-html="t.icon" />
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
    </div>
  </div>
</template>

<style scoped>
.cts-root { display: flex; flex-direction: column; gap: 20px; }
.cts-group-label {
  margin: 0 0 12px 0;
}
.cts-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 6px 14px;
  border-radius: 999px;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
}
.glass-badge-purple {
  background: linear-gradient(135deg, rgba(168, 85, 247, 0.15), rgba(236, 72, 153, 0.15));
  color: #c084fc;
  border-color: rgba(192, 132, 252, 0.3);
}
.glass-badge-blue {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(16, 185, 129, 0.15));
  color: #60a5fa;
  border-color: rgba(96, 165, 250, 0.3);
}
.cts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 12px;
}
.cts-card {
  --accent: var(--brand);
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 18px 12px;
  background: var(--surface);
  border: 2px solid var(--border);
  border-radius: var(--r-lg);
  cursor: pointer;
  transition: all 0.2s ease;
}
.cts-card:hover { border-color: var(--accent); background: color-mix(in srgb, var(--accent) 6%, transparent); }
.cts-card.active {
  border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
  box-shadow: 0 4px 14px color-mix(in srgb, var(--accent) 25%, transparent);
}
.cts-card.game .cts-icon { filter: drop-shadow(0 0 6px color-mix(in srgb, var(--accent) 40%, transparent)); }
.cts-radio { position: absolute; opacity: 0; width: 0; height: 0; }
.cts-icon { color: var(--muted); margin-bottom: 10px; transition: color 0.2s; display: flex; align-items: center; justify-content: center; }
.cts-card.active .cts-icon { color: var(--accent); }
.cts-name { font-size: 0.9rem; font-weight: 700; color: var(--dark); margin-bottom: 4px; display: block; }
.cts-desc { font-size: 0.73rem; color: var(--muted); line-height: 1.3; }
.cts-check {
  position: absolute; top: -8px; right: -8px;
  width: 22px; height: 22px;
  background: var(--accent);
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 2px 6px color-mix(in srgb, var(--accent) 40%, transparent);
}
</style>
