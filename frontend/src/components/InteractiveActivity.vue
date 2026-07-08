<script setup lang="ts">
/**
 * InteractiveActivity.vue — Minijuegos de aprendizaje
 *   5 - Memorama
 *   6 - Arrastrar y Soltar / Clasificar
 *   7 - Sopa de Letras
 *   8 - Completar Espacios
 *   9 - Ordenar Secuencia
 */
import { ref, watch, computed, onUnmounted, reactive } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'

const props = defineProps<{
  lesson: any
  cursoId: string
}>()

const emit = defineEmits<{
  (e: 'completed', data: { points: number; timeSecs: number }): void
}>()

// gameType se resuelve desde lesson_type (número) o type (string legacy)
const gameType = computed(() => {
  const lt = props.lesson?.lesson_type
  const t = props.lesson?.type
  if (lt !== undefined && lt !== null) return String(lt)
  if (t) {
    // mapear nombres string del enum al número
    const MAP: Record<string, string> = {
      LESSON_TYPE_GAME_MEMORY:     '5',
      LESSON_TYPE_GAME_DRAGDROP:   '6',
      LESSON_TYPE_GAME_WORDSEARCH: '7',
      LESSON_TYPE_GAME_FILLBLANK:  '8',
      LESSON_TYPE_GAME_ORDER:      '9',
    }
    if (MAP[t]) return MAP[t]
    // si viene como número string
    return String(t)
  }
  return '1'
})

const config = ref<any>({})
const isCompleted = ref(false)
const pointsEarned = ref(0)

// Temporizador
const startTime = ref(0)
const elapsedSecs = ref(0)
let timerInterval: ReturnType<typeof setInterval> | null = null

function startTimer() {
  if (timerInterval) clearInterval(timerInterval)
  startTime.value = Date.now()
  elapsedSecs.value = 0
  timerInterval = setInterval(() => {
    elapsedSecs.value = Math.floor((Date.now() - startTime.value) / 1000)
  }, 1000)
}
function stopTimer() {
  if (timerInterval) { clearInterval(timerInterval); timerInterval = null }
}
onUnmounted(() => stopTimer())

// ── Audio feedback (Web Audio API) ─────────────────────────────────────────
function playBeep(freq = 520, type: OscillatorType = 'sine', duration = 0.15) {
  try {
    const ctx = new (window.AudioContext || (window as any).webkitAudioContext)()
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.type = type
    osc.frequency.setValueAtTime(freq, ctx.currentTime)
    gain.gain.setValueAtTime(0.1, ctx.currentTime)
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + duration)
    osc.connect(gain)
    gain.connect(ctx.destination)
    osc.start()
    osc.stop(ctx.currentTime + duration)
  } catch {}
}
function soundSuccess() { playBeep(660, 'sine', 0.1); setTimeout(() => playBeep(880, 'triangle', 0.25), 100) }
function soundError() { playBeep(220, 'sawtooth', 0.2) }
function soundWin() {
  soundSuccess()
  setTimeout(() => playBeep(990, 'triangle', 0.3), 250)
  setTimeout(() => playBeep(1320, 'sine', 0.4), 450)
}

// ── Carga de configuración ──────────────────────────────────────────────────
function loadGame() {
  isCompleted.value = false
  pointsEarned.value = 0
  stopTimer()

  let parsed: any = {}
  try {
    const raw = props.lesson?.game_config_json
    if (raw) {
      parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
    }
  } catch (e) {
    console.warn('[InteractiveActivity] Error parseando game_config_json:', e)
  }

  config.value = parsed
  initSpecificGame()
  startTimer()
}

watch(
  () => [props.lesson?.id, props.lesson?.lesson_type, props.lesson?.type, props.lesson?.game_config_json],
  loadGame,
  { immediate: true }
)

async function handleGameWin() {
  if (isCompleted.value) return
  isCompleted.value = true
  stopTimer()
  soundWin()
  const pts = Number(props.lesson?.points_reward || 100)
  pointsEarned.value = pts

  try {
    if (!props.lesson?.completada) {
      await api.post(`/lecciones/${props.lesson.id}/game-score`, {
        curso_id: props.cursoId,
        points: pts,
        time_secs: elapsedSecs.value,
      })
      toast.success(`¡Felicidades! Has ganado +${pts} pts`)
    }
  } catch (e) {
    console.warn('[InteractiveActivity] Error al registrar puntaje:', e)
  }
  emit('completed', { points: pts, timeSecs: elapsedSecs.value })
}

// ─────────────────────────────────────────────────────────────────────────────
// 5: MEMORAMA
// ─────────────────────────────────────────────────────────────────────────────
const memoCards = ref<any[]>([])
const memoFlipped = ref<number[]>([])
const memoMatched = ref<number[]>([])  // pairIds encontrados (array para reactividad)
let memoLocked = false

