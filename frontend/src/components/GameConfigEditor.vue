<script setup lang="ts">
/**
 * GameConfigEditor.vue
 * Editor visual para la configuración de los 5 tipos de minijuego.
 * El instructor llena formularios estructurados; el componente emite el JSON final.
 */
import { ref, watch, computed } from 'vue'

const props = defineProps<{
  modelValue: string   // JSON string actual del game_config_json
  lessonType: string   // "5" | "6" | "7" | "8" | "9"
  pointsReward?: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: string): void
  (e: 'update:pointsReward', v: number): void
}>()

// ── Tipado de schemas ──────────────────────────────────────────────────────────
interface MemoryPair { front: string; back: string; front_img?: string; back_img?: string; image_url?: string }
interface MemoryConfig { instruction: string; pairs: MemoryPair[]; max_time_secs: number; show_labels: boolean }

interface DragItem { text: string; correct_category: string; image_url?: string }
interface DragConfig { instruction: string; categories: string[]; items: DragItem[] }

interface WSConfig { instruction: string; words: string[]; grid_size: number; difficulty: string; show_word_list: boolean }

interface FBItem { text: string; answer: string; options: string[] }
interface FBConfig { instruction: string; mode: string; sentences: FBItem[] }

interface OrderItem { text: string; correct_order: number; image_url?: string }
interface OrderConfig { instruction: string; items: OrderItem[]; show_numbers: boolean }

interface HangmanItem { word: string; hint: string }
interface HangmanConfig { instruction: string; max_errors: number; items: HangmanItem[] }

// ── Estado local ───────────────────────────────────────────────────────────────
const memory  = ref<MemoryConfig>({ instruction: '', pairs: [{ front: '', back: '' }], max_time_secs: 120, show_labels: true })
const drag    = ref<DragConfig>({ instruction: '', categories: ['Categoría A', 'Categoría B'], items: [{ text: '', correct_category: 'Categoría A' }] })
const ws      = ref<WSConfig>({ instruction: '', words: [''], grid_size: 12, difficulty: 'medium', show_word_list: true })
const fb      = ref<FBConfig>({ instruction: '', mode: 'select', sentences: [{ text: '', answer: '', options: ['', ''] }] })
const order   = ref<OrderConfig>({ instruction: '', items: [{ text: '', correct_order: 1 }, { text: '', correct_order: 2 }], show_numbers: false })
const hangman = ref<HangmanConfig>({ instruction: '', max_errors: 6, items: [{ word: 'JAVASCRIPT', hint: 'Lenguaje web dinámico' }, { word: 'SEGURIDAD', hint: 'Protección de sistemas' }] })

function defaultsFor(type: string) {
  if (type === '5') return { instruction: '', pairs: [{ front: '', back: '' }], max_time_secs: 120, show_labels: true }
  if (type === '6') return { instruction: '', categories: ['Categoría A', 'Categoría B'], items: [{ text: '', correct_category: 'Categoría A' }] }
  if (type === '7') return { instruction: '', words: [''], grid_size: 12, difficulty: 'medium', show_word_list: true }
  if (type === '8') return { instruction: '', mode: 'select', sentences: [{ text: '', answer: '', options: ['', ''] }] }
  if (type === '9') return { instruction: '', items: [{ text: '', correct_order: 1 }, { text: '', correct_order: 2 }], show_numbers: false }
  if (type === '10') return { instruction: '', max_errors: 6, items: [{ word: 'JAVASCRIPT', hint: 'Lenguaje web dinámico' }, { word: 'SEGURIDAD', hint: 'Protección de sistemas' }] }
  return {}
}

function parseOrDefault(json: string, type: string) {
  try { return json ? JSON.parse(json) : defaultsFor(type) }
  catch { return defaultsFor(type) }
}

// Cargar estado desde el JSON recibido
watch([() => props.modelValue, () => props.lessonType], ([json, type]) => {
  const parsed = parseOrDefault(json, type)
  if (type === '5') memory.value = { ...defaultsFor('5') as MemoryConfig, ...parsed }
  if (type === '6') drag.value   = { ...defaultsFor('6') as DragConfig,  ...parsed }
  if (type === '7') ws.value     = { ...defaultsFor('7') as WSConfig,    ...parsed }
  if (type === '8') fb.value     = { ...defaultsFor('8') as FBConfig,    ...parsed }
  if (type === '9') order.value  = { ...defaultsFor('9') as OrderConfig, ...parsed }
  if (type === '10') hangman.value = { ...defaultsFor('10') as HangmanConfig, ...parsed }
  if (!json) emitJson()
}, { immediate: true })

