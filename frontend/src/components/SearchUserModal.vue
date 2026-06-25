<script setup lang="ts">
import { ref, watch } from 'vue'
import api from '../api'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', user: { id: string; name: string; avatar_url: string }): void
}>()

const query = ref('')
const results = ref<any[]>([])
const loading = ref(false)

async function onSearch() {
  if (!query.value.trim()) {
    results.value = []
    return
  }
  loading.value = true
  try {
    const res = await api.get('/usuarios/search', { params: { q: query.value.trim() } })
    results.value = res.data || []
  } catch (e) {
    console.error('Error al buscar usuarios:', e)
  } finally {
    loading.value = false
  }
}

let timeout: ReturnType<typeof setTimeout> | null = null
watch(query, () => {
  if (timeout) clearTimeout(timeout)
  timeout = setTimeout(onSearch, 300)
})

function selectUser(u: any) {
  emit('select', { id: u.id, name: u.name, avatar_url: u.avatar_url })
  emit('close')
}
</script>

<template>
  <Transition name="modal">
    <div v-if="show" class="modal-mask" @click.self="emit('close')">
      <div class="modal-wrapper">
        <div class="modal-container">
          <div class="modal-header">
            <h3>Nueva Conversación</h3>
            <button class="close-btn" @click="emit('close')">✕</button>
          </div>
          
          <div class="modal-body">
            <input 
              v-model="query" 
              type="text" 
              class="search-input" 
              placeholder="Buscar por nombre o correo..." 
              autofocus 
            />
            
            <div v-if="loading" class="loading-state">Buscando...</div>
            <div v-else-if="results.length === 0 && query" class="empty-state">No se encontraron usuarios</div>
            
            <ul v-else class="user-list">
              <li v-for="user in results" :key="user.id" class="user-item" @click="selectUser(user)">
                <div class="avatar">
                  <img v-if="user.avatar_url" :src="user.avatar_url" alt="" />
                  <span v-else>{{ user.name.charAt(0).toUpperCase() }}</span>
                </div>
                <div class="user-info">
                  <div class="name">{{ user.name }}</div>
                  <div class="email">{{ user.email }}</div>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-mask {
  position: fixed;
  z-index: 9999;
  top: 0; left: 0; width: 100%; height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex; align-items: center; justify-content: center;
}
.modal-wrapper {
  width: 100%; max-width: 400px;
}
.modal-container {
  background: var(--surface);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.15);
  display: flex; flex-direction: column;
  overflow: hidden;
}
.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--border);
}
.modal-header h3 { margin: 0; font-size: 1.1rem; color: var(--text); }
.close-btn { background: none; border: none; font-size: 1.2rem; cursor: pointer; color: var(--text-muted); }
.close-btn:hover { color: var(--text); }

.modal-body { padding: 1rem 1.25rem; display: flex; flex-direction: column; gap: 1rem; max-height: 60vh; }
.search-input {
  width: 100%; padding: 0.75rem 1rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--background);
  color: var(--text); outline: none;
}
.search-input:focus { border-color: var(--primary); }

.loading-state, .empty-state { text-align: center; color: var(--text-muted); padding: 2rem 0; }

.user-list { list-style: none; padding: 0; margin: 0; overflow-y: auto; flex: 1; }
.user-item {
  display: flex; align-items: center; gap: 1rem; padding: 0.75rem;
  border-radius: 8px; cursor: pointer; transition: background 0.2s;
}
.user-item:hover { background: var(--surface-hover); }
.avatar {
  width: 40px; height: 40px; border-radius: 50%; background: var(--primary); color: #fff;
  display: flex; align-items: center; justify-content: center; font-weight: bold; overflow: hidden;
}
.avatar img { width: 100%; height: 100%; object-fit: cover; }
.user-info { display: flex; flex-direction: column; }
.name { font-weight: 500; color: var(--text); }
.email { font-size: 0.85rem; color: var(--text-muted); }

/* Transitions */
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
