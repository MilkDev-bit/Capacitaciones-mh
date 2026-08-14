<script setup lang="ts">
/**
 * Listado de constancias DC-3 emitidas del alumno.
 *
 * Cubre el hueco que hacía inútil todo lo anterior: el endpoint
 * `GET /mis-constancias` existía desde el principio pero ningún componente lo
 * consumía, así que una constancia emitida solo era accesible volviendo a
 * entrar en el curso concreto que la generó. Quien terminaba cinco cursos no
 * tenía forma de ver los cinco documentos juntos.
 *
 * Es además el destino al que apuntan los mensajes del panel DC-3 cuando dicen
 * "la verás en Mis constancias".
 */
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
import EmptyState from '../../components/EmptyState.vue'
import { useDC3Store } from '../../stores/dc3'

type Constancia = {
  capacitacion_id: string
  capacitacion_titulo: string
  archivo_url: string
  generada_at: string
}

const router = useRouter()
const dc3 = useDC3Store()

const constancias = ref<Constancia[]>([])
const cargando = ref(true)
const error = ref('')

function formatearFecha(valor: string) {
  if (!valor) return ''
  const d = new Date(valor)
  // Un timestamp inválido pintaría "Invalid Date" en la tarjeta.
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleDateString('es-MX', { day: '2-digit', month: 'long', year: 'numeric' })
}

async function cargar() {
  cargando.value = true
  error.value = ''
  try {
    const { data } = await api.get('/mis-constancias')
    constancias.value = Array.isArray(data) ? data : []
  } catch {
    // Se distingue del vacío legítimo: "no tienes constancias" y "no pudimos
    // consultarlas" llevan al alumno a hacer cosas muy distintas.
    error.value = 'No pudimos cargar tus constancias.'
  } finally {
    cargando.value = false
  }
}

onMounted(cargar)
</script>

<template>
  <div class="page-content">
    <header class="constancias-head">
      <h1>Mis constancias</h1>
      <p>
        Constancias DC-3 de habilidades laborales emitidas a tu nombre. Se generan
        solas al terminar una capacitación que las incluya.
      </p>
    </header>

    <div v-if="cargando" class="constancias-estado">Cargando…</div>

    <div v-else-if="error" class="constancias-estado">
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="cargar">Reintentar</button>
    </div>

    <EmptyState
      v-else-if="constancias.length === 0"
      title="Todavía no tienes constancias"
      description="Cuando completes una capacitación con constancia DC-3 la encontrarás aquí."
    >
      <template #action>
        <button class="btn btn-primary" @click="router.push('/usuario/capacitaciones')">
          Ver mis cursos
        </button>
      </template>
    </EmptyState>

    <ul v-else class="auto-grid constancias-lista" style="--auto-grid-max: 380px">
      <li v-for="c in constancias" :key="c.capacitacion_id" class="constancia-card">
        <div class="constancia-icono" aria-hidden="true">
          <svg width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
            <path d="M14 2v6h6M9 15l2 2 4-4" />
          </svg>
        </div>
        <div class="constancia-datos">
          <h2>{{ c.capacitacion_titulo || 'Capacitación' }}</h2>
          <p v-if="formatearFecha(c.generada_at)">
            Emitida el {{ formatearFecha(c.generada_at) }}
          </p>
        </div>
        <div class="constancia-acciones">
          <a :href="c.archivo_url" target="_blank" rel="noopener" class="btn btn-primary btn-sm" download>
            Descargar
          </a>
          <!-- Reabre el formulario por si hay que corregir un dato del documento. -->
          <button class="btn btn-secondary btn-sm"
                  @click="dc3.abrir(c.capacitacion_id, c.capacitacion_titulo)">
            Ver datos
          </button>
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.constancias-head { margin-bottom: 24px; }

.constancias-head h1 {
  margin: 0 0 6px;
  font-size: clamp(1.35rem, 1.1rem + 1vw, 1.75rem);
  color: var(--text);
}

.constancias-head p {
  margin: 0;
  color: var(--muted);
  font-size: 0.9rem;
  max-width: 62ch;
  line-height: 1.55;
}

.constancias-estado {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  color: var(--muted);
  padding: 32px 0;
}

.constancias-lista { list-style: none; margin: 0; padding: 0; }

.constancia-card {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 12px 14px;
  align-items: start;
  padding: 18px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--surface);
}

.constancia-icono {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: color-mix(in srgb, var(--brand) 12%, transparent);
  color: var(--brand);
}

.constancia-datos h2 {
  margin: 0 0 4px;
  font-size: 0.98rem;
  color: var(--text);
  /* El título del curso puede ser largo; sin esto desborda la tarjeta. */
  overflow-wrap: anywhere;
}

.constancia-datos p {
  margin: 0;
  font-size: 0.82rem;
  color: var(--muted);
}

.constancia-acciones {
  grid-column: 1 / -1;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.constancia-acciones > * {
  min-height: var(--touch-min);
  display: inline-flex;
  align-items: center;
}

@media (max-width: 479px) {
  .constancia-acciones > * { flex: 1 1 100%; justify-content: center; }
}
</style>
