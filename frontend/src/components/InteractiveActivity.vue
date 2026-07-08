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
  const raw = props.lesson?.lesson_type ?? props.lesson?.type
  if (raw !== undefined && raw !== null) {
    const s = String(raw).trim().toLowerCase()
    const MAP: Record<string, string> = {
      '5': '5', 'memory': '5', 'memorama': '5', 'lesson_type_game_memory': '5',
      '6': '6', 'dragdrop': '6', 'clasificar': '6', 'lesson_type_game_dragdrop': '6',
      '7': '7', 'wordsearch': '7', 'sopa': '7', 'lesson_type_game_wordsearch': '7',
      '8': '8', 'fillblank': '8', 'completar': '8', 'lesson_type_game_fillblank': '8',
      '9': '9', 'order': '9', 'ordenar': '9', 'lesson_type_game_order': '9',
      '10': '10', 'hangman': '10', 'ahorcado': '10', 'lesson_type_game_hangman': '10'
    }
    return MAP[s] ?? String(raw)
  }
  return '1'
})

const GAME_META: Record<string, { label: string; icon: string; gradient: string; accent: string }> = {
  '5': { label: 'Memorama',          icon: '🎴', gradient: 'linear-gradient(135deg,#6366f1,#8b5cf6)', accent: '#8b5cf6' },
  '6': { label: 'Clasificar',         icon: '🗂️', gradient: 'linear-gradient(135deg,#0ea5e9,#06b6d4)', accent: '#0ea5e9' },
  '7': { label: 'Sopa de Letras',     icon: '🔤', gradient: 'linear-gradient(135deg,#f59e0b,#f97316)', accent: '#f97316' },
  '8': { label: 'Completar Espacios', icon: '✍️', gradient: 'linear-gradient(135deg,#10b981,#059669)', accent: '#10b981' },
  '9': { label: 'Ordenar',            icon: '🔢', gradient: 'linear-gradient(135deg,#ec4899,#f43f5e)', accent: '#ec4899' },
  '10': { label: 'Ahorcado Cibernético', icon: '🎯', gradient: 'linear-gradient(135deg,#f43f5e,#e11d48)', accent: '#f43f5e' },
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

// Sonido de burla divertida / risa juguetona descendente cuando fallas una letra en Ahorcado
function soundTaunt() {
  try {
    const ctx = new (window.AudioContext || (window as any).webkitAudioContext)()
    const notes = [320, 280, 240, 180]
    notes.forEach((freq, i) => {
      setTimeout(() => {
        const osc = ctx.createOscillator()
        const gain = ctx.createGain()
        osc.type = i === 3 ? 'sawtooth' : 'triangle'
        osc.frequency.setValueAtTime(freq, ctx.currentTime)
        if (i === 3) osc.frequency.linearRampToValueAtTime(110, ctx.currentTime + 0.4)
        gain.gain.setValueAtTime(0.14, ctx.currentTime)
        gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + (i === 3 ? 0.45 : 0.16))
        osc.connect(gain); gain.connect(ctx.destination)
        osc.start(); osc.stop(ctx.currentTime + (i === 3 ? 0.45 : 0.16))
      }, i * 160)
    })
  } catch {}
}

function soundChuckle() {
  try {
    const ctx = new (window.AudioContext || (window as any).webkitAudioContext)()
    const pitches = [520, 470, 540, 440]
    pitches.forEach((freq, i) => {
      setTimeout(() => {
        const osc = ctx.createOscillator()
        const gain = ctx.createGain()
        osc.type = 'sine'
        osc.frequency.setValueAtTime(freq, ctx.currentTime)
        gain.gain.setValueAtTime(0.09, ctx.currentTime)
        gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.08)
        osc.connect(gain); gain.connect(ctx.destination)
        osc.start(); osc.stop(ctx.currentTime + 0.08)
      }, i * 95)
    })
  } catch {}
}

// Reproductor de música retro Arcade sintetizado
const isMusicOn = ref(false)
let musicInterval: ReturnType<typeof setInterval> | null = null

function toggleMusic() {
  isMusicOn.value = !isMusicOn.value
  if (isMusicOn.value) {
    playArcadeNote()
    musicInterval = setInterval(playArcadeNote, 1600)
  } else {
    stopMusic()
  }
}
function stopMusic() {
  isMusicOn.value = false
  if (musicInterval) { clearInterval(musicInterval); musicInterval = null }
}
function playArcadeNote() {
  if (!isMusicOn.value) return
  const melody = [523.25, 659.25, 783.99, 1046.50, 783.99, 659.25]
  const idx = Math.floor(Math.random() * melody.length)
  beep(melody[idx]!, 'sine', 0.18)
}
onUnmounted(() => stopMusic())

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
  isCompleted.value = !!props.lesson?.completada; pointsEarned.value = props.lesson?.completada ? (props.lesson?.points_reward || 100) : 0; stopTimer()
  let parsed: any = {}
  try {
    let raw = props.lesson?.game_config_json
    while (typeof raw === 'string' && (raw.trim().startsWith('{') || raw.trim().startsWith('"'))) {
      try {
        const next = JSON.parse(raw)
        if (next === raw) break
        raw = next
      } catch { break }
    }
    if (raw && typeof raw === 'object') parsed = raw
  } catch {}
  config.value = parsed || {}
  initSpecificGame(); startTimer()
}