// Emitir JSON al padre cada vez que cambie el estado
function emitJson() {
  const t = props.lessonType
  let obj: any
  if (t === '5') obj = memory.value
  else if (t === '6') obj = drag.value
  else if (t === '7') {
    const validWords = ws.value.words.map(w => w.trim().toUpperCase()).filter(w => w.length >= 2)
    obj = {
      ...ws.value,
      words: validWords.length > 0 ? validWords : ['CAPACITACION', 'SOPA', 'LETRA']
    }
  }
  else if (t === '8') obj = fb.value
  else if (t === '9') obj = order.value
  else if (t === '10') obj = hangman.value
  else return
  emit('update:modelValue', JSON.stringify(obj, null, 2))
}

function addHangmanItem() { hangman.value.items.push({ word: '', hint: '' }); emitJson() }
function removeHangmanItem(i: number) { hangman.value.items.splice(i, 1); emitJson() }

// ── Helpers de edición ─────────────────────────────────────────────────────────

// MEMORY
function addPair() { memory.value.pairs.push({ front: '', back: '', front_img: '', back_img: '' }); emitJson() }
function removePair(i: number) { memory.value.pairs.splice(i, 1); emitJson() }
function onUploadPairImg(e: Event, index: number, field: 'front_img' | 'back_img') {
  const input = e.target as HTMLInputElement
  if (!input.files || !input.files[0]) return
  const file = input.files[0]
  if (file.size > 2 * 1024 * 1024) {
    alert('La imagen no debe superar los 2 MB para las tarjetas de memorama.')
    return
  }
  const reader = new FileReader()
  reader.onload = () => {
    if (reader.result && memory.value.pairs[index]) {
      memory.value.pairs[index][field] = String(reader.result)
      emitJson()
    }
  }
  reader.readAsDataURL(file)
}

// DRAG & DROP
function addCategory() { drag.value.categories.push(`Categoría ${drag.value.categories.length + 1}`); emitJson() }
function removeCategory(i: number) {
  const cat = drag.value.categories[i]
  drag.value.items = drag.value.items.filter(it => it.correct_category !== cat)
  drag.value.categories.splice(i, 1)
  emitJson()
}
function addDragItem() { drag.value.items.push({ text: '', correct_category: drag.value.categories[0] || '' }); emitJson() }
function removeDragItem(i: number) { drag.value.items.splice(i, 1); emitJson() }

// WORD SEARCH
function addWord() { ws.value.words.push(''); emitJson() }
function removeWord(i: number) { ws.value.words.splice(i, 1); emitJson() }

// FILL BLANK
function addSentence() { fb.value.sentences.push({ text: '', answer: '', options: ['', ''] }); emitJson() }
function removeSentence(i: number) { fb.value.sentences.splice(i, 1); emitJson() }
function addOption(si: number) { const s = fb.value.sentences[si]; if (s) { s.options.push(''); emitJson() } }
function removeOption(si: number, oi: number) { const s = fb.value.sentences[si]; if (s) { s.options.splice(oi, 1); emitJson() } }

// ORDER
function addOrderItem() { order.value.items.push({ text: '', correct_order: order.value.items.length + 1 }); emitJson() }
function removeOrderItem(i: number) {
  order.value.items.splice(i, 1)
  order.value.items.forEach((it, idx) => (it.correct_order = idx + 1))
  emitJson()
}
function moveOrderItem(i: number, dir: -1 | 1) {
  const j = i + dir
  if (j < 0 || j >= order.value.items.length) return
  const temp = order.value.items[i];
  const target = order.value.items[j];
  if (temp && target) {
    order.value.items[i] = target;
    order.value.items[j] = temp;
  }
  order.value.items.forEach((it, idx) => (it.correct_order = idx + 1))
  emitJson()
}

const localPoints = ref(props.pointsReward ?? 100)
watch(() => props.pointsReward, v => { if (v !== undefined) localPoints.value = v })
function onPointsChange() { emit('update:pointsReward', localPoints.value); emitJson() }

const gameNames: Record<string, string> = {
  '5': 'Memorama', '6': 'Arrastrar y Soltar', '7': 'Sopa de Letras',
  '8': 'Completar Espacios', '9': 'Ordenar Secuencia', '10': 'Ahorcado',
}
const gameName = computed(() => gameNames[props.lessonType] ?? '')
</script>

