<script setup lang="ts">
/**
 * Seguimiento de alumnos de un curso, en dos modales encadenados.
 *
 * El primero lista a los inscritos con su avance; al pulsar uno, el segundo
 * muestra su detalle lección por lección, sus exámenes y sus puntos.
 *
 * Dos modales y no una vista aparte porque el instructor entra desde la tarjeta
 * del curso y vuelve a ella: navegar fuera le haría perder el sitio en una
 * parrilla que puede tener veinte cursos.
 */
import { ref, computed, watch } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'
import { listaDe } from '../utils/lista'
import { porcentaje, fechaHora, entero } from '../utils/dinero'
import { useScrollLock } from '../composables/useScrollLock'

const props = defineProps<{
  cursoId: string | null
  cursoTitulo?: string
}>()
const emit = defineEmits<{ cerrar: [] }>()

interface Inscrito {
  user_id: string
  nombre: string
  email: string
  avatar_url: string
  inscrito_at: string
  por_licencia: boolean
  completadas: number
  total_lecciones: number
  ultima_actividad: string
  puntos: number
}

interface LeccionDetalle {
  id: string
  title: string
  tipo: string
  orden: number
  duracion_min: number
  completada: boolean
}

interface ExamenDetalle {
  id: string
  title: string
  presentado: boolean
  puntaje?: number
  porcentaje?: number
  submitted_at?: string
}

interface Detalle {
  user_id: string
  nombre: string
  email: string
  completado: boolean
  completadas: number
  total_lecciones: number
  ultima_actividad: string
  puntos: number
  lecciones: LeccionDetalle[]
  examenes: ExamenDetalle[]
}

const inscritos = ref<Inscrito[]>([])
const cargando = ref(false)
const filtro = ref('')

const detalle = ref<Detalle | null>(null)
const cargandoDetalle = ref(false)

const abierto = computed(() => !!props.cursoId)
useScrollLock(abierto)

async function cargar(id: string) {
  cargando.value = true
  inscritos.value = []
  try {
    const { data } = await api.get(`/instructor/capacitaciones/${id}/inscritos`)
    inscritos.value = listaDe<Inscrito>(data, 'inscritos')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos cargar los alumnos')
  } finally {
    cargando.value = false
  }
}

watch(
  () => props.cursoId,
  (id) => {
    detalle.value = null
    filtro.value = ''
    if (id) cargar(id)
  },
  { immediate: true }
)

async function abrirDetalle(a: Inscrito) {
  if (!props.cursoId) return
  cargandoDetalle.value = true
  // Se abre con lo que ya sabemos de la fila para que la cabecera aparezca al
  // instante; el resto se rellena al llegar la respuesta.
  detalle.value = {
    user_id: a.user_id, nombre: a.nombre, email: a.email,
    completado: false, completadas: a.completadas, total_lecciones: a.total_lecciones,
    ultima_actividad: a.ultima_actividad, puntos: a.puntos, lecciones: [], examenes: [],
  }
  try {
    const { data } = await api.get(`/instructor/capacitaciones/${props.cursoId}/inscritos/${a.user_id}`)
    detalle.value = {
      ...data,
      lecciones: listaDe<LeccionDetalle>(data, 'lecciones'),
      examenes: listaDe<ExamenDetalle>(data, 'examenes'),
    }
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos cargar el detalle')
    detalle.value = null
  } finally {
    cargandoDetalle.value = false
  }
}

const visibles = computed(() => {
  const q = filtro.value.trim().toLowerCase()
  if (!q) return inscritos.value
  return inscritos.value.filter((a) =>
    [a.nombre, a.email].some((c) => (c || '').toLowerCase().includes(q))
  )
})

/** Cuántos terminaron el curso. Con cero lecciones nadie cuenta como completo. */
const terminados = computed(
  () => inscritos.value.filter((a) => a.total_lecciones > 0 && a.completadas >= a.total_lecciones).length
)

function iniciales(nombre: string): string {
  return (nombre || '?').split(' ').filter(Boolean).slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? '').join('')
}

/** Etiqueta del tipo de lección, que llega como el enum del proto. */
function tipoLegible(tipo: string): string {
  const t = (tipo || '').toUpperCase()
  if (t.includes('VIDEO')) return 'Video'
  if (t.includes('DOC')) return 'Documento'
  if (t.includes('GAME') || t.includes('JUEGO')) return 'Actividad'
  if (t.includes('TEXT') || t.includes('LECT')) return 'Lectura'
  return 'Lección'
}
</script>