function initMemo() {
  memoFlipped.value = []
  memoMatched.value = []
  memoLocked = false

  const pairs = (config.value?.pairs && config.value.pairs.length > 0)
    ? config.value.pairs
    : [
        { front: 'HTML', back: 'Estructura web' },
        { front: 'CSS', back: 'Estilos y diseño' },
        { front: 'JS', back: 'Interactividad' },
        { front: 'Vue', back: 'Framework reactivo' },
      ]

  const deck: any[] = []
  pairs.forEach((p: any, idx: number) => {
    const frontText = p.front || p.text_a || ''
    const frontImg  = p.front_img || p.img_a || ''
    const backText  = p.back || p.text_b || ''
    const backImg   = p.back_img || p.img_b || ''
    deck.push({ id: idx * 2,     pairId: idx, text: frontText, img: frontImg })
    deck.push({ id: idx * 2 + 1, pairId: idx, text: backText,  img: backImg  })
  })
  // Mezclar
  for (let i = deck.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[deck[i], deck[j]] = [deck[j], deck[i]]
  }
  memoCards.value = deck
}

function flipCard(card: any) {
  if (isCompleted.value || memoLocked) return
  if (memoMatched.value.includes(card.pairId)) return
  if (memoFlipped.value.includes(card.id)) return
  if (memoFlipped.value.length >= 2) return

  playBeep(440, 'sine', 0.08)
  memoFlipped.value = [...memoFlipped.value, card.id]

  if (memoFlipped.value.length === 2) {
    memoLocked = true
    const [id1, id2] = memoFlipped.value
    const c1 = memoCards.value.find(c => c.id === id1)
    const c2 = memoCards.value.find(c => c.id === id2)

    if (c1 && c2 && c1.pairId === c2.pairId) {
      setTimeout(() => {
        memoMatched.value = [...memoMatched.value, c1.pairId]
        memoFlipped.value = []
        memoLocked = false
        soundSuccess()
        // Win cuando todos los pares están encontrados
        const totalPairs = memoCards.value.length / 2
        if (memoMatched.value.length >= totalPairs) handleGameWin()
      }, 500)
    } else {
      setTimeout(() => {
        soundError()
        memoFlipped.value = []
        memoLocked = false
      }, 900)
    }
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// 6: CLASIFICAR / ARRASTRAR
// ─────────────────────────────────────────────────────────────────────────────
const dragCategories = ref<string[]>([])
const dragItems = ref<any[]>([])
const dragAssignments = reactive<Record<string, string>>({})
const selectedDragItem = ref<string | null>(null)

function initDrag() {
  dragCategories.value = config.value?.categories?.length
    ? config.value.categories
    : ['Frontend', 'Backend']

  const rawItems = config.value?.items?.length
    ? config.value.items
    : [
        { id: '1', text: 'Vue 3', correct_category: 'Frontend' },
        { id: '2', text: 'CSS3', correct_category: 'Frontend' },
        { id: '3', text: 'Go / gRPC', correct_category: 'Backend' },
        { id: '4', text: 'PostgreSQL', correct_category: 'Backend' },
      ]

  dragItems.value = rawItems
    .map((it: any, i: number) => ({ ...it, id: String(it.id ?? i), category: it.correct_category || it.category || dragCategories.value[0] }))
    .sort(() => Math.random() - 0.5)

  // Limpiar asignaciones
  Object.keys(dragAssignments).forEach(k => delete dragAssignments[k])
  selectedDragItem.value = null
}

function selectItem(id: string) {
  if (isCompleted.value) return
  selectedDragItem.value = selectedDragItem.value === id ? null : id
  playBeep(500, 'sine', 0.05)
}

function assignToCategory(cat: string) {
  if (!selectedDragItem.value || isCompleted.value) return
  dragAssignments[selectedDragItem.value] = cat
  selectedDragItem.value = null
  playBeep(550, 'sine', 0.1)
}

function removeAssignment(itemId: string) {
  delete dragAssignments[itemId]
}

function checkDrag() {
  if (Object.keys(dragAssignments).length < dragItems.value.length) {
    toast.error('Aún faltan elementos por clasificar.')
    return
  }
  const allCorrect = dragItems.value.every(it =>
    dragAssignments[it.id] === (it.correct_category || it.category)
  )
  if (allCorrect) {
    handleGameWin()
  } else {
    soundError()
    toast.error('Algunas clasificaciones no son correctas. ¡Revisa e inténtalo de nuevo!')
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// 7: SOPA DE LETRAS
// ─────────────────────────────────────────────────────────────────────────────
const wordGrid = ref<string[][]>([])
const wordsToFind = ref<string[]>([])
const foundWords = ref<string[]>([])
// Selección por arrastre
const wsSelecting = ref(false)
const wsStart = ref<[number, number] | null>(null)
const wsEnd = ref<[number, number] | null>(null)
// Palabras ya marcadas (celdas resaltadas permanentemente)
const wsFoundCells = ref<string[]>([])

function initWordSearch() {
  foundWords.value = []
  wsSelecting.value = false
  wsStart.value = null
  wsEnd.value = null
  wsFoundCells.value = []

  // Leer palabras desde config
  let rawWords: string[] = []
  const wsCfg = config.value
  if (wsCfg?.words && Array.isArray(wsCfg.words) && wsCfg.words.length > 0) {
    rawWords = wsCfg.words.filter((w: any) => String(w || '').trim().length > 0)
  } else if (wsCfg?.words && typeof wsCfg.words === 'string') {
    rawWords = String(wsCfg.words).split(',')
  }

  const validWords = rawWords
    .map(w => w.trim().normalize('NFD').replace(/[\u0300-\u036f]/g, '').replace(/[^a-zA-Z]/g, '').toUpperCase())
    .filter(w => w.length >= 2)

  wordsToFind.value = validWords.length > 0
    ? validWords
    : ['VUE', 'GOLANG', 'GRPC', 'HTML', 'CSS']

  buildGrid()
}

function buildGrid() {
  const words = wordsToFind.value
  const maxLen = Math.max(...words.map(w => w.length), 5)
  const size = Math.max(Number(config.value?.grid_size) || 12, maxLen + 2, 10)
  const grid: string[][] = Array.from({ length: size }, () => Array(size).fill(''))

  const DIRS = [
    [0, 1],   // →
    [1, 0],   // ↓
    [1, 1],   // ↘
    [0, -1],  // ←
    [-1, 0],  // ↑
    [-1, -1], // ↖
    [1, -1],  // ↙
    [-1, 1],  // ↗
  ]

  words.forEach(word => {
    let placed = false
    for (let attempt = 0; attempt < 300 && !placed; attempt++) {
      const [dr, dc] = DIRS[Math.floor(Math.random() * DIRS.length)]!
      const r = Math.floor(Math.random() * size)
      const c = Math.floor(Math.random() * size)
      const endR = r + dr * (word.length - 1)
      const endC = c + dc * (word.length - 1)
      if (endR < 0 || endR >= size || endC < 0 || endC >= size) continue

      let canPlace = true
      for (let i = 0; i < word.length && canPlace; i++) {
        const cell = grid[r + dr * i]?.[c + dc * i]
        if (cell !== '' && cell !== word[i]) canPlace = false
      }

      if (canPlace) {
        for (let i = 0; i < word.length; i++) {
          grid[r + dr * i]![c + dc * i] = word[i]!
        }
        placed = true
      }
    }
  })

  // Rellenar huecos
  const ALPHA = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  for (let r = 0; r < size; r++) {
    for (let c = 0; c < size; c++) {
      if (!grid[r]![c]) grid[r]![c] = ALPHA[Math.floor(Math.random() * ALPHA.length)]!
    }
  }

  wordGrid.value = grid
}

// Rango de celdas en línea recta entre start y end
function getCellsInLine(start: [number, number], end: [number, number]): string[] {
  const [r1, c1] = start
  const [r2, c2] = end
  const dr = Math.sign(r2 - r1)
  const dc = Math.sign(c2 - c1)
  const steps = Math.max(Math.abs(r2 - r1), Math.abs(c2 - c1))

  // Validar que sea una línea recta (horizontal, vertical o diagonal 45°)
  if (r1 !== r2 && c1 !== c2 && Math.abs(r2 - r1) !== Math.abs(c2 - c1)) return []

  const cells: string[] = []
  for (let i = 0; i <= steps; i++) {
    cells.push(`${r1 + dr * i},${c1 + dc * i}`)
  }
  return cells
}

const wsPreviewCells = computed(() => {
  if (!wsSelecting.value || !wsStart.value || !wsEnd.value) return []
  return getCellsInLine(wsStart.value, wsEnd.value)
})

function wsMouseDown(r: number, c: number) {
  if (isCompleted.value) return
  wsSelecting.value = true
  wsStart.value = [r, c]
  wsEnd.value = [r, c]
}

function wsMouseEnter(r: number, c: number) {
  if (!wsSelecting.value) return
  wsEnd.value = [r, c]
}

function wsMouseUp() {
  if (!wsSelecting.value || !wsStart.value || !wsEnd.value) { wsSelecting.value = false; return }

  const cells = getCellsInLine(wsStart.value, wsEnd.value)
  if (cells.length < 2) { wsSelecting.value = false; wsStart.value = null; wsEnd.value = null; return }

  const letters = cells.map(key => {
    const [r, c] = key.split(',').map(Number)
    return wordGrid.value[r!]?.[c!] ?? ''
  }).join('')

  const rev = letters.split('').reverse().join('')
  const found = wordsToFind.value.find(w => !foundWords.value.includes(w) && (letters === w || rev === w))

  if (found) {
    foundWords.value = [...foundWords.value, found]
    wsFoundCells.value = [...wsFoundCells.value, ...cells]
    soundSuccess()
    playBeep(600, 'sine', 0.08)
    if (foundWords.value.length >= wordsToFind.value.length) handleGameWin()
  } else {
    soundError()
  }

  wsSelecting.value = false
  wsStart.value = null
  wsEnd.value = null
}

function wsCellClass(r: number, c: number): string[] {
  const key = `${r},${c}`
  const cls = ['ws-cell']
  if (wsFoundCells.value.includes(key)) cls.push('found')
  else if (wsPreviewCells.value.includes(key)) cls.push('preview')
  return cls
}

// ─────────────────────────────────────────────────────────────────────────────
// 8: COMPLETAR ESPACIOS
// ─────────────────────────────────────────────────────────────────────────────
const fbSentences = ref<any[]>([])
const fbAnswers = ref<Record<number, string>>({})

function initFillBlank() {
  fbAnswers.value = {}
  fbSentences.value = (config.value?.sentences && config.value.sentences.length > 0)
    ? config.value.sentences
    : [
        { text: 'Para estilizar una página web usamos ___ y para la estructura ___', answer: 'CSS', options: ['CSS', 'Python', 'SQL'] },
        { text: 'El framework de JavaScript progresivo que usamos es ___', answer: 'Vue 3', options: ['Vue 3', 'Django', 'Laravel'] },
      ]
}

function checkFillBlank() {
  if (isCompleted.value) return
  const unanswered = fbSentences.value.some((_s: any, idx: number) => !fbAnswers.value[idx])
  if (unanswered) { toast.error('Responde todas las preguntas antes de validar.'); return }

  const allCorrect = fbSentences.value.every((s: any, idx: number) =>
    (fbAnswers.value[idx] ?? '').trim().toLowerCase() === (s.answer ?? '').trim().toLowerCase()
  )
  if (allCorrect) {
    handleGameWin()
  } else {
    soundError()
    toast.error('Hay respuestas incorrectas. ¡Revisa e inténtalo de nuevo!')
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// 9: ORDENAR SECUENCIA
// ─────────────────────────────────────────────────────────────────────────────
const orderItems = ref<any[]>([])

function initOrder() {
  const rawItems = (config.value?.items && config.value.items.length > 0)
    ? config.value.items
    : [
        { text: 'Diseñar los wireframes / UI', correct_order: 1 },
        { text: 'Configurar base de datos y repositorios', correct_order: 2 },
        { text: 'Desarrollar lógica de negocio en backend', correct_order: 3 },
        { text: 'Integrar componentes en frontend y probar', correct_order: 4 },
      ]
  // Mezclar y asignar ID interno
  orderItems.value = rawItems
    .map((it: any, idx: number) => ({ ...it, _id: idx }))
    .sort(() => Math.random() - 0.5)
}

function moveOrderItem(idx: number, dir: -1 | 1) {
  if (isCompleted.value) return
  const j = idx + dir
  if (j < 0 || j >= orderItems.value.length) return
  const arr = [...orderItems.value]
  ;[arr[idx]!, arr[j]!] = [arr[j]!, arr[idx]!]
  orderItems.value = arr
  playBeep(500, 'triangle', 0.08)
}

function checkOrder() {
  if (isCompleted.value) return
  const correct = orderItems.value.every((it, idx) => it.correct_order === idx + 1)
  if (correct) {
    handleGameWin()
  } else {
    soundError()
    toast.error('El orden no es correcto todavía. ¡Sigue intentando!')
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Dispatcher
// ─────────────────────────────────────────────────────────────────────────────
function initSpecificGame() {
  const t = gameType.value
  if      (t === '5') initMemo()
  else if (t === '6') initDrag()
  else if (t === '7') initWordSearch()
  else if (t === '8') initFillBlank()
  else if (t === '9') initOrder()
}

function restartGame() {
  isCompleted.value = false
  pointsEarned.value = 0
  stopTimer()
  initSpecificGame()
  startTimer()
}
</script>

<template>
  <div class="game-activity" @mouseup.self="wsMouseUp">
    <!-- Header del juego -->
    <div class="game-header">
      <div class="game-title-wrap">
        <span class="game-badge glass-badge-glow">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="6" width="20" height="12" rx="6"/><path d="M6 12h4m-2-2v4"/><circle cx="17" cy="11" r="1" fill="currentColor"/><circle cx="15" cy="13" r="1" fill="currentColor"/></svg>
          Actividad Interactiva
        </span>
        <h2 class="game-title">{{ lesson?.title }}</h2>
      </div>
      <div class="game-stats">
        <div class="stat-chip glass-chip">
          <span class="stat-icon-glass">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </span>
          <span class="stat-val">{{ elapsedSecs }}s</span>
        </div>
        <div class="stat-chip glass-chip stat-pts">
          <span class="stat-icon-glass star-glow">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
          </span>
          <span class="stat-val">+{{ lesson?.points_reward || 100 }} pts</span>
        </div>
      </div>
    </div>

    <p v-if="lesson?.description" class="game-desc">{{ lesson.description }}</p>

    <!-- ── Victoria ────────────────────────────────────────────────── -->
    <div v-if="isCompleted" class="game-win-banner slide-down">
      <div class="win-icon glass-trophy-box">
        <svg width="42" height="42" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/>
          <path d="M4 22h16"/><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20.24 7 22"/>
          <path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20.24 17 22"/>
          <path d="M18 2H6v7a6 6 0 0 0 12 0V2Z"/>
        </svg>
      </div>
      <div class="win-info">
        <h3>¡Reto Superado con Éxito!</h3>
        <p>Completado en <strong>{{ elapsedSecs }}s</strong> · <strong>+{{ pointsEarned }} puntos</strong> ganados</p>
      </div>
      <button class="btn btn-primary glass-btn-action" @click="restartGame">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right:6px"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
        Jugar de nuevo
      </button>
    </div>

    <!-- ── 5: MEMORAMA ───────────────────────────────────────────────── -->
    <div v-if="gameType === '5'" class="game-area memo-area">
      <p class="game-instruct">Encuentra todos los pares haciendo clic en las tarjetas para voltearlas:</p>
      <div class="memo-grid">
        <button
          v-for="card in memoCards"
          :key="card.id"
          :class="['memo-card',
            memoFlipped.includes(card.id) || memoMatched.includes(card.pairId) ? 'flipped' : '',
            memoMatched.includes(card.pairId) ? 'matched' : '']"
          :disabled="memoMatched.includes(card.pairId) || isCompleted"
          @click="flipCard(card)"
        >
          <div class="card-inner">
            <div class="card-front glass-card-front">
              <div class="glass-logo-circle">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
              </div>
            </div>
            <div class="card-back">
              <img v-if="card.img" :src="card.img" class="card-img-content" :alt="card.text || 'Imagen'" />
              <span v-if="card.text" class="card-text-content">{{ card.text }}</span>
            </div>
          </div>
        </button>
      </div>
      <div class="memo-progress">
        {{ memoMatched.length }} / {{ memoCards.length / 2 }} pares encontrados
      </div>
    </div>

    <!-- ── 6: CLASIFICAR / ARRASTRAR ────────────────────────────────── -->
    <div v-else-if="gameType === '6'" class="game-area drag-area">
      <p class="game-instruct">Selecciona un elemento y luego haz clic en la categoría donde pertenece:</p>

      <div class="drag-items-pool">
        <button
          v-for="it in dragItems"
          :key="it.id"
          :class="['drag-item-btn',
            selectedDragItem === it.id ? 'active' : '',
            dragAssignments[it.id] ? 'assigned' : '']"
          :disabled="isCompleted"
          @click="selectItem(it.id)"
        >
          <span>{{ it.text }}</span>
          <span v-if="dragAssignments[it.id]" class="assigned-badge">→ {{ dragAssignments[it.id] }}</span>
        </button>
      </div>

      <div class="drag-categories-grid">
        <div
          v-for="cat in dragCategories"
          :key="cat"
          :class="['category-box', selectedDragItem ? 'droppable' : '']"
          @click="assignToCategory(cat)"
        >
          <h4 class="cat-title glass-cat-header">
            <span class="glass-icon-circle">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
            </span>
            {{ cat }}
          </h4>
          <div class="cat-items-list">
            <div
              v-for="it in dragItems.filter(i => dragAssignments[i.id] === cat)"
              :key="it.id"
              class="cat-assigned-chip"
            >
              {{ it.text }}
              <button class="remove-assign" @click.stop="removeAssignment(it.id)" :disabled="isCompleted">✕</button>
            </div>
            <span v-if="!dragItems.some(i => dragAssignments[i.id] === cat)" class="cat-empty-hint">
              {{ selectedDragItem ? 'Haz clic aquí para asignar' : 'Vacío' }}
            </span>
          </div>
        </div>
      </div>

      <div class="game-actions">
        <button class="btn btn-primary btn-lg glass-action-btn" @click="checkDrag" :disabled="isCompleted">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right:8px"><polyline points="20 6 9 17 4 12"/></svg>
          Verificar Clasificación
        </button>
      </div>
    </div>

    <!-- ── 7: SOPA DE LETRAS ─────────────────────────────────────────── -->
    <div v-else-if="gameType === '7'" class="game-area ws-area" @mouseup="wsMouseUp" @mouseleave="wsMouseUp">
      <p class="game-instruct">Arrastra el ratón sobre las letras para seleccionar las palabras ocultas:</p>
      <div class="ws-layout">
        <div class="ws-grid-wrap">
          <div class="ws-grid" :style="{ userSelect: 'none' }">
            <div v-for="(row, r) in wordGrid" :key="r" class="ws-row">
              <button
                v-for="(char, c) in row"
                :key="c"
                :class="wsCellClass(r, c)"
                @mousedown.prevent="wsMouseDown(r, c)"
                @mouseenter="wsMouseEnter(r, c)"
                @touchstart.prevent="wsMouseDown(r, c)"
              >
                {{ char }}
              </button>
            </div>
          </div>
        </div>
        <div class="ws-sidebar">
          <h4 class="ws-sidebar-title glass-subhead">
            <span class="glass-icon-circle">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            </span>
            Palabras ({{ foundWords.length }}/{{ wordsToFind.length }}):
          </h4>
          <ul class="ws-word-list">
            <li v-for="w in wordsToFind" :key="w" :class="{ found: foundWords.includes(w) }">
              <span class="word-chk">
                <svg v-if="foundWords.includes(w)" class="chk-svg found-chk" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><rect x="3" y="3" width="18" height="18" rx="5" fill="rgba(16, 185, 129, 0.2)" stroke="#10B981"/><polyline points="8 12 11 15 16 9" stroke="#10B981"/></svg>
                <svg v-else class="chk-svg empty-chk" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="5" fill="rgba(255,255,255,0.05)" stroke="currentColor" stroke-opacity="0.3"/></svg>
              </span>
              <span class="word-txt">{{ w }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- ── 8: COMPLETAR ESPACIOS ─────────────────────────────────────── -->
    <div v-else-if="gameType === '8'" class="game-area fb-area">
      <p class="game-instruct">Lee atentamente cada oración y elige la respuesta correcta para llenar el espacio:</p>
      <div class="fb-list">
        <div v-for="(sent, idx) in fbSentences" :key="idx" class="fb-card">
          <div class="fb-sent-num">{{ idx + 1 }}</div>
          <div class="fb-sent-body">
            <p class="fb-sent-text">{{ sent.text }}</p>
            <div class="fb-options-grid">
              <label
                v-for="opt in sent.options"
                :key="opt"
                :class="['fb-option-chip', fbAnswers[idx] === opt ? 'selected' : '']"
              >
                <input type="radio" :name="`fb_${idx}`" :value="opt" v-model="fbAnswers[idx]" class="hidden-radio" :disabled="isCompleted" />
                <span>{{ opt }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>
      <div class="game-actions">
        <button class="btn btn-primary btn-lg glass-action-btn" @click="checkFillBlank" :disabled="isCompleted">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right:8px"><polyline points="20 6 9 17 4 12"/></svg>
          Validar Respuestas
        </button>
      </div>
    </div>

    <!-- ── 9: ORDENAR SECUENCIA ──────────────────────────────────────── -->
    <div v-else-if="gameType === '9'" class="game-area order-area">
      <p class="game-instruct">Ordena los siguientes pasos cronológicamente usando las flechas:</p>
      <div class="order-list">
        <div v-for="(it, idx) in orderItems" :key="it._id" class="order-card">
          <span class="order-pos">{{ idx + 1 }}</span>
          <span class="order-text">{{ it.text }}</span>
          <div class="order-controls">
            <button class="btn-order-move" :disabled="idx === 0 || isCompleted" @click="moveOrderItem(idx, -1)" title="Subir">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M18 15l-6-6-6 6"/></svg>
            </button>
            <button class="btn-order-move" :disabled="idx === orderItems.length - 1 || isCompleted" @click="moveOrderItem(idx, 1)" title="Bajar">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9l6 6 6-6"/></svg>
            </button>
          </div>
        </div>
      </div>
      <div class="game-actions">
        <button class="btn btn-primary btn-lg glass-action-btn" @click="checkOrder" :disabled="isCompleted">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right:8px"><polyline points="20 6 9 17 4 12"/></svg>
          Verificar Orden
        </button>
      </div>
    </div>

    <!-- Tipo de lección no es juego (fallback) -->
    <div v-else class="game-area">
      <p class="game-instruct">Esta lección no tiene un minijuego configurado aún.</p>
    </div>
  </div>
</template>

<style scoped>
.game-activity {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-xl);
  padding: 28px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.04);
}

/* ─── Header ────────────────────────────────────────────────────── */
.game-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.game-title-wrap { display: flex; flex-direction: column; gap: 6px; }
.game-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--brand);
  background: rgba(99,102,241,0.1);
  border: 1px solid rgba(99,102,241,0.25);
  border-radius: 20px;
  padding: 3px 10px;
  width: fit-content;
  letter-spacing: 0.04em;
}
.game-title { font-size: 1.3rem; font-weight: 800; color: var(--dark); margin: 0; }
.game-desc { font-size: 0.9rem; color: var(--muted); margin-bottom: 20px; }
.game-stats { display: flex; gap: 10px; flex-wrap: wrap; }
.stat-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 20px;
  background: var(--surface-soft);
  border: 1px solid var(--border);
  font-size: 0.85rem;
}
.stat-icon-glass { display: flex; align-items: center; color: var(--muted); }
.stat-pts .stat-icon-glass { color: #f59e0b; }
.star-glow { filter: drop-shadow(0 0 4px rgba(245,158,11,0.5)); }
.stat-val { font-weight: 700; color: var(--dark); }

/* ─── Instruct & Area ──────────────────────────────────────────── */
.game-instruct { font-size: 0.9rem; color: var(--muted); margin-bottom: 20px; }
.game-area { margin-top: 16px; }
.game-actions { margin-top: 24px; display: flex; justify-content: center; }

/* ─── Victory Banner ────────────────────────────────────────────── */
.game-win-banner {
  display: flex;
  align-items: center;
  gap: 16px;
  background: linear-gradient(135deg, rgba(16,185,129,0.12), rgba(99,102,241,0.1));
  border: 1px solid rgba(16,185,129,0.3);
  border-radius: var(--r-lg);
  padding: 20px 24px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.win-icon { display: flex; align-items: center; justify-content: center; }
.glass-trophy-box {
  width: 64px; height: 64px;
  background: rgba(245,158,11,0.12);
  border: 1.5px solid rgba(245,158,11,0.3);
  border-radius: 16px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.win-info { flex: 1; }
.win-info h3 { font-size: 1.1rem; font-weight: 800; color: var(--dark); margin: 0 0 4px 0; }
.win-info p { font-size: 0.9rem; color: var(--muted); margin: 0; }
.glass-btn-action {
  background: var(--brand);
  color: white;
  border: none;
  border-radius: var(--r-md);
  padding: 10px 20px;
  font-weight: 700;
  cursor: pointer;
  display: flex; align-items: center;
  transition: opacity 0.2s;
}
.glass-btn-action:hover { opacity: 0.9; }

/* ─── Memorama ─────────────────────────────────────────────────── */
.memo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 12px;
}
.memo-card {
  width: 100%;
  aspect-ratio: 3/4;
  perspective: 1000px;
  cursor: pointer;
  border: none;
  background: transparent;
  padding: 0;
}
.card-inner {
  position: relative;
  width: 100%;
  height: 100%;
  transition: transform 0.5s cubic-bezier(0.25, 0.8, 0.25, 1);
  transform-style: preserve-3d;
  border-radius: var(--r-lg);
}
.memo-card.flipped .card-inner { transform: rotateY(180deg); }
.card-front,
.card-back {
  position: absolute;
  inset: 0;
  backface-visibility: hidden;
  border-radius: var(--r-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px;
  border: 1.5px solid var(--border);
  overflow: hidden;
}
.card-front {
  background: var(--surface-soft);
}
.glass-card-front {
  background: linear-gradient(135deg, rgba(99,102,241,0.12), rgba(139,92,246,0.1));
}
.glass-logo-circle {
  width: 44px; height: 44px;
  border-radius: 50%;
  background: rgba(99,102,241,0.15);
  border: 1.5px solid rgba(99,102,241,0.3);
  display: flex; align-items: center; justify-content: center;
  color: var(--brand);
}
.card-back {
  background: white;
  transform: rotateY(180deg);
  gap: 6px;
}
.card-img-content { max-width: 100%; max-height: 60%; object-fit: contain; border-radius: 6px; }
.card-text-content { font-size: 0.85rem; font-weight: 700; color: var(--dark); text-align: center; word-break: break-word; }
.memo-card.matched .card-back { background: rgba(16,185,129,0.12); border-color: #10b981; }
.memo-card:disabled { cursor: default; }
.memo-progress {
  text-align: center;
  margin-top: 16px;
  font-size: 0.85rem;
  color: var(--muted);
  font-weight: 600;
}

/* ─── Drag & Drop ─────────────────────────────────────────────── */
.drag-items-pool {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 20px;
  padding: 16px;
  background: var(--surface-soft);
  border-radius: var(--r-lg);
  border: 1px solid var(--border);
  min-height: 60px;
}
.drag-item-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 20px;
  border: 1.5px solid var(--border);
  background: var(--surface);
  cursor: pointer;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--dark);
  transition: all 0.18s;
}
.drag-item-btn:hover:not(:disabled) { border-color: var(--brand); background: rgba(99,102,241,0.06); }
.drag-item-btn.active { border-color: var(--brand); background: rgba(99,102,241,0.14); box-shadow: 0 0 0 3px rgba(99,102,241,0.18); }
.drag-item-btn.assigned { opacity: 0.55; }
.assigned-badge { font-size: 0.75rem; font-weight: 700; color: var(--brand); }
.drag-categories-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 14px;
}
.category-box {
  background: var(--surface-soft);
  border: 1.5px solid var(--border);
  border-radius: var(--r-lg);
  padding: 14px;
  cursor: pointer;
  transition: border-color 0.18s, box-shadow 0.18s;
  min-height: 100px;
}
.category-box.droppable { border-color: var(--brand); box-shadow: 0 0 0 3px rgba(99,102,241,0.12); }
.cat-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  font-weight: 800;
  color: var(--dark);
  margin-bottom: 10px;
}
.glass-cat-header { color: var(--dark); }
.cat-items-list { display: flex; flex-direction: column; gap: 6px; }
.cat-assigned-chip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  background: rgba(99,102,241,0.1);
  border-radius: 8px;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--brand);
}
.remove-assign {
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--muted);
  font-size: 0.75rem;
  line-height: 1;
  padding: 2px 4px;
}
.remove-assign:hover { color: #ef4444; }
.cat-empty-hint { font-size: 0.8rem; color: var(--muted); font-style: italic; }

/* ─── Word Search ──────────────────────────────────────────────── */
.ws-area { position: relative; }
.ws-layout { display: flex; gap: 24px; flex-wrap: wrap; }
.ws-grid-wrap { flex: 1; min-width: 260px; overflow-x: auto; }
.ws-grid {
  display: inline-flex;
  flex-direction: column;
  gap: 3px;
  background: var(--surface-soft);
  padding: 12px;
  border-radius: var(--r-lg);
  border: 1px solid var(--border);
  cursor: crosshair;
}
.ws-row { display: flex; gap: 3px; }
.ws-cell {
  width: 32px;
  height: 32px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 5px;
  font-weight: 800;
  font-size: 0.88rem;
  color: var(--dark);
  cursor: crosshair;
  transition: background 0.1s, border-color 0.1s;
  display: flex;
  align-items: center;
  justify-content: center;
  user-select: none;
  -webkit-user-select: none;
  touch-action: none;
}
.ws-cell:hover { background: rgba(99,102,241,0.08); border-color: var(--brand); }
.ws-cell.preview { background: rgba(99,102,241,0.2); border-color: var(--brand); }
.ws-cell.found { background: rgba(16,185,129,0.2); border-color: #10b981; color: #065f46; font-weight: 900; }
.ws-sidebar {
  width: 200px;
  background: var(--surface-soft);
  border-radius: var(--r-lg);
  padding: 18px;
  border: 1px solid var(--border);
  align-self: flex-start;
}
.ws-sidebar-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  font-weight: 800;
  margin: 0 0 12px 0;
  color: var(--dark);
}
.ws-word-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 8px; }
.ws-word-list li { display: flex; align-items: center; gap: 8px; font-size: 0.88rem; font-weight: 700; color: var(--dark); }
.ws-word-list li.found { text-decoration: line-through; color: #10b981; }
.word-chk { display: flex; align-items: center; flex-shrink: 0; }

/* ─── Fill Blank ───────────────────────────────────────────────── */
.fb-list { display: flex; flex-direction: column; gap: 16px; }
.fb-card {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  background: var(--surface-soft);
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  padding: 16px 20px;
}
.fb-sent-num {
  width: 28px; height: 28px;
  background: var(--brand);
  color: white;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.8rem; font-weight: 800;
  flex-shrink: 0;
}
.fb-sent-body { flex: 1; }
.fb-sent-text { font-size: 0.9rem; color: var(--dark); margin: 0 0 12px 0; font-weight: 500; }
.fb-options-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.fb-option-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 20px;
  border: 1.5px solid var(--border);
  background: var(--surface);
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--dark);
  transition: all 0.15s;
}
.fb-option-chip:hover { border-color: var(--brand); background: rgba(99,102,241,0.06); }
.fb-option-chip.selected { border-color: var(--brand); background: rgba(99,102,241,0.14); color: var(--brand); }
.hidden-radio { display: none; }

