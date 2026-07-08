<script setup lang="ts">
/**
 * InteractiveActivity.vue — Minijuegos de aprendizaje (versión premium)
 *   5 - Memorama      6 - Clasificar   7 - Sopa de Letras
 *   8 - Completar Espacios             9 - Ordenar Secuencia
 */
import { ref, watch, computed, onUnmounted, reactive, nextTick } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'

const props = defineProps<{ lesson: any; cursoId: string }>()
const emit = defineEmits<{ (e: 'completed', data: { points: number; timeSecs: number }): void }>()

// ── Tipo del juego ──────────────────────────────────────────────────────────
const gameType = computed(() => {
  const lt = props.lesson?.lesson_type
  const t  = props.lesson?.type
  if (lt !== undefined && lt !== null) return String(lt)
  if (t) {
    const MAP: Record<string, string> = {
      LESSON_TYPE_GAME_MEMORY: '5', LESSON_TYPE_GAME_DRAGDROP: '6',
      LESSON_TYPE_GAME_WORDSEARCH: '7', LESSON_TYPE_GAME_FILLBLANK: '8',
      LESSON_TYPE_GAME_ORDER: '9',
    }
    return MAP[t] ?? String(t)
  }
  return '1'
})

const GAME_META: Record<string, { label: string; icon: string; gradient: string; accent: string }> = {
  '5': { label: 'Memorama',          icon: '🎴', gradient: 'linear-gradient(135deg,#6366f1,#8b5cf6)', accent: '#8b5cf6' },
  '6': { label: 'Clasificar',         icon: '🗂️', gradient: 'linear-gradient(135deg,#0ea5e9,#06b6d4)', accent: '#0ea5e9' },
  '7': { label: 'Sopa de Letras',     icon: '🔤', gradient: 'linear-gradient(135deg,#f59e0b,#f97316)', accent: '#f97316' },
  '8': { label: 'Completar Espacios', icon: '✍️', gradient: 'linear-gradient(135deg,#10b981,#059669)', accent: '#10b981' },
  '9': { label: 'Ordenar',            icon: '🔢', gradient: 'linear-gradient(135deg,#ec4899,#f43f5e)', accent: '#ec4899' },
}
const meta = computed(() => GAME_META[gameType.value] ?? { label: 'Juego', icon: '🎮', gradient: 'linear-gradient(135deg,#6366f1,#8b5cf6)', accent: '#6366f1' })

// ── Estado global ───────────────────────────────────────────────────────────
const config       = ref<any>({})
const isCompleted  = ref(false)
const pointsEarned = ref(0)
const showConfetti = ref(false)
const startTime    = ref(0)
const elapsedSecs  = ref(0)
let timerInterval: ReturnType<typeof setInterval> | null = null

function startTimer() {
  if (timerInterval) clearInterval(timerInterval)
  startTime.value = Date.now(); elapsedSecs.value = 0
  timerInterval = setInterval(() => { elapsedSecs.value = Math.floor((Date.now() - startTime.value) / 1000) }, 1000)
}
function stopTimer() { if (timerInterval) { clearInterval(timerInterval); timerInterval = null } }
onUnmounted(() => stopTimer())

// ── Audio ───────────────────────────────────────────────────────────────────
function beep(freq = 520, type: OscillatorType = 'sine', dur = 0.15) {
  try {
    const ctx = new (window.AudioContext || (window as any).webkitAudioContext)()
    const osc = ctx.createOscillator(); const gain = ctx.createGain()
    osc.type = type; osc.frequency.setValueAtTime(freq, ctx.currentTime)
    gain.gain.setValueAtTime(0.12, ctx.currentTime)
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + dur)
    osc.connect(gain); gain.connect(ctx.destination); osc.start(); osc.stop(ctx.currentTime + dur)
  } catch {}
}
const soundOk  = () => { beep(660,'sine',0.1); setTimeout(() => beep(880,'triangle',0.2), 100) }
const soundBad = () => { beep(220,'sawtooth',0.2) }
const soundWin = () => { soundOk(); setTimeout(() => beep(990,'triangle',0.3),250); setTimeout(() => beep(1320,'sine',0.5),450) }

// ── Confetti ────────────────────────────────────────────────────────────────
const confettiPieces = ref<{ id: number; x: number; color: string; delay: number; size: number; dur: number }[]>([])
function launchConfetti() {
  const colors = ['#6366f1','#f59e0b','#10b981','#ec4899','#0ea5e9','#f97316','#8b5cf6','#fbbf24']
  confettiPieces.value = Array.from({ length: 48 }, (_, i) => ({
    id: i,
    x: Math.random() * 100,
    color: colors[i % colors.length]!,
    delay: Math.random() * 0.8,
    size: 6 + Math.random() * 8,
    dur: 1.8 + Math.random() * 1.2,
  }))
  showConfetti.value = true
  setTimeout(() => { showConfetti.value = false; confettiPieces.value = [] }, 3500)
}

// ── loadGame ────────────────────────────────────────────────────────────────
function loadGame() {
  isCompleted.value = false; pointsEarned.value = 0; stopTimer()
  let parsed: any = {}
  try {
    const raw = props.lesson?.game_config_json
    if (raw) parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
  } catch {}
  config.value = parsed
  initSpecificGame(); startTimer()
}
watch(() => [props.lesson?.id, props.lesson?.lesson_type, props.lesson?.type, props.lesson?.game_config_json], loadGame, { immediate: true })

async function handleWin() {
  if (isCompleted.value) return
  isCompleted.value = true; stopTimer(); soundWin()
  const pts = Number(props.lesson?.points_reward || 100); pointsEarned.value = pts
  await nextTick(); launchConfetti()
  try {
    if (!props.lesson?.completada) {
      await api.post(`/lecciones/${props.lesson.id}/game-score`, { curso_id: props.cursoId, points: pts, time_secs: elapsedSecs.value })
      toast.success(`¡+${pts} puntos ganados!`)
    }
  } catch {}
  emit('completed', { points: pts, timeSecs: elapsedSecs.value })
}

function restartGame() { isCompleted.value = false; pointsEarned.value = 0; stopTimer(); initSpecificGame(); startTimer() }

// ═══════════════════════════════════════════════════════════════════════════
//  5: MEMORAMA
// ═══════════════════════════════════════════════════════════════════════════
const PAIR_COLORS = ['#6366f1','#f59e0b','#10b981','#ec4899','#0ea5e9','#f97316','#8b5cf6','#ef4444','#14b8a6','#84cc16']
const memoCards   = ref<any[]>([])
const memoFlipped = ref<number[]>([])
const memoMatched = ref<number[]>([])
const memoBounce  = ref<number[]>([])   // pairIds para animación bounce on match
let memoLocked = false