<template>
  <Teleport to="body">
    <!-- ── Modal 1: lista de alumnos ─────────────────────────────────── -->
    <Transition name="sa-fade">
      <div v-if="abierto" class="sa-overlay" @click.self="emit('cerrar')">
        <div class="sa-modal" role="dialog" aria-modal="true" aria-labelledby="sa-titulo">
          <header class="sa-head">
            <div>
              <h2 id="sa-titulo">Avance y puntuaciones</h2>
              <p class="sa-sub">{{ cursoTitulo }}</p>
            </div>
            <button class="sa-cerrar" aria-label="Cerrar" @click="emit('cerrar')">✕</button>
          </header>

          <div class="sa-resumen">
            <span><strong>{{ entero(inscritos.length) }}</strong> inscritos</span>
            <span><strong>{{ entero(terminados) }}</strong> completaron</span>
          </div>

          <input
            v-model="filtro" class="field-input sa-buscar" type="search"
            placeholder="Buscar por nombre o correo…" aria-label="Buscar alumno"
          />

          <div class="sa-body">
            <p v-if="cargando" class="sa-vacio">Cargando…</p>
            <p v-else-if="!inscritos.length" class="sa-vacio">
              Todavía no hay nadie inscrito a esta capacitación.
            </p>
            <p v-else-if="!visibles.length" class="sa-vacio">
              Ningún alumno coincide con «{{ filtro }}».
            </p>

            <ul v-else class="sa-lista">
              <li v-for="a in visibles" :key="a.user_id">
                <button class="sa-fila" @click="abrirDetalle(a)">
                  <span class="sa-avatar" aria-hidden="true">{{ iniciales(a.nombre) }}</span>
                  <span class="sa-quien">
                    <span class="sa-nombre">{{ a.nombre }}</span>
                    <span class="sa-meta">{{ a.email }}</span>
                  </span>
                  <span class="sa-avance">
                    <span class="sa-avance-txt">
                      {{ a.completadas }} / {{ a.total_lecciones }} lecciones
                    </span>
                    <span class="sa-barra">
                      <span
                        class="sa-barra-fill"
                        :class="{ lleno: a.total_lecciones > 0 && a.completadas >= a.total_lecciones }"
                        :style="{ width: `${porcentaje(a.completadas, a.total_lecciones)}%` }"
                      ></span>
                    </span>
                  </span>
                  <span class="sa-puntos">{{ entero(a.puntos) }} pts</span>
                </button>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </Transition>

    <!-- ── Modal 2: detalle del alumno ───────────────────────────────── -->
    <Transition name="sa-fade">
      <div v-if="detalle" class="sa-overlay sa-overlay-2" @click.self="detalle = null">
        <div class="sa-modal sa-modal-detalle" role="dialog" aria-modal="true">
          <header class="sa-head">
            <div class="sa-head-quien">
              <span class="sa-avatar grande" aria-hidden="true">{{ iniciales(detalle.nombre) }}</span>
              <div>
                <h2>{{ detalle.nombre }}</h2>
                <p class="sa-sub">{{ detalle.email }}</p>
              </div>
            </div>
            <button class="sa-cerrar" aria-label="Cerrar" @click="detalle = null">✕</button>
          </header>

          <div class="sa-kpis">
            <div class="sa-kpi" :class="{ ok: detalle.completado }">
              <span class="sa-kpi-val">{{ detalle.completado ? 'Sí' : 'No' }}</span>
              <span class="sa-kpi-lbl">Curso completado</span>
            </div>
            <div class="sa-kpi">
              <span class="sa-kpi-val">{{ detalle.completadas }} / {{ detalle.total_lecciones }}</span>
              <span class="sa-kpi-lbl">Lecciones</span>
            </div>
            <div class="sa-kpi">
              <span class="sa-kpi-val">{{ entero(detalle.puntos) }}</span>
              <span class="sa-kpi-lbl">Puntos</span>
            </div>
          </div>

          <p v-if="detalle.ultima_actividad" class="sa-ultima">
            Última lección completada: {{ fechaHora(detalle.ultima_actividad) }}
          </p>

          <div class="sa-body">
            <p v-if="cargandoDetalle" class="sa-vacio">Cargando detalle…</p>

            <template v-else>
              <h3 class="sa-seccion">Lecciones</h3>
              <p v-if="!detalle.lecciones.length" class="sa-vacio-min">
                Esta capacitación todavía no tiene lecciones.
              </p>
              <ul v-else class="sa-lecciones">
                <li v-for="l in detalle.lecciones" :key="l.id" class="sa-leccion">
                  <span class="sa-check" :class="{ ok: l.completada }" aria-hidden="true">
                    {{ l.completada ? '✓' : '' }}
                  </span>
                  <span class="sa-leccion-txt">
                    <span class="sa-leccion-title">{{ l.title }}</span>
                    <span class="sa-meta">
                      {{ tipoLegible(l.tipo) }}
                      <template v-if="l.duracion_min"> · {{ l.duracion_min }} min</template>
                    </span>
                  </span>
                  <span class="sa-leccion-estado" :class="{ ok: l.completada }">
                    {{ l.completada ? 'Completada' : 'Pendiente' }}
                  </span>
                </li>
              </ul>

              <h3 class="sa-seccion">Exámenes</h3>
              <p v-if="!detalle.examenes.length" class="sa-vacio-min">
                Esta capacitación no tiene exámenes.
              </p>
              <ul v-else class="sa-lecciones">
                <li v-for="e in detalle.examenes" :key="e.id" class="sa-leccion">
                  <span class="sa-check" :class="{ ok: e.presentado }" aria-hidden="true">
                    {{ e.presentado ? '✓' : '' }}
                  </span>
                  <span class="sa-leccion-txt">
                    <span class="sa-leccion-title">{{ e.title }}</span>
                    <span v-if="e.presentado && e.submitted_at" class="sa-meta">
                      {{ fechaHora(e.submitted_at) }}
                    </span>
                  </span>
                  <span class="sa-leccion-estado" :class="{ ok: e.presentado }">
                    <template v-if="e.presentado">
                      {{ (e.porcentaje ?? 0).toFixed(0) }}%
                    </template>
                    <template v-else>Sin presentar</template>
                  </span>
                </li>
              </ul>
            </template>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.sa-overlay {
  position: fixed; inset: 0; z-index: 1200;
  background: rgba(0, 0, 0, 0.55);
  display: flex; align-items: center; justify-content: center;
  padding: var(--gutter, 16px);
}