/* ─── Order ────────────────────────────────────────────────────── */
.order-list { display: flex; flex-direction: column; gap: 10px; }
.order-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  background: var(--surface-soft);
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  transition: box-shadow 0.15s;
}
.order-card:hover { box-shadow: 0 2px 12px rgba(0,0,0,0.06); }
.order-pos {
  width: 28px; height: 28px;
  background: var(--brand);
  color: white;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.8rem; font-weight: 800;
  flex-shrink: 0;
}
.order-text { flex: 1; font-size: 0.9rem; font-weight: 500; color: var(--dark); }
.order-controls { display: flex; flex-direction: column; gap: 4px; }
.btn-order-move {
  width: 26px; height: 26px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
  color: var(--muted);
}
.btn-order-move:hover:not(:disabled) { background: rgba(99,102,241,0.1); border-color: var(--brand); color: var(--brand); }
.btn-order-move:disabled { opacity: 0.3; cursor: not-allowed; }

/* ─── Common ────────────────────────────────────────────────────── */
.glass-icon-circle {
  width: 24px; height: 24px;
  background: rgba(99,102,241,0.12);
  border: 1px solid rgba(99,102,241,0.25);
  border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  color: var(--brand);
  flex-shrink: 0;
}
.glass-subhead { color: var(--dark); }
.btn-lg { padding: 12px 28px; font-size: 0.95rem; }
.glass-action-btn {
  background: var(--brand);
  color: white;
  border: none;
  border-radius: var(--r-md);
  padding: 12px 28px;
  font-weight: 700;
  font-size: 0.95rem;
  cursor: pointer;
  display: flex; align-items: center;
  transition: opacity 0.2s, transform 0.1s;
}
.glass-action-btn:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.glass-action-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.slide-down { animation: slideDown 0.35s cubic-bezier(0.25,0.8,0.25,1); }
@keyframes slideDown {
  from { transform: translateY(-12px); opacity: 0; }
  to   { transform: translateY(0);     opacity: 1; }
}

@media (max-width: 600px) {
  .game-activity { padding: 16px; }
  .ws-cell { width: 26px; height: 26px; font-size: 0.75rem; }
  .memo-grid { grid-template-columns: repeat(auto-fill, minmax(90px, 1fr)); }
  .ws-sidebar { width: 100%; }
}
</style>
