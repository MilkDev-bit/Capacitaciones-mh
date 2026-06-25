<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'

interface Notificacion {
  id: string
  user_id: string
  tipo: string
  titulo: string
  mensaje: string
  leida: boolean
  enlace: string
  created_at: string
}

const router = useRouter()
const notificaciones = ref<Notificacion[]>([])
const unreadCount = computed(() => notificaciones.value.filter(n => !n.leida).length)
const isOpen = ref(false)

let pollInterval: ReturnType<typeof setInterval>

async function fetchNotificaciones() {
  try {
    const res = await api.get('/notificaciones')
    notificaciones.value = res.data
  } catch (err) {
    console.error('Error fetching notificaciones', err)
  }
}

async function markAsRead(n: Notificacion) {
  if (!n.leida) {
    try {
      await api.post('/notificaciones/marcar-leidas', { ids: [n.id] })
      n.leida = true
    } catch (err) {
      console.error('Error marking as read', err)
    }
  }
  isOpen.value = false
  if (n.enlace) {
    router.push(n.enlace)
  }
}

async function markAllAsRead() {
  const unreadIds = notificaciones.value.filter(n => !n.leida).map(n => n.id)
  if (unreadIds.length === 0) return

  try {
    await api.post('/notificaciones/marcar-leidas', { ids: unreadIds })
    notificaciones.value.forEach(n => n.leida = true)
  } catch (err) {
    console.error('Error marking all as read', err)
  }
}

function toggleDropdown() {
  isOpen.value = !isOpen.value
}

function closeDropdown(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.notification-wrapper')) {
    isOpen.value = false
  }
}

function timeAgo(dateString: string) {
  const date = new Date(dateString)
  const now = new Date()
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000)

  let interval = seconds / 31536000
  if (interval > 1) return Math.floor(interval) + ' años'
  interval = seconds / 2592000
  if (interval > 1) return Math.floor(interval) + ' meses'
  interval = seconds / 86400
  if (interval > 1) return Math.floor(interval) + 'd'
  interval = seconds / 3600
  if (interval > 1) return Math.floor(interval) + 'h'
  interval = seconds / 60
  if (interval > 1) return Math.floor(interval) + 'm'
  return Math.floor(seconds) + 's'
}

onMounted(() => {
  fetchNotificaciones()
  pollInterval = setInterval(fetchNotificaciones, 30000)
  document.addEventListener('click', closeDropdown)
})

onUnmounted(() => {
  clearInterval(pollInterval)
  document.removeEventListener('click', closeDropdown)
})
</script>

<template>
  <div class="notification-wrapper">
    <button class="icon-btn" @click="toggleDropdown" data-tooltip="Notificaciones">
      <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
      <span v-if="unreadCount > 0" class="badge"></span>
    </button>

    <Transition name="slide-down">
      <div v-if="isOpen" class="dropdown-menu">
        <div class="dropdown-header">
          <strong>Notificaciones</strong>
          <button v-if="unreadCount > 0" class="mark-read-btn" @click.stop="markAllAsRead">
            Marcar todas como leídas
          </button>
        </div>
        <div class="dropdown-body">
          <div v-if="notificaciones.length === 0" class="empty-state">
            No tienes notificaciones
          </div>
          <div v-else 
               v-for="n in notificaciones" 
               :key="n.id" 
               class="notification-item" 
               :class="{ unread: !n.leida }"
               @click="markAsRead(n)">
            <div class="notif-content">
              <strong>{{ n.titulo }}</strong>
              <p>{{ n.mensaje }}</p>
              <span class="time">{{ timeAgo(n.created_at) }}</span>
            </div>
            <div v-if="!n.leida" class="unread-dot"></div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.notification-wrapper {
  position: relative;
}

.icon-btn {
  border: none;
  background: transparent;
  position: relative;
  width: 36px;
  height: 36px;
  cursor: pointer;
  color: var(--text);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.icon-btn:hover {
  background-color: var(--surface-hover);
}

.badge {
  position: absolute;
  top: 6px;
  right: 8px;
  width: 8px;
  height: 8px;
  background: var(--danger);
  border-radius: 50%;
  border: 2px solid var(--surface);
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: -10px;
  width: 320px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0,0,0,0.1);
  z-index: 100;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.dropdown-header {
  padding: 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dropdown-header strong {
  font-size: 1rem;
  color: var(--text);
}

.mark-read-btn {
  background: none;
  border: none;
  color: var(--brand);
  font-size: 0.8rem;
  cursor: pointer;
  padding: 0;
}

.mark-read-btn:hover {
  text-decoration: underline;
}

.dropdown-body {
  max-height: 400px;
  overflow-y: auto;
}

.empty-state {
  padding: 32px 16px;
  text-align: center;
  color: var(--muted);
  font-size: 0.9rem;
}

.notification-item {
  padding: 16px;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  transition: background-color 0.2s;
}

.notification-item:last-child {
  border-bottom: none;
}

.notification-item:hover {
  background-color: var(--surface-hover);
}

.notification-item.unread {
  background-color: rgba(var(--brand-rgb), 0.05);
}

.notif-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.notif-content strong {
  font-size: 0.9rem;
  color: var(--text);
}

.notif-content p {
  margin: 0;
  font-size: 0.85rem;
  color: var(--muted);
  line-height: 1.4;
}

.time {
  font-size: 0.75rem;
  color: var(--muted);
  margin-top: 4px;
}

.unread-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--brand);
  flex-shrink: 0;
  margin-left: 12px;
}

/* Animations */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