/* El detalle se dibuja por encima de la lista, no en su lugar: al cerrarlo se
 * vuelve donde estabas sin recargar nada. */
.sa-overlay-2 { z-index: 1250; }

.sa-modal {
  display: flex; flex-direction: column;
  width: 100%; max-width: 720px;
  /* Alto acotado con el contenido desplazándose dentro; en móvil ocupa casi
   * toda la pantalla. */
  max-height: min(86vh, 900px);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 18px;
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}
.sa-modal-detalle { max-width: 640px; }

.sa-head {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 16px; padding: 20px 22px 14px;
  border-bottom: 1px solid var(--border-light);
}
.sa-head h2 { margin: 0; font-size: 1.05rem; font-weight: 800; color: var(--text); }
.sa-head-quien { display: flex; align-items: center; gap: 12px; min-width: 0; }
.sa-sub { margin: 2px 0 0; font-size: 0.85rem; color: var(--muted); overflow-wrap: anywhere; }

.sa-cerrar {
  flex-shrink: 0; width: 34px; height: 34px; border-radius: 50%;
  border: 1px solid var(--border); background: var(--surface);
  color: var(--muted); font-size: 0.9rem; cursor: pointer; line-height: 1;
}

.sa-resumen {
  display: flex; gap: 20px; padding: 12px 22px 0;
  font-size: 0.85rem; color: var(--muted);
}
.sa-resumen strong { color: var(--text); font-size: 1rem; }

.sa-buscar { margin: 12px 22px 0; min-height: var(--touch-min, 44px); }

.sa-body { flex: 1; overflow-y: auto; padding: 16px 22px 22px; }

.sa-vacio { padding: 32px 0; text-align: center; color: var(--muted); font-size: 0.92rem; }
.sa-vacio-min { margin: 0 0 14px; color: var(--muted); font-size: 0.86rem; }

.sa-lista, .sa-lecciones { list-style: none; margin: 0; padding: 0; }

