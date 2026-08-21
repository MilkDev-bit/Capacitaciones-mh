<script setup lang="ts">
/**
 * Empresas / Licencias (solo administradores).
 *
 * Esta pantalla no existía. El menú lateral enlazaba a /admin/licencias, pero
 * esa ruta nunca se declaró y el panel salía en blanco: no era un fallo de
 * carga, era una pantalla que faltaba entera.
 *
 * Solo se listan las licencias con comprador, que son las vendidas a una
 * empresa. Las que no lo tienen son plantillas que un instructor dejó creadas
 * y todavía no ha comprado nadie; en una pantalla de "Empresas" solo harían
 * ruido.
 */
import { ref, computed, onMounted } from 'vue'
import api from '../../api'
import { toast } from '../../utils/toast'
import { pesos, porcentaje, fechaHora } from '../../utils/dinero'

interface Licencia {
  id: string
  nombre: string
  capacitacion_id: string
  capacitacion_titulo: string
  comprador_id: string
  comprador_nombre: string
  comprador_email: string
  precio_centavos: number
  capacidad_maxima: number
  usadas: number
  codigo_acceso: string
  created_at: string
}

const licencias = ref<Licencia[]>([])
const totalAsientos = ref(0)
const asientosUsados = ref(0)
const cargando = ref(true)
const filtro = ref('')

async function cargar() {
  cargando.value = true
  try {
    const { data } = await api.get('/admin/licencias')
    licencias.value = data.licencias || []
    totalAsientos.value = data.total_asientos || 0
    asientosUsados.value = data.asientos_usados || 0
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos cargar las licencias')
  } finally {
    cargando.value = false
  }
}

onMounted(cargar)

const visibles = computed(() => {
  const q = filtro.value.trim().toLowerCase()
  if (!q) return licencias.value
  return licencias.value.filter((l) =>
    [l.comprador_nombre, l.comprador_email, l.capacitacion_titulo, l.nombre, l.codigo_acceso]
      .some((c) => (c || '').toLowerCase().includes(q))
  )
})

const facturado = computed(() =>
  licencias.value.reduce((suma, l) => suma + (l.precio_centavos || 0), 0)
)

const ocupacion = computed(() => porcentaje(asientosUsados.value, totalAsientos.value))

async function copiar(codigo: string) {
  try {
    await navigator.clipboard.writeText(codigo)
    toast.success('Código copiado')
  } catch {
    // Sin permiso de portapapeles (o sin HTTPS) no se puede copiar; se dice,
    // en vez de dejar al usuario creyendo que sí.
    toast.error('Tu navegador no permitió copiar. Selecciónalo a mano.')
  }
}
</script>

<template>
  <div class="lic">
    <header class="lic-head">
      <div>
        <p class="lic-eyebrow">Administración</p>
        <h1 class="lic-title">Empresas y licencias</h1>
      </div>
      <input
        v-model="filtro"
        class="lic-buscar field-input"
        type="search"
        placeholder="Buscar empresa, curso o código…"
        aria-label="Buscar licencias"
      />
    </header>

    <section class="lic-resumen">
      <article class="lic-kpi">
        <p class="lic-kpi-label">Licencias vendidas</p>
        <p class="lic-kpi-valor">{{ licencias.length }}</p>
      </article>
      <article class="lic-kpi">
        <p class="lic-kpi-label">Facturado</p>
        <p class="lic-kpi-valor">{{ pesos(facturado, false) }}</p>
      </article>
      <article class="lic-kpi">
        <p class="lic-kpi-label">Asientos ocupados</p>
        <p class="lic-kpi-valor">{{ asientosUsados }} / {{ totalAsientos }}</p>
        <div class="lic-barra" role="presentation">
          <div class="lic-barra-fill" :style="{ width: `${ocupacion}%` }"></div>
        </div>
      </article>
    </section>

    <p v-if="cargando" class="lic-vacio">Cargando…</p>
    <p v-else-if="!licencias.length" class="lic-vacio">
      Todavía no se ha vendido ninguna licencia a empresas.
    </p>
    <p v-else-if="!visibles.length" class="lic-vacio">
      Ninguna licencia coincide con «{{ filtro }}».
    </p>

    <ul v-else class="lic-lista">
      <li v-for="l in visibles" :key="l.id" class="lic-item">
        <div class="lic-item-main">
          <p class="lic-empresa">{{ l.comprador_nombre }}</p>
          <p class="lic-meta">{{ l.comprador_email }}</p>
          <p class="lic-curso">{{ l.capacitacion_titulo }} · {{ l.nombre }}</p>
        </div>

        <div class="lic-item-cupos">
          <p class="lic-cupos-txt">{{ l.usadas }} / {{ l.capacidad_maxima }} asientos</p>
          <div class="lic-barra" role="presentation">
            <div
              class="lic-barra-fill"
              :class="{ lleno: l.usadas >= l.capacidad_maxima }"
              :style="{ width: `${porcentaje(l.usadas, l.capacidad_maxima)}%` }"
            ></div>
          </div>
        </div>

        <div class="lic-item-lado">
          <p class="lic-precio">{{ pesos(l.precio_centavos) }}</p>
          <p class="lic-meta">{{ fechaHora(l.created_at) }}</p>
          <button v-if="l.codigo_acceso" class="lic-codigo" @click="copiar(l.codigo_acceso)">
            {{ l.codigo_acceso }}
          </button>
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.lic { display: flex; flex-direction: column; gap: 20px; }

