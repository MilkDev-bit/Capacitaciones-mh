<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useSuscripcionStore, precioDesdeCentavos } from '../../stores/suscripcion'
import { toast } from '../../utils/toast'
import logoSrc from '../../assets/logo-capacitaciones.png'

const router = useRouter()
const auth = useAuthStore()
const susc = useSuscripcionStore()

/** Mensual / anual: el mismo catálogo de planes filtrado por intervalo. */
const intervalo = ref<'mes' | 'anio'>('mes')
const asientos = ref(5)
const procesando = ref('')

onMounted(() => {
  susc.cargarPlanes()
  if (auth.isLoggedIn) susc.cargarMia()
})

const individuales = computed(() =>
  susc.planesIndividuales.filter((p) => p.intervalo === intervalo.value)
)
const equipo = computed(() => susc.planesEquipo.filter((p) => p.intervalo === intervalo.value))

const hayAnual = computed(() => susc.planes.some((p) => p.intervalo === 'anio'))

/** Ahorro del anual frente a pagar 12 meses del plan equivalente. */
function ahorroAnual(modalidad: string) {
  const mensual = susc.planes.find((p) => p.modalidad === modalidad && p.intervalo === 'mes')
  const anual = susc.planes.find((p) => p.modalidad === modalidad && p.intervalo === 'anio')
  if (!mensual || !anual) return 0
  const doceMeses = mensual.precio_centavos * 12
  if (doceMeses <= anual.precio_centavos) return 0
  return Math.round((100 * (doceMeses - anual.precio_centavos)) / doceMeses)
}

async function suscribirse(codigo: string, modalidad: string) {
  if (!auth.isLoggedIn) {
    // Se guarda la intención para retomarla después de iniciar sesión.
    router.push(`/login?redirect=${encodeURIComponent('/planes')}`)
    return
  }
  if (susc.tieneSuscripcion) {
    toast.info('Ya tienes una suscripción. Cámbiala desde el portal de facturación.')
    return
  }
  procesando.value = codigo
  try {
    await susc.checkout(codigo, modalidad === 'asientos' ? asientos.value : 1)
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'No pudimos abrir el pago. Intenta de nuevo.')
    procesando.value = ''
  }
}

const BENEFICIOS_INDIVIDUAL = [
  'Acceso a todo el catálogo mientras tu plan esté activo',
  'Cursos nuevos incluidos sin costo adicional',
  'Constancia DC-3 al completar cada curso',
  'Cancela cuando quieras desde tu panel',
]

const BENEFICIOS_EQUIPO = [
  'Un lugar por colaborador, con acceso a todo el catálogo',
  'Reparte los accesos por correo en un clic',
  'Reasigna lugares cuando alguien deja el equipo',
  'Facturación mensual o anual con recibo automático',
]
</script>

