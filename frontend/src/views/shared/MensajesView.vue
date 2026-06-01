<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import api from '../../api'

// ─── Tipos ───────────────────────────────────────────────────────────────────
interface Conversacion {
  peer_id: string
  peer_name: string
  last_message: string
  last_time: string
  unread_count: number
}

interface Mensaje {
  id: string
  emisor_id: string
  emisor_name: string
  receptor_id: string
  receptor_name: string
  contenido: string
  leido: boolean
  created_at: string
}

// ─── Estado ───────────────────────────────────────────────────────────────────
const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()

const convs       = ref<Conversacion[]>([])
const msgs        = ref<Mensaje[]>([])
const peerName    = ref('')
const newMsg      = ref('')
const loadingConvs = ref(false)
const loadingMsgs  = ref(false)
const sending      = ref(false)
const threadRef    = ref<HTMLElement | null>(null)

const activePeerId = computed(() => route.params.peer_id as string | undefined)

// ─── Helpers ──────────────────────────────────────────────────────────────────
function formatTime(iso: string) {
  const d = new Date(iso)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffDays = Math.floor(diffMs / 86400000)
  if (diffDays === 0) return d.toLocaleTimeString('es', { hour: '2-digit', minute: '2-digit' })
  if (diffDays === 1) return 'ayer'
  if (diffDays < 7)  return d.toLocaleDateString('es', { weekday: 'short' })
  return d.toLocaleDateString('es', { day: '2-digit', month: 'short' })
}

function formatDateSep(iso: string) {
  return new Date(iso).toLocaleDateString('es', { weekday: 'long', day: '2-digit', month: 'long' })
}

function isSameDay(a: string, b: string) {
  return new Date(a).toDateString() === new Date(b).toDateString()
}

function initials(name: string) {
  return (name || '?').split(' ').map((w: string) => w[0]).join('').toUpperCase().slice(0, 2)
}

// ─── API ─────────────────────────────────────────────────────────────────────
async function loadConversaciones() {
  loadingConvs.value = true
  try {
    const res = await api.get('/mensajes/conversaciones')
    convs.value = res.data ?? []
  } catch { /* silencioso */ } finally {
    loadingConvs.value = false
  }
}

async function loadMensajes(peerId: string) {
  loadingMsgs.value = true
  msgs.value = []
  try {
    // Intentar nombre desde history.state (navegación directa) si aún no se tiene
    if (!peerName.value && (history.state as { peerName?: string })?.peerName) {
      peerName.value = (history.state as { peerName: string }).peerName
    }
    // Si el peer aún no tiene nombre, cargarlo del perfil público
    if (!peerName.value) {
      try {
        const perfil = await api.get(`/usuarios/${peerId}/perfil`)
        peerName.value = perfil.data?.user?.name ?? ''
      } catch { /* ignorar */ }
    }
    const res = await api.get(`/mensajes/${peerId}`)
    msgs.value = res.data ?? []
    // Actualizar el unread de esa conversacion en la lista
    const conv = convs.value.find(c => c.peer_id === peerId)
    if (conv) conv.unread_count = 0
    await scrollToBottom()
  } catch { /* silencioso */ } finally {
    loadingMsgs.value = false
  }
}

async function sendMensaje() {
  if (!activePeerId.value || !newMsg.value.trim() || sending.value) return
  sending.value = true
  const text = newMsg.value.trim()
  newMsg.value = ''
  try {
    const res = await api.post(`/mensajes/${activePeerId.value}`, {
      contenido: text,
      peer_name: peerName.value,
    })
    msgs.value.push(res.data)
    // Actualizar última conversación
    const conv = convs.value.find(c => c.peer_id === activePeerId.value)
    if (conv) {
      conv.last_message = text
      conv.last_time    = res.data.created_at
      convs.value = [conv, ...convs.value.filter(c => c.peer_id !== activePeerId.value)]
    } else {
      convs.value.unshift({
        peer_id: activePeerId.value,
        peer_name: peerName.value,
        last_message: text,
        last_time: res.data.created_at,
        unread_count: 0,
      })
    }
    await scrollToBottom()
  } catch {
    newMsg.value = text // restaurar si falla
  } finally {
    sending.value = false
  }
}

async function scrollToBottom() {
  await nextTick()
  if (threadRef.value) {
    threadRef.value.scrollTop = threadRef.value.scrollHeight
  }
}

function openConversacion(conv: Conversacion) {
  peerName.value = conv.peer_name
  // Determinar la ruta base según el rol
  const base = auth.isInstructor ? '/instructor' : '/usuario'
  router.push(`${base}/mensajes/${conv.peer_id}`)
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMensaje()
  }
}

// ─── Watcher: cambio de conversación activa ───────────────────────────────────
watch(activePeerId, async (peerId) => {
  if (peerId) {
    await loadMensajes(peerId)
  } else {
    msgs.value = []
    peerName.value = ''
  }
}, { immediate: true })

