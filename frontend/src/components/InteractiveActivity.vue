<script setup lang="ts">
/**
 * InteractiveActivity.vue
 * Componente que renderiza los 5 minijuegos de aprendizaje en VerCapacitacion:
 *   5 - Memorama
 *   6 - Arrastrar y Soltar / Clasificar
 *   7 - Sopa de Letras
 *   8 - Completar Espacios
 *   9 - Ordenar Secuencia
 */
import { ref, watch, computed, onMounted, onUnmounted } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'

const props = defineProps<{
  lesson: any
  cursoId: string
}>()

const emit = defineEmits<{
  (e: 'completed', data: { points: number; timeSecs: number }): void
}>()

const gameType = computed(() => String(props.lesson?.lesson_type ?? props.lesson?.type ?? '1'))
const config = ref<any>({})
const isCompleted = ref(false)
const pointsEarned = ref(0)

// Temporizador
const startTime = ref(0)
const elapsedSecs = ref(0)
let timerInterval: any = null

function startTimer() {
  clearInterval(timerInterval)
  startTime.value = Date.now()
  elapsedSecs.value = 0
  timerInterval = setInterval(() => {
    elapsedSecs.value = Math.floor((Date.now() - startTime.value) / 1000)
  }, 1000)
}
function stopTimer() {
  clearInterval(timerInterval)
}

onUnmounted(() => stopTimer())

// Sonidos sintetizados simples (Web Audio API) para feedback inmediato
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

// Cargar y parsear configuración
function loadGame() {
  isCompleted.value = false
  pointsEarned.value = 0
  stopTimer()
  
  let parsed = {}
  try {
    if (props.lesson?.game_config_json) {
      parsed = typeof props.lesson.game_config_json === 'string'
        ? JSON.parse(props.lesson.game_config_json)
        : props.lesson.game_config_json
    }
  } catch (e) {
    console.warn('Error parseando game_config_json:', e)
  }
  
  config.value = parsed
  initSpecificGame()
  startTimer()
}

watch(() => [props.lesson?.id, props.lesson?.game_config_json, props.lesson?.lesson_type, props.lesson?.type], loadGame, { immediate: true })

async function handleGameWin() {
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
      toast.success(`¡Felicidades! Has ganado +${pts} pts 🌟`)
    }
  } catch (e) {
    console.warn('Error al registrar puntaje:', e)
  }
  emit('completed', { points: pts, timeSecs: elapsedSecs.value })
}

// ── 5: MEMORAMA ──────────────────────────────────────────────────────────────
const memoCards = ref<any[]>([])
const memoFlipped = ref<number[]>([])
const memoMatched = ref<Set<number>>(new Set())

function initMemo() {
  memoFlipped.value = []
  memoMatched.value = new Set()
  const pairs = config.value?.pairs || [
    { front: 'HTML', back: 'Estructura web' },
    { front: 'CSS', back: 'Estilos y diseño' },
    { front: 'JS', back: 'Interactividad' },
    { front: 'Vue', back: 'Framework reactivo' },
  ]
  const deck: any[] = []
  pairs.forEach((p: any, idx: number) => {
    const textA = p.front || p.text_a || ''
    const imgA = p.front_img || p.image_url || p.img_a || ''
    const textB = p.back || p.text_b || ''
    const imgB = p.back_img || p.img_b || ''
    deck.push({ id: idx * 2, pairId: idx, text: textA, img: imgA, type: 'A' })
    deck.push({ id: idx * 2 + 1, pairId: idx, text: textB, img: imgB, type: 'B' })
  })
  // Mezclar
  memoCards.value = deck.sort(() => Math.random() - 0.5)
}

function flipCard(card: any) {
  if (isCompleted.value || memoMatched.value.has(card.pairId)) return
  if (memoFlipped.value.includes(card.id) || memoFlipped.value.length >= 2) return
  
  playBeep(440, 'sine', 0.08)
  memoFlipped.value.push(card.id)
  
  if (memoFlipped.value.length === 2) {
    const c1 = memoCards.value.find(c => c.id === memoFlipped.value[0])
    const c2 = memoCards.value.find(c => c.id === memoFlipped.value[1])
    if (c1 && c2 && c1.pairId === c2.pairId) {
      setTimeout(() => {
        memoMatched.value.add(c1.pairId)
        memoFlipped.value = []
        soundSuccess()
        if (memoMatched.value.size * 2 === memoCards.value.length) handleGameWin()
      }, 400)
    } else {
      setTimeout(() => {
        soundError()
        memoFlipped.value = []
      }, 800)
    }
  }
}

