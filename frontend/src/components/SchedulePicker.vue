<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  schedules: any[]
  modelValue: string
}>()

const emit = defineEmits(['update:modelValue'])

const selectedDate = ref<string>('')

const groupedSchedules = computed(() => {
  const groups: Record<string, any[]> = {}
  props.schedules.forEach(s => {
    const d = new Date(s.start_time)
    const dateStr = d.toLocaleDateString()
    if (!groups[dateStr]) groups[dateStr] = []
    groups[dateStr].push(s)
  })
  return groups
})

const availableDates = computed(() => Object.keys(groupedSchedules.value))

watch(availableDates, (newDates) => {
  if (newDates.length > 0 && !selectedDate.value) {
    selectedDate.value = newDates[0]
  }
}, { immediate: true })

const currentTimes = computed(() => {
  if (!selectedDate.value) return []
  return groupedSchedules.value[selectedDate.value] || []
})

function formatTime(iso: string) {
  const d = new Date(iso)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function selectTime(id: string) {
  emit('update:modelValue', id)
}
</script>

<template>
  <div class="schedule-picker">
    <div v-if="schedules.length === 0" class="no-schedules">
      <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="8" x2="12" y2="12"></line>
        <line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      No hay horarios disponibles por el momento.
    </div>
    
    <div v-else class="picker-container">
      <div class="dates-list">
        <button
          v-for="d in availableDates"
          :key="d"
          class="date-btn"
          :class="{ active: selectedDate === d }"
          @click="selectedDate = d"
          type="button"
        >
          {{ d }}
        </button>
      </div>
      
      <div class="times-list" v-if="currentTimes.length > 0">
        <button
          v-for="t in currentTimes"
          :key="t.id"
          class="time-btn"
          :class="{ selected: modelValue === t.id }"
          @click="selectTime(t.id)"
          type="button"
        >
          {{ formatTime(t.start_time) }} - {{ formatTime(t.end_time) }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.schedule-picker {
  margin-top: 15px;
  margin-bottom: 20px;
}
.no-schedules {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--danger, #dc2626);
  font-size: 0.9rem;
  background: rgba(220, 38, 38, 0.1);
  padding: 12px;
  border-radius: 8px;
  border: 1px solid rgba(220, 38, 38, 0.2);
}
.picker-container {
  display: flex;
  flex-direction: column;
  gap: 15px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 16px;
}
.dates-list {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 8px;
}
.date-btn {
  white-space: nowrap;
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: 20px;
  background: var(--bg-color);
  color: var(--text-color);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
.date-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
}
.date-btn.active {
  background: var(--primary);
  color: #fff;
  border-color: var(--primary);
}
.times-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 10px;
}
.time-btn {
  padding: 12px 8px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg-color);
  color: var(--text-color);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: center;
}
.time-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
  transform: translateY(-2px);
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}
.time-btn.selected {
  background: var(--primary);
  color: #fff;
  border-color: var(--primary);
}

/* Custom Scrollbar for dates */
.dates-list::-webkit-scrollbar {
  height: 4px;
}
.dates-list::-webkit-scrollbar-track {
  background: transparent;
}
.dates-list::-webkit-scrollbar-thumb {
  background-color: var(--border);
  border-radius: 4px;
}
</style>