.lic-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.lic-eyebrow {
  margin: 0 0 2px;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--muted);
}

.lic-title {
  margin: 0;
  font-size: clamp(1.4rem, 1.1rem + 1.2vw, 2rem);
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text);
}

.lic-buscar { flex: 1 1 260px; max-width: 340px; min-height: var(--touch-min, 44px); }

.lic-resumen {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.lic-kpi {
  padding: 18px;
  border: 1px solid var(--border);
  border-radius: 16px;
  background: var(--surface);
  box-shadow: var(--shadow-sm);
}

.lic-kpi-label {
  margin: 0 0 6px;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
}

.lic-kpi-valor {
  margin: 0;
  font-size: 1.6rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text);
  overflow-wrap: anywhere;
}

.lic-barra {
  margin-top: 10px;
  height: 6px;
  border-radius: 9999px;
  background: var(--surface-soft);
  overflow: hidden;
}

.lic-barra-fill {
  height: 100%;
  border-radius: 9999px;
  background: var(--brand);
  transition: width 0.3s ease;
}

/* Sin cupo libre: el naranja de marca ya no informa, y el rojo sí dice que esa
 * empresa no puede dar de alta a nadie más. */
.lic-barra-fill.lleno { background: var(--danger); }

.lic-vacio {
  padding: 32px;
  text-align: center;
  color: var(--muted);
  font-size: 0.92rem;
}

.lic-lista { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 12px; }

.lic-item {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(0, 1.2fr) auto;
  align-items: center;
  gap: 18px;
  padding: 16px 18px;
  border: 1px solid var(--border);
  border-radius: 16px;
  background: var(--surface);
}

.lic-item-main { min-width: 0; }

.lic-empresa {
  margin: 0 0 2px;
  font-weight: 700;
  font-size: 0.98rem;
  color: var(--text);
  overflow-wrap: anywhere;
}

.lic-curso { margin: 4px 0 0; font-size: 0.82rem; color: var(--text); opacity: 0.75; overflow-wrap: anywhere; }
.lic-meta { margin: 0; font-size: 0.78rem; color: var(--muted); overflow-wrap: anywhere; }

.lic-cupos-txt { margin: 0; font-size: 0.82rem; font-weight: 600; color: var(--text); }

.lic-item-lado { text-align: right; }

.lic-precio {
  margin: 0 0 2px;
  font-size: 1.05rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text);
  white-space: nowrap;
}

.lic-codigo {
  margin-top: 8px;
  min-height: 32px;
  padding: 0 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface-soft);
  color: var(--brand);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-weight: 700;
  font-size: 0.8rem;
  letter-spacing: 0.08em;
  cursor: pointer;
}

@media (max-width: 767px) {
  .lic-item { grid-template-columns: 1fr; gap: 12px; }
  .lic-item-lado { text-align: left; }
}
</style>