// ── 6: ARRASTRAR Y SOLTAR / CLASIFICAR ────────────────────────────────────────
const dragCategories = ref<string[]>([])
const dragItems = ref<any[]>([])
const dragAssignments = ref<Record<string, string>>({}) // itemId -> category
const selectedDragItem = ref<string | null>(null)

function initDrag() {
  dragCategories.value = config.value?.categories || ['Frontend', 'Backend']
  const rawItems = config.value?.items || [
    { id: '1', text: 'Vue 3', category: 'Frontend' },
    { id: '2', text: 'CSS3', category: 'Frontend' },
    { id: '3', text: 'Go / gRPC', category: 'Backend' },
    { id: '4', text: 'PostgreSQL', category: 'Backend' },
  ]
  dragItems.value = rawItems.map((it: any, i: number) => ({ ...it, id: String(it.id || i) })).sort(() => Math.random() - 0.5)
  dragAssignments.value = {}
  selectedDragItem.value = null
}

function assignCategory(cat: string) {
  if (!selectedDragItem.value) return
  dragAssignments.value[selectedDragItem.value] = cat
  selectedDragItem.value = null
  playBeep(550, 'sine', 0.1)
  checkDragWin()
}

function selectItemForCategory(id: string) {
  selectedDragItem.value = selectedDragItem.value === id ? null : id
}

function checkDragWin() {
  if (Object.keys(dragAssignments.value).length < dragItems.value.length) return
  const allCorrect = dragItems.value.every(it => dragAssignments.value[it.id] === it.category)
  if (allCorrect) {
    handleGameWin()
  } else {
    soundError()
    toast.error('Algunas clasificaciones no son correctas. Revisa tus respuestas.')
  }
}

// ── 7: SOPA DE LETRAS ────────────────────────────────────────────────────────
const wordGrid = ref<string[][]>([])
const wordsToFind = ref<string[]>([])
const foundWords = ref<Set<string>>(new Set())
const selectedCells = ref<string[]>([]) // "r,c"

function initWordSearch() {
  foundWords.value = new Set()
  selectedCells.value = []
  
  let rawWords: any[] = ['VUE', 'GOLANG', 'GRPC', 'STRIPE', 'CLOUD']
  if (config.value && Array.isArray(config.value.words) && config.value.words.length > 0) {
    rawWords = config.value.words
  } else if (config.value && typeof config.value.words === 'string') {
    rawWords = String(config.value.words).split(',')
  }
  
  const validWords = rawWords
    .map((w: any) => String(w || '').trim().replace(/[^a-zA-ZáéíóúñÁÉÍÓÚÑ]/g, '').toUpperCase())
    .filter((w: string) => w.length >= 2)
  
  wordsToFind.value = validWords.length > 0 ? validWords : ['VUE', 'GOLANG', 'GRPC', 'STRIPE', 'CLOUD']
  
  const maxWordLen = Math.max(...wordsToFind.value.map(w => w.length), 8)
  const size = Math.max(Number(config.value?.grid_size) || 12, maxWordLen + 1)
  const grid: string[][] = Array(size).fill(0).map(() => Array(size).fill(''))
  
  // Colocar palabras en horizontal/vertical simple
  wordsToFind.value.forEach(word => {
    let placed = false
    let attempts = 0
    while (!placed && attempts < 150) {
      attempts++
      const dir = Math.random() > 0.5 ? 'H' : 'V'
      const r = Math.floor(Math.random() * (dir === 'V' ? size - word.length + 1 : size))
      const c = Math.floor(Math.random() * (dir === 'H' ? size - word.length + 1 : size))
      
      let canPlace = true
      for (let i = 0; i < word.length; i++) {
        const rr = dir === 'V' ? r + i : r
        const cc = dir === 'H' ? c + i : c
        const row = grid[rr]
        if (!row || (row[cc] !== '' && row[cc] !== word[i])) { canPlace = false; break }
      }
      if (canPlace) {
        for (let i = 0; i < word.length; i++) {
          const rr = dir === 'V' ? r + i : r
          const cc = dir === 'H' ? c + i : c
          const row = grid[rr]
          if (row) row[cc] = word[i] || ''
        }
        placed = true
      }
    }
  })
  
  // Rellenar espacios con letras aleatorias
  const ALPHA = 'ABCDEFGHIJKLMNÑOPQRSTUVWXYZ'
  for (let r = 0; r < size; r++) {
    const row = grid[r]
    if (row) {
      for (let c = 0; c < size; c++) {
        if (!row[c]) row[c] = ALPHA[Math.floor(Math.random() * ALPHA.length)] || 'A'
      }
    }
  }
  wordGrid.value = grid
}