<template>
  <div class="gce-root" v-if="['5','6','7','8','9','10'].includes(lessonType)">
    <!-- Header del editor -->
    <div class="gce-header">
      <span class="gce-badge glass-badge-editor">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="6" width="20" height="12" rx="6"/><path d="M6 12h4m-2-2v4"/><circle cx="17" cy="11" r="1" fill="currentColor"/><circle cx="15" cy="13" r="1" fill="currentColor"/></svg>
        {{ gameName }}
      </span>
      <span class="gce-hint">Configura el minijuego que verán los alumnos</span>
    </div>

    <!-- Puntos de recompensa (común a todos) -->
    <div class="gce-field">
      <label class="gce-label gce-label-icon">
        <span class="glass-icon-xs star-glow">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
        </span>
        Puntos de recompensa
        <span class="gce-sublabel">Los alumnos ganan estos puntos al completar el juego</span>
      </label>
      <div class="gce-points-row">
        <input
          type="number"
          v-model.number="localPoints"
          @change="onPointsChange"
          min="0"
          max="1000"
          class="gce-input-sm"
        />
        <span class="gce-unit">pts</span>
      </div>
    </div>

    <!-- ─────────────── MEMORAMA ─────────────── -->
    <template v-if="lessonType === '5'">
      <div class="gce-field">
        <label class="gce-label">Instrucción para el alumno</label>
        <input v-model="memory.instruction" @input="emitJson" class="gce-input" placeholder="Ej: Encuentra los pares de conceptos relacionados" />
      </div>

      <div class="gce-field">
        <label class="gce-label">Tiempo límite (0 = sin límite)</label>
        <div class="gce-row">
          <input type="number" v-model.number="memory.max_time_secs" @change="emitJson" min="0" class="gce-input-sm" />
          <span class="gce-unit">segundos</span>
        </div>
      </div>

      <div class="gce-field">
        <label class="gce-label">Mostrar texto en tarjetas</label>
        <label class="gce-toggle">
          <input type="checkbox" v-model="memory.show_labels" @change="emitJson" />
          <span class="gce-toggle-track"><span class="gce-toggle-thumb" /></span>
          <span>{{ memory.show_labels ? 'Sí' : 'No' }}</span>
        </label>
      </div>

      <div class="gce-section-title">Pares de tarjetas <span class="gce-count">{{ memory.pairs.length }}/24</span></div>
      <p class="gce-hint-sm">Puedes escribir texto, añadir imagen (por URL o subiendo archivo local), o ambos en cada tarjeta.</p>
      <div class="gce-pairs-list">
        <div v-for="(pair, i) in memory.pairs" :key="i" class="gce-pair-block">
          <div class="gce-pair-header">
            <span class="gce-pair-title">Par #{{ i + 1 }}</span>
            <button class="gce-remove-sm" @click="removePair(i)" :disabled="memory.pairs.length <= 2" title="Eliminar par">✕ Eliminar par</button>
          </div>
          <div class="gce-pair-cols">
            <!-- Cara A -->
            <div class="gce-pair-col">
              <label class="gce-sublabel gce-icon-sublabel">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
                Cara A (Concepto / Imagen)
              </label>
              <input v-model="pair.front" @input="emitJson" class="gce-input" placeholder="Texto Cara A (opcional si hay imagen)" />
              <div class="gce-img-picker">
                <input v-model="pair.front_img" @input="emitJson" class="gce-input-xs" placeholder="URL de imagen o DataURL..." />
                <label class="gce-btn-upload" title="Subir imagen desde tu equipo">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                  <span>Subir</span>
                  <input type="file" accept="image/*" class="hidden-file" @change="e => onUploadPairImg(e, i, 'front_img')" />
                </label>
              </div>
              <div v-if="pair.front_img" class="gce-img-preview">
                <img :src="pair.front_img" alt="Preview Cara A" />
                <button class="gce-remove-preview" @click="pair.front_img = ''; emitJson()" title="Quitar imagen">✕</button>
              </div>
            </div>

            <div class="gce-pair-sep">↔</div>

            <!-- Cara B -->
            <div class="gce-pair-col">
              <label class="gce-sublabel gce-icon-sublabel">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
                Cara B (Pareja / Definición)
              </label>
              <input v-model="pair.back" @input="emitJson" class="gce-input" placeholder="Texto Cara B (opcional si hay imagen)" />
              <div class="gce-img-picker">
                <input v-model="pair.back_img" @input="emitJson" class="gce-input-xs" placeholder="URL de imagen o DataURL..." />
                <label class="gce-btn-upload" title="Subir imagen desde tu equipo">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                  <span>Subir</span>
                  <input type="file" accept="image/*" class="hidden-file" @change="e => onUploadPairImg(e, i, 'back_img')" />
                </label>
              </div>
              <div v-if="pair.back_img" class="gce-img-preview">
                <img :src="pair.back_img" alt="Preview Cara B" />
                <button class="gce-remove-preview" @click="pair.back_img = ''; emitJson()" title="Quitar imagen">✕</button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <button class="gce-add-btn" @click="addPair" :disabled="memory.pairs.length >= 24">
        <span>+</span> Agregar par de tarjetas
      </button>
    </template>

    <!-- ─────────────── DRAG & DROP ─────────────── -->
    <template v-if="lessonType === '6'">
      <div class="gce-field">
        <label class="gce-label">Instrucción para el alumno</label>
        <input v-model="drag.instruction" @input="emitJson" class="gce-input" placeholder="Ej: Clasifica cada elemento en su categoría" />
      </div>

      <div class="gce-section-title">Categorías (zonas de destino)</div>
      <div class="gce-tags-list">
        <div v-for="(cat, i) in drag.categories" :key="i" class="gce-tag-row">
          <input v-model="drag.categories[i]" @input="emitJson" class="gce-input" :placeholder="`Categoría ${i + 1}`" />
          <button class="gce-remove" @click="removeCategory(i)" :disabled="drag.categories.length <= 2">✕</button>
        </div>
      </div>
      <button class="gce-add-btn" @click="addCategory">+ Agregar categoría</button>

      <div class="gce-section-title" style="margin-top:16px">Ítems a clasificar</div>
      <div class="gce-items-list">
        <div v-for="(item, i) in drag.items" :key="i" class="gce-item-row">
          <input v-model="item.text" @input="emitJson" class="gce-input" :placeholder="`Ítem ${i + 1}`" />
          <select v-model="item.correct_category" @change="emitJson" class="gce-select">
            <option v-for="cat in drag.categories" :key="cat" :value="cat">{{ cat }}</option>
          </select>
          <button class="gce-remove" @click="removeDragItem(i)" :disabled="drag.items.length <= 1">✕</button>
        </div>
      </div>
      <button class="gce-add-btn" @click="addDragItem">+ Agregar ítem</button>
    </template>

    <!-- ─────────────── SOPA DE LETRAS ─────────────── -->
    <template v-if="lessonType === '7'">
      <div class="gce-field">
        <label class="gce-label">Instrucción para el alumno</label>
        <input v-model="ws.instruction" @input="emitJson" class="gce-input" placeholder="Ej: Encuentra los conceptos de redes en la sopa de letras" />
      </div>
      <div class="gce-row-2">
        <div class="gce-field">
          <label class="gce-label">Tamaño de cuadrícula</label>
          <select v-model.number="ws.grid_size" @change="emitJson" class="gce-select">
            <option :value="8">8×8 (Fácil)</option>
            <option :value="10">10×10</option>
            <option :value="12">12×12 (Normal)</option>
            <option :value="15">15×15 (Difícil)</option>
            <option :value="20">20×20 (Experto)</option>
          </select>
        </div>
        <div class="gce-field">
          <label class="gce-label">Dificultad (direcciones)</label>
          <select v-model="ws.difficulty" @change="emitJson" class="gce-select">
            <option value="easy">Fácil (H + V)</option>
            <option value="medium">Media (H + V + Diag)</option>
            <option value="hard">Difícil (+ Invertidas)</option>
          </select>
        </div>
      </div>
      <div class="gce-field">
        <label class="gce-toggle">
          <input type="checkbox" v-model="ws.show_word_list" @change="emitJson" />
          <span class="gce-toggle-track"><span class="gce-toggle-thumb" /></span>
          <span>Mostrar lista de palabras a buscar</span>
        </label>
      </div>
      <div class="gce-section-title">Palabras a ocultar <span class="gce-count">{{ ws.words.filter(w=>w).length }}/20</span></div>
      <p class="gce-hint-sm">Usa solo letras sin acentos ni espacios. El sistema las insertará en la cuadrícula automáticamente.</p>
      <div class="gce-words-grid">
        <div v-for="(_, i) in ws.words" :key="i" class="gce-word-row">
          <input
            v-model="ws.words[i]"
            @input="emitJson"
            class="gce-input"
            :placeholder="`Palabra ${i + 1}`"
            style="text-transform:uppercase"
          />
          <button class="gce-remove" @click="removeWord(i)" :disabled="ws.words.length <= 1">✕</button>
        </div>
      </div>
      <button class="gce-add-btn" @click="addWord" :disabled="ws.words.length >= 20">+ Agregar palabra</button>
    </template>

    <!-- ─────────────── COMPLETAR ESPACIOS ─────────────── -->
    <template v-if="lessonType === '8'">
      <div class="gce-field">
        <label class="gce-label">Instrucción para el alumno</label>
        <input v-model="fb.instruction" @input="emitJson" class="gce-input" placeholder="Ej: Completa las definiciones" />
      </div>
      <div class="gce-field">
        <label class="gce-label">Modo de respuesta</label>
        <div class="gce-radio-group">
          <label class="gce-radio-opt" :class="{ active: fb.mode === 'select' }">
            <input type="radio" value="select" v-model="fb.mode" @change="emitJson" />
            <span class="gce-opt-span">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="4" fill="currentColor"/></svg>
              Selección múltiple
            </span>
          </label>
          <label class="gce-radio-opt" :class="{ active: fb.mode === 'type' }">
            <input type="radio" value="type" v-model="fb.mode" @change="emitJson" />
            <span class="gce-opt-span">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="4" width="20" height="16" rx="2"/><line x1="6" y1="8" x2="6.01" y2="8"/><line x1="10" y1="8" x2="10.01" y2="8"/><line x1="14" y1="8" x2="14.01" y2="8"/><line x1="18" y1="8" x2="18.01" y2="8"/><line x1="6" y1="12" x2="6.01" y2="12"/><line x1="10" y1="12" x2="10.01" y2="12"/><line x1="14" y1="12" x2="14.01" y2="12"/><line x1="18" y1="12" x2="18.01" y2="12"/><line x1="7" y1="16" x2="17" y2="16"/></svg>
              Escribir respuesta
            </span>
          </label>
        </div>
      </div>

      <div class="gce-hint-sm">Usa <code class="gce-code">___</code> (tres guiones bajos) para marcar el espacio en blanco.</div>

      <div v-for="(sentence, si) in fb.sentences" :key="si" class="gce-sentence-block">
        <div class="gce-sentence-header">
          <span class="gce-sentence-num">Oración {{ si + 1 }}</span>
          <button class="gce-remove-sm" @click="removeSentence(si)" :disabled="fb.sentences.length <= 1">Eliminar</button>
        </div>
        <div class="gce-field">
          <label class="gce-sublabel">Texto con ___ como espacio en blanco</label>
          <textarea v-model="sentence.text" @input="emitJson" class="gce-input" rows="2"
            placeholder="El protocolo ___ se usa para transferir páginas web." />
        </div>
        <div class="gce-field">
          <label class="gce-sublabel">Respuesta correcta</label>
          <input v-model="sentence.answer" @input="emitJson" class="gce-input" placeholder="HTTP" />
        </div>
        <div v-if="fb.mode === 'select'" class="gce-field">
          <label class="gce-sublabel">Opciones de respuesta (incluye la correcta)</label>
          <div v-for="(_, oi) in sentence.options" :key="oi" class="gce-option-row">
            <input
              v-model="sentence.options[oi]"
              @input="emitJson"
              class="gce-input"
              :placeholder="`Opción ${oi + 1}`"
            />
            <button class="gce-remove" @click="removeOption(si, oi)" :disabled="sentence.options.length <= 2">✕</button>
          </div>
          <button class="gce-add-btn-sm" @click="addOption(si)">+ Opción</button>
        </div>
      </div>
      <button class="gce-add-btn" @click="addSentence">+ Agregar oración</button>
    </template>

    <!-- ─────────────── ORDENAR SECUENCIA ─────────────── -->
    <template v-if="lessonType === '9'">
      <div class="gce-field">
        <label class="gce-label">Instrucción para el alumno</label>
        <input v-model="order.instruction" @input="emitJson" class="gce-input" placeholder="Ej: Ordena los pasos del proceso de desarrollo" />
      </div>
      <div class="gce-field">
        <label class="gce-toggle">
          <input type="checkbox" v-model="order.show_numbers" @change="emitJson" />
          <span class="gce-toggle-track"><span class="gce-toggle-thumb" /></span>
          <span>Mostrar números de posición como pistas</span>
        </label>
      </div>

      <div class="gce-section-title">Ítems en orden correcto (este es el orden que debe lograr el alumno)</div>
      <div class="gce-order-list">
        <div v-for="(item, i) in order.items" :key="i" class="gce-order-row">
          <div class="gce-order-controls">
            <button class="gce-arrow" @click="moveOrderItem(i, -1)" :disabled="i === 0">
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M18 15l-6-6-6 6"/></svg>
            </button>
            <span class="gce-order-pos">{{ i + 1 }}</span>
            <button class="gce-arrow" @click="moveOrderItem(i, 1)" :disabled="i === order.items.length - 1">
              <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9l6 6 6-6"/></svg>
            </button>
          </div>
          <input v-model="item.text" @input="emitJson" class="gce-input" :placeholder="`Paso ${i + 1}`" />
          <button class="gce-remove" @click="removeOrderItem(i)" :disabled="order.items.length <= 2">✕</button>
        </div>
      </div>
      <button class="gce-add-btn" @click="addOrderItem">+ Agregar paso</button>
    </template>

    <!-- ─────────────── AHORCADO ─────────────── -->
    <template v-if="lessonType === '10'">
      <div class="gce-field">
        <label class="gce-label">Instrucción del Ahorcado</label>
        <input v-model="hangman.instruction" @input="emitJson" class="gce-input" placeholder="Ej: Adivina la palabra oculta antes de perder tus vidas" />
      </div>
      <div class="gce-field">
        <label class="gce-label">Límite de errores permitidos</label>
        <input type="number" v-model.number="hangman.max_errors" @input="emitJson" class="gce-input-sm" min="3" max="10" />
      </div>
      <div class="gce-field">
        <label class="gce-label">Palabras y pistas</label>
        <div class="gce-list">
          <div v-for="(item, i) in hangman.items" :key="i" class="gce-item-row">
            <input v-model="item.word" @input="emitJson" class="gce-input" placeholder="Palabra (Ej: KUBERNETES)" />
            <input v-model="item.hint" @input="emitJson" class="gce-input" placeholder="Pista (Ej: Orquestador de contenedores)" />
            <button class="gce-remove" @click="removeHangmanItem(i)" :disabled="hangman.items.length <= 1">✕</button>
          </div>
        </div>
      </div>
      <button class="gce-add-btn" @click="addHangmanItem">+ Agregar palabra</button>
    </template>

  </div>
