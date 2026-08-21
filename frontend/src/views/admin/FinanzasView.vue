<script setup lang="ts">
/**
 * Panel financiero de Stripe (solo administradores).
 *
 * Todas las cifras salen de lo que se cobró de verdad —`ordenes` y
 * `suscripcion_facturas`— y no del precio de catálogo. La comisión es la que
 * Stripe reporta en el BalanceTransaction de cada cobro, no una tarifa
 * estimada por nosotros.
 *
 * Cuando quedan cobros sin comisión traída de Stripe, la pantalla lo dice en
 * vez de callarlo: en ese caso la ganancia neta es un PISO, porque falta
 * descontar comisiones, y presentarla como definitiva ante la directiva sería
 * enseñar un número mejor que el real.
 */
import { ref, computed, onMounted } from 'vue'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import api from '../../api'
import { toast } from '../../utils/toast'
import { useTheme } from '../../composables/useTheme'
import { pesos, porcentaje, mesCorto, fechaHora } from '../../utils/dinero'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const { isDark } = useTheme()

interface Punto { mes: string; bruto_centavos: number; comision_centavos: number; neto_centavos: number }
interface Movimiento {
  id: string
  cliente: string
  email: string
  fecha: string
  bruto_centavos: number
  comision_centavos: number
  neto_centavos: number
  comision_conocida: boolean
  origen: string
  concepto: string
}
interface Finanzas {
  bruto_centavos: number
  comision_centavos: number
  neto_centavos: number
  moneda: string
  transacciones: number
  sin_comision: number
  serie: Punto[]
  transacciones_recientes: Movimiento[]
}

const datos = ref<Finanzas | null>(null)
const cargando = ref(true)
const rellenando = ref(false)

async function cargar() {
  cargando.value = true
  try {
    const { data } = await api.get('/admin/finanzas')
    datos.value = data
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos cargar las finanzas')
  } finally {
    cargando.value = false
  }
}

/**
 * Trae de Stripe las comisiones que faltan.
 *
 * Va por lotes porque cada cobro es una petición a Stripe. Se puede pulsar
 * varias veces: solo mira los que aún no la tienen.
 */
async function rellenarComisiones() {
  rellenando.value = true
  try {
    const { data } = await api.post('/admin/comisiones/rellenar')
    if (data.restantes > 0) {
      toast.success(`Actualizados ${data.rellenadas}. Quedan ${data.restantes}; vuelve a pulsar.`)
    } else {
      toast.success(`Actualizados ${data.rellenadas}. Ya no falta ninguno.`)
    }
    await cargar()
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos consultar Stripe')
  } finally {
    rellenando.value = false
  }
}

onMounted(cargar)

const bruto = computed(() => datos.value?.bruto_centavos ?? 0)
const comision = computed(() => datos.value?.comision_centavos ?? 0)
const neto = computed(() => datos.value?.neto_centavos ?? 0)
const faltanComisiones = computed(() => (datos.value?.sin_comision ?? 0) > 0)

/**
 * Los anillos miden cada parte contra el bruto, que es el único total con
 * sentido: la comisión y el neto son trozos de él y juntos suman el 100%.
 */
const pctComision = computed(() => porcentaje(comision.value, bruto.value))
const pctNeto = computed(() => porcentaje(neto.value, bruto.value))

/** Ticket medio. Con cero transacciones se enseña un guion, no una división. */
const ticketMedio = computed(() => {
  const n = datos.value?.transacciones ?? 0
  return n > 0 ? pesos(Math.round(bruto.value / n)) : '—'
})

// ── Gráfico ────────────────────────────────────────────────────────────────
const ACENTO = '#22d3ee'

const datosGrafico = computed(() => {
  const serie = datos.value?.serie ?? []
  return {
    labels: serie.map((p) => mesCorto(p.mes)),
    datasets: [
      {
        label: 'Ganancia neta',
        // A pesos para el eje: aquí ya no se suma nada, solo se dibuja.
        data: serie.map((p) => p.neto_centavos / 100),
        borderColor: ACENTO,
        borderWidth: 3,
        // Curva suave como en la referencia, sin llegar a inventar picos que
        // los datos no tienen.
        tension: 0.4,
        fill: true,
        pointBackgroundColor: ACENTO,
        pointBorderColor: isDark.value ? '#1d1d1f' : '#ffffff',
        pointBorderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 7,
        backgroundColor: (ctx: any) => {
          const { chart } = ctx
          if (!chart.chartArea) return 'rgba(34,211,238,0.15)'
          const g = chart.ctx.createLinearGradient(0, chart.chartArea.top, 0, chart.chartArea.bottom)
          g.addColorStop(0, 'rgba(34,211,238,0.38)')
          g.addColorStop(1, 'rgba(34,211,238,0)')
          return g
        },
      },
    ],
  }
})