function initMemo() {
  memoFlipped.value = []; memoMatched.value = []; memoBounce.value = []; memoLocked = false
  const pairs = config.value?.pairs?.length ? config.value.pairs : [
    { front: 'HTML', back: 'Estructura web' }, { front: 'CSS', back: 'Estilos' },
    { front: 'JS',   back: 'Interactividad' }, { front: 'Vue', back: 'Reactividad' },
  ]
  const deck: any[] = []
  pairs.forEach((p: any, idx: number) => {
    const color = PAIR_COLORS[idx % PAIR_COLORS.length]!
    deck.push({ id: idx*2,   pairId: idx, text: p.front || p.text_a || '', img: p.front_img || p.img_a || '', color })
    deck.push({ id: idx*2+1, pairId: idx, text: p.back  || p.text_b || '', img: p.back_img  || p.img_b || '', color })
  })
  for (let i = deck.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [deck[i], deck[j]] = [deck[j]!, deck[i]!]
  }
  memoCards.value = deck
}

function flipCard(card: any) {
  if (isCompleted.value || memoLocked || memoMatched.value.includes(card.pairId) || memoFlipped.value.includes(card.id) || memoFlipped.value.length >= 2) return
  beep(440, 'sine', 0.08)
  memoFlipped.value = [...memoFlipped.value, card.id]
  if (memoFlipped.value.length === 2) {
    memoLocked = true
    const c1 = memoCards.value.find(c => c.id === memoFlipped.value[0])
    const c2 = memoCards.value.find(c => c.id === memoFlipped.value[1])
    if (c1 && c2 && c1.pairId === c2.pairId) {
      setTimeout(() => {
        memoMatched.value = [...memoMatched.value, c1.pairId]
        memoBounce.value  = [...memoBounce.value,  c1.pairId]
        memoFlipped.value = []; memoLocked = false; soundOk()
        setTimeout(() => { memoBounce.value = memoBounce.value.filter(id => id !== c1.pairId) }, 700)
        if (memoMatched.value.length >= memoCards.value.length / 2) handleWin()
      }, 500)
    } else {
      setTimeout(() => { soundBad(); memoFlipped.value = []; memoLocked = false }, 950)
    }
  }
}

// ═══════════════════════════════════════════════════════════════════════════
//  6: CLASIFICAR
// ═══════════════════════════════════════════════════════════════════════════
const CAT_COLORS  = ['#6366f1','#f59e0b','#10b981','#ec4899','#0ea5e9','#f97316']
const dragCategories = ref<string[]>([])
const dragItems      = ref<any[]>([])
const dragAssignments = reactive<Record<string, string>>({})
const selectedDrag   = ref<string | null>(null)
const shakeItems     = ref<string[]>([])

function initDrag() {
  dragCategories.value = config.value?.categories?.length ? config.value.categories : ['Opción A', 'Opción B']
  const raw = config.value?.items?.length ? config.value.items : [
    { text: 'Vue 3', correct_category: 'Opción A' }, { text: 'CSS3', correct_category: 'Opción A' },
    { text: 'Go',    correct_category: 'Opción B' }, { text: 'SQL',  correct_category: 'Opción B' },
  ]
  dragItems.value = raw.map((it: any, i: number) => ({ ...it, id: String(it.id ?? i), category: it.correct_category || it.category || dragCategories.value[0] })).sort(() => Math.random() - 0.5)
  Object.keys(dragAssignments).forEach(k => delete dragAssignments[k])
  selectedDrag.value = null
}

function pickItem(id: string) { if (!isCompleted.value) { selectedDrag.value = selectedDrag.value === id ? null : id; beep(500,'sine',0.05) } }
function dropInCategory(cat: string) {
  if (!selectedDrag.value || isCompleted.value) return
  dragAssignments[selectedDrag.value] = cat; selectedDrag.value = null; beep(580,'sine',0.1)
}
function removeAssignment(id: string) { delete dragAssignments[id] }

function checkDrag() {
  if (Object.keys(dragAssignments).length < dragItems.value.length) { toast.error('Falta clasificar algunos elementos'); return }
  const wrong = dragItems.value.filter(it => dragAssignments[it.id] !== (it.correct_category || it.category))
  if (wrong.length === 0) {
    handleWin()
  } else {
    soundBad(); shakeItems.value = wrong.map((it: any) => it.id)
    toast.error(`${wrong.length} elemento${wrong.length > 1 ? 's' : ''} incorrectos — intenta de nuevo`)
    setTimeout(() => { shakeItems.value = [] }, 700)
  }
}

// ═══════════════════════════════════════════════════════════════════════════
//  7: SOPA DE LETRAS
// ═══════════════════════════════════════════════════════════════════════════
const WS_WORD_COLORS = ['#6366f1','#f59e0b','#10b981','#ec4899','#0ea5e9','#f97316','#8b5cf6','#ef4444']
const wordGrid      = ref<string[][]>([])
const wordsToFind   = ref<string[]>([])
const foundWords    = ref<string[]>([])
const wsFoundCells  = ref<Record<string, string>>({}) // key → color
const wsSelecting   = ref(false)
const wsStart       = ref<[number, number] | null>(null)
const wsEnd         = ref<[number, number] | null>(null)

function initWordSearch() {
  foundWords.value = []; wsFoundCells.value = {}; wsSelecting.value = false; wsStart.value = null; wsEnd.value = null
  let raw: string[] = []
  if (config.value?.words?.length) raw = Array.isArray(config.value.words) ? config.value.words : String(config.value.words).split(',')
  const valid = raw.map((w: string) => w.trim().normalize('NFD').replace(/[\u0300-\u036f]/g,'').replace(/[^a-zA-Z]/g,'').toUpperCase()).filter(w => w.length >= 2)
  wordsToFind.value = valid.length ? valid : ['VUE','GOLANG','GRPC','HTML','CSS']
  buildGrid()
}