function toggleCell(r: number, c: number) {
  if (isCompleted.value) return
  const key = `${r},${c}`
  const idx = selectedCells.value.indexOf(key)
  if (idx >= 0) selectedCells.value.splice(idx, 1)
  else selectedCells.value.push(key)
  
  playBeep(480, 'sine', 0.05)
  checkWordMatch()
}

function checkWordMatch() {
  if (selectedCells.value.length < 2) return
  // Obtener letras seleccionadas en orden
  const chars = selectedCells.value.map(k => {
    const [r = 0, c = 0] = k.split(',').map(Number)
    const row = wordGrid.value[r]
    return row ? (row[c] || '') : ''
  }).join('')
  
  const rev = chars.split('').reverse().join('')
  wordsToFind.value.forEach(w => {
    if (!foundWords.value.has(w) && (chars === w || rev === w)) {
      foundWords.value.add(w)
      selectedCells.value = []
      soundSuccess()
      if (foundWords.value.size === wordsToFind.value.length) handleGameWin()
    }
  })
}

// ── 8: COMPLETAR ESPACIOS ────────────────────────────────────────────────────
const fbSentences = ref<any[]>([])
const fbAnswers = ref<Record<number, string>>({})

function initFillBlank() {
  fbAnswers.value = {}
  fbSentences.value = config.value?.sentences || [
    { text: 'Para estilizar una página web usamos ___ y para la estructura ___', answer: 'CSS', options: ['CSS', 'Python', 'SQL'] },
    { text: 'El framework de JavaScript progresivo que usamos es ___', answer: 'Vue 3', options: ['Vue 3', 'Django', 'Laravel'] },
  ]
}

function checkFillBlank() {
  let allCorrect = true
  fbSentences.value.forEach((s, idx) => {
    if ((fbAnswers.value[idx] || '').trim().toLowerCase() !== (s.answer || '').trim().toLowerCase()) {
      allCorrect = false
    }
  })
  if (allCorrect) {
    handleGameWin()
  } else {
    soundError()
    toast.error('Hay respuestas incorrectas. Revisa e inténtalo de nuevo.')
  }
}

// ── 9: ORDENAR SECUENCIA ─────────────────────────────────────────────────────
const orderItems = ref<any[]>([])

function initOrder() {
  const items = config.value?.items || [
    { text: 'Diseñar los wireframes / UI', correct_order: 1 },
    { text: 'Configurar base de datos y repositorios', correct_order: 2 },
    { text: 'Desarrollar lógica de negocio en backend', correct_order: 3 },
    { text: 'Integrar componentes en frontend y probar', correct_order: 4 },
  ]
  orderItems.value = items.map((it: any, idx: number) => ({ ...it, origId: idx })).sort(() => Math.random() - 0.5)
}

function moveOrderItem(idx: number, dir: -1 | 1) {
  const j = idx + dir
  if (j < 0 || j >= orderItems.value.length) return
  const tmp = orderItems.value[idx]
  const tgt = orderItems.value[j]
  if (tmp && tgt) {
    orderItems.value[idx] = tgt
    orderItems.value[j] = tmp
    playBeep(500, 'triangle', 0.08)
  }
}

