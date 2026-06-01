<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import api from '../../api'

// ─── Tipos ─────────────────────────────────────────────────────────────────
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
  _status?: 'sending' | 'sent' | 'error'
  _tempId?: string
}

// ─── Estado ────────────────────────────────────────────────────────────────
const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()

const convs          = ref<Conversacion[]>([])
const msgs           = ref<Mensaje[]>([])
const peerName       = ref('')
const newMsg         = ref('')
const loadingConvs   = ref(false)
const loadingMsgs    = ref(false)
const sending        = ref(false)
const hasMore        = ref(false)
const loadingMore    = ref(false)
const showTyping     = ref(false)
const typingPeerName = ref('')
const threadRef      = ref<HTMLElement | null>(null)
const sentinelRef    = ref<HTMLElement | null>(null)
const textareaRef    = ref<HTMLTextAreaElement | null>(null)
const errorMsg       = ref('')

const activePeerId = computed(() => route.params.peer_id as string | undefined)

// ─── WebSocket ─────────────────────────────────────────────────────────────
let ws: WebSocket | null = null
let wsReconnectTimer: ReturnType<typeof setTimeout> | null = null
let wsReconnectDelay = 1000
let wsShouldReconnect = true
let typingHideTimer: ReturnType<typeof setTimeout> | null = null
let lastTypingSent = 0

function connectWs() {
  if (!wsShouldReconnect) return
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${proto}//${location.host}/api/ws`)
  ws.onopen = () => { wsReconnectDelay = 1000 }
  ws.onmessage = (ev: MessageEvent) => {
    try { handleWsEvent(JSON.parse(ev.data as string)) } catch { /* ignorar */ }
  }
  ws.onclose = () => {
    if (!wsShouldReconnect) return
    wsReconnectTimer = setTimeout(() => {
      wsReconnectDelay = Math.min(wsReconnectDelay * 2, 30000)
      connectWs()
    }, wsReconnectDelay)
  }
  ws.onerror = () => ws?.close()
}

function disconnectWs() {
  wsShouldReconnect = false
  if (wsReconnectTimer) clearTimeout(wsReconnectTimer)
  ws?.close()
  ws = null
}

function handleWsEvent(ev: { type: string; msg?: Mensaje; peer_id?: string; peer_name?: string }) {
  switch (ev.type) {
    case 'new_message': {
      if (!ev.msg) break
      if (activePeerId.value === ev.msg.emisor_id) {
        msgs.value.push(ev.msg)
        scrollToBottom()
        markRead(ev.msg.id)
      }
      refreshConvEntry(ev.msg)
      break
    }
    case 'message_read': {
      msgs.value.forEach(m => { if (m.emisor_id === auth.user?.id) m.leido = true })
      break
    }
    case 'typing': {
      if (ev.peer_id !== activePeerId.value) break
      typingPeerName.value = ev.peer_name ?? ''
      showTyping.value = true
      if (typingHideTimer) clearTimeout(typingHideTimer)
      typingHideTimer = setTimeout(() => { showTyping.value = false }, 3000)
      break
    }
  }
}

function sendTyping() {
  if (!activePeerId.value || !ws || ws.readyState !== WebSocket.OPEN) return
  const now = Date.now()
  if (now - lastTypingSent < 2000) return
  lastTypingSent = now
  ws.send(JSON.stringify({ type: 'typing', peer_id: activePeerId.value }))
}

async function markRead(msgId: string) {
  try { await api.post(`/mensajes/leido/${msgId}`) } catch { /* silencioso */ }
}

function refreshConvEntry(msg: Mensaje) {
  const peerId = msg.emisor_id === auth.user?.id ? msg.receptor_id : msg.emisor_id
  const peerN  = msg.emisor_id === auth.user?.id ? msg.receptor_name : msg.emisor_name
  const conv = convs.value.find(c => c.peer_id === peerId)
  const unread = activePeerId.value !== peerId && msg.emisor_id !== auth.user?.id ? 1 : 0
  if (conv) {
    conv.last_message = msg.contenido
    conv.last_time    = msg.created_at
    conv.unread_count += unread
    convs.value = [conv, ...convs.value.filter(c => c.peer_id !== peerId)]
  } else {
    convs.value.unshift({ peer_id: peerId, peer_name: peerN, last_message: msg.contenido, last_time: msg.created_at, unread_count: unread })
  }
}