async function handleWin() {
  if (isCompleted.value) return
  isCompleted.value = true; stopTimer(); soundWin()
  const pts = Number(props.lesson?.points_reward || 100); pointsEarned.value = pts
  await nextTick(); launchConfetti()
  try {
    await api.post(`/lecciones/${props.lesson.id}/game-score`, { curso_id: props.cursoId, points: pts, time_secs: elapsedSecs.value })
    toast.success(`¡+${pts} puntos ganados!`)
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
const dragOverCat    = ref<string | null>(null)
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
  dragOverCat.value = null
}

function pickItem(id: string) { if (!isCompleted.value) { selectedDrag.value = selectedDrag.value === id ? null : id; beep(500,'sine',0.05) } }
function dropInCategory(cat: string) {
  if (!selectedDrag.value || isCompleted.value) return
  dragAssignments[selectedDrag.value] = cat; selectedDrag.value = null; beep(580,'sine',0.1)
}
function removeAssignment(id: string) { delete dragAssignments[id] }

function handleDragStart(e: DragEvent, id: string) {
  if (isCompleted.value) return
  selectedDrag.value = id
  if (e.dataTransfer) {
    e.dataTransfer.setData('text/plain', id)
    e.dataTransfer.effectAllowed = 'move'
  }
}
function handleDragOver(e: DragEvent, cat: string) {
  if (isCompleted.value) return
  dragOverCat.value = cat
}
function handleDragLeave(cat: string) {
  if (dragOverCat.value === cat) {
    dragOverCat.value = null
  }
}
function handleDrop(e: DragEvent, cat: string) {
  dragOverCat.value = null
  if (isCompleted.value) return
  let id = selectedDrag.value
  if (!id && e.dataTransfer) {
    id = e.dataTransfer.getData('text/plain')
  }
  if (id) {
    selectedDrag.value = id
    dropInCategory(cat)
  }
}

function catColor(catName: string | undefined): string {
  if (!catName) return '#6366f1'
  const idx = dragCategories.value.indexOf(catName)
  return CAT_COLORS[idx % CAT_COLORS.length] ?? '#6366f1'
}

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

// ═══════════════════════════════════════════════════════════════════════════
//  10: AHORCADO CIBERNÉTICO
// ═══════════════════════════════════════════════════════════════════════════
const hmItems = ref<{ word: string; hint: string }[]>([])
const hmCurrentIdx = ref(0)
const hmWord = ref('')
const hmHint = ref('')
const hmGuessed = ref<string[]>([])
const hmErrors = ref(0)
const hmMaxErrors = ref(6)
const hmTauntMessage = ref('')
const HM_ALPHABET = ['A','B','C','D','E','F','G','H','I','J','K','L','M','N','Ñ','O','P','Q','R','S','T','U','V','W','X','Y','Z']

function initHangman() {
  hmGuessed.value = []
  hmErrors.value = 0
  hmCurrentIdx.value = 0
  hmTauntMessage.value = ''
  let items: { word: string; hint: string }[] = []
  if (config.value?.items && Array.isArray(config.value.items) && config.value.items.length > 0) {
    items = config.value.items.map((it: any) => ({
      word: String(it.word || '').trim().normalize('NFD').replace(/[\u0300-\u036f]/g,'').replace(/[^a-zA-Z]/g,'').toUpperCase(),
      hint: String(it.hint || 'Adivina la palabra oculta')
    })).filter((it: any) => it.word.length >= 2)
  }
  if (!items.length) {
    items = [
      { word: 'JAVASCRIPT', hint: 'Lenguaje web dinámico y moderno' },
      { word: 'CIBERSEGURIDAD', hint: 'Protección de sistemas contra intrusos' },
      { word: 'MICROSERVICIOS', hint: 'Arquitectura modular distribuida en la nube' },
      { word: 'DOCKER', hint: 'Plataforma para contenerizar aplicaciones' }
    ]
  }
  hmItems.value = items
  hmMaxErrors.value = Math.max(Number(config.value?.max_errors) || 6, 3)
  loadHangmanWord()
}

function loadHangmanWord() {
  const current = hmItems.value[hmCurrentIdx.value] || hmItems.value[0]
  if (!current) return
  hmWord.value = current.word
  hmHint.value = current.hint
  hmGuessed.value = []
  hmErrors.value = 0
  hmTauntMessage.value = ''
}

function hmGuessLetter(letter: string) {
  if (isCompleted.value || hmGuessed.value.includes(letter)) return
  hmGuessed.value.push(letter)
  if (hmWord.value.includes(letter)) {
    soundOk()
    const allFound = hmWord.value.split('').every(ch => hmGuessed.value.includes(ch))
    if (allFound) {
      if (hmCurrentIdx.value < hmItems.value.length - 1) {
        toast.success(`¡Excelente! "${hmWord.value}" completada. Pasando a la siguiente...`)
        setTimeout(() => {
          hmCurrentIdx.value++
          loadHangmanWord()
        }, 1200)
      } else {
        handleWin()
      }
    }
  } else {
    hmErrors.value++
    const taunts = [
      '¡Uy, ni cerca! 😏',
      '¡Esa letra no existe en este servidor! 🤖',
      '¡Cuidado con el cortocircuito! ⚡',
      '¡Una pieza más al androide! HAHAHA 😂',
      '¡Estás sudando frío! 😅',
      '¡PELIGRO DE REINICIO! 🚨'
    ]
    hmTauntMessage.value = taunts[hmErrors.value % taunts.length]!
    soundTaunt()
    if (hmErrors.value >= hmMaxErrors.value) {
      toast.error(`¡Sistemas caídos! La palabra era "${hmWord.value}". Intentemos de nuevo.`)
      setTimeout(() => {
        loadHangmanWord()
      }, 2000)
    }
  }
}

// ── Dispatcher ──────────────────────────────────────────────────────────────
function initSpecificGame() {
  const t = gameType.value
  if      (t==='5') initMemo()
  else if (t==='6') initDrag()
  else if (t==='7') initWordSearch()
  else if (t==='8') initFillBlank()
  else if (t==='9') initOrder()
  else if (t==='10') initHangman()
}

const fmt = (s: number) => s < 60 ? `${s}s` : `${Math.floor(s/60)}m ${s%60}s`

// Cargar juego al final del setup cuando todas las refs ya están inicializadas
watch(() => [props.lesson?.id, props.lesson?.lesson_type, props.lesson?.type, props.lesson?.game_config_json], loadGame, { immediate: true })
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
          <span class="ia-badge-icon">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="12 2 2 7 12 12 22 7 12 2"/>
              <polyline points="2 17 12 22 22 17"/>
              <polyline points="2 12 12 17 22 12"/>
            </svg>
          </span>
          <span>{{ meta.label }}</span>
        </div>
        <h2 class="ia-title">{{ lesson?.title }}</h2>
        <p v-if="lesson?.description" class="ia-desc">{{ lesson.description }}</p>
      </div>
      <div class="ia-stats">
        <button class="ia-music-btn" @click="toggleMusic" :class="{ 'ia-music-on': isMusicOn }">
          <span>{{ isMusicOn ? '🎵 Música ON' : '🔇 Música OFF' }}</span>
        </button>
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

    <!-- ── Menú Superior de Estado: Juego Completado ─────────────────────── -->
    <transition name="pop">
      <div v-if="isCompleted" class="ia-completed-menu">
        <div class="completed-topbar">
          <div class="completed-status-badge">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
            <span>Juego Completado · Progreso Guardado</span>
          </div>
          <span class="completed-pts-pill">+{{ pointsEarned || lesson?.points_reward || 100 }} pts</span>
        </div>

        <div class="completed-content">
          <p class="completed-info-text">
            ¡Excelente trabajo! Tu progreso y tu puntaje ya han sido registrados exitosamente en el sistema. <strong>No necesitas repetirlo</strong> para conservar tu avance, pero puedes volver a jugar por práctica si lo deseas.
          </p>

          <div class="completed-metrics">
            <div class="metric-box">
              <span class="metric-title">Estado del reto</span>
              <span class="metric-value status-saved">✓ Guardado</span>
            </div>
            <div class="metric-box">
              <span class="metric-title">Puntaje obtenido</span>
              <span class="metric-value">+{{ pointsEarned || lesson?.points_reward || 100 }} pts</span>
            </div>
            <div class="metric-box" v-if="elapsedSecs > 0">
              <span class="metric-title">Tiempo realizado</span>
              <span class="metric-value">{{ fmt(elapsedSecs) }}</span>
            </div>
          </div>
        </div>

        <div class="completed-footer">
          <button class="btn-completed-replay" @click="restartGame">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
            Jugar de nuevo (Practicar)
          </button>
        </div>
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
            <!-- Dorso Naipe Glassmorphism -->
            <div class="card-front" :style="{ background: meta.gradient }">
              <div class="naipe-frame"></div>
              <div class="naipe-corner top-left">
                <span>MH</span>
                <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z"/></svg>
              </div>
              <div class="naipe-center-emblem">
                <div class="glass-emblem-circle">
                  <svg width="38" height="38" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="rgba(255,255,255,0.22)"/>
                    <path d="M2 17L12 22L22 17"/>
                    <path d="M2 12L12 17L22 12"/>
                  </svg>
                </div>
              </div>
              <div class="naipe-corner bottom-right">
                <span>MH</span>
                <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z"/></svg>
              </div>
              <div class="card-shine" />
            </div>
            <!-- Cara Naipe Glassmorphism -->
            <div class="card-back" :style="memoMatched.includes(card.pairId)
              ? { background: 'rgba(255, 255, 255, 0.94)', borderColor: card.color, boxShadow: `0 14px 35px ${card.color}35, inset 0 0 18px ${card.color}20` }
              : {}">
              <div class="card-face-corner top-left" :style="{ color: card.color }">
                <div class="corner-dot" :style="{ background: card.color }"></div>
              </div>

              <div v-if="memoMatched.includes(card.pairId)" class="card-match-mark" :style="{ background: card.color }">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="3.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
              </div>

              <img v-if="card.img" :src="card.img" class="card-img" :alt="card.text" />
              <span class="card-txt" :style="memoMatched.includes(card.pairId) ? { color: card.color } : {}">{{ card.text }}</span>

              <div class="card-face-corner bottom-right" :style="{ color: card.color }">
                <div class="corner-dot" :style="{ background: card.color }"></div>
              </div>
            </div>
          </div>
        </button>
      </div>
    </div>

    <!-- ════════════════════════════════════════════════════════════════════
         6: CLASIFICAR
    ═════════════════════════════════════════════════════════════════════ -->
    <div v-else-if="gameType === '6'" class="game-wrap">
      <p class="game-hint">Arrastra y suelta cada elemento hacia su categoría, o haz clic en un elemento y luego en la categoría donde pertenece</p>

      <!-- Pool de items -->
      <div class="drag-pool">
        <button
          v-for="it in dragItems" :key="it.id"
          :class="['drag-item',
            selectedDrag === it.id ? 'picked' : '',
            dragAssignments[it.id] ? 'done' : '',
            shakeItems.includes(it.id) ? 'shake' : '']"
          :disabled="isCompleted"
          draggable="true"
          @dragstart="handleDragStart($event, it.id)"
          @click="pickItem(it.id)"
        >
          <span class="drag-item-dot" :style="{ background: dragAssignments[it.id]
            ? catColor(dragAssignments[it.id])
            : selectedDrag === it.id ? 'var(--game-accent)' : '#cbd5e1' }" />
          {{ it.text }}
          <span v-if="dragAssignments[it.id]" class="drag-item-tag" :style="{ background: catColor(dragAssignments[it.id]) + '22', color: catColor(dragAssignments[it.id]) }">
            {{ dragAssignments[it.id] }}
          </span>
        </button>
      </div>

      <!-- Categorías -->
      <div class="drag-cats">
        <div
          v-for="(cat, ci) in dragCategories" :key="cat"
          :class="['drag-cat', selectedDrag ? 'droppable' : '', dragOverCat === cat ? 'drag-over' : '']"
          :style="{ '--cat-color': CAT_COLORS[ci % CAT_COLORS.length] }"
          @dragover.prevent="handleDragOver($event, cat)"
          @dragenter.prevent="handleDragOver($event, cat)"
          @dragleave="handleDragLeave(cat)"
          @drop.prevent="handleDrop($event, cat)"
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
              draggable="true"
              @dragstart="handleDragStart($event, it.id)"
              @click.stop="pickItem(it.id)"
            >
              {{ it.text }}
              <button @click.stop="removeAssignment(it.id)" class="cat-chip-del">✕</button>
            </div>
            <span v-if="!dragItems.some(i => dragAssignments[i.id] === cat)" class="cat-empty">
              {{ selectedDrag ? '← Haz clic o suelta aquí' : 'Arrastra o haz clic para asignar aquí' }}
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

    <!-- ═══════════════ 10: AHORCADO CIBERNÉTICO ═══════════════ -->
    <div v-else-if="gameType === '10'" class="game-wrap">
      <!-- Barra superior de estado -->
      <div class="hm-header-bar">
        <div class="hm-progress">
          <span class="hm-pill">Palabra {{ hmCurrentIdx + 1 }} de {{ hmItems.length }}</span>
        </div>
        <div class="hm-lives">
          <span class="hm-lives-label">Energía del Sistema:</span>
          <div class="hm-hearts">
            <span v-for="n in hmMaxErrors" :key="n" class="hm-heart" :class="{ 'hm-heart-lost': n <= hmErrors }">
              ⚡
            </span>
          </div>
        </div>
      </div>

      <div class="hm-stage">
        <!-- Visualizador SVG Ciber-Ahorcado -->
        <div class="hm-visualizer glass-badge-editor">
          <svg viewBox="0 0 200 220" class="hm-svg">
            <!-- Plataforma y Horca -->
            <line x1="20" y1="200" x2="180" y2="200" stroke="#475569" stroke-width="6" stroke-linecap="round" />
            <line x1="60" y1="200" x2="60" y2="20" stroke="#64748b" stroke-width="6" stroke-linecap="round" />
            <line x1="60" y1="20" x2="130" y2="20" stroke="#64748b" stroke-width="6" stroke-linecap="round" />
            <line x1="130" y1="20" x2="130" y2="50" stroke="#f43f5e" stroke-width="4" stroke-dasharray="4,4" />

            <!-- 1. Cabeza Holográmica -->
            <g v-if="hmErrors >= 1" class="hm-part">
              <circle cx="130" cy="70" r="18" fill="rgba(244,63,94,0.15)" stroke="#f43f5e" stroke-width="3" />
              <!-- Ojos -->
              <text x="122" y="74" font-size="10" fill="#f43f5e" font-weight="bold">{{ hmErrors >= hmMaxErrors ? 'X' : '•' }}</text>
              <text x="133" y="74" font-size="10" fill="#f43f5e" font-weight="bold">{{ hmErrors >= hmMaxErrors ? 'X' : '•' }}</text>
            </g>

            <!-- 2. Cuerpo Core -->
            <line v-if="hmErrors >= 2" x1="130" y1="88" x2="130" y2="135" stroke="#f43f5e" stroke-width="4" stroke-linecap="round" class="hm-part" />

            <!-- 3. Brazo Izquierdo -->
            <line v-if="hmErrors >= 3" x1="130" y1="100" x2="105" y2="120" stroke="#f43f5e" stroke-width="4" stroke-linecap="round" class="hm-part" />

            <!-- 4. Brazo Derecho -->
            <line v-if="hmErrors >= 4" x1="130" y1="100" x2="155" y2="120" stroke="#f43f5e" stroke-width="4" stroke-linecap="round" class="hm-part" />

            <!-- 5. Pierna Izquierda -->
            <line v-if="hmErrors >= 5" x1="130" y1="135" x2="110" y2="170" stroke="#f43f5e" stroke-width="4" stroke-linecap="round" class="hm-part" />

            <!-- 6. Pierna Derecha -->
            <line v-if="hmErrors >= 6" x1="130" y1="135" x2="150" y2="170" stroke="#f43f5e" stroke-width="4" stroke-linecap="round" class="hm-part" />
          </svg>

          <!-- Globo de burla cuando falla -->
          <transition name="fade">
            <div v-if="hmTauntMessage" class="hm-taunt-bubble">
              {{ hmTauntMessage }}
            </div>
          </transition>
        </div>

        <!-- Pista de la palabra -->
        <div class="hm-hint-box">
          <span class="hm-hint-tag">PISTA:</span>
          <p class="hm-hint-text">{{ hmHint }}</p>
        </div>
      </div>

      <!-- Palabra Oculta -->
      <div class="hm-word-container">
        <div v-for="(char, idx) in hmWord.split('')" :key="idx" class="hm-letter-slot" :class="{ 'hm-letter-revealed': hmGuessed.includes(char) || isCompleted }">
          <span v-if="hmGuessed.includes(char) || isCompleted">{{ char }}</span>
          <span v-else class="hm-mystery">?</span>
        </div>
      </div>

      <!-- Teclado en pantalla -->
      <div class="hm-keyboard">
        <button
          v-for="char in HM_ALPHABET"
          :key="char"
          class="hm-key"
          :class="{
            'hm-key-ok': hmGuessed.includes(char) && hmWord.includes(char),
            'hm-key-bad': hmGuessed.includes(char) && !hmWord.includes(char)
          }"
          :disabled="hmGuessed.includes(char) || isCompleted"
          @click="hmGuessLetter(char)"
        >
          {{ char }}
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

