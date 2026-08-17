<script setup lang="ts">
/**
 * Verificación pública de una constancia DC-3 por folio.
 *
 * Página SIN sesión: la usa quien recibe una constancia en papel —un patrón, un
 * inspector, un cliente— y no tiene cuenta aquí ni motivo para crearse una.
 *
 * Existe porque el PDF por sí solo no impide falsificar: se puede editar y el
 * diseño se puede copiar. Lo que hace detectable un documento inventado es que
 * su folio no aparezca aquí.
 *
 * Muestra a nombre de quién, de qué curso, de qué empresa y de cuándo. Nada más:
 * quien teclea un folio que no es suyo no debe llevarse la CURP ni el RFC de esa
 * persona, ni un enlace de descarga al documento.
 */
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../../api'

type Resultado = {
  valida: boolean
  nombre_trabajador?: string
  capacitacion_titulo?: string
  razon_social?: string
  generada_at?: string
}

const route = useRoute()
const router = useRouter()

const folio = ref(String(route.params.folio || ''))
const resultado = ref<Resultado | null>(null)
const cargando = ref(false)
const error = ref('')

function formatearFecha(valor?: string) {
  if (!valor) return ''
  const d = new Date(valor + 'T00:00:00')
  if (Number.isNaN(d.getTime())) return valor
  return d.toLocaleDateString('es-MX', { day: '2-digit', month: 'long', year: 'numeric' })
}

async function verificar() {
  const codigo = folio.value.trim().toUpperCase()
  if (!codigo) return
  cargando.value = true
  error.value = ''
  resultado.value = null
  try {
    const { data } = await api.get(`/constancias/verificar/${encodeURIComponent(codigo)}`)
    resultado.value = data
    // El folio se refleja en la URL para que el resultado sea compartible y
    // para que recargar no borre la consulta.
    if (route.params.folio !== codigo) {
      router.replace({ path: `/verificar/${codigo}` })
    }
  } catch {
    // Se distingue del folio inexistente: "no existe" y "no pudimos consultar"
    // llevan a quien verifica a conclusiones opuestas sobre el documento.
    error.value = 'No pudimos completar la consulta. Inténtalo de nuevo en un momento.'
  } finally {
    cargando.value = false
  }
}

onMounted(() => { if (folio.value) verificar() })
</script>

<template>
  <div class="verificar-page">
    <div class="verificar-caja">
      <header class="verificar-head">
        <h1>Verificar constancia DC-3</h1>
        <p>
          Escribe el folio impreso al pie del documento para comprobar que fue
          emitido por MH Soluciones Empresariales.
        </p>
      </header>

      <form class="verificar-form" @submit.prevent="verificar">
        <label class="sr-only" for="folio">Folio</label>
        <input id="folio" v-model="folio" class="field-input"
               placeholder="MH-XXXX-XXXX-XXXX" autocapitalize="characters"
               spellcheck="false" />
        <button class="btn btn-primary" type="submit" :disabled="cargando || !folio.trim()">
          {{ cargando ? 'Consultando…' : 'Verificar' }}
        </button>
      </form>

      <p v-if="error" class="verificar-error">{{ error }}</p>

      <!-- Válida -->
      <div v-else-if="resultado?.valida" class="verificar-resultado valida">
        <div class="verificar-icono" aria-hidden="true">
          <svg width="26" height="26" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
            <path d="M20 6L9 17l-5-5" />
          </svg>
        </div>
        <div>
          <h2>Constancia válida</h2>
          <dl class="verificar-datos">
            <div><dt>Trabajador</dt><dd>{{ resultado.nombre_trabajador || '—' }}</dd></div>
            <div><dt>Capacitación</dt><dd>{{ resultado.capacitacion_titulo || '—' }}</dd></div>
            <div><dt>Empresa</dt><dd>{{ resultado.razon_social || '—' }}</dd></div>
            <div><dt>Fecha de emisión</dt><dd>{{ formatearFecha(resultado.generada_at) || '—' }}</dd></div>
          </dl>
          <p class="verificar-nota">
            Compara estos datos con los del documento que tienes. Si no coinciden,
            el documento no corresponde a este folio.
          </p>
        </div>
      </div>

      <!-- No encontrada -->
      <div v-else-if="resultado && !resultado.valida" class="verificar-resultado invalida">
        <div class="verificar-icono" aria-hidden="true">
          <svg width="26" height="26" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </div>
        <div>
          <h2>No encontramos ese folio</h2>
          <!--
            No se afirma que el documento sea falso: un folio mal tecleado da el
            mismo resultado, y acusar a alguien por una letra cambiada sería peor
            que no decir nada.
          -->
          <p class="verificar-nota">
            Revisa que lo hayas escrito tal cual aparece en el documento. Si el
            folio es correcto y aun así no aparece, esa constancia no fue emitida
            por nosotros.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.verificar-page {
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--gutter, 16px);
  background: var(--bg);
}

.verificar-caja {
  width: 100%;
  max-width: 560px;
  padding: clamp(20px, 4vw, 32px);
  border: 1px solid var(--border);
  border-radius: 14px;
  background: var(--surface);
}

.verificar-head h1 {
  margin: 0 0 8px;
  font-size: clamp(1.25rem, 1.1rem + 1vw, 1.6rem);
  color: var(--text);
}

.verificar-head p {
  margin: 0 0 20px;
  color: var(--muted);
  font-size: 0.9rem;
  line-height: 1.55;
}

.verificar-form {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.verificar-form .field-input {
  flex: 1 1 240px;
  min-height: var(--touch-min);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.verificar-form .btn { min-height: var(--touch-min); }

@media (max-width: 479px) {
  .verificar-form > * { flex: 1 1 100%; }
}

.verificar-error {
  margin: 18px 0 0;
  color: var(--danger);
  font-size: 0.88rem;
}

.verificar-resultado {
  display: flex;
  gap: 14px;
  margin-top: 22px;
  padding: 18px;
  border-radius: 12px;
  border: 1px solid var(--border);
}

.verificar-resultado.valida {
  border-color: color-mix(in srgb, #10b981 45%, transparent);
  background: color-mix(in srgb, #10b981 8%, transparent);
}

.verificar-resultado.invalida {
  border-color: color-mix(in srgb, var(--danger) 45%, transparent);
  background: color-mix(in srgb, var(--danger) 8%, transparent);
}

.verificar-icono {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border-radius: 999px;
}

.valida .verificar-icono { background: #10b981; color: #fff; }
.invalida .verificar-icono { background: var(--danger); color: #fff; }

.verificar-resultado h2 {
  margin: 0 0 10px;
  font-size: 1rem;
  color: var(--text);
}

.verificar-datos { margin: 0; }

.verificar-datos > div {
  display: grid;
  grid-template-columns: 130px 1fr;
  gap: 8px;
  padding: 5px 0;
}

.verificar-datos dt {
  font-size: 0.74rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--muted);
}

.verificar-datos dd {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text);
  overflow-wrap: anywhere;
}

/* En móvil la etiqueta encima: 130px de columna fija dejan el valor en un
   canal de dos palabras por línea. */
@media (max-width: 479px) {
  .verificar-datos > div { grid-template-columns: 1fr; gap: 2px; }
}

.verificar-nota {
  margin: 12px 0 0;
  font-size: 0.82rem;
  color: var(--muted);
  line-height: 1.55;
}

.sr-only {
  position: absolute;
  width: 1px; height: 1px;
  padding: 0; margin: -1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
  white-space: nowrap;
  border: 0;
}
</style>