const opcionesGrafico = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { intersect: false, mode: 'index' as const },
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: isDark.value ? 'rgba(255,255,255,0.95)' : 'rgba(0,0,0,0.85)',
      titleColor: isDark.value ? '#000' : '#fff',
      bodyColor: isDark.value ? '#000' : '#fff',
      padding: 12,
      cornerRadius: 10,
      displayColors: false,
      callbacks: {
        label: (item: any) => `Neto: ${pesos(Math.round(item.parsed.y * 100))}`,
      },
    },
  },
  scales: {
    x: {
      grid: { display: false },
      border: { display: false },
      ticks: { color: isDark.value ? '#aeaeb2' : '#86868b', font: { size: 12 } },
    },
    y: {
      // Arranca en cero: un eje recortado exagera visualmente las subidas y en
      // una pantalla de resultados eso es engañoso.
      beginAtZero: true,
      grid: { color: isDark.value ? 'rgba(255,255,255,0.06)' : 'rgba(0,0,0,0.05)' },
      border: { display: false },
      ticks: {
        color: isDark.value ? '#aeaeb2' : '#86868b',
        font: { size: 12 },
        callback: (v: any) => pesos(Number(v) * 100, false),
      },
    },
  },
}))

/** Iniciales para el avatar del movimiento. */
function iniciales(nombre: string): string {
  return (nombre || '?')
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? '')
    .join('')
}
</script>

<template>
  <div class="fin">
    <header class="fin-head">
      <div>
        <p class="fin-eyebrow">Stripe</p>
        <h1 class="fin-title">Resultados financieros</h1>
      </div>
      <button class="fin-refresh" :disabled="cargando" @click="cargar">
        {{ cargando ? 'Actualizando…' : 'Actualizar' }}
      </button>
    </header>

    <div v-if="cargando && !datos" class="fin-skeleton">Cargando…</div>

    <template v-else-if="datos">
      <!-- Advertencia honesta: sin todas las comisiones, el neto es un piso. -->
      <div v-if="faltanComisiones" class="fin-aviso">
        <div>
          <strong>Faltan comisiones por traer de Stripe.</strong>
          Hay {{ datos.sin_comision }} cobro(s) liquidado(s) sin su comisión registrada, así
          que la ganancia neta de abajo es un <em>mínimo</em>: al descontarlas bajará.
        </div>
        <button class="fin-aviso-btn" :disabled="rellenando" @click="rellenarComisiones">
          {{ rellenando ? 'Consultando a Stripe…' : 'Traer de Stripe' }}
        </button>
      </div>

      <!-- ── KPIs ─────────────────────────────────────────────────────── -->
      <section class="fin-kpis">
        <article class="fin-card">
          <div class="fin-card-txt">
            <p class="fin-card-label">Volumen bruto</p>
            <p class="fin-card-valor">{{ pesos(bruto, false) }}</p>
            <p class="fin-card-pie">{{ datos.transacciones }} transacciones · ticket medio {{ ticketMedio }}</p>
          </div>
          <div class="fin-anillo" style="--pct: 100; --tono: var(--fin-cyan)">
            <span>100%</span>
          </div>
        </article>

        <article class="fin-card">
          <div class="fin-card-txt">
            <p class="fin-card-label">Comisiones Stripe</p>
            <p class="fin-card-valor alerta">−{{ pesos(comision, false) }}</p>
            <p class="fin-card-pie">{{ pctComision.toFixed(1) }}% del volumen bruto</p>
          </div>
          <div class="fin-anillo" :style="{ '--pct': pctComision, '--tono': 'var(--fin-naranja)' }">
            <span>{{ pctComision.toFixed(0) }}%</span>
          </div>
        </article>

        <article class="fin-card">
          <div class="fin-card-txt">
            <p class="fin-card-label">Ganancia neta</p>
            <p class="fin-card-valor positivo">{{ pesos(neto, false) }}</p>
            <p class="fin-card-pie">
              {{ pctNeto.toFixed(1) }}% del bruto
              <span v-if="faltanComisiones" class="fin-pie-aviso">· es un mínimo</span>
            </p>
          </div>
          <div class="fin-anillo" :style="{ '--pct': pctNeto, '--tono': 'var(--fin-lima)' }">
            <span>{{ pctNeto.toFixed(0) }}%</span>
          </div>
        </article>
      </section>

      <!-- ── Tendencia ────────────────────────────────────────────────── -->
      <section class="fin-panel">
        <header class="fin-panel-head">
          <h2>Ganancia neta</h2>
          <span class="fin-panel-nota">Últimos 6 meses</span>
        </header>
        <div class="fin-grafico">
          <Line :data="datosGrafico" :options="opcionesGrafico" />
        </div>
      </section>

      <!-- ── Movimientos ──────────────────────────────────────────────── -->
      <section class="fin-panel">
        <header class="fin-panel-head">
          <h2>Transacciones recientes</h2>
          <span class="fin-panel-nota">Importe neto</span>
        </header>

        <p v-if="!datos.transacciones_recientes.length" class="fin-vacio">
          Todavía no hay cobros registrados.
        </p>

        <ul v-else class="fin-movs">
          <li v-for="m in datos.transacciones_recientes" :key="m.id" class="fin-mov">
            <div class="fin-mov-avatar" aria-hidden="true">{{ iniciales(m.cliente) }}</div>
            <div class="fin-mov-quien">
              <p class="fin-mov-nombre">{{ m.cliente }}</p>
              <p class="fin-mov-meta">{{ m.concepto }}</p>
            </div>
            <p class="fin-mov-fecha">{{ fechaHora(m.fecha) }}</p>
            <div class="fin-mov-importe">
              <p class="fin-mov-neto">{{ pesos(m.neto_centavos) }}</p>
              <p class="fin-mov-bruto">
                <template v-if="m.comision_conocida">
                  {{ pesos(m.bruto_centavos) }} − {{ pesos(m.comision_centavos) }}
                </template>
                <!-- Sin comisión traída, enseñar el bruto como si fuera neto
                     sería mentir por omisión. -->
                <template v-else>{{ pesos(m.bruto_centavos) }} · comisión pendiente</template>
              </p>
            </div>
          </li>
        </ul>
      </section>
    </template>
  </div>