/* ─── Menú Superior de Estado: Juego Completado (Glassmorphism) ──────────── */
.ia-completed-menu {
  margin: 24px 28px 8px;
  background: rgba(240, 253, 244, 0.88);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 2px solid rgba(16, 185, 129, 0.45);
  border-radius: 24px;
  padding: 24px 28px;
  box-shadow: 0 16px 36px rgba(16, 185, 129, 0.12), inset 0 1px 2px rgba(255, 255, 255, 0.95);
  display: flex;
  flex-direction: column;
  gap: 18px;
  animation: completedPulseIn 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes completedPulseIn {
  from { opacity: 0; transform: translateY(-12px) scale(0.97); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
.completed-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  border-bottom: 1px solid rgba(16, 185, 129, 0.2);
  padding-bottom: 14px;
}
.completed-status-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #059669;
  font-size: 1.05rem;
  font-weight: 900;
  letter-spacing: 0.2px;
}
.completed-pts-pill {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  font-size: 0.92rem;
  font-weight: 900;
  padding: 6px 16px;
  border-radius: 30px;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}
.completed-info-text {
  font-size: 0.96rem;
  line-height: 1.55;
  color: #1e293b;
  margin: 0;
}
.completed-info-text strong {
  color: #047857;
}
.completed-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: 4px;
}
.metric-box {
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(16, 185, 129, 0.25);
  border-radius: 16px;
  padding: 10px 18px;
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 130px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.03);
}
.metric-title {
  font-size: 0.75rem;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.metric-value {
  font-size: 1.05rem;
  font-weight: 900;
  color: #0f172a;
}
.metric-value.status-saved {
  color: #059669;
}
.completed-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 6px;
}
.btn-completed-replay {
  display: flex; align-items: center; gap: 8px;
  background: rgba(255, 255, 255, 0.9);
  color: #047857;
  border: 1.5px solid rgba(16, 185, 129, 0.45);
  border-radius: 14px; padding: 11px 22px;
  font-weight: 800; font-size: 0.9rem;
  cursor: pointer; transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  box-shadow: 0 4px 14px rgba(0,0,0,0.04);
}
.btn-completed-replay:hover {
  background: #10b981;
  color: white;
  border-color: #10b981;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(16, 185, 129, 0.3);
}

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
  grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
  gap: 22px;
}
.memo-card {
  height: 285px; cursor: pointer; border: none; background: transparent; padding: 0;
  perspective: 1200px;
  transition: transform 0.22s cubic-bezier(0.16, 1, 0.3, 1);
}
.memo-card:not(.matched):hover { transform: scale(1.04) translateY(-6px); }
.memo-card:not(.matched):active { transform: scale(0.98); }
.memo-card.bounce { animation: cardBounce 0.6s cubic-bezier(0.36, 0.07, 0.19, 0.97); }
@keyframes cardBounce { 0%,100%{transform:scale(1)} 30%{transform:scale(1.08)} 60%{transform:scale(0.96)} }

