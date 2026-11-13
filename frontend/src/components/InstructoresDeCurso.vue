<script setup lang="ts">
/**
 * Instructores que revisan un curso.
 *
 * Lo abre el administrador desde el listado de capacitaciones. Asignar a alguien
 * aquí le deja VER el avance de los alumnos y los resultados de los exámenes del
 * curso; no le deja editar el contenido ni borrarlo, eso sigue siendo solo del
 * instructor que lo creó.
 *
 * Esa distinción se dice en la propia pantalla a propósito: «asignar un curso a
 * un instructor» se puede entender como entregárselo, y lo que hace es otra cosa.
 */
import { ref, watch, computed } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'

const props = defineProps<{
  cursoId: string | null
  cursoTitulo: string
}>()
const emit = defineEmits<{ (e: 'close'): void }>()

interface Asignado {
  instructor_id: string
  nombre: string
  email: string
  avatar_url?: string
  asignado_at: string
}

const asignados = ref<Asignado[]>([])
const cargando = ref(false)
const guardando = ref(false)

const busqueda = ref('')
const resultados = ref<Array<{ id: string; name: string; email: string; role: string }>>([])
const buscando = ref(false)
let debounce: ReturnType<typeof setTimeout> | null = null

/** Ya asignados: se excluyen de los resultados para no ofrecer duplicados. */
const yaAsignados = computed(() => new Set(asignados.value.map(a => a.instructor_id)))

const candidatos = computed(() =>
  resultados.value.filter(u => !yaAsignados.value.has(u.id)),
)

async function cargar() {
  if (!props.cursoId) return
  cargando.value = true
  try {
    const { data } = await api.get(`/admin/capacitaciones/${props.cursoId}/instructores`)
    asignados.value = data?.instructores ?? []
  } catch (e: unknown) {
    toast.error(mensajeDe(e, 'No se pudieron cargar los instructores'))
  } finally {
    cargando.value = false
  }
}

watch(() => props.cursoId, cargar, { immediate: true })

function buscar() {
  if (debounce) clearTimeout(debounce)
  const q = busqueda.value.trim()
  if (q.length < 2) { resultados.value = []; return }
  // Se espera a que deje de teclear: sin esto son cinco consultas para escribir
  // un nombre de cinco letras.
  debounce = setTimeout(async () => {
    buscando.value = true
    try {
      const { data } = await api.get('/usuarios/search', { params: { q } })
      // El buscador devuelve a todo el mundo; aquí solo sirven instructores.
      resultados.value = (data ?? []).filter(
        (u: { role: string }) => u.role === 'instructor' || u.role === 'admin',
      )
    } catch {
      resultados.value = []
    } finally {
      buscando.value = false
    }
  }, 300)
}

async function asignar(instructorID: string) {
  if (!props.cursoId) return
  guardando.value = true
  try {
    await api.post(`/admin/capacitaciones/${props.cursoId}/instructores`, {
      instructor_id: instructorID,
    })
    busqueda.value = ''
    resultados.value = []
    await cargar()
    toast.success('Instructor asignado')
  } catch (e: unknown) {
    toast.error(mensajeDe(e, 'No se pudo asignar'))
  } finally {
    guardando.value = false
  }
}

async function quitar(a: Asignado) {
  if (!props.cursoId) return
  if (!await toast.confirm(
    `${a.nombre} dejará de ver el avance y los exámenes de esta capacitación.`,
    '¿Quitar el acceso?',
  )) return
  try {
    await api.delete(`/admin/capacitaciones/${props.cursoId}/instructores/${a.instructor_id}`)
    await cargar()
    toast.success('Acceso retirado')
  } catch (e: unknown) {
    toast.error(mensajeDe(e, 'No se pudo quitar'))
  }
}

/** Saca el motivo del servidor; el backend explica por qué rechazó. */
function mensajeDe(e: unknown, respaldo: string): string {
  return (e as { response?: { data?: { error?: string } } })?.response?.data?.error || respaldo
}

function iniciales(nombre: string) {
  return (nombre || '?').split(' ').map(p => p[0]).join('').toUpperCase().slice(0, 2)
}

function fecha(iso: string) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString('es', { day: '2-digit', month: 'short', year: 'numeric' })
}
</script>