</template>

<style scoped>
/* Acentos vivos solo para los datos, sobre la paleta del proyecto. Se aclaran
 * en tema claro porque un neón sobre blanco pierde contraste y deja de cumplir
 * el mínimo de legibilidad. */
.fin {
  --fin-cyan: #0891b2;
  --fin-naranja: #ea580c;
  --fin-lima: #15803d;
  --fin-rosa: #db2777;
}

:global(html.dark-theme) .fin {
  --fin-cyan: #22d3ee;
  --fin-naranja: #fb923c;
  --fin-lima: #a3e635;
  --fin-rosa: #f472b6;
}

.fin { display: flex; flex-direction: column; gap: 20px; }

.fin-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.fin-eyebrow {
  margin: 0 0 2px;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--muted);
}

.fin-title {
  margin: 0;
  font-size: clamp(1.4rem, 1.1rem + 1.2vw, 2rem);
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text);
}

.fin-refresh {
  min-height: var(--touch-min, 44px);
  padding: 0 18px;
  border: 1px solid var(--border);
  border-radius: 9999px;
  background: var(--surface);
  color: var(--text);
  font-weight: 600;
  font-size: 0.88rem;
  cursor: pointer;
}

.fin-refresh:disabled { opacity: 0.6; cursor: default; }

.fin-skeleton, .fin-vacio {
  padding: 32px;
  text-align: center;
  color: var(--muted);
  font-size: 0.92rem;
}

.fin-aviso {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  padding: 14px 16px;
  border: 1px solid var(--warning);
  border-radius: 14px;
  background: var(--warning-bg);
  color: var(--dark);
  font-size: 0.88rem;
  line-height: 1.55;
}

.fin-aviso > div { flex: 1 1 320px; }

.fin-aviso-btn {
  min-height: var(--touch-min, 44px);
  padding: 0 18px;
  border: none;
  border-radius: 9999px;
  background: var(--warning);
  color: #1d1d1f;
  font-weight: 700;
  font-size: 0.85rem;
  cursor: pointer;
}