</template>

<style scoped>
* { box-sizing: border-box; }
.gce-root {
  display: flex; flex-direction: column; gap: 16px;
  padding: 20px; background: var(--surface-soft);
  border-radius: var(--r-lg); border: 1px solid var(--border);
  margin-top: 8px; width: 100%; max-width: 100%; box-sizing: border-box;
}
.gce-header {
  display: flex; align-items: center; gap: 12px; flex-wrap: wrap;
  padding-bottom: 14px; border-bottom: 1px solid var(--border);
}
.gce-badge {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white; padding: 4px 12px; border-radius: 20px;
  font-size: 0.8rem; font-weight: 700;
}
.gce-hint { font-size: 0.82rem; color: var(--muted); }
.gce-field { display: flex; flex-direction: column; gap: 6px; width: 100%; max-width: 100%; box-sizing: border-box; }
.gce-label { font-size: 0.85rem; font-weight: 700; color: var(--dark); }
.gce-sublabel { font-size: 0.78rem; color: var(--muted); }
.gce-input {
  width: 100%; max-width: 100%; padding: 9px 12px;
  background: var(--surface); border: 1.5px solid var(--border);
  border-radius: var(--r-md); font-size: 0.9rem; color: var(--text);
  transition: border-color 0.2s; box-sizing: border-box;
}
.gce-input:focus { outline: none; border-color: var(--brand); }
.gce-input-sm { width: 100px; max-width: 100%; padding: 8px 10px; background: var(--surface); border: 1.5px solid var(--border); border-radius: var(--r-md); font-size: 0.9rem; color: var(--text); box-sizing: border-box; }
.gce-input-sm:focus { outline: none; border-color: var(--brand); }
.gce-select { padding: 8px 10px; background: var(--surface); border: 1.5px solid var(--border); border-radius: var(--r-md); font-size: 0.9rem; color: var(--text); cursor: pointer; max-width: 100%; }
.gce-unit { font-size: 0.85rem; color: var(--muted); align-self: center; }
.gce-points-row, .gce-row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; max-width: 100%; }
.gce-row-2 { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 16px; width: 100%; }
.gce-section-title {
  font-size: 0.82rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.06em; color: var(--muted);
  border-bottom: 1px solid var(--border); padding-bottom: 6px;
  margin-top: 4px;
}
.gce-count { font-weight: 400; margin-left: 6px; }
.gce-hint-sm { font-size: 0.78rem; color: var(--muted); margin: -8px 0 4px; }
.gce-code { background: var(--surface); border: 1px solid var(--border); border-radius: 4px; padding: 1px 5px; font-family: monospace; font-size: 0.82rem; }