<template>
  <Teleport to="body">
    <div class="ic-overlay" @click.self="emit('close')">
      <div class="ic-panel" role="dialog" aria-labelledby="ic-titulo">
        <header class="ic-header">
          <div>
            <h2 id="ic-titulo">Instructores que revisan</h2>
            <p class="ic-sub">{{ cursoTitulo }}</p>
          </div>
          <button class="ic-cerrar" type="button" aria-label="Cerrar" @click="emit('close')">
            <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/>
            </svg>
          </button>
        </header>

        <!--
          Se explica el alcance arriba del todo. "Asignar un curso a un
          instructor" suena a entregárselo, y lo que hace es dar acceso de
          lectura: sin este aviso, un administrador puede creer que transfirió
          el curso y que el otro ya puede editarlo.
        -->
        <p class="ic-alcance">
          Pueden <strong>ver el avance de los alumnos y los resultados de los exámenes</strong>
          de esta capacitación. No pueden editar el contenido ni eliminarla:
          eso sigue siendo solo de quien la creó.
        </p>

        <div class="ic-buscar">
          <label class="ic-label" for="ic-q">Agregar instructor</label>
          <input
            id="ic-q"
            v-model="busqueda"
            class="ic-input"
            type="search"
            placeholder="Busca por nombre o correo…"
            autocomplete="off"
            @input="buscar"
          />
          <p v-if="busqueda.trim().length === 1" class="ic-pista">Escribe al menos dos letras.</p>

          <ul v-if="candidatos.length" class="ic-resultados">
            <li v-for="u in candidatos" :key="u.id">
              <button type="button" class="ic-resultado" :disabled="guardando" @click="asignar(u.id)">
                <span class="ic-avatar">{{ iniciales(u.name) }}</span>
                <span class="ic-resultado-datos">
                  <span class="ic-nombre">{{ u.name }}</span>
                  <span class="ic-email">{{ u.email }}</span>
                </span>
                <span class="ic-agregar">Agregar</span>
              </button>
            </li>
          </ul>
          <p v-else-if="buscando" class="ic-pista">Buscando…</p>
          <p v-else-if="busqueda.trim().length >= 2" class="ic-pista">
            Ningún instructor coincide. Solo aparecen cuentas con perfil de instructor.
          </p>
        </div>

        <div class="ic-lista">
          <p v-if="cargando" class="ic-pista">Cargando…</p>
          <p v-else-if="asignados.length === 0" class="ic-vacio">
            Nadie más revisa esta capacitación todavía.
          </p>
          <ul v-else>
            <li v-for="a in asignados" :key="a.instructor_id" class="ic-fila">
              <img v-if="a.avatar_url" :src="a.avatar_url" :alt="a.nombre" class="ic-avatar-img" />
              <span v-else class="ic-avatar">{{ iniciales(a.nombre) }}</span>
              <span class="ic-fila-datos">
                <span class="ic-nombre">{{ a.nombre }}</span>
                <span class="ic-email">{{ a.email }}</span>
              </span>
              <span class="ic-desde">desde {{ fecha(a.asignado_at) }}</span>
              <button
                type="button"
                class="ic-quitar"
                :aria-label="`Quitar a ${a.nombre}`"
                @click="quitar(a)"
              >
                <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path d="M18 6L6 18M6 6l12 12" stroke-linecap="round"/>
                </svg>
              </button>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.ic-overlay {
  position: fixed; inset: 0; z-index: 70;
  background: rgba(0,0,0,.5);
  display: flex; align-items: center; justify-content: center;
  padding: 1rem;
}
.ic-panel {
  background: var(--surface, #fff); color: var(--text);
  border-radius: .9rem; width: 100%; max-width: 560px;
  max-height: 86vh; display: flex; flex-direction: column;
  box-shadow: 0 24px 60px rgba(0,0,0,.35);
}
.ic-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 1rem; padding: 1.1rem 1.25rem .5rem;
}
.ic-header h2 { margin: 0; font-size: 1.05rem; }
.ic-sub { margin: .15rem 0 0; font-size: .85rem; color: var(--text-muted); }
.ic-cerrar {
  background: none; border: none; cursor: pointer; color: var(--text-muted);
  padding: 4px; border-radius: 6px; flex-shrink: 0;
}
.ic-cerrar:hover { background: var(--surface-hover); color: var(--text); }

.ic-alcance {
  margin: .25rem 1.25rem .75rem;
  padding: .6rem .75rem;
  border-radius: .5rem;
  font-size: .8rem; line-height: 1.45;
  color: var(--text-muted);
  background: var(--bg);
  border: 1px solid var(--border);
}

.ic-buscar { padding: 0 1.25rem; }
.ic-label { display: block; font-size: .78rem; font-weight: 600; margin-bottom: .3rem; }
.ic-input {
  width: 100%; padding: .55rem .7rem;
  border: 1px solid var(--border); border-radius: .5rem;
  background: var(--bg); color: var(--text); font-size: .88rem;
}
.ic-pista { margin: .4rem 0 0; font-size: .78rem; color: var(--text-muted); }

.ic-resultados { list-style: none; margin: .5rem 0 0; padding: 0; }
.ic-resultado {
  display: flex; align-items: center; gap: .6rem; width: 100%;
  padding: .5rem; border: 1px solid var(--border); border-radius: .5rem;
  background: var(--bg); cursor: pointer; margin-bottom: .35rem;
  color: var(--text); text-align: left;
}
.ic-resultado:hover:not(:disabled) { background: var(--surface-hover); }
.ic-resultado:disabled { opacity: .6; cursor: progress; }
.ic-resultado-datos, .ic-fila-datos { display: flex; flex-direction: column; min-width: 0; flex: 1; }
.ic-agregar { font-size: .76rem; font-weight: 600; color: var(--primary, #3b82f6); flex-shrink: 0; }

/*
 * La lista crece y el panel tiene altura máxima: sin flex-shrink el contenedor
 * se comprime en vez de desplazarse, que es el mismo fallo que tuvieron las
 * tarjetas de examen y el hilo de mensajes.
 */
.ic-lista { padding: .9rem 1.25rem 1.25rem; overflow-y: auto; flex-shrink: 1; min-height: 0; }
.ic-lista ul { list-style: none; margin: 0; padding: 0; }
.ic-vacio { margin: 0; font-size: .85rem; color: var(--text-muted); text-align: center; padding: 1rem 0; }

.ic-fila {
  display: flex; align-items: center; gap: .6rem;
  padding: .55rem 0; border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.ic-fila:last-child { border-bottom: none; }
.ic-avatar, .ic-avatar-img {
  width: 34px; height: 34px; border-radius: 50%; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-size: .74rem; font-weight: 700;
  background: var(--primary, #3b82f6); color: #fff;
}
.ic-avatar-img { object-fit: cover; }
.ic-nombre { font-size: .88rem; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.ic-email { font-size: .76rem; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.ic-desde { font-size: .72rem; color: var(--text-muted); white-space: nowrap; flex-shrink: 0; }
.ic-quitar {
  background: none; border: none; cursor: pointer; color: var(--text-muted);
  padding: 4px; border-radius: 6px; flex-shrink: 0;
}
.ic-quitar:hover { background: rgba(239,68,68,.12); color: #ef4444; }
</style>