<template>
  <div class="planes">
    <header class="ph">
      <button class="ph__brand" @click="router.push('/')" aria-label="Ir al inicio">
        <img :src="logoSrc" alt="MH Capacitaciones" />
      </button>
      <button class="ph__back" @click="router.push('/')">← Volver a la tienda</button>
    </header>

    <section class="hero">
      <h1 v-reveal>Dos formas de capacitarte</h1>
      <p class="hero__sub" v-reveal="1">
        Suscríbete y abre <strong>todo el catálogo</strong> mientras tu plan esté activo,
        o compra <strong>solo los cursos que necesitas</strong> y consérvalos para siempre.
      </p>

      <div v-if="susc.tieneSuscripcion" class="aviso" :class="susc.enPeriodoDeGracia && 'aviso--alerta'">
        <strong>{{ susc.mia?.plan_nombre }}</strong>
        <span v-if="susc.enPeriodoDeGracia">
          — tu último cobro falló. Actualiza tu tarjeta para no perder el acceso.
        </span>
        <span v-else>— tu plan está activo, ya tienes acceso a todo el catálogo.</span>
        <button class="aviso__btn" @click="susc.abrirPortal()">Gestionar suscripción</button>
      </div>

      <div v-if="hayAnual" v-reveal="2" class="switch" role="tablist" aria-label="Periodicidad">
        <button
          role="tab"
          :aria-selected="intervalo === 'mes'"
          :class="['switch__opt', intervalo === 'mes' && 'is-on']"
          @click="intervalo = 'mes'"
        >
          Mensual
        </button>
        <button
          role="tab"
          :aria-selected="intervalo === 'anio'"
          :class="['switch__opt', intervalo === 'anio' && 'is-on']"
          @click="intervalo = 'anio'"
        >
          Anual
          <span v-if="ahorroAnual('individual')" class="switch__tag">−{{ ahorroAnual('individual') }}%</span>
        </button>
      </div>
    </section>

    <!-- ── Planes ─────────────────────────────────────────── -->
    <section class="grid" v-if="!susc.cargandoPlanes">
      <article v-for="p in individuales" :key="p.id" class="card card--destacada" v-reveal>
        <span class="card__flag">Para ti</span>
        <h2 class="card__nombre">{{ p.nombre }}</h2>
        <p class="card__desc">{{ p.descripcion || 'Acceso completo al catálogo, a tu ritmo.' }}</p>
        <p class="card__precio">
          {{ precioDesdeCentavos(p.precio_centavos, p.moneda) }}
          <small>/{{ p.intervalo === 'anio' ? 'año' : 'mes' }}</small>
        </p>
        <p v-if="p.dias_prueba" class="card__prueba">{{ p.dias_prueba }} días de prueba gratis</p>
        <ul class="card__lista">
          <li v-for="b in BENEFICIOS_INDIVIDUAL" :key="b">{{ b }}</li>
        </ul>
        <button
          class="card__cta"
          :disabled="procesando === p.codigo || susc.tieneSuscripcion"
          @click="suscribirse(p.codigo, p.modalidad)"
        >
          <span v-if="procesando === p.codigo">Abriendo pago…</span>
          <span v-else-if="susc.tieneSuscripcion">Ya tienes un plan</span>
          <span v-else>Suscribirme</span>
        </button>
      </article>

      <article v-for="p in equipo" :key="p.id" class="card" v-reveal="1">
        <span class="card__flag card__flag--alt">Para tu equipo</span>
        <h2 class="card__nombre">{{ p.nombre }}</h2>
        <p class="card__desc">{{ p.descripcion || 'Un lugar por colaborador, con acceso a todo.' }}</p>
        <p class="card__precio">
          {{ precioDesdeCentavos(p.precio_centavos, p.moneda) }}
          <small>/lugar /{{ p.intervalo === 'anio' ? 'año' : 'mes' }}</small>
        </p>

        <label class="asientos">
          <span>¿Cuántos lugares?</span>
          <input v-model.number="asientos" type="number" min="1" max="500" inputmode="numeric" />
        </label>
        <p class="asientos__total">
          Total: <strong>{{ precioDesdeCentavos(p.precio_centavos * Math.max(1, asientos), p.moneda) }}</strong>
          /{{ p.intervalo === 'anio' ? 'año' : 'mes' }}
        </p>

        <ul class="card__lista">
          <li v-for="b in BENEFICIOS_EQUIPO" :key="b">{{ b }}</li>
        </ul>
        <button
          class="card__cta card__cta--alt"
          :disabled="procesando === p.codigo || susc.tieneSuscripcion"
          @click="suscribirse(p.codigo, p.modalidad)"
        >
          <span v-if="procesando === p.codigo">Abriendo pago…</span>
          <span v-else-if="susc.tieneSuscripcion">Ya tienes un plan</span>
          <span v-else>Contratar {{ Math.max(1, asientos) }} lugares</span>
        </button>
      </article>

      <!-- Tercera columna: la compra individual, con el mismo peso visual -->
      <article class="card card--suelta" v-reveal="2">
        <span class="card__flag card__flag--neutro">Sin suscripción</span>
        <h2 class="card__nombre">Compra por curso</h2>
        <p class="card__desc">
          ¿Solo necesitas una capacitación puntual? Cómprala suelta y es tuya para siempre,
          sin pagos recurrentes.
        </p>
        <p class="card__precio card__precio--var">Desde el precio de cada curso</p>
        <ul class="card__lista">
          <li>Pago único, acceso de por vida a ese curso</li>
          <li>Constancia DC-3 igual que en la suscripción</li>
          <li>También en versión corporativa por licencias</li>
          <li>Sin tarjeta guardada ni renovaciones</li>
        </ul>
        <button class="card__cta card__cta--fantasma" @click="router.push('/tienda')">
          Ver catálogo y precios
        </button>
      </article>
    </section>

    <div v-else class="cargando">Cargando planes…</div>

    <p v-if="!susc.cargandoPlanes && !susc.planes.length" class="vacio">
      Todavía no hay planes de suscripción disponibles. Mientras tanto, puedes
      <button class="enlace" @click="router.push('/tienda')">comprar cursos por separado</button>.
    </p>

    <!-- ── Comparativa ────────────────────────────────────── -->
    <section class="comparativa">
      <h2 v-reveal>¿Cuál me conviene?</h2>
      <div class="comparativa__tabla" role="table" v-reveal="1">
        <div class="fila fila--head" role="row">
          <span role="columnheader"></span>
          <span role="columnheader">Suscripción</span>
          <span role="columnheader">Compra individual</span>
        </div>
        <div class="fila" role="row">
          <span role="rowheader">Qué incluye</span>
          <span>Todo el catálogo</span>
          <span>Solo los cursos que compras</span>
        </div>
        <div class="fila" role="row">
          <span role="rowheader">Cuánto dura</span>
          <span>Mientras el plan esté activo</span>
          <span>Para siempre</span>
        </div>
        <div class="fila" role="row">
          <span role="rowheader">Cursos nuevos</span>
          <span>Incluidos sin costo</span>
          <span>Se compran aparte</span>
        </div>
        <div class="fila" role="row">
          <span role="rowheader">Constancia DC-3</span>
          <span>Sí</span>
          <span>Sí</span>
        </div>
        <div class="fila" role="row">
          <span role="rowheader">Ideal si…</span>
          <span>Te capacitas de forma continua</span>
          <span>Necesitas un tema puntual</span>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.planes {
  min-height: 100vh;
  background: var(--bg);
  color: var(--text);
  padding-bottom: 5rem;
}

