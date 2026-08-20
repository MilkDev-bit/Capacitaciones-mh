<script setup lang="ts">
/**
 * Editor de texto largo con formato, para instructores.
 *
 * Markdown y no un editor visual a propósito: lo que se guarda es texto plano.
 * Se lee tal cual en la base de datos, no arrastra HTML de un editor concreto y,
 * si algún día se cambia de herramienta, el contenido migra sin limpiar nada.
 * El precio es enseñar cuatro marcas, y por eso hay pestaña de vista previa y
 * botones que las insertan.
 */
import { ref, computed } from 'vue'
import { renderMarkdown } from '../utils/markdown'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  /** Filas del área de texto en la pestaña de escritura. */
  filas?: number
}>(), { placeholder: '', filas: 8 })

const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const pestana = ref<'escribir' | 'vista'>('escribir')
const areaRef = ref<HTMLTextAreaElement | null>(null)

const htmlPrevio = computed(() => renderMarkdown(props.modelValue))

function actualizar(e: Event) {
  emit('update:modelValue', (e.target as HTMLTextAreaElement).value)
}

/**
 * Envuelve la selección con una marca, o la inserta donde esté el cursor.
 *
 * Se conserva la posición del cursor después de escribir: sin eso, el cursor
 * salta al final del texto en cada pulsación y da formato a la mitad de un
 * párrafo obliga a volver a buscar el sitio con el ratón.
 */
function envolver(antes: string, despues = antes) {
  const area = areaRef.value
  if (!area) return
  const { selectionStart: ini, selectionEnd: fin } = area
  const texto = props.modelValue
  const sel = texto.slice(ini, fin)
  const nuevo = texto.slice(0, ini) + antes + sel + despues + texto.slice(fin)
  emit('update:modelValue', nuevo)
  requestAnimationFrame(() => {
    area.focus()
    // Sin selección, el cursor queda entre las marcas, listo para escribir.
    const pos = sel ? ini + antes.length + sel.length + despues.length : ini + antes.length
    area.setSelectionRange(sel ? ini + antes.length : pos, sel ? ini + antes.length + sel.length : pos)
  })
}

/** Antepone una marca a cada línea seleccionada (listas, cita). */
function prefijarLineas(marca: string) {
  const area = areaRef.value
  if (!area) return
  const { selectionStart: ini, selectionEnd: fin } = area
  const texto = props.modelValue
  const desde = texto.lastIndexOf('\n', ini - 1) + 1
  const hasta = texto.indexOf('\n', fin) === -1 ? texto.length : texto.indexOf('\n', fin)
  const bloque = texto.slice(desde, hasta)
    .split('\n')
    .map(l => (l.startsWith(marca) ? l.slice(marca.length) : marca + l))
    .join('\n')
  emit('update:modelValue', texto.slice(0, desde) + bloque + texto.slice(hasta))
  requestAnimationFrame(() => area.focus())
}
</script>

<template>
  <div class="md">
    <div class="md-barra">
      <div class="md-tabs" role="tablist">
        <button type="button" role="tab" :aria-selected="pestana === 'escribir'"
                :class="{ activa: pestana === 'escribir' }" @click="pestana = 'escribir'">
          Escribir
        </button>
        <button type="button" role="tab" :aria-selected="pestana === 'vista'"
                :class="{ activa: pestana === 'vista' }" @click="pestana = 'vista'">
          Vista previa
        </button>
      </div>

      <div v-if="pestana === 'escribir'" class="md-acciones">
        <button type="button" title="Negrita" @click="envolver('**')"><b>B</b></button>
        <button type="button" title="Cursiva" @click="envolver('*')"><i>I</i></button>
        <button type="button" title="Lista" @click="prefijarLineas('- ')">☰</button>
        <button type="button" title="Lista numerada" @click="prefijarLineas('1. ')">№</button>
        <button type="button" title="Subtítulo" @click="prefijarLineas('### ')">H</button>
      </div>
    </div>

    <textarea v-if="pestana === 'escribir'" ref="areaRef" class="field-input md-area"
              :value="modelValue" :rows="filas" :placeholder="placeholder"
              @input="actualizar" />

    <!--
      v-html sobre la salida de renderMarkdown, que sanea siempre.
      No se pinta `modelValue` directamente en ningún sitio.
    -->
    <div v-else class="md-vista md-render" v-html="htmlPrevio || '<p class=\'md-vacio\'>Nada que previsualizar todavía.</p>'" />

    <small v-if="pestana === 'escribir'" class="md-ayuda">
      Usa <code>**negrita**</code>, <code>*cursiva*</code>, <code>- lista</code> y
      una línea en blanco para separar párrafos.
    </small>
  </div>
</template>

<style scoped>
.md { display: flex; flex-direction: column; gap: 8px; }

.md-barra {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
}

.md-tabs { display: flex; gap: 4px; }

.md-tabs button {
  padding: 6px 12px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: var(--muted);
  font-size: 0.82rem;
  font-weight: 700;
  cursor: pointer;
}

.md-tabs button.activa {
  background: var(--surface);
  border-color: var(--border);
  color: var(--text);
}

.md-acciones { display: flex; gap: 4px; }

.md-acciones button {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface);
  color: var(--text);
  font-size: 0.86rem;
  cursor: pointer;
}

.md-acciones button:hover { background: var(--bg); }

.md-area {
  resize: vertical;
  min-height: 140px;
  font-family: inherit;
  line-height: 1.55;
}

.md-vista {
  min-height: 140px;
  padding: 12px 14px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface);
}

.md-ayuda { font-size: 0.74rem; color: var(--muted); }
.md-ayuda code {
  background: var(--bg);
  border-radius: 4px;
  padding: 1px 4px;
  font-size: 0.92em;
}
</style>
