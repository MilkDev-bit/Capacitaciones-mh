<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import {
  esTipoVisible,
  perfilDesdeRol,
  resolverEnlace as resolverEnlaceBase,
} from '../composables/notificaciones'
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

// Las reglas de filtrado y de reescritura de enlaces viven en
// ../composables/notificaciones para poder probarlas como funciones puras.
const router = useRouter()
const auth = useAuthStore()

const notificaciones = ref<Notificacion[]>([])
const isOpen = ref(false)
const verTodas = ref(false)

let pollInterval: ReturnType<typeof setInterval>

const perfil = computed(() => perfilDesdeRol(auth.user?.role))

function esVisible(n: Notificacion) {
  if (verTodas.value) return true
  return esTipoVisible(n.tipo, perfil.value)
}

const noLeidas = computed(() => notificaciones.value.filter(n => !n.leida))

/**
 * Lista pintada en el desplegable.
 *
 * La deduplicación por (título, mensaje, enlace) se conserva aunque el backend
 * ya deduplica al insertar: las filas creadas antes de ese cambio siguen en la
 * base y se verían repetidas.
 */
const displayNotificaciones = computed(() => {
  const vistos = new Set<string>()
  return noLeidas.value.filter(n => {
    if (!esVisible(n)) return false
    const clave = `${n.tipo}|${n.titulo}|${n.mensaje}|${n.enlace}`
    if (vistos.has(clave)) return false
    vistos.add(clave)
    return true
  })
})

const unreadCount = computed(() => displayNotificaciones.value.length)

/** Cuántas quedan fuera por el filtro de perfil (para ofrecer "ver todas"). */
const ocultasPorPerfil = computed(() => {
  if (verTodas.value) return 0
  return noLeidas.value.filter(n => !esVisible(n)).length
})

/** Comprueba contra el router real si una ruta existe. */
function existeRuta(path: string) {
  return router.resolve(path).matched.length > 0
}

function resolverEnlace(enlace: string): string | null {
  return resolverEnlaceBase(enlace, perfil.value, existeRuta)
}

async function fetchNotificaciones() {
  try {
    const res = await api.get('/notificaciones')
    notificaciones.value = Array.isArray(res.data) ? res.data : []
  } catch (err) {
    console.error('Error fetching notificaciones', err)
  }
}

/** Marca como leídas todas las copias equivalentes que la lista había agrupado. */
async function markAsReadWithoutNavigating(n: Notificacion) {
  const duplicadas = notificaciones.value.filter(
    x => !x.leida && x.tipo === n.tipo && x.titulo === n.titulo && x.mensaje === n.mensaje,
  )
  const ids = duplicadas.map(x => x.id)
  if (ids.length === 0) return

  try {
    await api.post('/notificaciones/marcar-leidas', { ids })
    duplicadas.forEach(x => { x.leida = true })
  } catch (err) {
    console.error('Error marking as read', err)
  }
}

async function markAsRead(n: Notificacion) {
  await markAsReadWithoutNavigating(n)
  isOpen.value = false

  const destino = resolverEnlace(n.enlace)
  if (destino) router.push(destino)
}

async function markAllAsRead() {
  const ids = noLeidas.value.map(n => n.id)
  if (ids.length === 0) return

  try {
    await api.post('/notificaciones/marcar-leidas', { ids })
    notificaciones.value.forEach(n => { n.leida = true })
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
    verTodas.value = false
  }
}

function timeAgo(dateString: string) {
  if (!dateString) return ''
  let date = new Date(dateString)
  if (isNaN(date.getTime())) {
    date = new Date(dateString.replace(' ', 'T'))
  }
  if (isNaN(date.getTime())) return ''

  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)

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
  return Math.max(0, Math.floor(seconds)) + 's'
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

defineExpose({ resolverEnlace, displayNotificaciones, unreadCount })
</script>

<template>
  <div class="notification-wrapper">
    <button class="icon-btn" @click="toggleDropdown" data-tooltip="Notificaciones" :aria-label="`Notificaciones (${unreadCount} sin leer)`">
      <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
      <span v-if="unreadCount > 0" class="badge">{{ unreadCount > 9 ? '9+' : unreadCount }}</span>
    </button>

    <Transition name="slide-down">
      <div v-if="isOpen" class="dropdown-menu">
        <div class="dropdown-header">
          <strong>Notificaciones</strong>
          <button v-if="noLeidas.length > 0" class="mark-read-btn" @click.stop="markAllAsRead">
            Marcar todas como leídas
          </button>
        </div>
        <div class="dropdown-body">
          <div v-if="displayNotificaciones.length === 0" class="empty-state">
            No tienes notificaciones
          </div>
          <div v-else
               v-for="n in displayNotificaciones"
               :key="n.id"
               class="notification-item unread"
               @click="markAsRead(n)">
            <div class="notif-content">
              <strong>{{ n.titulo || 'Notificación' }}</strong>
              <p v-if="n.mensaje">{{ n.mensaje }}</p>
              <span class="time">{{ timeAgo(n.created_at) }}</span>
            </div>
            <button class="mark-btn" @click.stop="markAsReadWithoutNavigating(n)" title="Marcar como leída">
              <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"></path>
              </svg>
            </button>
          </div>
        </div>
        <!--
          El filtro por perfil oculta, nunca borra. Este pie es la salida de
          emergencia: sin él, una notificación mal clasificada sería invisible
          para el usuario y no habría forma de saber que existe.
        -->
        <button v-if="ocultasPorPerfil > 0" class="ver-todas" @click.stop="verTodas = true">
          Ver {{ ocultasPorPerfil }} de otros perfiles
        </button>
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
  top: 2px;
  right: 2px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  background: var(--danger);
  color: #fff;
  border-radius: 8px;
  border: 2px solid var(--surface);
  font-size: 0.65rem;
  font-weight: 700;
  line-height: 16px;
  text-align: center;
  box-sizing: content-box;
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

.mark-btn {
  background: none;
  border: none;
  color: var(--muted);
  cursor: pointer;
  padding: 6px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
  margin-left: 12px;
}

.mark-btn:hover {
  background: var(--surface);
  color: var(--brand);
}

.ver-todas {
  width: 100%;
  padding: 10px 16px;
  border: none;
  border-top: 1px solid var(--border);
  background: transparent;
  color: var(--brand);
  font-size: 0.8rem;
  cursor: pointer;
}

.ver-todas:hover {
  background-color: var(--surface-hover);
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