/* ── Cabecera ─────────────────────────────────────────── */
.ph {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem clamp(1rem, 5vw, 3rem);
  border-bottom: 1px solid var(--border);
}
.ph__brand {
  background: none;
  border: 0;
  cursor: pointer;
  padding: 0;
}
.ph__brand img { height: 38px; width: auto; display: block; }
.ph__back {
  background: none;
  border: 0;
  color: var(--muted);
  font-size: 0.92rem;
  cursor: pointer;
}
.ph__back:hover { color: var(--brand); }

/* ── Hero ─────────────────────────────────────────────── */
.hero {
  max-width: 760px;
  margin: 0 auto;
  padding: clamp(2.5rem, 7vw, 4.5rem) 1.25rem 2rem;
  text-align: center;
}
.hero h1 {
  font-size: clamp(2rem, 5vw, 3rem);
  font-weight: 800;
  letter-spacing: -0.035em;
  margin: 0 0 0.9rem;
}
.hero__sub {
  font-size: clamp(1rem, 2.2vw, 1.15rem);
  color: var(--muted);
  line-height: 1.65;
  margin: 0;
}
.hero__sub strong { color: var(--text); font-weight: 650; }

.aviso {
  margin-top: 1.5rem;
  padding: 0.85rem 1.1rem;
  border-radius: var(--r);
  background: var(--brand-light);
  border: 1px solid var(--brand-border);
  font-size: 0.93rem;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}
.aviso--alerta { background: #fef3c7; border-color: #fcd34d; color: #78350f; }
.aviso__btn {
  background: var(--brand);
  color: #fff;
  border: 0;
  border-radius: 999px;
  padding: 0.4rem 0.9rem;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
}

.switch {
  display: inline-flex;
  margin-top: 1.8rem;
  padding: 4px;
  border-radius: 999px;
  background: var(--surface-soft);
  border: 1px solid var(--border);
}
.switch__opt {
  border: 0;
  background: none;
  padding: 0.5rem 1.2rem;
  border-radius: 999px;
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--muted);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}
.switch__opt.is-on { background: var(--surface); color: var(--text); box-shadow: var(--shadow-sm); }
.switch__tag {
  background: #10b981;
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 999px;
}

/* ── Tarjetas ─────────────────────────────────────────── */
.grid {
  max-width: 1180px;
  margin: 0 auto;
  padding: 1rem clamp(1rem, 5vw, 3rem);
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(290px, 1fr));
  gap: 1.4rem;
  align-items: start;
}
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  padding: 1.8rem 1.5rem 1.5rem;
  transition: transform 0.3s var(--ease-apple), box-shadow 0.3s var(--ease-apple);
}
.card:hover { transform: translateY(-4px); box-shadow: var(--shadow-md); }
.card--destacada { border-color: var(--brand-border); box-shadow: var(--shadow-sm); }
.card--suelta { background: var(--surface-soft); }