// ─── Helpers ───────────────────────────────────────────────────────────────
function formatTime(iso: string) {
  const d = new Date(iso)
  const now = new Date()
  const diffDays = Math.floor((now.getTime() - d.getTime()) / 86400000)
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
function isContinued(idx: number) {
  if (idx === 0) return false
  return msgs.value[idx - 1]?.emisor_id === msgs.value[idx]?.emisor_id
}
function isLastInGroup(idx: number) {
  return msgs.value[idx + 1]?.emisor_id !== msgs.value[idx]?.emisor_id
}

// ─── API ───────────────────────────────────────────────────────────────────
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
  hasMore.value = false
  peerName.value = ''
  try {
    if ((history.state as { peerName?: string })?.peerName) {
      peerName.value = (history.state as { peerName: string }).peerName
    }
    if (!peerName.value) {
      try {
        const perfil = await api.get(`/usuarios/${peerId}/perfil`)
        peerName.value = perfil.data?.user?.name ?? ''
      } catch { /* ignorar */ }
    }
    const res = await api.get(`/mensajes/${peerId}`, { params: { limit: 50 } })
    msgs.value = (res.data?.mensajes ?? []).map((m: Mensaje) => ({ ...m, _status: 'sent' as const }))
    hasMore.value = res.data?.has_more ?? false
    const conv = convs.value.find(c => c.peer_id === peerId)
    if (conv) conv.unread_count = 0
    await scrollToBottom()
  } catch { /* silencioso */ } finally {
    loadingMsgs.value = false
  }
}

async function loadMoreMensajes() {
  if (!activePeerId.value || !hasMore.value || loadingMore.value || msgs.value.length === 0) return
  loadingMore.value = true
  const oldestId = msgs.value[0]?.id
  const prevScrollHeight = threadRef.value?.scrollHeight ?? 0
  try {
    const res = await api.get(`/mensajes/${activePeerId.value}`, { params: { limit: 50, before_id: oldestId } })
    const older: Mensaje[] = (res.data?.mensajes ?? []).map((m: Mensaje) => ({ ...m, _status: 'sent' as const }))
    hasMore.value = res.data?.has_more ?? false
    msgs.value = [...older, ...msgs.value]
    await nextTick()
    if (threadRef.value) threadRef.value.scrollTop = threadRef.value.scrollHeight - prevScrollHeight
  } catch { /* silencioso */ } finally {
    loadingMore.value = false
  }
}

async function sendMensaje() {
  if (!activePeerId.value || !newMsg.value.trim() || sending.value) return
  sending.value = true
  const text = newMsg.value.trim()
  newMsg.value = ''
  resetTextarea()

  const tempId = `tmp-${Date.now()}`
  const tempMsg: Mensaje = {
    id: tempId, _tempId: tempId, _status: 'sending',
    emisor_id: auth.user?.id ?? '', emisor_name: auth.user?.name ?? '',
    receptor_id: activePeerId.value, receptor_name: peerName.value,
    contenido: text, leido: false, created_at: new Date().toISOString(),
  }
  msgs.value.push(tempMsg)
  await scrollToBottom()

  try {
    const res = await api.post(`/mensajes/${activePeerId.value}`, { contenido: text, peer_name: peerName.value })
    const idx = msgs.value.findIndex(m => m._tempId === tempId)
    if (idx !== -1) msgs.value.splice(idx, 1, { ...res.data, _status: 'sent' as const })
    const conv = convs.value.find(c => c.peer_id === activePeerId.value)
    if (conv) {
      conv.last_message = text
      conv.last_time    = res.data.created_at
      convs.value = [conv, ...convs.value.filter(c => c.peer_id !== activePeerId.value)]
    } else {
      convs.value.unshift({ peer_id: activePeerId.value!, peer_name: peerName.value, last_message: text, last_time: res.data.created_at, unread_count: 0 })
    }
  } catch {
    const idx = msgs.value.findIndex(m => m._tempId === tempId)
    if (idx !== -1) msgs.value[idx]!._status = 'error'
    errorMsg.value = 'No se pudo enviar el mensaje'
    setTimeout(() => { errorMsg.value = '' }, 4000)
  } finally {
    sending.value = false
  }
}