function buildGrid() {
  const words = wordsToFind.value
  const maxLen = Math.max(...words.map(w => w.length), 5)
  const size = Math.max(Number(config.value?.grid_size) || 12, maxLen + 2, 10)
  const grid: string[][] = Array.from({ length: size }, () => Array(size).fill(''))
  const DIRS: [number, number][] = [[0,1],[1,0],[1,1],[0,-1],[-1,0],[-1,-1],[1,-1],[-1,1]]
  words.forEach(word => {
    let placed = false
    for (let a = 0; a < 300 && !placed; a++) {
      const dir = DIRS[Math.floor(Math.random() * DIRS.length)]!; const dr = dir[0]; const dc = dir[1]
      const r = Math.floor(Math.random() * size); const c = Math.floor(Math.random() * size)
      const eR = r + dr * (word.length - 1); const eC = c + dc * (word.length - 1)
      if (eR < 0 || eR >= size || eC < 0 || eC >= size) continue
      let ok = true
      for (let i = 0; i < word.length && ok; i++) { const cell = grid[r+dr*i]?.[c+dc*i]; if (cell !== '' && cell !== word[i]) ok = false }
      if (ok) { for (let i = 0; i < word.length; i++) grid[r+dr*i]![c+dc*i] = word[i]!; placed = true }
    }
  })
  const AL = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  for (let r = 0; r < size; r++) for (let c = 0; c < size; c++) if (!grid[r]![c]) grid[r]![c] = AL[Math.floor(Math.random() * AL.length)]!
  wordGrid.value = grid
}

function lineCells(s: [number,number], e: [number,number]): string[] {
  const [r1,c1] = s; const [r2,c2] = e
  const dr = Math.sign(r2-r1); const dc = Math.sign(c2-c1)
  if (r1!==r2 && c1!==c2 && Math.abs(r2-r1)!==Math.abs(c2-c1)) return []
  const steps = Math.max(Math.abs(r2-r1), Math.abs(c2-c1))
  return Array.from({ length: steps+1 }, (_,i) => `${r1+dr*i},${c1+dc*i}`)
}

const wsPreview = computed(() => (wsSelecting.value && wsStart.value && wsEnd.value) ? lineCells(wsStart.value, wsEnd.value) : [])

function wsDown(r: number, c: number) { if (isCompleted.value) return; wsSelecting.value = true; wsStart.value = [r,c]; wsEnd.value = [r,c] }
function wsEnter(r: number, c: number) { if (wsSelecting.value) wsEnd.value = [r,c] }
function wsUp() {
  if (!wsSelecting.value || !wsStart.value || !wsEnd.value) { wsSelecting.value = false; return }
  const cells = lineCells(wsStart.value, wsEnd.value)
  if (cells.length >= 2) {
    const letters = cells.map(k => { const [r,c] = k.split(',').map(Number); return wordGrid.value[r!]?.[c!] ?? '' }).join('')
    const rev = letters.split('').reverse().join('')
    const found = wordsToFind.value.find(w => !foundWords.value.includes(w) && (letters === w || rev === w))
    if (found) {
      const color = WS_WORD_COLORS[foundWords.value.length % WS_WORD_COLORS.length]!
      foundWords.value = [...foundWords.value, found]
      cells.forEach(k => { wsFoundCells.value[k] = color }); soundOk()
      if (foundWords.value.length >= wordsToFind.value.length) handleWin()
    } else soundBad()
  }
  wsSelecting.value = false; wsStart.value = null; wsEnd.value = null
}

function wsCellStyle(r: number, c: number): Record<string, string> {
  const key = `${r},${c}`
  if (wsFoundCells.value[key]) return { background: wsFoundCells.value[key]!, color: 'white', borderColor: wsFoundCells.value[key]! }
  if (wsPreview.value.includes(key)) return { background: 'rgba(99,102,241,0.3)', borderColor: '#6366f1' }
  return {}
}

// ═══════════════════════════════════════════════════════════════════════════
//  8: COMPLETAR ESPACIOS
// ═══════════════════════════════════════════════════════════════════════════
const fbSentences   = ref<any[]>([])
const fbAnswers     = ref<Record<number, string>>({})
const fbResult      = ref<Record<number, boolean | null>>({})
const fbChecked     = ref(false)

function initFillBlank() {
  fbAnswers.value = {}; fbResult.value = {}; fbChecked.value = false
  fbSentences.value = config.value?.sentences?.length ? config.value.sentences : [
    { text: 'Para estilizar una página web usamos ___', answer: 'CSS', options: ['CSS', 'Python', 'SQL'] },
    { text: 'El framework reactivo para JS es ___', answer: 'Vue 3', options: ['Vue 3', 'Django', 'Laravel'] },
  ]
}

function checkFillBlank() {
  if (fbSentences.value.some((_s: any, i: number) => !fbAnswers.value[i])) { toast.error('Responde todas las preguntas'); return }
  fbChecked.value = true
  const results: Record<number, boolean> = {}
  fbSentences.value.forEach((s: any, i: number) => {
    results[i] = (fbAnswers.value[i] ?? '').trim().toLowerCase() === (s.answer ?? '').trim().toLowerCase()
  })
  fbResult.value = results
  const allOk = Object.values(results).every(v => v)
  if (allOk) { handleWin() } else { soundBad(); toast.error('Algunas respuestas son incorrectas') }
}

// ═══════════════════════════════════════════════════════════════════════════
//  9: ORDENAR SECUENCIA
// ═══════════════════════════════════════════════════════════════════════════
const orderItems    = ref<any[]>([])
const orderChecked  = ref(false)
const orderResult   = ref(false)

function initOrder() {
  orderChecked.value = false; orderResult.value = false
  const raw = config.value?.items?.length ? config.value.items : [
    { text: 'Diseñar wireframes', correct_order: 1 },
    { text: 'Configurar base de datos', correct_order: 2 },
    { text: 'Desarrollar backend', correct_order: 3 },
    { text: 'Integrar frontend', correct_order: 4 },
  ]
  orderItems.value = raw.map((it: any, i: number) => ({ ...it, _id: i })).sort(() => Math.random() - 0.5)
}

function moveItem(idx: number, dir: -1|1) {
  if (isCompleted.value) return
  const j = idx + dir
  if (j < 0 || j >= orderItems.value.length) return
  const arr = [...orderItems.value];
  [arr[idx]!, arr[j]!] = [arr[j]!, arr[idx]!]; orderItems.value = arr; beep(500,'triangle',0.08)
}

function checkOrder() {
  orderChecked.value = true
  const ok = orderItems.value.every((it, i) => it.correct_order === i + 1)
  orderResult.value = ok
  if (ok) { handleWin() } else { soundBad(); toast.error('El orden no es correcto aún') }
}

// ── Dispatcher ──────────────────────────────────────────────────────────────
function initSpecificGame() {
  const t = gameType.value
  if      (t==='5') initMemo()
  else if (t==='6') initDrag()
  else if (t==='7') initWordSearch()
  else if (t==='8') initFillBlank()
  else if (t==='9') initOrder()
}