.fin-aviso-btn:disabled { opacity: 0.6; cursor: default; }

/* ── KPIs ─────────────────────────────────────────────────────────────── */
.fin-kpis {
  display: grid;
  /* auto-fit y no tres columnas fijas: en tableta caben dos y en móvil una,
   * sin necesidad de una consulta de medios por cada salto. */
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 16px;
}

.fin-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px;
  border: 1px solid var(--border);
  border-radius: 18px;
  background: var(--surface);
  box-shadow: var(--shadow-sm);
}

.fin-card-txt { min-width: 0; }

.fin-card-label {
  margin: 0 0 6px;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
}

.fin-card-valor {
  margin: 0 0 6px;
  font-size: clamp(1.5rem, 1.2rem + 1.4vw, 2.1rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  color: var(--text);
  /* Un importe largo no debe ensanchar la tarjeta ni sacar scroll lateral. */
  overflow-wrap: anywhere;
}

.fin-card-valor.alerta { color: var(--fin-naranja); }
.fin-card-valor.positivo { color: var(--fin-lima); }

.fin-card-pie { margin: 0; font-size: 0.78rem; color: var(--muted); }
.fin-pie-aviso { color: var(--warning); font-weight: 700; }

/* Anillo de progreso con conic-gradient: sin SVG ni librería, y el porcentaje
 * entra por una variable CSS. */
.fin-anillo {
  position: relative;
  flex-shrink: 0;
  width: 84px;
  height: 84px;
  border-radius: 50%;
  background: conic-gradient(
    var(--tono) calc(var(--pct) * 1%),
    color-mix(in srgb, var(--tono) 14%, transparent) 0
  );
  display: grid;
  place-items: center;
}

/* El hueco central se hace con una máscara y no con un círculo encima, para
 * que el fondo de la tarjeta se vea a través en cualquier tema. */
.fin-anillo::before {
  content: '';
  position: absolute;
  inset: 10px;
  border-radius: 50%;
  background: var(--surface);
}

.fin-anillo span {
  position: relative;
  font-size: 0.9rem;
  font-weight: 800;
  color: var(--tono);
}

/* ── Paneles ──────────────────────────────────────────────────────────── */
.fin-panel {
  padding: 20px;
  border: 1px solid var(--border);
  border-radius: 18px;
  background: var(--surface);
  box-shadow: var(--shadow-sm);
}

.fin-panel-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.fin-panel-head h2 { margin: 0; font-size: 1rem; font-weight: 700; color: var(--text); }
.fin-panel-nota { font-size: 0.78rem; color: var(--muted); }

.fin-grafico { height: 320px; }

/* ── Movimientos ──────────────────────────────────────────────────────── */
.fin-movs { list-style: none; margin: 0; padding: 0; }

.fin-mov {
  display: grid;
  grid-template-columns: 44px minmax(0, 1.6fr) minmax(0, 1fr) auto;
  align-items: center;
  gap: 14px;
  padding: 14px 0;
  border-bottom: 1px solid var(--border-light);
}

.fin-mov:last-child { border-bottom: none; }

.fin-mov-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: color-mix(in srgb, var(--fin-cyan) 16%, transparent);
  color: var(--fin-cyan);
  font-weight: 800;
  font-size: 0.82rem;
}

.fin-mov-quien { min-width: 0; }

.fin-mov-nombre {
  margin: 0 0 2px;
  font-weight: 700;
  font-size: 0.92rem;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fin-mov-meta, .fin-mov-fecha {
  margin: 0;
  font-size: 0.78rem;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fin-mov-importe { text-align: right; }

.fin-mov-neto {
  margin: 0 0 2px;
  font-size: 1.05rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text);
  white-space: nowrap;
}

.fin-mov-bruto { margin: 0; font-size: 0.74rem; color: var(--muted); white-space: nowrap; }

/* En móvil la rejilla de cuatro columnas deja el importe sin sitio: pasa a dos
 * filas con el importe alineado bajo el nombre. */
@media (max-width: 639px) {
  .fin-mov {
    grid-template-columns: 40px minmax(0, 1fr) auto;
    row-gap: 4px;
  }
  .fin-mov-avatar { width: 40px; height: 40px; }
  .fin-mov-fecha { grid-column: 2 / -1; }
}
</style>