async function retrySend(tempId: string) {
  const m = msgs.value.find(m => m._tempId === tempId)
  if (!m) return
  const text = m.contenido
  msgs.value = msgs.value.filter(x => x._tempId !== tempId)
  newMsg.value = text
  await sendMensaje()
}

async function scrollToBottom() {
  await nextTick()
  if (threadRef.value) threadRef.value.scrollTop = threadRef.value.scrollHeight
}

function resetTextarea() {
  nextTick(() => { if (textareaRef.value) textareaRef.value.style.height = 'auto' })
}

function autoResizeTextarea() {
  if (textareaRef.value) {
    textareaRef.value.style.height = 'auto'
    textareaRef.value.style.height = Math.min(textareaRef.value.scrollHeight, 120) + 'px'
  }
  sendTyping()
}

function openConversacion(conv: Conversacion) {
  peerName.value = conv.peer_name
  const base = auth.isInstructor ? '/instructor' : '/usuario'
  router.push(`${base}/mensajes/${conv.peer_id}`)
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMensaje() }
}

// ─── IntersectionObserver para paginación ──────────────────────────────────
let sentinelObserver: IntersectionObserver | null = null

function setupSentinel() {
  if (sentinelObserver) sentinelObserver.disconnect()
  sentinelObserver = new IntersectionObserver(
    (entries) => { if (entries[0]?.isIntersecting) loadMoreMensajes() },
    { threshold: 0.1 }
  )
  if (sentinelRef.value) sentinelObserver.observe(sentinelRef.value)
}

// ─── Watcher ───────────────────────────────────────────────────────────────
watch(activePeerId, async (peerId) => {
  if (peerId) {
    await loadMensajes(peerId)
    await nextTick()
    setupSentinel()
  } else {
    msgs.value = []
    peerName.value = ''
    hasMore.value = false
  }
}, { immediate: true })

// ─── Lifecycle ─────────────────────────────────────────────────────────────
onMounted(async () => {
  await loadConversaciones()
  connectWs()
})

onUnmounted(() => {
  disconnectWs()
  sentinelObserver?.disconnect()
  if (typingHideTimer) clearTimeout(typingHideTimer)
})
</script>