const fmt = (s: number) => s < 60 ? `${s}s` : `${Math.floor(s/60)}m ${s%60}s`
</script>

<template>
  <div class="ia-root" :style="{ '--game-accent': meta.accent, '--game-gradient': meta.gradient }">

    <!-- ── Confetti ───────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showConfetti" class="confetti-overlay" aria-hidden="true">
        <div
          v-for="p in confettiPieces" :key="p.id"
          class="confetti-piece"
          :style="{ left: `${p.x}%`, background: p.color, width: `${p.size}px`, height: `${p.size * 1.3}px`, animationDelay: `${p.delay}s`, animationDuration: `${p.dur}s` }"
        />
      </div>
    </Teleport>

    <!-- ── Game Header ───────────────────────────────────────────────────── -->
    <div class="ia-header">
      <div class="ia-header-left">
        <div class="ia-badge" :style="{ background: meta.gradient }">
          <span class="ia-badge-icon">{{ meta.icon }}</span>
          <span>{{ meta.label }}</span>
        </div>
        <h2 class="ia-title">{{ lesson?.title }}</h2>
        <p v-if="lesson?.description" class="ia-desc">{{ lesson.description }}</p>
      </div>
      <div class="ia-stats">
        <div class="ia-stat">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          <span>{{ fmt(elapsedSecs) }}</span>
        </div>
        <div class="ia-stat ia-stat-pts">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="0.5"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
          <span>{{ lesson?.points_reward || 100 }} pts</span>
        </div>
      </div>
    </div>

    <!-- ── Win Banner ────────────────────────────────────────────────────── -->
    <transition name="pop">
      <div v-if="isCompleted" class="ia-win">
        <div class="win-trophy">🏆</div>
        <div class="win-body">
          <h3>¡Reto Superado!</h3>
          <p>Completado en <strong>{{ fmt(elapsedSecs) }}</strong> · ganaste <strong>+{{ pointsEarned }} pts</strong></p>
        </div>
        <button class="ia-btn-replay" @click="restartGame">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
          Jugar de nuevo
        </button>
      </div>
    </transition>

    <!-- ════════════════════════════════════════════════════════════════════
         5: MEMORAMA
    ═════════════════════════════════════════════════════════════════════ -->
    <div v-if="gameType === '5'" class="game-wrap">
      <div class="game-topbar">
        <span class="topbar-hint">Encuentra todos los <strong>{{ memoCards.length / 2 }}</strong> pares</span>
        <div class="topbar-progress">
          <span class="prog-badge" :style="{ background: meta.gradient }">{{ memoMatched.length }}/{{ memoCards.length / 2 }}</span>
          <div class="prog-track">
            <div class="prog-fill" :style="{ width: `${memoCards.length > 0 ? (memoMatched.length / (memoCards.length / 2)) * 100 : 0}%`, background: meta.gradient }" />
          </div>
        </div>
      </div>
      <div class="memo-grid">
        <button
          v-for="card in memoCards" :key="card.id"
          :class="['memo-card',
            (memoFlipped.includes(card.id) || memoMatched.includes(card.pairId)) ? 'flipped' : '',
            memoMatched.includes(card.pairId) ? 'matched' : '',
            memoBounce.includes(card.pairId) ? 'bounce' : '']"
          :disabled="memoMatched.includes(card.pairId) || isCompleted"
          @click="flipCard(card)"
        >
          <div class="card-inner">
            <!-- Dorso -->
            <div class="card-front" :style="{ background: meta.gradient }">
              <span class="card-front-icon">{{ meta.icon }}</span>
              <div class="card-shine" />
            </div>
            <!-- Cara -->
            <div class="card-back" :style="memoMatched.includes(card.pairId) ? { background: card.color + '18', borderColor: card.color } : {}">
              <div v-if="memoMatched.includes(card.pairId)" class="card-match-mark" :style="{ background: card.color }">✓</div>
              <img v-if="card.img" :src="card.img" class="card-img" :alt="card.text" />
              <span class="card-txt" :style="memoMatched.includes(card.pairId) ? { color: card.color } : {}">{{ card.text }}</span>
            </div>
          </div>
        </button>
      </div>
    </div>

    <!-- ════════════════════════════════════════════════════════════════════
         6: CLASIFICAR
    ═════════════════════════════════════════════════════════════════════ -->
    <div v-else-if="gameType === '6'" class="game-wrap">
      <p class="game-hint">Selecciona un elemento y luego la categoría donde pertenece</p>

      <!-- Pool de items -->
      <div class="drag-pool">
        <button
          v-for="it in dragItems" :key="it.id"
          :class="['drag-item',
            selectedDrag === it.id ? 'picked' : '',
            dragAssignments[it.id] ? 'done' : '',
            shakeItems.includes(it.id) ? 'shake' : '']"
          :disabled="isCompleted"
          @click="pickItem(it.id)"
        >
          <span class="drag-item-dot" :style="{ background: dragAssignments[it.id]
            ? CAT_COLORS[dragCategories.indexOf(dragAssignments[it.id]) % CAT_COLORS.length]
            : selectedDrag === it.id ? 'var(--game-accent)' : '#cbd5e1' }" />
          {{ it.text }}
          <span v-if="dragAssignments[it.id]" class="drag-item-tag" :style="{ background: CAT_COLORS[dragCategories.indexOf(dragAssignments[it.id]) % CAT_COLORS.length] + '22', color: CAT_COLORS[dragCategories.indexOf(dragAssignments[it.id]) % CAT_COLORS.length] }">
            {{ dragAssignments[it.id] }}
          </span>
        </button>
      </div>

      <!-- Categorías -->
      <div class="drag-cats">
        <div
          v-for="(cat, ci) in dragCategories" :key="cat"
          :class="['drag-cat', selectedDrag ? 'droppable' : '']"
          :style="{ '--cat-color': CAT_COLORS[ci % CAT_COLORS.length] }"
          @click="dropInCategory(cat)"
        >
          <div class="cat-header">
            <div class="cat-dot" />
            <span class="cat-name">{{ cat }}</span>
            <span class="cat-count">{{ dragItems.filter(i => dragAssignments[i.id] === cat).length }}</span>
          </div>
          <div class="cat-body">
            <div
              v-for="it in dragItems.filter(i => dragAssignments[i.id] === cat)" :key="it.id"
              class="cat-chip"
            >
              {{ it.text }}
              <button @click.stop="removeAssignment(it.id)" class="cat-chip-del">✕</button>
            </div>
            <span v-if="!dragItems.some(i => dragAssignments[i.id] === cat)" class="cat-empty">
              {{ selectedDrag ? '← Haz clic para asignar aquí' : 'Sin elementos' }}
            </span>
          </div>
        </div>
      </div>

      <div class="game-actions">
        <button class="ia-btn-primary" @click="checkDrag" :disabled="isCompleted">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
          Verificar Clasificación
        </button>
      </div>
    </div>

    <!-- ════════════════════════════════════════════════════════════════════
         7: SOPA DE LETRAS
    ═════════════════════════════════════════════════════════════════════ -->
    <div v-else-if="gameType === '7'" class="game-wrap ws-wrap" @mouseup="wsUp" @mouseleave="wsUp">
      <p class="game-hint">Arrastra sobre las letras para seleccionar las palabras ocultas</p>
      <div class="ws-layout">
        <!-- Grid -->
        <div class="ws-board-wrap">
          <div class="ws-board" :style="{ userSelect: 'none' }">
            <div v-for="(row, r) in wordGrid" :key="r" class="ws-row">
              <button
                v-for="(ch, c) in row" :key="c"
                class="ws-cell"
                :style="wsCellStyle(r, c)"
                :class="{ 'ws-preview': wsPreview.includes(`${r},${c}`) && !wsFoundCells[`${r},${c}`] }"
                @mousedown.prevent="wsDown(r, c)"
                @mouseenter="wsEnter(r, c)"
                @touchstart.prevent="wsDown(r, c)"
              >{{ ch }}</button>
            </div>
          </div>
        </div>

        <!-- Word list -->
        <div class="ws-panel">
          <div class="ws-panel-title">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            Palabras <span class="ws-count">{{ foundWords.length }}/{{ wordsToFind.length }}</span>
          </div>
          <div class="ws-progress-bar">
            <div class="ws-progress-fill" :style="{ width: `${(foundWords.length/wordsToFind.length)*100}%`, background: meta.gradient }" />
          </div>
          <ul class="ws-list">
            <li v-for="(w, wi) in wordsToFind" :key="w" :class="['ws-word', foundWords.includes(w) ? 'found' : '']"
              :style="foundWords.includes(w) ? { color: WS_WORD_COLORS[wi % WS_WORD_COLORS.length], textDecorationColor: WS_WORD_COLORS[wi % WS_WORD_COLORS.length] } : {}">
              <span class="ws-bullet" :style="{ background: foundWords.includes(w) ? WS_WORD_COLORS[wi % WS_WORD_COLORS.length] : '#e2e8f0' }" />
              {{ w }}
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- ════════════════════════════════════════════════════════════════════
         8: COMPLETAR ESPACIOS
    ═════════════════════════════════════════════════════════════════════ -->
    <div v-else-if="gameType === '8'" class="game-wrap">
      <p class="game-hint">Lee cada pregunta y selecciona la respuesta correcta</p>
      <div class="fb-list">
        <div v-for="(sent, idx) in fbSentences" :key="idx"
          :class="['fb-card',
            fbChecked && fbResult[idx] === true  ? 'fb-correct' : '',
            fbChecked && fbResult[idx] === false ? 'fb-wrong'   : '']">
          <div class="fb-num" :style="{ background: meta.gradient }">{{ idx + 1 }}</div>
          <div class="fb-body">
            <p class="fb-text">{{ sent.text }}</p>
            <div class="fb-opts">
              <label
                v-for="opt in sent.options" :key="opt"
                :class="['fb-opt', fbAnswers[idx] === opt ? 'fb-opt-sel' : '',
                  fbChecked && opt === sent.answer ? 'fb-opt-correct' : '',
                  fbChecked && fbAnswers[idx] === opt && opt !== sent.answer ? 'fb-opt-wrong' : '']"
              >
                <input type="radio" :name="`fb_${idx}`" :value="opt" v-model="fbAnswers[idx]" class="sr-only" :disabled="isCompleted" />
                <span class="fb-opt-dot" />
                {{ opt }}
                <span v-if="fbChecked && opt === sent.answer" class="fb-opt-check">✓</span>
                <span v-if="fbChecked && fbAnswers[idx] === opt && opt !== sent.answer" class="fb-opt-x">✗</span>
              </label>
            </div>
          </div>
        </div>
      </div>
      <div class="game-actions">
        <button class="ia-btn-primary" @click="checkFillBlank" :disabled="isCompleted">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
          Validar Respuestas
        </button>
      </div>
    </div>

    <!-- ════════════════════════════════════════════════════════════════════
         9: ORDENAR SECUENCIA
    ═════════════════════════════════════════════════════════════════════ -->
    <div v-else-if="gameType === '9'" class="game-wrap">
      <p class="game-hint">Usa las flechas para ordenar los pasos correctamente</p>
      <div class="order-list">
        <transition-group name="order-move">
          <div v-for="(it, idx) in orderItems" :key="it._id"
            :class="['order-card',
              orderChecked && it.correct_order === idx + 1 ? 'order-ok' : '',
              orderChecked && it.correct_order !== idx + 1 ? 'order-bad' : '']">
            <div class="order-num" :style="{ background: meta.gradient }">{{ idx + 1 }}</div>
            <span class="order-txt">{{ it.text }}</span>
            <div class="order-arrows">
              <button class="order-arrow" :disabled="idx === 0 || isCompleted" @click="moveItem(idx, -1)">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M18 15l-6-6-6 6"/></svg>
              </button>
              <button class="order-arrow" :disabled="idx === orderItems.length - 1 || isCompleted" @click="moveItem(idx, 1)">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M6 9l6 6 6-6"/></svg>
              </button>
            </div>
            <span v-if="orderChecked" class="order-result-icon">{{ it.correct_order === idx + 1 ? '✓' : '✗' }}</span>
          </div>
        </transition-group>
      </div>
      <div class="game-actions">
        <button class="ia-btn-primary" @click="checkOrder" :disabled="isCompleted">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
          Verificar Orden
        </button>
      </div>
    </div>

    <div v-else class="game-wrap">
      <p class="game-hint">Este juego aún no tiene configuración.</p>
    </div>
  </div>