.card__flag {
  position: absolute;
  top: -0.7rem;
  left: 1.5rem;
  background: var(--brand);
  color: #fff;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding: 0.25rem 0.7rem;
  border-radius: 999px;
}
.card__flag--alt { background: #6366f1; }
.card__flag--neutro { background: var(--muted); }

.card__nombre { font-size: 1.3rem; font-weight: 750; margin: 0.4rem 0 0.4rem; letter-spacing: -0.02em; }
.card__desc { color: var(--muted); font-size: 0.93rem; line-height: 1.55; margin: 0 0 1.1rem; }
.card__precio { font-size: 2rem; font-weight: 800; letter-spacing: -0.03em; margin: 0 0 0.3rem; }
.card__precio small { font-size: 0.9rem; font-weight: 500; color: var(--muted); }
.card__precio--var { font-size: 1.1rem; font-weight: 600; color: var(--muted); }
.card__prueba { font-size: 0.85rem; color: #10b981; font-weight: 600; margin: 0 0 0.6rem; }

.card__lista { list-style: none; padding: 0; margin: 1rem 0 1.5rem; display: grid; gap: 0.6rem; }
.card__lista li {
  position: relative;
  padding-left: 1.5rem;
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--muted);
}
.card__lista li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0.42em;
  width: 0.85rem;
  height: 0.45rem;
  border-left: 2px solid var(--brand);
  border-bottom: 2px solid var(--brand);
  transform: rotate(-45deg);
}

.asientos {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  font-size: 0.9rem;
  color: var(--muted);
  margin-top: 0.6rem;
}
.asientos input {
  width: 84px;
  padding: 0.45rem 0.6rem;
  border-radius: var(--r-sm);
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  font-size: 0.95rem;
  text-align: center;
}
.asientos__total { font-size: 0.9rem; color: var(--muted); margin: 0.5rem 0 0; }

.card__cta {
  margin-top: auto;
  width: 100%;
  padding: 0.8rem 1rem;
  border: 0;
  border-radius: var(--r);
  background: var(--brand);
  color: #fff;
  font-size: 0.98rem;
  font-weight: 650;
  cursor: pointer;
  transition: filter 0.2s;
}
.card__cta:hover:not(:disabled) { filter: brightness(1.08); }
.card__cta:disabled { opacity: 0.55; cursor: not-allowed; }
.card__cta--alt { background: #6366f1; }
.card__cta--fantasma {
  background: transparent;
  color: var(--text);
  border: 1px solid var(--border);
}
.card__cta--fantasma:hover { border-color: var(--brand); color: var(--brand); }

.cargando,
.vacio { text-align: center; color: var(--muted); padding: 3rem 1rem; }
.enlace { background: none; border: 0; color: var(--brand); cursor: pointer; text-decoration: underline; font: inherit; }

/* ── Comparativa ──────────────────────────────────────── */
.comparativa { max-width: 900px; margin: 4rem auto 0; padding: 0 clamp(1rem, 5vw, 3rem); }
.comparativa h2 {
  text-align: center;
  font-size: clamp(1.4rem, 3vw, 1.9rem);
  font-weight: 750;
  letter-spacing: -0.025em;
  margin: 0 0 1.6rem;
}
.comparativa__tabla {
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  overflow: hidden;
  background: var(--surface);
}
.fila {
  display: grid;
  grid-template-columns: 1.1fr 1fr 1fr;
  gap: 0.75rem;
  padding: 0.9rem 1.1rem;
  font-size: 0.92rem;
  border-top: 1px solid var(--border);
}
.fila:first-child { border-top: 0; }
.fila--head { background: var(--surface-soft); font-weight: 700; }
.fila > span:first-child { color: var(--muted); font-weight: 600; }

@media (max-width: 620px) {
  .fila { grid-template-columns: 1fr; gap: 0.25rem; }
  .fila > span:first-child { margin-top: 0.2rem; }
  .fila--head { display: none; }
  /* Sin encabezado de columna, cada celda dice a qué modalidad pertenece. */
  .fila > span:nth-child(2)::before { content: 'Suscripción: '; color: var(--muted); }
  .fila > span:nth-child(3)::before { content: 'Compra individual: '; color: var(--muted); }
}
</style>