<template>
  <div class="mensajes-shell">

    <!-- ── Panel izquierdo ──────────────────────────────────────────────── -->
    <aside :class="['convs-panel', activePeerId ? 'hidden-mobile' : '']">
      <div class="convs-header"><h2>Mensajes</h2></div>

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
          v-for="conv in convs" :key="conv.peer_id"
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

    <!-- ── Panel derecho ────────────────────────────────────────────────── -->
    <section :class="['thread-panel', !activePeerId ? 'hidden-mobile' : '']">

      <div v-if="!activePeerId" class="thread-empty">
        <svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.2" viewBox="0 0 24 24">
          <path d="M8 10h.01M12 10h.01M16 10h.01M21 12c0 4.418-4.03 8-9 8a9.862 9.862 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
        </svg>
        <p>Selecciona una conversación para leer tus mensajes</p>
      </div>

      <template v-else>
        <!-- Header -->
        <div class="thread-header">
          <button class="back-btn" @click="router.back()">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M15 19l-7-7 7-7"/></svg>
          </button>
          <div class="thread-avatar">{{ initials(peerName) }}</div>
          <span class="thread-peername">{{ peerName || '...' }}</span>
        </div>

        <!-- Toast de error -->
        <Transition name="toast">
          <div v-if="errorMsg" class="error-toast">{{ errorMsg }}</div>
        </Transition>

        <!-- Hilo -->
        <div ref="threadRef" class="thread-body">
          <!-- Sentinel de paginación -->
          <div ref="sentinelRef" class="load-sentinel">
            <span v-if="loadingMore" class="spinner spinner-sm"></span>
          </div>

          <div v-if="loadingMsgs" class="thread-loading"><span class="spinner"></span></div>
          <div v-else-if="msgs.length === 0" class="thread-no-msgs"><p>Sé el primero en escribir 👋</p></div>

          <TransitionGroup v-else name="list" tag="div" class="msgs-list">
            <template v-for="(msg, idx) in msgs" :key="msg._tempId ?? msg.id">
              <!-- Separador de fecha -->
              <div
                v-if="idx === 0 || !isSameDay(msgs[idx - 1]!.created_at, msg.created_at)"
                class="date-sep"
              >
                <span>{{ formatDateSep(msg.created_at) }}</span>
              </div>

              <!-- Burbuja -->
              <div :class="[
                'bubble-wrap',
                msg.emisor_id === auth.user?.id ? 'mine' : 'theirs',
                isContinued(idx) ? 'continued' : '',
                isLastInGroup(idx) ? 'last-in-group' : '',
                msg._status === 'error' ? 'has-error' : '',
              ]">
                <div class="bubble">
                  <p>{{ msg.contenido }}</p>
                  <span class="bubble-meta">
                    <span class="bubble-time">{{ formatTime(msg.created_at) }}</span>
                    <span v-if="msg.emisor_id === auth.user?.id" class="status-icon">
                      <span v-if="msg._status === 'sending'" title="Enviando">⏱</span>
                      <span v-else-if="msg.leido" class="read" title="Leído">✓✓</span>
                      <span v-else title="Enviado">✓</span>
                    </span>
                  </span>
                  <button
                    v-if="msg._status === 'error' && msg._tempId"
                    class="retry-btn"
                    @click="retrySend(msg._tempId)"
                  >Reintentar</button>
                </div>
              </div>
            </template>
          </TransitionGroup>

          <!-- Typing indicator -->
          <Transition name="fade">
            <div v-if="showTyping" class="typing-indicator">
              <div class="theirs-bubble">
                <span class="typing-dots"><span></span><span></span><span></span></span>
              </div>
            </div>
          </Transition>
        </div>

        <!-- Input -->
        <form class="thread-input" @submit.prevent="sendMensaje">
          <textarea
            ref="textareaRef"
            v-model="newMsg"
            rows="1"
            placeholder="Escribe un mensaje…"
            :disabled="sending"
            @keydown="handleKeydown"
            @input="autoResizeTextarea"
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
/* ── Layout ────────────────────────────────────────────────────────────── */
.mensajes-shell { display: flex; height: 100%; overflow: hidden; background: var(--surface); }