.sa-fila {
  display: grid;
  grid-template-columns: 40px minmax(0, 1.4fr) minmax(0, 1.2fr) auto;
  align-items: center; gap: 14px; width: 100%;
  padding: 12px 8px; border: none; border-bottom: 1px solid var(--border-light);
  background: none; text-align: left; cursor: pointer;
}
.sa-fila:hover { background: var(--surface-soft); }

.sa-avatar {
  width: 40px; height: 40px; border-radius: 50%;
  display: grid; place-items: center; flex-shrink: 0;
  background: color-mix(in srgb, var(--brand) 16%, transparent);
  color: var(--brand); font-weight: 800; font-size: 0.8rem;
}
.sa-avatar.grande { width: 46px; height: 46px; font-size: 0.9rem; }

.sa-quien { min-width: 0; display: flex; flex-direction: column; }
.sa-nombre {
  font-weight: 700; font-size: 0.92rem; color: var(--text);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.sa-meta {
  font-size: 0.78rem; color: var(--muted);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}

.sa-avance { min-width: 0; display: flex; flex-direction: column; gap: 5px; }
.sa-avance-txt { font-size: 0.78rem; color: var(--muted); white-space: nowrap; }
.sa-barra { height: 6px; border-radius: 9999px; background: var(--surface-soft); overflow: hidden; }
.sa-barra-fill { display: block; height: 100%; border-radius: 9999px; background: var(--brand); transition: width .3s ease; }
.sa-barra-fill.lleno { background: var(--success); }

.sa-puntos { font-size: 0.82rem; font-weight: 700; color: var(--text); white-space: nowrap; }

/* ── Detalle ──────────────────────────────────────────────────────────── */
.sa-kpis {
  display: grid; grid-template-columns: repeat(3, 1fr);
  gap: 10px; padding: 16px 22px 0;
}
.sa-kpi {
  padding: 12px; border: 1px solid var(--border); border-radius: 12px;
  text-align: center; background: var(--surface-soft);
}
.sa-kpi.ok { border-color: var(--success); background: var(--success-bg); }
.sa-kpi-val { display: block; font-size: 1.1rem; font-weight: 800; color: var(--text); }
.sa-kpi-lbl { display: block; margin-top: 2px; font-size: 0.72rem; color: var(--muted); }

.sa-ultima { margin: 12px 22px 0; font-size: 0.8rem; color: var(--muted); }

.sa-seccion {
  margin: 6px 0 10px; font-size: 0.78rem; font-weight: 800;
  letter-spacing: 0.08em; text-transform: uppercase; color: var(--muted);
}
.sa-seccion + * + .sa-seccion, .sa-lecciones + .sa-seccion { margin-top: 22px; }

.sa-leccion {
  display: grid; grid-template-columns: 24px minmax(0, 1fr) auto;
  align-items: center; gap: 12px;
  padding: 10px 0; border-bottom: 1px solid var(--border-light);
  /* Igual que en el detalle de exámenes: en un contenedor flex con scroll los
   * hijos se encogen por defecto y el texto se recorta. */
  flex-shrink: 0;
}
.sa-leccion:last-child { border-bottom: none; }

.sa-check {
  width: 22px; height: 22px; border-radius: 50%;
  display: grid; place-items: center; flex-shrink: 0;
  border: 1.5px solid var(--border); color: transparent;
  font-size: 0.72rem; font-weight: 800;
}
.sa-check.ok { background: var(--success); border-color: var(--success); color: #fff; }

.sa-leccion-txt { min-width: 0; display: flex; flex-direction: column; }
.sa-leccion-title {
  font-size: 0.9rem; font-weight: 600; color: var(--text);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.sa-leccion-estado { font-size: 0.78rem; font-weight: 700; color: var(--muted); white-space: nowrap; }
.sa-leccion-estado.ok { color: var(--success); }

.sa-fade-enter-active, .sa-fade-leave-active { transition: opacity .2s ease; }
.sa-fade-enter-from, .sa-fade-leave-to { opacity: 0; }

@media (prefers-reduced-motion: reduce) {
  .sa-fade-enter-active, .sa-fade-leave-active { transition: none; }
  .sa-barra-fill { transition: none; }
}

@media (max-width: 639px) {
  .sa-fila { grid-template-columns: 36px minmax(0, 1fr) auto; row-gap: 8px; }
  .sa-avance { grid-column: 2 / -1; }
  .sa-kpis { grid-template-columns: 1fr; }
}
</style>