.card-inner {
  position: relative; width: 100%; height: 100%;
  transform-style: preserve-3d; transition: transform 0.65s cubic-bezier(0.23, 1, 0.32, 1);
  border-radius: 20px;
}
.memo-card.flipped .card-inner { transform: rotateY(180deg); }

.card-front, .card-back {
  position: absolute; inset: 0; backface-visibility: hidden;
  border-radius: 20px; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 10px;
}
.card-front {
  box-shadow: 0 16px 36px rgba(0,0,0,0.22), inset 0 1px 2px rgba(255,255,255,0.65);
  position: relative;
  overflow: hidden;
  border: 1.5px solid rgba(255, 255, 255, 0.42);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
}
.naipe-frame {
  position: absolute;
  inset: 12px;
  border: 1.5px dashed rgba(255, 255, 255, 0.45);
  border-radius: 12px;
  pointer-events: none;
}
.naipe-corner {
  position: absolute;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.5px;
  line-height: 1;
}
.naipe-corner.top-left { top: 16px; left: 16px; }
.naipe-corner.bottom-right { bottom: 16px; right: 16px; transform: rotate(180deg); }
.naipe-center-emblem {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
}
.glass-emblem-circle {
  width: 76px;
  height: 76px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.16);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1.5px solid rgba(255, 255, 255, 0.52);
  box-shadow: 0 8px 24px rgba(0,0,0,0.18), inset 0 0 18px rgba(255,255,255,0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}