/* ── Panel izquierdo ────────────────────────────────────────────────────── */
.convs-panel {
  width: 320px; min-width: 260px;
  border-right: 1px solid var(--border);
  display: flex; flex-direction: column;
  background: var(--surface); overflow: hidden;
}
.convs-header {
  padding: 1.25rem 1.25rem 1rem;
  border-bottom: 1px solid var(--border);
}
.convs-header h2 { font-size: 1.15rem; font-weight: 700; margin: 0; color: var(--text); }
.convs-loading, .convs-empty {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  gap: .75rem; color: var(--text-muted); font-size: .9rem; padding: 2rem;
}
.convs-empty svg { opacity: .35; }
.convs-list { flex: 1; overflow-y: auto; list-style: none; margin: 0; padding: 0; }
.conv-item {
  display: flex; align-items: center; gap: .75rem;
  padding: .85rem 1.25rem; cursor: pointer;
  border-bottom: 1px solid var(--border); transition: background .15s;
}
.conv-item:hover { background: var(--surface-hover); }
.conv-item.active { background: var(--primary-soft, rgba(59,130,246,.08)); }
.conv-avatar {
  width: 42px; height: 42px; border-radius: 50%;
  background: var(--primary, #3b82f6); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: .8rem; flex-shrink: 0;
}
.conv-info { flex: 1; min-width: 0; }
.conv-row { display: flex; align-items: center; justify-content: space-between; gap: .5rem; }
.conv-name { font-weight: 600; font-size: .9rem; color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.conv-time { font-size: .75rem; color: var(--text-muted); white-space: nowrap; flex-shrink: 0; }
.conv-preview { font-size: .82rem; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.conv-badge {
  background: var(--primary, #3b82f6); color: #fff;
  font-size: .7rem; font-weight: 700;
  border-radius: 999px; padding: .1rem .45rem;
  min-width: 18px; text-align: center; flex-shrink: 0;
}

/* ── Panel derecho ──────────────────────────────────────────────────────── */
.thread-panel { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-width: 0; position: relative; }
.thread-empty {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  gap: 1rem; color: var(--text-muted); font-size: .95rem;
  padding: 2rem; text-align: center;
}
.thread-empty svg { opacity: .3; }

.thread-header {
  display: flex; align-items: center; gap: .75rem;
  padding: .9rem 1.25rem;
  border-bottom: 1px solid var(--border);
  background: var(--surface);
}
.back-btn {
  background: none; border: none; cursor: pointer;
  color: var(--text-muted); padding: .25rem;
  border-radius: .4rem; display: flex; align-items: center; transition: color .15s;
}
.back-btn:hover { color: var(--text); }
.thread-avatar {
  width: 36px; height: 36px; border-radius: 50%;
  background: var(--primary, #3b82f6); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: .75rem; flex-shrink: 0;
}
.thread-peername { font-weight: 600; font-size: 1rem; color: var(--text); }

.error-toast {
  position: absolute; top: 60px; left: 50%; transform: translateX(-50%);
  background: #ef4444; color: #fff;
  padding: .45rem 1rem; border-radius: .5rem;
  font-size: .85rem; z-index: 20; white-space: nowrap;
}
.toast-enter-active, .toast-leave-active { transition: opacity .3s, transform .3s; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateX(-50%) translateY(-8px); }

.thread-body {
  flex: 1; overflow-y: auto;
  padding: .5rem 1.25rem .5rem;
  display: flex; flex-direction: column; gap: .1rem;
}
.thread-loading, .thread-no-msgs {
  flex: 1; display: flex; align-items: center; justify-content: center;
  color: var(--text-muted); font-size: .9rem;
}

.load-sentinel { display: flex; justify-content: center; min-height: 24px; padding: .25rem 0; }

.date-sep { text-align: center; margin: .75rem 0; }
.date-sep span {
  background: var(--surface-soft); color: var(--text-muted);
  font-size: .72rem; padding: .25rem .75rem; border-radius: 999px;
}

.msgs-list { display: flex; flex-direction: column; gap: .15rem; }

/* Burbujas */
.bubble-wrap { display: flex; margin-bottom: .05rem; }
.bubble-wrap.mine    { justify-content: flex-end; }
.bubble-wrap.theirs  { justify-content: flex-start; }
.bubble-wrap.continued  { margin-top: .05rem; }
.bubble-wrap.last-in-group { margin-bottom: .4rem; }

.bubble {
  max-width: 72%; padding: .55rem .85rem;
  border-radius: 1.2rem; font-size: .9rem;
  line-height: 1.5; word-break: break-word;
}
.bubble p { margin: 0 0 .2rem; }
.bubble-meta { display: flex; align-items: center; justify-content: flex-end; gap: .3rem; }
.bubble-time { font-size: .68rem; color: rgba(255,255,255,.65); }
.status-icon { font-size: .7rem; }
.status-icon .read { color: #93c5fd; }

.bubble-wrap.mine .bubble {
  background: var(--primary, #3b82f6); color: #fff;
  border-bottom-right-radius: .3rem;
}
.bubble-wrap.mine.continued .bubble { border-top-right-radius: .4rem; }
.bubble-wrap.theirs .bubble {
  background: var(--surface-soft); color: var(--text);
  border-bottom-left-radius: .3rem;
}
.bubble-wrap.theirs.continued .bubble { border-top-left-radius: .4rem; }
.bubble-wrap.theirs .bubble-time, .bubble-wrap.theirs .status-icon { color: var(--text-muted); }

.bubble-wrap.has-error .bubble { opacity: .7; }
.retry-btn {
  display: block; margin-top: .35rem;
  background: rgba(255,255,255,.2); border: 1px solid rgba(255,255,255,.4);
  border-radius: .4rem; color: #fff; font-size: .75rem;
  padding: .2rem .6rem; cursor: pointer; transition: background .15s;
}
.retry-btn:hover { background: rgba(255,255,255,.35); }

/* Typing indicator */
.typing-indicator { display: flex; justify-content: flex-start; padding: .25rem 0; }
.theirs-bubble {
  background: var(--surface-soft);
  padding: .5rem .8rem; border-radius: 1.2rem; border-bottom-left-radius: .3rem;
  display: inline-flex; align-items: center;
}
.typing-dots { display: flex; gap: .25rem; align-items: center; }
.typing-dots span {
  width: 6px; height: 6px; border-radius: 50%;
  background: var(--text-muted);
  animation: typingBounce 1.2s infinite ease-in-out;
}
.typing-dots span:nth-child(2) { animation-delay: .2s; }
.typing-dots span:nth-child(3) { animation-delay: .4s; }
@keyframes typingBounce {
  0%, 60%, 100% { transform: translateY(0); }
  30%            { transform: translateY(-5px); }
}

/* Transiciones */
.list-enter-active { transition: all .3s ease; }
.list-leave-active { transition: all .2s ease; }
.list-enter-from   { opacity: 0; transform: translateY(15px); }
.list-leave-to     { opacity: 0; transform: translateY(5px); }

.fade-enter-active, .fade-leave-active { transition: opacity .25s; }
.fade-enter-from, .fade-leave-to       { opacity: 0; }

/* Input */
.thread-input {
  display: flex; align-items: flex-end; gap: .6rem;
  padding: .75rem 1.25rem 1rem;
  border-top: 1px solid var(--border);
  background: var(--surface);
}
.thread-input textarea {
  flex: 1; resize: none;
  border: 1.5px solid var(--border); border-radius: .75rem;
  padding: .6rem .9rem; font-size: .9rem; font-family: inherit;
  background: var(--surface-soft); color: var(--text);
  outline: none; max-height: 120px; overflow-y: auto;
  line-height: 1.5; transition: border-color .2s;
}
.thread-input textarea:focus { border-color: var(--primary, #3b82f6); }
.thread-input textarea:disabled { opacity: .5; cursor: not-allowed; }
.thread-input button[type="submit"] {
  width: 42px; height: 42px; border-radius: 50%;
  background: var(--primary, #3b82f6); color: #fff;
  border: none; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; transition: opacity .2s, transform .1s;
}
.thread-input button[type="submit"]:hover  { opacity: .85; }
.thread-input button[type="submit"]:active { transform: scale(.92); }
.thread-input button[type="submit"]:disabled { opacity: .4; cursor: not-allowed; }

/* Spinner */
.spinner {
  display: inline-block; width: 28px; height: 28px;
  border: 3px solid var(--border); border-top-color: var(--primary, #3b82f6);
  border-radius: 50%; animation: spin .7s linear infinite;
}
.spinner-sm { width: 18px; height: 18px; border-width: 2px; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Responsive */
@media (max-width: 640px) {
  .convs-panel { width: 100%; border-right: none; }
  .hidden-mobile { display: none; }
}
</style>