function checkOrder() {
  let correct = true
  orderItems.value.forEach((it, idx) => {
    if (it.correct_order !== idx + 1) correct = false
  })
  if (correct) {
    handleGameWin()
  } else {
    soundError()
    toast.error('El orden no es correcto todavía. ¡Sigue intentando!')
  }
}

function initSpecificGame() {
  const t = gameType.value
  if (t === '5') initMemo()
  else if (t === '6') initDrag()
  else if (t === '7') initWordSearch()
  else if (t === '8') initFillBlank()
  else if (t === '9') initOrder()
}
</script>

<template>
  <div class="game-activity">
    <!-- Header del juego -->
    <div class="game-header">
      <div class="game-title-wrap">
        <span class="game-badge">🎮 Actividad Interactiva</span>
        <h2 class="game-title">{{ lesson?.title }}</h2>
      </div>
      <div class="game-stats">
        <div class="stat-chip">
          <span class="stat-icon">⏱️</span>
          <span class="stat-val">{{ elapsedSecs }}s</span>
        </div>
        <div class="stat-chip stat-pts">
          <span class="stat-icon">⭐</span>
          <span class="stat-val">+{{ lesson?.points_reward || 100 }} pts</span>
        </div>
      </div>
    </div>

    <!-- Descripción si la hay -->
    <p v-if="lesson?.description" class="game-desc">{{ lesson.description }}</p>

    <!-- Estado Completado (Victoria) -->
    <div v-if="isCompleted" class="game-win-banner slide-down">
      <div class="win-icon">🏆</div>
      <div class="win-info">
        <h3>¡Reto Superado con Éxito!</h3>
        <p>Has completado esta actividad en <strong>{{ elapsedSecs }} segundos</strong> y obtuviste <strong>+{{ pointsEarned }} puntos</strong> de experiencia.</p>
      </div>
      <button class="btn btn-primary" @click="initSpecificGame(); isCompleted = false">
        🔄 Jugar de nuevo
      </button>
    </div>

    <!-- ── 5: MEMORAMA ────────────────────────────────────────────── -->
    <div v-if="gameType === '5'" class="game-area memo-area">
      <p class="game-instruct">Encuentra todos los pares haciendo clic en las tarjetas para voltearlas:</p>
      <div class="memo-grid">
        <button
          v-for="card in memoCards"
          :key="card.id"
          :class="['memo-card', memoFlipped.includes(card.id) ? 'flipped' : '', memoMatched.has(card.pairId) ? 'matched' : '']"
          @click="flipCard(card)"
        >
          <div class="card-inner">
            <div class="card-front">❓</div>
            <div class="card-back">
              <img v-if="card.img" :src="card.img" class="card-img-content" :alt="card.text || 'Imagen'" />
              <span v-if="card.text" class="card-text-content">{{ card.text }}</span>
            </div>
          </div>
        </button>
      </div>
    </div>

    <!-- ── 6: CLASIFICAR / ARRASTRAR ──────────────────────────────── -->
    <div v-else-if="gameType === '6'" class="game-area drag-area">
      <p class="game-instruct">Selecciona un elemento y luego haz clic en la categoría a la que pertenece:</p>
      
      <!-- Items para asignar -->
      <div class="drag-items-pool">
        <button
          v-for="it in dragItems"
          :key="it.id"
          :class="['drag-item-btn', selectedDragItem === it.id ? 'active' : '', dragAssignments[it.id] ? 'assigned' : '']"
          @click="selectItemForCategory(it.id)"
        >
          <span>{{ it.text }}</span>
          <span v-if="dragAssignments[it.id]" class="assigned-badge">{{ dragAssignments[it.id] }}</span>
        </button>
      </div>

      <!-- Categorías (destinos) -->
      <div class="drag-categories-grid">
        <div
          v-for="cat in dragCategories"
          :key="cat"
          class="category-box"
          @click="assignCategory(cat)"
        >
          <h4 class="cat-title">📁 {{ cat }}</h4>
          <div class="cat-items-list">
            <div
              v-for="it in dragItems.filter(i => dragAssignments[i.id] === cat)"
              :key="it.id"
              class="cat-assigned-chip"
            >
              {{ it.text }}
              <button class="remove-assign" @click.stop="delete dragAssignments[it.id]">✕</button>
            </div>
            <span v-if="!dragItems.some(i => dragAssignments[i.id] === cat)" class="cat-empty-hint">Haz clic aquí para asignar el elemento seleccionado</span>
          </div>
        </div>
      </div>
      
      <div class="game-actions">
        <button class="btn btn-primary btn-lg" @click="checkDragWin">✅ Verificar Clasificación</button>
      </div>
    </div>

    <!-- ── 7: SOPA DE LETRAS ──────────────────────────────────────── -->
    <div v-else-if="gameType === '7'" class="game-area ws-area">
      <div class="ws-layout">
        <div class="ws-grid-wrap">
          <p class="game-instruct">Haz clic en las letras de la cuadrícula para formar las palabras secretas:</p>
          <div class="ws-grid">
            <div v-for="(row, r) in wordGrid" :key="r" class="ws-row">
              <button
                v-for="(char, c) in row"
                :key="c"
                :class="['ws-cell', selectedCells.includes(`${r},${c}`) ? 'selected' : '']"
                @click="toggleCell(r, c)"
              >
                {{ char }}
              </button>
            </div>
          </div>
        </div>
        <div class="ws-sidebar">
          <h4>🔍 Palabras a encontrar:</h4>
          <ul class="ws-word-list">
            <li v-for="w in wordsToFind" :key="w" :class="{ found: foundWords.has(w) }">
              <span class="word-chk">{{ foundWords.has(w) ? '☑️' : '◻️' }}</span>
              <span class="word-txt">{{ w }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- ── 8: COMPLETAR ESPACIOS ──────────────────────────────────── -->
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
                <input type="radio" :name="`fb_${idx}`" :value="opt" v-model="fbAnswers[idx]" class="hidden-radio" />
                <span>{{ opt }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>
      <div class="game-actions">
        <button class="btn btn-primary btn-lg" @click="checkFillBlank">✅ Validar Respuestas</button>
      </div>
    </div>

    <!-- ── 9: ORDENAR SECUENCIA ───────────────────────────────────── -->
    <div v-else-if="gameType === '9'" class="game-area order-area">
      <p class="game-instruct">Ordena los siguientes pasos o eventos cronológicamente usando las flechas:</p>
      <div class="order-list">
        <div v-for="(it, idx) in orderItems" :key="it.origId" class="order-card">
          <span class="order-pos">{{ idx + 1 }}</span>
          <span class="order-text">{{ it.text }}</span>
          <div class="order-controls">
            <button class="btn-order-move" :disabled="idx === 0" @click="moveOrderItem(idx, -1)" title="Subir">▲</button>
            <button class="btn-order-move" :disabled="idx === orderItems.length - 1" @click="moveOrderItem(idx, 1)" title="Bajar">▼</button>
          </div>
        </div>
      </div>
      <div class="game-actions">
        <button class="btn btn-primary btn-lg" @click="checkOrder">✅ Verificar Orden</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.game-activity {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-xl);
  padding: 28px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.04);
  display: flex; flex-direction: column; gap: 24px;
}