/* Pairs */
.gce-pairs-list, .gce-items-list, .gce-tags-list, .gce-words-grid { display: flex; flex-direction: column; gap: 8px; width: 100%; max-width: 100%; box-sizing: border-box; }
.gce-pair-row, .gce-item-row, .gce-tag-row, .gce-word-row, .gce-option-row {
  display: flex; align-items: center; gap: 8px; flex-wrap: wrap; width: 100%; max-width: 100%; box-sizing: border-box;
}
.gce-pair-num { font-size: 0.78rem; color: var(--muted); min-width: 20px; text-align: right; }
.gce-pair-sep { display: none; }
.gce-words-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 8px; }
.gce-word-row { flex-direction: row; }

/* Advanced Memorama Pair Blocks */
.gce-pair-block {
  padding: 14px; background: var(--surface); border: 1.5px solid var(--border);
  border-radius: var(--r-md); display: flex; flex-direction: column; gap: 10px;
  width: 100%; max-width: 100%; box-sizing: border-box; overflow: hidden;
}
.gce-pair-header { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px dashed var(--border); padding-bottom: 6px; }
.gce-pair-title { font-size: 0.85rem; font-weight: 800; color: var(--brand); }
.gce-pair-cols {
  display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  align-items: start; gap: 14px; width: 100%; max-width: 100%; box-sizing: border-box;
}
@media (max-width: 640px) {
  .gce-pair-cols { grid-template-columns: minmax(0, 1fr); gap: 16px; }
}
.gce-pair-col { display: flex; flex-direction: column; gap: 6px; min-width: 0; width: 100%; max-width: 100%; box-sizing: border-box; }
.gce-img-picker { display: flex; gap: 6px; align-items: center; width: 100%; max-width: 100%; box-sizing: border-box; }
.gce-input-xs { flex: 1 1 0%; min-width: 0; width: 100%; max-width: 100%; box-sizing: border-box; padding: 6px 10px; background: var(--surface-soft); border: 1px solid var(--border); border-radius: 6px; font-size: 0.8rem; color: var(--text); }
.gce-btn-upload { display: inline-flex; align-items: center; justify-content: center; gap: 4px; padding: 6px 12px; background: var(--brand-light); color: var(--brand); font-size: 0.8rem; font-weight: 700; border-radius: 6px; cursor: pointer; border: 1px solid rgba(99,102,241,0.2); transition: all 0.15s; white-space: nowrap; flex-shrink: 0; }
.gce-btn-upload:hover { background: var(--brand); color: white; }
.hidden-file { display: none; }
.gce-img-preview { position: relative; width: fit-content; max-width: 100%; border: 1px solid var(--border); border-radius: 6px; padding: 4px; background: var(--surface-soft); margin-top: 2px; }
.gce-img-preview img { max-height: 80px; max-width: 160px; object-fit: contain; border-radius: 4px; display: block; }
.gce-remove-preview { position: absolute; top: -6px; right: -6px; width: 20px; height: 20px; background: #ef4444; color: white; border: none; border-radius: 50%; font-size: 0.7rem; font-weight: 800; cursor: pointer; display: flex; align-items: center; justify-content: center; box-shadow: 0 2px 4px rgba(0,0,0,0.2); }

/* Sentences */
.gce-sentence-block {
  padding: 16px; background: var(--surface); border-radius: var(--r-md);
  border: 1px solid var(--border); display: flex; flex-direction: column; gap: 10px;
}
.gce-sentence-header { display: flex; justify-content: space-between; align-items: center; }
.gce-sentence-num { font-weight: 700; font-size: 0.85rem; color: var(--dark); }
.gce-remove-sm { font-size: 0.78rem; color: var(--danger); background: none; border: none; cursor: pointer; }

/* Order */
.gce-order-list { display: flex; flex-direction: column; gap: 8px; }
.gce-order-row { display: flex; align-items: center; gap: 8px; }
.gce-order-controls { display: flex; flex-direction: column; align-items: center; gap: 2px; }
.gce-arrow { background: none; border: 1px solid var(--border); border-radius: 4px; width: 22px; height: 22px; cursor: pointer; font-size: 0.7rem; display: flex; align-items: center; justify-content: center; transition: all 0.15s; }
.gce-arrow:hover:not(:disabled) { background: var(--brand); color: white; border-color: var(--brand); }
.gce-arrow:disabled { opacity: 0.3; cursor: not-allowed; }
.gce-order-pos { font-size: 0.75rem; font-weight: 700; color: var(--brand); }

/* Radio group */
.gce-radio-group { display: flex; gap: 10px; flex-wrap: wrap; }
.gce-radio-opt {
  display: flex; align-items: center; gap: 8px; cursor: pointer;
  padding: 8px 16px; border: 2px solid var(--border);
  border-radius: var(--r-md); font-size: 0.88rem; transition: all 0.2s;
}
.gce-radio-opt.active { border-color: var(--brand); background: var(--brand-light); color: var(--brand); font-weight: 600; }
.gce-radio-opt input { display: none; }

/* Toggle */
.gce-toggle { display: flex; align-items: center; gap: 10px; cursor: pointer; font-size: 0.88rem; color: var(--dark); }
.gce-toggle input { display: none; }
.gce-toggle-track { width: 40px; height: 22px; background: var(--border); border-radius: 20px; position: relative; transition: background 0.2s; }
.gce-toggle input:checked ~ .gce-toggle-track { background: var(--brand); }
.gce-toggle-thumb { width: 18px; height: 18px; background: white; border-radius: 50%; position: absolute; top: 2px; left: 2px; transition: transform 0.2s; box-shadow: 0 1px 4px rgba(0,0,0,0.2); }
.gce-toggle input:checked ~ .gce-toggle-track .gce-toggle-thumb { transform: translateX(18px); }

/* Buttons */
.gce-add-btn {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 18px; border: 2px dashed var(--border);
  background: transparent; color: var(--muted);
  border-radius: var(--r-md); cursor: pointer; font-size: 0.87rem;
  font-weight: 600; transition: all 0.2s; width: fit-content;
}
.gce-add-btn:hover:not(:disabled) { border-color: var(--brand); color: var(--brand); background: var(--brand-light); }
.gce-add-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.gce-add-btn-sm { font-size: 0.82rem; color: var(--brand); background: none; border: none; cursor: pointer; padding: 4px 0; font-weight: 600; }
.gce-remove {
  background: none; border: none; cursor: pointer; color: var(--muted);
  padding: 4px; border-radius: 4px; font-size: 0.85rem; transition: all 0.15s;
  flex-shrink: 0;
}
.gce-remove:hover:not(:disabled) { color: var(--danger); background: var(--danger-bg); }
.gce-remove:disabled { opacity: 0.2; cursor: not-allowed; }

/* Glassmorphism Editor Styles */
.glass-badge-editor {
  display: inline-flex; align-items: center; gap: 8px; font-size: 0.8rem; font-weight: 700;
  color: white; background: linear-gradient(135deg, rgba(99, 102, 241, 0.9), rgba(139, 92, 246, 0.9));
  padding: 6px 14px; border-radius: 999px; backdrop-filter: blur(8px);
  box-shadow: 0 4px 14px rgba(139, 92, 246, 0.25); border: 1px solid rgba(255, 255, 255, 0.2);
}
.gce-label-icon { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.gce-icon-sublabel { display: flex; align-items: center; gap: 6px; font-weight: 600; color: var(--dark); }
.gce-opt-span { display: flex; align-items: center; gap: 6px; }
.glass-icon-xs { display: inline-flex; align-items: center; justify-content: center; }
.star-glow { color: #f59e0b; filter: drop-shadow(0 0 4px rgba(245, 158, 11, 0.4)); }
</style>