// ─── Polling automático cada 10 s ────────────────────────────────────────────
let pollTimer: ReturnType<typeof setInterval>
onMounted(async () => {
  await loadConversaciones()
  pollTimer = setInterval(async () => {
    await loadConversaciones()
    if (activePeerId.value) await loadMensajes(activePeerId.value)
  }, 10000)
})
onUnmounted(() => clearInterval(pollTimer))
</script>

<template>
  <div class="mensajes-shell">
    <!-- ── Panel izquierdo: lista de conversaciones ───────────────────────── -->
    <aside :class="['convs-panel', activePeerId ? 'hidden-mobile' : '']">
      <div class="convs-header">
        <h2>Mensajes</h2>
      </div>

      <div v-if="loadingConvs && convs.length === 0" class="convs-loading">
        <span class="spinner"></span>
      </div>
      <div v-else-if="convs.length === 0" class="convs-empty">
        <svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <path d="M8 10h.01M12 10h.01M16 10h.01M21 12c0 4.418-4.03 8-9 8a9.862 9.862 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
        </svg>
        <p>Aún no tienes mensajes</p>
      </div>

      <ul v-else class="convs-list">
        <li
          v-for="conv in convs"
          :key="conv.peer_id"
          :class="['conv-item', conv.peer_id === activePeerId ? 'active' : '']"
          @click="openConversacion(conv)"
        >
          <div class="conv-avatar">{{ initials(conv.peer_name) }}</div>
          <div class="conv-info">
            <div class="conv-row">
              <span class="conv-name">{{ conv.peer_name }}</span>
              <span class="conv-time">{{ formatTime(conv.last_time) }}</span>
            </div>
            <div class="conv-row">
              <span class="conv-preview">{{ conv.last_message }}</span>
              <span v-if="conv.unread_count > 0" class="conv-badge">{{ conv.unread_count }}</span>
            </div>
          </div>
        </li>
      </ul>
    </aside>

    <!-- ── Panel derecho: hilo de mensajes ────────────────────────────────── -->
    <section :class="['thread-panel', !activePeerId ? 'hidden-mobile' : '']">
      <!-- Sin conversación seleccionada -->
      <div v-if="!activePeerId" class="thread-empty">
        <svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.2" viewBox="0 0 24 24">
          <path d="M8 10h.01M12 10h.01M16 10h.01M21 12c0 4.418-4.03 8-9 8a9.862 9.862 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
        </svg>
        <p>Selecciona una conversación para leer tus mensajes</p>
      </div>

      <!-- Conversación activa -->
      <template v-else>
        <div class="thread-header">
          <button class="back-btn" @click="router.back()">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M15 19l-7-7 7-7"/></svg>
          </button>
          <div class="thread-avatar">{{ initials(peerName) }}</div>
          <span class="thread-peername">{{ peerName || '...' }}</span>
        </div>

        <div ref="threadRef" class="thread-body">
          <div v-if="loadingMsgs" class="thread-loading">
            <span class="spinner"></span>
          </div>
          <div v-else-if="msgs.length === 0" class="thread-no-msgs">
            <p>Sé el primero en escribir 👋</p>
          </div>
          <template v-else>
            <template v-for="(msg, idx) in msgs" :key="msg.id">
              <!-- Separador de fecha -->
              <div
                v-if="idx === 0 || !isSameDay(msgs[idx - 1].created_at, msg.created_at)"
                class="date-sep"
              >
                <span>{{ formatDateSep(msg.created_at) }}</span>
              </div>

              <!-- Burbuja -->
              <div :class="['bubble-wrap', msg.emisor_id === auth.user?.id ? 'mine' : 'theirs']">
                <div class="bubble">
                  <p>{{ msg.contenido }}</p>
                  <span class="bubble-time">{{ formatTime(msg.created_at) }}</span>
                </div>
              </div>
            </template>
          </template>
        </div>

        <!-- Input de envío -->
        <form class="thread-input" @submit.prevent="sendMensaje">
          <textarea
            v-model="newMsg"
            rows="1"
            placeholder="Escribe un mensaje…"
            :disabled="sending"
            @keydown="handleKeydown"
          ></textarea>
          <button type="submit" :disabled="!newMsg.trim() || sending" aria-label="Enviar">
            <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/>
            </svg>
          </button>
        </form>
      </template>
    </section>
  </div>
</template>

<style scoped>
/* ── Layout general ──────────────────────────────────────────────────────────── */
.mensajes-shell {
  display: flex;
  height: 100%;
  overflow: hidden;
  background: var(--surface);
}

/* ── Panel izquierdo ─────────────────────────────────────────────────────────── */
.convs-panel {
  width: 320px;
  min-width: 260px;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  background: var(--surface);
  overflow: hidden;
}

.convs-header {
  padding: 1.25rem 1.25rem 1rem;
  border-bottom: 1px solid var(--border);
}
.convs-header h2 {
  font-size: 1.15rem;
  font-weight: 700;
  margin: 0;
  color: var(--text);
}

.convs-loading,
.convs-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: .75rem;
  color: var(--text-muted);
  font-size: .9rem;
  padding: 2rem;
}
.convs-empty svg { opacity: .35; }