.game-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px; border-bottom: 1px solid var(--border-light); padding-bottom: 20px; }
.game-title-wrap { display: flex; flex-direction: column; gap: 6px; }
.game-badge { font-size: 0.78rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.08em; color: var(--brand); background: var(--brand-light); padding: 4px 10px; border-radius: 20px; width: fit-content; }
.game-title { font-size: 1.5rem; font-weight: 800; color: var(--dark); margin: 0; }
.game-desc { font-size: 0.95rem; color: var(--muted); margin: -8px 0 0 0; }

.game-stats { display: flex; gap: 10px; }
.stat-chip { display: flex; align-items: center; gap: 6px; background: var(--surface-soft); border: 1px solid var(--border); padding: 8px 14px; border-radius: 20px; font-weight: 700; color: var(--dark); font-size: 0.9rem; }
.stat-pts { background: #fffbeb; border-color: #fef3c7; color: #d97706; }

.game-instruct { font-size: 0.95rem; font-weight: 600; color: var(--dark); margin-bottom: 16px; }

/* Win banner */
.game-win-banner { display: flex; align-items: center; gap: 18px; background: linear-gradient(135deg, #ecfdf5, #d1fae5); border: 2px solid #10b981; border-radius: var(--r-lg); padding: 20px 24px; color: #065f46; flex-wrap: wrap; }
.win-icon { font-size: 2.8rem; }
.win-info { flex: 1; min-width: 240px; }
.win-info h3 { font-size: 1.25rem; font-weight: 800; margin: 0 0 4px 0; }
.win-info p { margin: 0; font-size: 0.92rem; opacity: 0.9; }

/* Memorama */
.memo-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(135px, 1fr)); gap: 14px; }
.memo-card { background: none; border: none; padding: 0; height: 130px; cursor: pointer; perspective: 600px; }
.card-inner { width: 100%; height: 100%; position: relative; transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1); transform-style: preserve-3d; }
.memo-card.flipped .card-inner, .memo-card.matched .card-inner { transform: rotateY(180deg); }
.card-front, .card-back { position: absolute; inset: 0; backface-visibility: hidden; border-radius: var(--r-md); display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 8px; font-weight: 700; text-align: center; box-shadow: 0 2px 8px rgba(0,0,0,0.08); gap: 4px; overflow: hidden; }
.card-front { background: linear-gradient(135deg, var(--brand), #ea580c); color: white; font-size: 2.2rem; border: 2px solid rgba(255,255,255,0.2); }
.card-back { background: var(--surface); border: 2px solid var(--brand); color: var(--dark); font-size: 0.88rem; transform: rotateY(180deg); }
.memo-card.matched .card-back { background: #ecfdf5; border-color: #10b981; color: #065f46; }
.card-img-content { max-width: 95%; max-height: 75px; object-fit: contain; border-radius: 4px; flex-shrink: 0; }
.card-text-content { font-size: 0.82rem; line-height: 1.15; word-break: break-word; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

/* Drag / Clasificar */
.drag-items-pool { display: flex; flex-wrap: wrap; gap: 10px; padding: 16px; background: var(--surface-soft); border-radius: var(--r-lg); border: 1.5px dashed var(--border); }
.drag-item-btn { display: flex; align-items: center; gap: 8px; padding: 10px 18px; background: var(--surface); border: 1.5px solid var(--border); border-radius: 24px; font-weight: 600; font-size: 0.9rem; color: var(--dark); cursor: pointer; transition: all 0.2s; box-shadow: 0 2px 4px rgba(0,0,0,0.03); }
.drag-item-btn:hover { border-color: var(--brand); transform: translateY(-2px); }
.drag-item-btn.active { background: var(--brand); border-color: var(--brand); color: white; box-shadow: 0 4px 12px rgba(249,115,22,0.3); }
.drag-item-btn.assigned { opacity: 0.5; }
.assigned-badge { background: rgba(0,0,0,0.1); padding: 2px 8px; border-radius: 10px; font-size: 0.72rem; }

.drag-categories-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin-top: 16px; }
.category-box { background: var(--surface); border: 2px solid var(--border); border-radius: var(--r-lg); padding: 16px; cursor: pointer; transition: all 0.2s; min-height: 160px; display: flex; flex-direction: column; }
.category-box:hover { border-color: var(--brand); background: var(--surface-soft); }
.cat-title { font-size: 1rem; font-weight: 800; color: var(--dark); margin: 0 0 12px 0; border-bottom: 1px solid var(--border-light); padding-bottom: 8px; }
.cat-items-list { display: flex; flex-direction: column; gap: 8px; flex: 1; }
.cat-assigned-chip { display: flex; align-items: center; justify-content: space-between; background: #eff6ff; border: 1px solid #bfdbfe; color: #1e3a8a; padding: 8px 12px; border-radius: var(--r-sm); font-size: 0.88rem; font-weight: 600; }
.remove-assign { background: none; border: none; color: #60a5fa; cursor: pointer; font-weight: 800; font-size: 0.9rem; }
.remove-assign:hover { color: #dc2626; }
.cat-empty-hint { font-size: 0.8rem; color: var(--muted); font-style: italic; margin: auto 0; text-align: center; }

/* Sopa letras */
.ws-layout { display: flex; gap: 24px; flex-wrap: wrap; }
.ws-grid-wrap { flex: 1; min-width: 280px; }
.ws-grid { display: inline-flex; flex-direction: column; gap: 4px; background: var(--surface-soft); padding: 12px; border-radius: var(--r-lg); border: 1px solid var(--border); }
.ws-row { display: flex; gap: 4px; }
.ws-cell { width: 36px; height: 36px; background: var(--surface); border: 1px solid var(--border); border-radius: 6px; font-weight: 800; font-size: 1rem; color: var(--dark); cursor: pointer; transition: all 0.15s; display: flex; align-items: center; justify-content: center; }
.ws-cell:hover { border-color: var(--brand); background: var(--brand-light); }
.ws-cell.selected { background: var(--brand); border-color: var(--brand); color: white; transform: scale(1.08); }
.ws-sidebar { width: 200px; background: var(--surface-soft); border-radius: var(--r-lg); padding: 18px; border: 1px solid var(--border); }
.ws-sidebar h4 { font-size: 0.95rem; font-weight: 800; margin: 0 0 12px 0; color: var(--dark); }
.ws-word-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 8px; }
.ws-word-list li { display: flex; align-items: center; gap: 8px; font-size: 0.9rem; font-weight: 700; color: var(--dark); }
.ws-word-list li.found { text-decoration: line-through; color: #10b981; }

/* Fill Blank */
.fb-list { display: flex; flex-direction: column; gap: 16px; }
.fb-card { display: flex; gap: 14px; background: var(--surface-soft); border: 1px solid var(--border); border-radius: var(--r-lg); padding: 18px; }
.fb-sent-num { width: 32px; height: 32px; background: var(--brand); color: white; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 800; flex-shrink: 0; }
.fb-sent-body { flex: 1; }
.fb-sent-text { font-size: 1.05rem; font-weight: 600; color: var(--dark); margin: 0 0 14px 0; }
.fb-options-grid { display: flex; flex-wrap: wrap; gap: 10px; }
.fb-option-chip { display: flex; align-items: center; padding: 8px 16px; background: var(--surface); border: 1.5px solid var(--border); border-radius: 20px; font-weight: 600; font-size: 0.9rem; cursor: pointer; transition: all 0.2s; }
.fb-option-chip:hover { border-color: var(--brand); }
.fb-option-chip.selected { background: var(--brand); border-color: var(--brand); color: white; }
.hidden-radio { display: none; }

/* Order sequence */
.order-list { display: flex; flex-direction: column; gap: 12px; }
.order-card { display: flex; align-items: center; gap: 14px; background: var(--surface-soft); border: 1px solid var(--border); border-radius: var(--r-lg); padding: 14px 18px; transition: transform 0.2s; }
.order-card:hover { border-color: var(--border-dark); }
.order-pos { width: 32px; height: 32px; background: var(--dark); color: white; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 1rem; flex-shrink: 0; }
.order-text { flex: 1; font-size: 0.98rem; font-weight: 600; color: var(--dark); }
.order-controls { display: flex; flex-direction: column; gap: 4px; }
.btn-order-move { width: 28px; height: 28px; background: var(--surface); border: 1px solid var(--border); border-radius: 6px; cursor: pointer; font-size: 0.7rem; display: flex; align-items: center; justify-content: center; color: var(--dark); transition: all 0.15s; }
.btn-order-move:hover:not(:disabled) { background: var(--brand); color: white; border-color: var(--brand); }
.btn-order-move:disabled { opacity: 0.3; cursor: not-allowed; }

.game-actions { display: flex; justify-content: flex-end; margin-top: 10px; }
.btn-lg { padding: 12px 28px; font-size: 1rem; font-weight: 700; border-radius: var(--r-lg); }

@keyframes slideDown { from { opacity: 0; transform: translateY(-10px); } to { opacity: 1; transform: translateY(0); } }
.slide-down { animation: slideDown 0.3s cubic-bezier(0.16, 1, 0.3, 1); }
</style>