</template>

<style scoped>
/* ─── Variables del tema ─────────────────────────────────────────────────── */
.ia-root {
  --accent: var(--game-accent, #6366f1);
  font-family: inherit;
}

/* ─── Confetti ───────────────────────────────────────────────────────────── */
.confetti-overlay {
  position: fixed; inset: 0; pointer-events: none; z-index: 9999; overflow: hidden;
}
.confetti-piece {
  position: absolute; top: -20px; border-radius: 3px;
  animation: confettiFall linear forwards;
}
@keyframes confettiFall {
  0%   { transform: translateY(0) rotate(0deg) scale(1); opacity: 1; }
  80%  { opacity: 1; }
  100% { transform: translateY(110vh) rotate(720deg) scale(0.4); opacity: 0; }
}

/* ─── Raíz del componente ───────────────────────────────────────────────── */
.ia-root {
  background: var(--surface, #fff);
  border-radius: 20px;
  border: 1px solid var(--border, #e2e8f0);
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(0,0,0,0.06);
}

/* ─── Header ────────────────────────────────────────────────────────────── */
.ia-header {
  display: flex; justify-content: space-between; align-items: flex-start; gap: 16px;
  padding: 24px 28px 0; flex-wrap: wrap;
}
.ia-header-left { display: flex; flex-direction: column; gap: 6px; }
.ia-badge {
  display: inline-flex; align-items: center; gap: 6px;
  color: white; font-size: 0.72rem; font-weight: 800; letter-spacing: 0.06em;
  padding: 4px 12px; border-radius: 20px; width: fit-content;
  box-shadow: 0 2px 10px rgba(0,0,0,0.2);
}
.ia-badge-icon { font-size: 1rem; }
.ia-title { font-size: 1.4rem; font-weight: 900; color: var(--dark, #0f172a); margin: 0; }
.ia-desc  { font-size: 0.88rem; color: var(--muted, #64748b); margin: 0; }

.ia-stats { display: flex; gap: 10px; flex-wrap: wrap; align-items: center; }
.ia-stat {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px; border-radius: 20px;
  background: var(--surface-soft, #f8fafc);
  border: 1px solid var(--border, #e2e8f0);
  font-size: 0.88rem; font-weight: 700; color: var(--dark, #0f172a);
}
.ia-stat svg { color: var(--muted, #64748b); }
.ia-stat-pts { background: linear-gradient(135deg, #fef3c7, #fde68a); border-color: #f59e0b; }
.ia-stat-pts svg { color: #d97706; }

/* ─── Win Banner ────────────────────────────────────────────────────────── */
.ia-win {
  margin: 20px 28px 0;
  background: linear-gradient(135deg, rgba(16,185,129,0.1), rgba(99,102,241,0.1));
  border: 2px solid rgba(16,185,129,0.35);
  border-radius: 16px; padding: 20px 24px;
  display: flex; align-items: center; gap: 16px; flex-wrap: wrap;
}
.win-trophy { font-size: 2.5rem; filter: drop-shadow(0 4px 8px rgba(245,158,11,0.4)); animation: trophyPulse 1.5s ease-in-out infinite; }
@keyframes trophyPulse { 0%,100% { transform: scale(1) rotate(-5deg); } 50% { transform: scale(1.1) rotate(5deg); } }
.win-body { flex: 1; }
.win-body h3 { font-size: 1.15rem; font-weight: 900; color: var(--dark, #0f172a); margin: 0 0 4px; }
.win-body p  { font-size: 0.9rem; color: var(--muted, #64748b); margin: 0; }
.win-body strong { color: var(--dark, #0f172a); }
.ia-btn-replay {
  display: flex; align-items: center; gap: 8px;
  background: var(--game-gradient, linear-gradient(135deg,#6366f1,#8b5cf6));
  color: white; border: none; border-radius: 12px; padding: 10px 20px;
  font-weight: 700; font-size: 0.88rem; cursor: pointer; transition: opacity 0.2s, transform 0.15s;
  box-shadow: 0 4px 14px rgba(99,102,241,0.35);
}
.ia-btn-replay:hover { opacity: 0.9; transform: translateY(-1px); }

/* ─── Game Wrapper ──────────────────────────────────────────────────────── */
.game-wrap { padding: 20px 28px 28px; }
.game-topbar {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 16px; gap: 12px; flex-wrap: wrap;
}
.topbar-hint { font-size: 0.88rem; color: var(--muted, #64748b); font-weight: 500; }
.topbar-progress { display: flex; align-items: center; gap: 10px; }
.prog-badge { color: white; font-size: 0.78rem; font-weight: 800; padding: 3px 10px; border-radius: 20px; }
.prog-track { width: 100px; height: 7px; background: var(--border, #e2e8f0); border-radius: 10px; overflow: hidden; }
.prog-fill { height: 100%; border-radius: 10px; transition: width 0.5s cubic-bezier(0.25,0.8,0.25,1); }
.game-hint { font-size: 0.88rem; color: var(--muted, #64748b); margin: 0 0 18px; }

/* ─── Botón Principal ───────────────────────────────────────────────────── */
.game-actions { margin-top: 24px; display: flex; justify-content: center; }
.ia-btn-primary {
  display: flex; align-items: center; gap: 8px;
  background: var(--game-gradient, linear-gradient(135deg,#6366f1,#8b5cf6));
  color: white; border: none; border-radius: 14px;
  padding: 13px 32px; font-weight: 800; font-size: 0.95rem;
  cursor: pointer; transition: all 0.2s;
  box-shadow: 0 4px 18px color-mix(in srgb, var(--accent) 40%, transparent);
}
.ia-btn-primary:hover:not(:disabled) { transform: translateY(-2px); opacity: 0.92; box-shadow: 0 8px 24px color-mix(in srgb, var(--accent) 45%, transparent); }
.ia-btn-primary:active:not(:disabled) { transform: translateY(0); }
.ia-btn-primary:disabled { opacity: 0.45; cursor: not-allowed; }

/* ─── Pop transition ────────────────────────────────────────────────────── */
.pop-enter-active { animation: popIn 0.4s cubic-bezier(0.25,0.8,0.25,1); }
.pop-leave-active { animation: popIn 0.3s reverse; }
@keyframes popIn { from { transform: scale(0.8) translateY(-10px); opacity: 0; } to { transform: scale(1) translateY(0); opacity: 1; } }

/* ══════════════════════════════════════════════════════════════════
   5: MEMORAMA
══════════════════════════════════════════════════════════════════ */
.memo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 12px;
}
.memo-card {
  height: 150px; cursor: pointer; border: none; background: transparent; padding: 0;
  perspective: 800px;
  transition: transform 0.12s;
}
.memo-card:not(.matched):hover { transform: scale(1.04) rotate(-1deg); }
.memo-card:not(.matched):active { transform: scale(0.97); }
.memo-card.bounce { animation: cardBounce 0.6s cubic-bezier(0.36, 0.07, 0.19, 0.97); }
@keyframes cardBounce { 0%,100%{transform:scale(1)} 30%{transform:scale(1.15)} 60%{transform:scale(0.95)} }

.card-inner {
  position: relative; width: 100%; height: 100%;
  transform-style: preserve-3d; transition: transform 0.55s cubic-bezier(0.25,0.8,0.25,1);
  border-radius: 14px;
}
.memo-card.flipped .card-inner { transform: rotateY(180deg); }

.card-front, .card-back {
  position: absolute; inset: 0; backface-visibility: hidden;
  border-radius: 14px; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 6px;
  border: 2px solid transparent; overflow: hidden;
}
.card-front {
  box-shadow: 0 4px 16px rgba(0,0,0,0.15);
  position: relative;
}
.card-front-icon { font-size: 2rem; filter: drop-shadow(0 2px 6px rgba(0,0,0,0.2)); }
.card-shine {
  position: absolute; inset: 0;
  background: linear-gradient(135deg, rgba(255,255,255,0.25) 0%, transparent 60%);
  pointer-events: none;
}
.card-back {
  background: var(--surface, #fff);
  border-color: var(--border, #e2e8f0);
  transform: rotateY(180deg);
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
  transition: background 0.3s, border-color 0.3s;
  padding: 10px;
}
.card-match-mark {
  position: absolute; top: 8px; right: 8px;
  width: 22px; height: 22px; border-radius: 50%;
  color: white; font-size: 0.75rem; font-weight: 900;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
}
.card-img { max-width: 80%; max-height: 60%; object-fit: contain; border-radius: 8px; }
.card-txt { font-size: 0.85rem; font-weight: 800; text-align: center; color: var(--dark, #0f172a); word-break: break-word; transition: color 0.3s; }

/* ══════════════════════════════════════════════════════════════════
   6: CLASIFICAR
══════════════════════════════════════════════════════════════════ */
.drag-pool {
  display: flex; flex-wrap: wrap; gap: 10px; margin-bottom: 20px;
  padding: 16px; background: var(--surface-soft, #f8fafc);
  border-radius: 16px; border: 1.5px dashed var(--border, #e2e8f0);
  min-height: 64px;
}
.drag-item {
  display: flex; align-items: center; gap: 8px;
  padding: 9px 16px; border-radius: 30px;
  border: 1.5px solid var(--border, #e2e8f0); background: white;
  font-size: 0.88rem; font-weight: 600; color: var(--dark, #0f172a);
  cursor: pointer; transition: all 0.18s; position: relative;
}
.drag-item:hover:not(:disabled) { border-color: var(--accent); transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.drag-item.picked { border-color: var(--accent); background: color-mix(in srgb, var(--accent) 10%, white); box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent) 25%, transparent); transform: scale(1.04); }
.drag-item.done { opacity: 0.55; }
.drag-item.shake { animation: shake 0.5s; }
@keyframes shake { 0%,100%{transform:translateX(0)} 20%{transform:translateX(-5px)} 40%{transform:translateX(5px)} 60%{transform:translateX(-5px)} 80%{transform:translateX(5px)} }
.drag-item-dot { width: 9px; height: 9px; border-radius: 50%; flex-shrink: 0; transition: background 0.3s; }
.drag-item-tag { font-size: 0.7rem; font-weight: 800; padding: 2px 8px; border-radius: 10px; }

.drag-cats { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 14px; }
.drag-cat {
  border-radius: 16px; border: 2px solid color-mix(in srgb, var(--cat-color) 30%, transparent);
  background: color-mix(in srgb, var(--cat-color) 5%, white);
  transition: all 0.2s; overflow: hidden;
}
.drag-cat.droppable { border-color: var(--cat-color); cursor: pointer; box-shadow: 0 0 0 4px color-mix(in srgb, var(--cat-color) 15%, transparent); }
.drag-cat.droppable:hover { transform: scale(1.02); }

.cat-header { display: flex; align-items: center; gap: 8px; padding: 12px 14px 8px; }
.cat-dot { width: 12px; height: 12px; border-radius: 50%; background: var(--cat-color); flex-shrink: 0; }
.cat-name { flex: 1; font-size: 0.9rem; font-weight: 800; color: var(--dark, #0f172a); }
.cat-count { font-size: 0.75rem; font-weight: 700; color: var(--cat-color); background: color-mix(in srgb, var(--cat-color) 15%, transparent); padding: 2px 8px; border-radius: 10px; }
.cat-body { padding: 4px 14px 14px; display: flex; flex-direction: column; gap: 6px; min-height: 50px; }
.cat-chip { display: flex; align-items: center; justify-content: space-between; padding: 6px 10px; background: white; border-radius: 8px; font-size: 0.82rem; font-weight: 600; color: var(--dark, #0f172a); border: 1px solid color-mix(in srgb, var(--cat-color) 20%, transparent); animation: chipIn 0.25s; }
@keyframes chipIn { from{opacity:0;transform:scale(0.85)} to{opacity:1;transform:scale(1)} }
.cat-chip-del { border: none; background: transparent; cursor: pointer; color: var(--muted, #94a3b8); font-size: 0.7rem; padding: 2px 4px; border-radius: 4px; }
.cat-chip-del:hover { background: #fee2e2; color: #ef4444; }
.cat-empty { font-size: 0.8rem; color: var(--muted, #94a3b8); font-style: italic; }

/* ══════════════════════════════════════════════════════════════════
   7: SOPA DE LETRAS
══════════════════════════════════════════════════════════════════ */
.ws-wrap { }
.ws-layout { display: flex; gap: 20px; flex-wrap: wrap; align-items: flex-start; }
.ws-board-wrap { flex: 1; min-width: 250px; overflow-x: auto; }
.ws-board {
  display: inline-flex; flex-direction: column; gap: 3px;
  background: var(--surface-soft, #f8fafc); padding: 12px;
  border-radius: 16px; border: 1.5px solid var(--border, #e2e8f0);
  cursor: crosshair; box-shadow: inset 0 2px 8px rgba(0,0,0,0.04);
}
.ws-row { display: flex; gap: 3px; }
.ws-cell {
  width: 33px; height: 33px;
  background: white; border: 1.5px solid #e2e8f0; border-radius: 7px;
  font-weight: 900; font-size: 0.85rem; color: var(--dark, #0f172a);
  cursor: crosshair; display: flex; align-items: center; justify-content: center;
  transition: background 0.12s, border-color 0.12s, color 0.12s, transform 0.08s;
  user-select: none; -webkit-user-select: none; touch-action: none;
}
.ws-cell:hover { background: rgba(99,102,241,0.07); border-color: #a5b4fc; }
.ws-preview { background: rgba(99,102,241,0.22) !important; border-color: #6366f1 !important; transform: scale(1.08); }

.ws-panel {
  width: 190px; background: var(--surface-soft, #f8fafc); border-radius: 16px;
  padding: 16px; border: 1.5px solid var(--border, #e2e8f0); align-self: flex-start; flex-shrink: 0;
}
.ws-panel-title { display: flex; align-items: center; gap: 7px; font-size: 0.88rem; font-weight: 800; color: var(--dark, #0f172a); margin-bottom: 10px; }
.ws-count { margin-left: auto; font-size: 0.78rem; background: var(--border, #e2e8f0); padding: 1px 8px; border-radius: 10px; }
.ws-progress-bar { height: 5px; background: #e2e8f0; border-radius: 10px; overflow: hidden; margin-bottom: 12px; }
.ws-progress-fill { height: 100%; border-radius: 10px; transition: width 0.5s; }
.ws-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 7px; }
.ws-word { display: flex; align-items: center; gap: 8px; font-size: 0.85rem; font-weight: 700; color: var(--dark, #0f172a); transition: all 0.3s; }
.ws-word.found { text-decoration: line-through; animation: wordFound 0.4s; }
@keyframes wordFound { 0%{transform:scale(1)} 40%{transform:scale(1.12)} 100%{transform:scale(1)} }
.ws-bullet { width: 9px; height: 9px; border-radius: 50%; flex-shrink: 0; transition: background 0.3s; }

/* ══════════════════════════════════════════════════════════════════
   8: COMPLETAR ESPACIOS
══════════════════════════════════════════════════════════════════ */
.fb-list { display: flex; flex-direction: column; gap: 14px; }
.fb-card {
  display: flex; gap: 14px; align-items: flex-start;
  background: var(--surface-soft, #f8fafc); border: 2px solid var(--border, #e2e8f0);
  border-radius: 16px; padding: 16px 18px; transition: border-color 0.3s, background 0.3s;
}
.fb-card.fb-correct { border-color: #10b981; background: rgba(16,185,129,0.06); }
.fb-card.fb-wrong   { border-color: #ef4444; background: rgba(239,68,68,0.05); }
.fb-num {
  width: 32px; height: 32px; border-radius: 10px; color: white;
  font-size: 0.82rem; font-weight: 900; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
}
.fb-body { flex: 1; }
.fb-text { font-size: 0.92rem; color: var(--dark, #0f172a); font-weight: 500; margin: 0 0 12px; line-height: 1.5; }
.fb-opts { display: flex; flex-wrap: wrap; gap: 8px; }
.fb-opt {
  display: flex; align-items: center; gap: 7px;
  padding: 8px 16px; border-radius: 30px;
  border: 1.5px solid var(--border, #e2e8f0); background: white;
  cursor: pointer; font-size: 0.86rem; font-weight: 600; color: var(--dark, #0f172a);
  transition: all 0.15s; position: relative;
}
.fb-opt:hover:not([for]) { border-color: var(--accent); transform: translateY(-1px); }
.fb-opt-sel { border-color: var(--accent); background: color-mix(in srgb, var(--accent) 10%, white); color: var(--accent); }
.fb-opt-correct { border-color: #10b981 !important; background: rgba(16,185,129,0.1) !important; color: #059669 !important; }
.fb-opt-wrong   { border-color: #ef4444 !important; background: rgba(239,68,68,0.08) !important; color: #dc2626 !important; }
.fb-opt-dot {
  width: 9px; height: 9px; border-radius: 50%;
  background: var(--border, #e2e8f0); flex-shrink: 0; transition: background 0.2s;
}
.fb-opt-sel .fb-opt-dot { background: var(--accent); }
.fb-opt-correct .fb-opt-dot { background: #10b981; }
.fb-opt-wrong .fb-opt-dot { background: #ef4444; }
.fb-opt-check { font-weight: 900; color: #10b981; }
.fb-opt-x     { font-weight: 900; color: #ef4444; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; }

/* ══════════════════════════════════════════════════════════════════
   9: ORDENAR
══════════════════════════════════════════════════════════════════ */
.order-list { display: flex; flex-direction: column; gap: 10px; }
.order-card {
  display: flex; align-items: center; gap: 14px;
  background: var(--surface-soft, #f8fafc); border: 2px solid var(--border, #e2e8f0);
  border-radius: 14px; padding: 14px 16px;
  transition: border-color 0.3s, background 0.3s, transform 0.2s, box-shadow 0.2s;
}
.order-card:hover:not(.order-ok):not(.order-bad) { transform: translateX(3px); box-shadow: 0 4px 16px rgba(0,0,0,0.07); }
.order-card.order-ok  { border-color: #10b981; background: rgba(16,185,129,0.07); }
.order-card.order-bad { border-color: #ef4444; background: rgba(239,68,68,0.05); animation: shake 0.5s; }
.order-num {
  width: 34px; height: 34px; border-radius: 10px; color: white;
  font-size: 0.85rem; font-weight: 900; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2); transition: all 0.3s;
}
.order-txt { flex: 1; font-size: 0.9rem; font-weight: 500; color: var(--dark, #0f172a); }
.order-arrows { display: flex; flex-direction: column; gap: 4px; }
.order-arrow {
  width: 28px; height: 28px; border-radius: 8px;
  border: 1.5px solid var(--border, #e2e8f0); background: white;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  color: var(--muted, #64748b); transition: all 0.15s;
}
.order-arrow:hover:not(:disabled) { background: var(--accent); color: white; border-color: var(--accent); transform: scale(1.1); }
.order-arrow:disabled { opacity: 0.25; cursor: not-allowed; }
.order-result-icon { font-size: 1.2rem; font-weight: 900; flex-shrink: 0; }
.order-card.order-ok  .order-result-icon { color: #10b981; }
.order-card.order-bad .order-result-icon { color: #ef4444; }

/* order-move transition */
.order-move-move { transition: transform 0.3s cubic-bezier(0.25,0.8,0.25,1); }

/* ─── Responsive ─────────────────────────────────────────────────────────── */
@media (max-width: 600px) {
  .ia-header { padding: 16px 16px 0; }
  .game-wrap  { padding: 14px 16px 20px; }
  .memo-grid  { grid-template-columns: repeat(auto-fill, minmax(80px, 1fr)); gap: 8px; }
  .memo-card  { height: 110px; }
  .ws-cell    { width: 26px; height: 26px; font-size: 0.72rem; }
  .ws-panel   { width: 100%; }
  .drag-cats  { grid-template-columns: 1fr 1fr; }
  .ia-title   { font-size: 1.1rem; }
}
</style>
