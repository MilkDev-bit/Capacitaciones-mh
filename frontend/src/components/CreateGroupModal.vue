<script setup lang="ts">
import { ref, watch } from 'vue'
import api from '../api'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', group: { id: string; name: string }): void
}>()

const name = ref('')
const query = ref('')
const results = ref<any[]>([])
const selectedUsers = ref<Map<string, any>>(new Map())
const loading = ref(false)
const creating = ref(false)

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

function toggleUser(user: any) {
  if (selectedUsers.value.has(user.id)) {
    selectedUsers.value.delete(user.id)
  } else {
    selectedUsers.value.set(user.id, user)
  }
}

async function createGroup() {
  if (!name.value.trim()) return
  creating.value = true
  try {
    const res = await api.post('/mensajes/grupos', {
      nombre: name.value.trim(),
      members: Array.from(selectedUsers.value.keys())
    })
    emit('created', { id: res.data.grupo_id, name: res.data.nombre })
    emit('close')
  } catch (e) {
    console.error('Error creando grupo', e)
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <Transition name="modal">
    <div v-if="show" class="modal-mask" @click.self="emit('close')">
      <div class="modal-wrapper">
        <div class="modal-container">
          <div class="modal-header">
            <h3>Crear Grupo</h3>
            <button class="close-btn" @click="emit('close')">✕</button>
          </div>
          
          <div class="modal-body">
            <div class="form-group">
              <label>Nombre del Grupo</label>
              <input v-model="name" type="text" class="input-field" placeholder="Ej: Equipo de Desarrollo" />
            </div>

            <div class="form-group">
              <label>Añadir Miembros</label>
              <input v-model="query" type="text" class="input-field search-input" placeholder="Buscar usuarios..." />
            </div>

            <!-- Seleccionados -->
            <div v-if="selectedUsers.size > 0" class="selected-tags">
              <span v-for="[id, user] in selectedUsers" :key="id" class="tag">
                {{ user.name.split(' ')[0] }}
                <button type="button" @click="toggleUser(user)">✕</button>
              </span>
            </div>

            <div v-if="loading" class="loading-state">Buscando...</div>
            <ul v-else-if="results.length > 0" class="user-list">
              <li v-for="user in results" :key="user.id" class="user-item" @click="toggleUser(user)">
                <div class="avatar">
                  <img v-if="user.avatar_url" :src="user.avatar_url" alt="" />
                  <span v-else>{{ user.name.charAt(0).toUpperCase() }}</span>
                </div>
                <div class="user-info">
                  <div class="name">{{ user.name }}</div>
                  <div class="email">{{ user.email }}</div>
                </div>
                <div class="check-icon" v-if="selectedUsers.has(user.id)">
                  <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M20 6L9 17l-5-5"/></svg>
                </div>
              </li>
            </ul>
          </div>
          
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="emit('close')">Cancelar</button>
            <button class="btn btn-primary" :disabled="!name.trim() || creating" @click="createGroup">
              {{ creating ? 'Creando...' : 'Crear Grupo' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-mask {
  position: fixed; z-index: 9999;
  top: 0; left: 0; width: 100%; height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex; align-items: center; justify-content: center;
}
.modal-wrapper { width: 100%; max-width: 450px; }
.modal-container {
  background: var(--surface); border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.15);
  display: flex; flex-direction: column; overflow: hidden;
}
.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--border);
}
.modal-header h3 { margin: 0; font-size: 1.1rem; color: var(--text); }
.close-btn { background: none; border: none; font-size: 1.2rem; cursor: pointer; color: var(--text-muted); }
.close-btn:hover { color: var(--text); }

.modal-body { padding: 1.25rem; display: flex; flex-direction: column; gap: 1rem; max-height: 60vh; overflow-y: auto; }
.form-group { display: flex; flex-direction: column; gap: 0.5rem; }
.form-group label { font-size: 0.9rem; font-weight: 500; color: var(--text-muted); }
.input-field {
  width: 100%; padding: 0.75rem 1rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--background);
  color: var(--text); outline: none;
}
.input-field:focus { border-color: var(--primary); }

.selected-tags { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.tag {
  background: var(--primary-soft, rgba(59,130,246,0.15)); color: var(--primary);
  padding: 0.25rem 0.75rem; border-radius: 16px; font-size: 0.85rem; font-weight: 500;
  display: flex; align-items: center; gap: 0.5rem;
}
.tag button { background: none; border: none; color: inherit; cursor: pointer; font-size: 0.9rem; }

.loading-state { text-align: center; color: var(--text-muted); padding: 1rem 0; }
.user-list { list-style: none; padding: 0; margin: 0; }
.user-item {
  display: flex; align-items: center; gap: 1rem; padding: 0.75rem;
  border-radius: 8px; cursor: pointer; transition: background 0.2s; position: relative;
}
.user-item:hover { background: var(--surface-hover); }
.avatar {
  width: 40px; height: 40px; border-radius: 50%; background: var(--primary); color: #fff;
  display: flex; align-items: center; justify-content: center; font-weight: bold; overflow: hidden;
}
.avatar img { width: 100%; height: 100%; object-fit: cover; }
.user-info { display: flex; flex-direction: column; flex: 1; }
.name { font-weight: 500; color: var(--text); }
.email { font-size: 0.85rem; color: var(--text-muted); }
.check-icon { color: var(--primary); }

.modal-footer {
  padding: 1rem 1.25rem; border-top: 1px solid var(--border);
  display: flex; justify-content: flex-end; gap: 1rem;
}
.btn { padding: 0.5rem 1.25rem; border-radius: 8px; font-weight: 600; cursor: pointer; transition: opacity 0.2s; border: none; }
.btn:hover { opacity: 0.9; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: var(--surface-hover); color: var(--text); }
.btn-primary { background: var(--primary); color: #fff; }

.modal-enter-active, .modal-leave-active { transition: opacity 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