.card-shine {
  position: absolute; inset: 0;
  background: linear-gradient(135deg, rgba(255,255,255,0.32) 0%, transparent 55%);
  pointer-events: none;
}
.card-back {
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1.5px solid rgba(255, 255, 255, 0.75);
  transform: rotateY(180deg);
  box-shadow: 0 16px 36px rgba(0,0,0,0.14), inset 0 1px 2px rgba(255,255,255,0.9);
  transition: all 0.35s ease;
  padding: 22px 18px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(148,163,184,0.4) transparent;
}
.card-face-corner {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
}
.card-face-corner.top-left { top: 12px; left: 12px; }
.card-face-corner.bottom-right { bottom: 12px; right: 12px; }
.corner-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  box-shadow: 0 0 8px currentColor;
}
.card-match-mark {
  position: absolute; top: 10px; right: 10px; z-index: 5;
  width: 26px; height: 26px; border-radius: 50%;
  color: white; font-size: 0.8rem; font-weight: 900;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 4px 12px rgba(0,0,0,0.22);
}
.card-img {
  max-width: 95%; max-height: 145px; width: auto; height: auto;
  object-fit: contain; border-radius: 8px; flex-shrink: 0; margin: auto;
}
.card-txt {
  font-size: 0.94rem; line-height: 1.38; font-weight: 700;
  text-align: center; color: var(--dark, #0f172a);
  word-break: break-word; overflow-wrap: break-word;
  transition: color 0.3s; width: 100%; margin: auto 0;
}

/* ══════════════════════════════════════════════════════════════════
   6: CLASIFICAR
══════════════════════════════════════════════════════════════════ */
.drag-pool {
  display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 24px;
  padding: 20px;
  background: rgba(248, 250, 252, 0.65);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border-radius: 20px; border: 1.5px dashed rgba(226, 232, 240, 0.85);
  min-height: 76px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.03);
}
.drag-item {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 20px; border-radius: 30px;
  border: 1.5px solid rgba(255, 255, 255, 0.85);
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  font-size: 0.94rem; font-weight: 700; color: var(--dark, #0f172a);
  cursor: grab; transition: all 0.22s cubic-bezier(0.16, 1, 0.3, 1); position: relative;
  user-select: none;
  box-shadow: 0 4px 14px rgba(0,0,0,0.06);
}
.drag-item:active { cursor: grabbing; }
.drag-item:hover:not(:disabled) { border-color: var(--accent); transform: translateY(-3px) scale(1.02); box-shadow: 0 8px 20px rgba(0,0,0,0.1); }
.drag-item.picked { border-color: var(--accent); background: color-mix(in srgb, var(--accent) 15%, white); box-shadow: 0 0 0 4px color-mix(in srgb, var(--accent) 25%, transparent); transform: scale(1.04); }
.drag-item.done { opacity: 0.55; }
.drag-item.shake { animation: shake 0.5s; }
@keyframes shake { 0%,100%{transform:translateX(0)} 20%{transform:translateX(-5px)} 40%{transform:translateX(5px)} 60%{transform:translateX(-5px)} 80%{transform:translateX(5px)} }
.drag-item-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; transition: background 0.3s; box-shadow: 0 0 6px currentColor; }
.drag-item-tag { font-size: 0.72rem; font-weight: 800; padding: 3px 9px; border-radius: 10px; }