.convs-list {
  flex: 1;
  overflow-y: auto;
  list-style: none;
  margin: 0;
  padding: 0;
}

.conv-item {
  display: flex;
  align-items: center;
  gap: .75rem;
  padding: .85rem 1.25rem;
  cursor: pointer;
  border-bottom: 1px solid var(--border);
  transition: background .15s;
}
.conv-item:hover { background: var(--surface-hover); }
.conv-item.active { background: var(--primary-soft, rgba(59,130,246,.08)); }

.conv-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: var(--primary, #3b82f6);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: .8rem;
  flex-shrink: 0;
}

.conv-info { flex: 1; min-width: 0; }

.conv-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
}

.conv-name {
  font-weight: 600;
  font-size: .9rem;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.conv-time {
  font-size: .75rem;
  color: var(--text-muted);
  white-space: nowrap;
  flex-shrink: 0;
}
.conv-preview {
  font-size: .82rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.conv-badge {
  background: var(--primary, #3b82f6);
  color: #fff;
  font-size: .7rem;
  font-weight: 700;
  border-radius: 999px;
  padding: .1rem .45rem;
  min-width: 18px;
  text-align: center;
  flex-shrink: 0;
}

/* ── Panel derecho ───────────────────────────────────────────────────────────── */
.thread-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

.thread-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  color: var(--text-muted);
  font-size: .95rem;
  padding: 2rem;
  text-align: center;
}
.thread-empty svg { opacity: .3; }

.thread-header {
  display: flex;
  align-items: center;
  gap: .75rem;
  padding: .9rem 1.25rem;
  border-bottom: 1px solid var(--border);
  background: var(--surface);
}

.back-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: .25rem;
  border-radius: .4rem;
  display: flex;
  align-items: center;
  transition: color .15s;
}
.back-btn:hover { color: var(--text); }

.thread-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--primary, #3b82f6);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: .75rem;
  flex-shrink: 0;
}
.thread-peername {
  font-weight: 600;
  font-size: 1rem;
  color: var(--text);
}

.thread-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem 1.25rem .5rem;
  display: flex;
  flex-direction: column;
  gap: .25rem;
}

.thread-loading,
.thread-no-msgs {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: .9rem;
}

.date-sep {
  text-align: center;
  margin: .75rem 0;
}
.date-sep span {
  background: var(--surface-soft);
  color: var(--text-muted);
  font-size: .72rem;
  padding: .25rem .75rem;
  border-radius: 999px;
}

.bubble-wrap {
  display: flex;
  margin-bottom: .2rem;
}
.bubble-wrap.mine    { justify-content: flex-end; }
.bubble-wrap.theirs  { justify-content: flex-start; }

.bubble {
  max-width: 72%;
  padding: .55rem .85rem;
  border-radius: 1.2rem;
  font-size: .9rem;
  line-height: 1.5;
  word-break: break-word;
}
.bubble p { margin: 0 0 .25rem; }
.bubble-time {
  font-size: .68rem;
  color: rgba(255,255,255,.65);
  display: block;
  text-align: right;
}

.bubble-wrap.mine .bubble {
  background: var(--primary, #3b82f6);
  color: #fff;
  border-bottom-right-radius: .3rem;
}
.bubble-wrap.theirs .bubble {
  background: var(--surface-soft);
  color: var(--text);
  border-bottom-left-radius: .3rem;
}
.bubble-wrap.theirs .bubble-time {
  color: var(--text-muted);
}

/* ── Input de envío ─────────────────────────────────────────────────────────── */
.thread-input {
  display: flex;
  align-items: flex-end;
  gap: .6rem;
  padding: .75rem 1.25rem 1rem;
  border-top: 1px solid var(--border);
  background: var(--surface);
}

.thread-input textarea {
  flex: 1;
  resize: none;
  border: 1.5px solid var(--border);
  border-radius: .75rem;
  padding: .6rem .9rem;
  font-size: .9rem;
  font-family: inherit;
  background: var(--surface-soft);
  color: var(--text);
  outline: none;
  max-height: 120px;
  overflow-y: auto;
  line-height: 1.5;
  transition: border-color .2s;
}
.thread-input textarea:focus { border-color: var(--primary, #3b82f6); }
.thread-input textarea:disabled { opacity: .5; cursor: not-allowed; }

.thread-input button[type="submit"] {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: var(--primary, #3b82f6);
  color: #fff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: opacity .2s, transform .1s;
}
.thread-input button[type="submit"]:hover  { opacity: .85; }
.thread-input button[type="submit"]:active { transform: scale(.92); }
.thread-input button[type="submit"]:disabled { opacity: .4; cursor: not-allowed; }

/* ── Spinner ─────────────────────────────────────────────────────────────────── */
.spinner {
  display: inline-block;
  width: 28px;
  height: 28px;
  border: 3px solid var(--border);
  border-top-color: var(--primary, #3b82f6);
  border-radius: 50%;
  animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Responsive: móvil ───────────────────────────────────────────────────────── */
@media (max-width: 640px) {
  .convs-panel { width: 100%; border-right: none; }
  .hidden-mobile { display: none; }
}
</style>