.drag-cats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}
.drag-cat {
  border-radius: 24px;
  border: 2px solid color-mix(in srgb, var(--cat-color) 45%, rgba(255,255,255,0.7));
  background: color-mix(in srgb, var(--cat-color) 8%, rgba(255, 255, 255, 0.72));
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  display: flex;
  flex-direction: column;
  box-shadow: 0 10px 30px rgba(0,0,0,0.06), inset 0 1px 2px rgba(255,255,255,0.85);
}
.drag-cat.droppable {
  border-color: var(--cat-color);
  cursor: pointer;
  box-shadow: 0 0 0 5px color-mix(in srgb, var(--cat-color) 20%, transparent);
}
.drag-cat.droppable:hover { transform: scale(1.015); }
.drag-cat.drag-over {
  border-color: var(--cat-color) !important;
  background: color-mix(in srgb, var(--cat-color) 18%, rgba(255,255,255,0.9)) !important;
  box-shadow: 0 0 0 8px color-mix(in srgb, var(--cat-color) 35%, transparent) !important;
  transform: scale(1.025);
}

.cat-header {
  display: flex; align-items: center; gap: 12px;
  padding: 18px 20px 14px;
  border-bottom: 1px solid color-mix(in srgb, var(--cat-color) 18%, transparent);
}
.cat-dot {
  width: 14px; height: 14px; border-radius: 50%;
  background: var(--cat-color); flex-shrink: 0;
  box-shadow: 0 0 10px var(--cat-color);
}
.cat-name { flex: 1; font-size: 1.15rem; font-weight: 800; color: var(--dark, #0f172a); }
.cat-count {
  font-size: 0.88rem; font-weight: 800;
  color: var(--cat-color);
  background: color-mix(in srgb, var(--cat-color) 18%, white);
  padding: 4px 14px; border-radius: 20px;
}
.cat-body {
  padding: 18px 20px 20px;
  display: flex; flex-direction: column; gap: 12px;
  min-height: 175px;
  flex: 1;
}
.cat-chip {
  display: flex; align-items: center; justify-content: space-between; gap: 10px;
  padding: 12px 16px; background: rgba(255, 255, 255, 0.88); border-radius: 14px;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  font-size: 0.95rem; font-weight: 700; color: var(--dark, #0f172a);
  border: 1.5px solid color-mix(in srgb, var(--cat-color) 35%, rgba(255,255,255,0.8));
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  animation: chipIn 0.25s;
}
@keyframes chipIn { from{opacity:0;transform:scale(0.85)} to{opacity:1;transform:scale(1)} }
.cat-chip-del {
  border: none; background: transparent; cursor: pointer;
  color: var(--muted, #94a3b8); font-size: 0.8rem; padding: 4px 6px; border-radius: 6px;
}
.cat-chip-del:hover { background: #fee2e2; color: #ef4444; }
.cat-empty {
  font-size: 0.95rem; color: var(--muted, #94a3b8); font-style: italic;
  display: flex; align-items: center; justify-content: center;
  flex: 1; min-height: 100px;
  border: 2px dashed color-mix(in srgb, var(--cat-color) 25%, #e2e8f0);
  border-radius: 12px;
  background: color-mix(in srgb, var(--cat-color) 3%, white);
  padding: 12px; text-align: center;
}

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

/* ─── Botón de Música Arcade ─────────────────────────────────────────────── */
.ia-music-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 14px; border-radius: 20px;
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(8px);
  border: 1.5px solid var(--border, #e2e8f0);
  font-size: 0.85rem; font-weight: 800; color: var(--dark, #0f172a);
  cursor: pointer; transition: all 0.25s ease;
}
.ia-music-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}
.ia-music-on {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white; border-color: #059669;
  box-shadow: 0 0 14px rgba(16,185,129,0.4);
}

/* ─── 10: AHORCADO CIBERNÉTICO (Glassmorphism + Cyber Neon) ──────────────── */
.hm-header-bar {
  display: flex; justify-content: space-between; align-items: center;
  flex-wrap: wrap; gap: 12px; margin-bottom: 20px;
}
.hm-pill {
  background: rgba(244,63,94,0.1); color: #e11d48;
  padding: 6px 16px; border-radius: 20px; font-weight: 800; font-size: 0.85rem;
  border: 1px solid rgba(244,63,94,0.3);
}
.hm-lives {
  display: flex; align-items: center; gap: 8px;
}
.hm-lives-label { font-size: 0.85rem; font-weight: 700; color: var(--dark); }
.hm-hearts { display: flex; gap: 4px; }
.hm-heart {
  font-size: 1.2rem; filter: drop-shadow(0 0 4px rgba(245,158,11,0.8));
  transition: all 0.3s ease;
}
.hm-heart-lost {
  opacity: 0.2; filter: grayscale(1); transform: scale(0.8);
}

.hm-stage {
  display: flex; flex-direction: column; align-items: center; gap: 16px;
  margin-bottom: 24px;
}
.hm-visualizer {
  position: relative; width: 220px; height: 230px;
  background: rgba(255, 255, 255, 0.45);
  backdrop-filter: blur(14px);
  border-radius: 24px; border: 1.5px solid rgba(255, 255, 255, 0.7);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  display: flex; align-items: center; justify-content: center;
  padding: 10px;
}
.hm-svg { width: 100%; height: 100%; overflow: visible; }
.hm-part {
  animation: partAppear 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}
@keyframes partAppear {
  from { opacity: 0; transform: scale(0.6); }
  to { opacity: 1; transform: scale(1); }
}

.hm-taunt-bubble {
  position: absolute; bottom: 12px; left: 50%; transform: translateX(-50%);
  width: 90%; text-align: center;
  background: rgba(244, 63, 94, 0.95); color: white;
  padding: 8px 12px; border-radius: 14px;
  font-size: 0.78rem; font-weight: 800;
  box-shadow: 0 4px 16px rgba(244, 63, 94, 0.4);
  backdrop-filter: blur(6px);
  pointer-events: none;
}

.hm-hint-box {
  display: inline-flex; align-items: center; gap: 10px;
  background: rgba(244,63,94,0.06); border: 1px dashed rgba(244,63,94,0.35);
  padding: 10px 20px; border-radius: 16px; max-width: 500px;
}
.hm-hint-tag {
  font-weight: 900; font-size: 0.75rem; color: #e11d48;
  background: rgba(244,63,94,0.15); padding: 3px 8px; border-radius: 8px;
}
.hm-hint-text { margin: 0; font-size: 0.92rem; font-weight: 600; color: var(--dark); }

.hm-word-container {
  display: flex; flex-wrap: wrap; justify-content: center; gap: 10px;
  margin-bottom: 28px;
}
.hm-letter-slot {
  width: 44px; height: 54px;
  background: rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(10px);
  border-radius: 14px;
  border: 2px solid var(--border);
  display: flex; align-items: center; justify-content: center;
  font-size: 1.4rem; font-weight: 900; color: var(--dark);
  box-shadow: 0 4px 14px rgba(0,0,0,0.05);
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.hm-letter-revealed {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15), rgba(5, 150, 105, 0.25));
  border-color: #10b981; color: #065f46;
  transform: translateY(-4px) scale(1.05);
  box-shadow: 0 6px 18px rgba(16, 185, 129, 0.25);
}
.hm-mystery {
  color: var(--muted); opacity: 0.4;
}

.hm-keyboard {
  display: flex; flex-wrap: wrap; justify-content: center; gap: 8px;
  max-width: 680px; margin: 0 auto;
}
.hm-key {
  width: 44px; height: 46px; border-radius: 12px;
  background: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(8px);
  border: 1.5px solid var(--border);
  font-size: 0.95rem; font-weight: 800; color: var(--dark);
  cursor: pointer; transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 3px 10px rgba(0,0,0,0.06);
}
.hm-key:hover:not(:disabled) {
  transform: translateY(-3px) scale(1.08);
  border-color: #f43f5e; color: #f43f5e;
  box-shadow: 0 6px 16px rgba(244, 63, 94, 0.25);
}
.hm-key-ok {
  background: linear-gradient(135deg, #10b981, #059669) !important;
  color: white !important; border-color: #059669 !important;
  box-shadow: 0 4px 14px rgba(16,185,129,0.4) !important;
}
.hm-key-bad {
  background: linear-gradient(135deg, #f43f5e, #e11d48) !important;
  color: white !important; border-color: #e11d48 !important;
  opacity: 0.4; cursor: not-allowed;
}

/* ─── Responsive ─────────────────────────────────────────────────────────── */
@media (max-width: 600px) {
  .ia-header { padding: 16px 16px 0; }
  .game-wrap  { padding: 14px 16px 20px; }
  .memo-grid  { grid-template-columns: repeat(auto-fill, minmax(145px, 1fr)); gap: 12px; }
  .memo-card  { height: 185px; }
  .card-txt   { font-size: 0.88rem; }
  .ws-cell    { width: 26px; height: 26px; font-size: 0.72rem; }
  .ws-panel   { width: 100%; }
  .drag-cats  { grid-template-columns: 1fr; }
  .ia-title   { font-size: 1.1rem; }
}
</style>
