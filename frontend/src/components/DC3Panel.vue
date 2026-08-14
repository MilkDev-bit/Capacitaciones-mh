<script setup lang="ts">
/**
 * Panel de la constancia DC-3 dentro de una capacitación.
 *
 * Tiene cuatro estados y cada uno responde a una pregunta distinta del alumno:
 *
 *   lista        → "ya está, descárgala"
 *   faltan-mios  → "faltan TUS datos, captúralos" (el único accionable por él)
 *   falta-empresa→ "faltan los del instructor, no puedes hacer nada"
 *   no-aplica    → el curso no emite constancia, no se muestra nada
 *
 * Distinguir los dos "faltan" importa: mandar al alumno a rellenar un formulario
 * que no va a desbloquear nada es peor que decirle a quién reclamar.
 */
import { ref, computed, onMounted, watch } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'

const props = defineProps<{
  cursoId: string
  /** El panel solo aparece cuando el curso está terminado. */
  completado: boolean
  /** dc3_enabled de la capacitación. */
  habilitado?: boolean
}>()

const cargando = ref(false)
const guardando = ref(false)
const constanciaUrl = ref('')
const trabajadorCompleto = ref(false)
const empresaCompleta = ref(false)
const consultado = ref(false)

const form = ref({
  curp: '', puesto: '', ocupacion_especifica: '',
  razon_social: '', rfc: '', nombre_patron: '', representante_trabajadores: '',
})

/** Empresa a la que saldrá la constancia si el alumno no declara la suya. */
const empresaRespaldo = ref('')
const tieneEmpresaPropia = ref(false)

/**
 * El bloque de empresa se acepta entero o vacío.
 *
 * Mezclar la razón social del alumno con el representante del instructor
 * produciría un documento que no corresponde a ninguna empresa real, así que el
 * backend lo rechaza y aquí se avisa antes de enviarlo.
 */
const camposEmpresa = computed(() => [
  form.value.razon_social, form.value.rfc,
  form.value.nombre_patron, form.value.representante_trabajadores,
].map(v => v.trim()))

// Ojo con el nombre: `empresaCompleta` (más abajo) es la bandera del backend
// sobre la empresa YA RESUELTA. Estas dos miran solo lo que el alumno teclea.
const empresaPropiaCompleta = computed(() => camposEmpresa.value.every(v => v !== ''))
const empresaPropiaVacia = computed(() => camposEmpresa.value.every(v => v === ''))
const empresaAMedias = computed(() => !empresaPropiaCompleta.value && !empresaPropiaVacia.value)

const estado = computed(() => {
  if (!consultado.value) return 'cargando'
  if (constanciaUrl.value) return 'lista'
  if (!trabajadorCompleto.value) return 'faltan-mios'
  if (!empresaCompleta.value) return 'falta-empresa'
  // Datos completos pero sin documento: suele ser un curso sin duración
  // registrada, que el backend rechaza al armar el documento.
  return 'falta-empresa'
})

const curpValida = computed(() => form.value.curp.trim().length === 18)

async function consultar() {
  if (!props.completado || props.habilitado === false) return
  cargando.value = true
  try {
    const { data } = await api.get(`/capacitaciones/${props.cursoId}/dc3`)
    constanciaUrl.value = data.constancia_url || ''
    trabajadorCompleto.value = !!data.trabajador_completo
    empresaCompleta.value = !!data.empresa_completa
    // De dónde saldrá la empresa y cuál es, para que el alumno sepa a nombre de
    // quién va su constancia antes de decidir si declara la suya.
    tieneEmpresaPropia.value = data.empresa_origen === 'alumno'
    empresaRespaldo.value = data.empresa?.razon_social || ''

    if (data.trabajador) {
      // Se precargan para que corregir un dato no obligue a teclear todo.
      form.value = {
        ...form.value,
        curp: data.trabajador.curp || '',
        puesto: data.trabajador.puesto || '',
        ocupacion_especifica: data.trabajador.ocupacion_especifica || '',
      }
    }
    if (tieneEmpresaPropia.value && data.empresa) {
      form.value.razon_social = data.empresa.razon_social || ''
      form.value.rfc = data.empresa.rfc || ''
      form.value.nombre_patron = data.empresa.nombre_patron || ''
      form.value.representante_trabajadores = data.empresa.representante_trabajadores || ''
    }
  } catch {
    // Un curso sin DC-3 responde con error; no es algo que reportar al alumno.
  } finally {
    cargando.value = false
    consultado.value = true
  }
}

async function enviar() {
  if (!curpValida.value) {
    toast.error('La CURP debe tener 18 caracteres')
    return
  }
  if (empresaAMedias.value) {
    toast.error('Completa los cuatro datos de tu empresa o déjalos todos vacíos')
    return
  }
  guardando.value = true
  try {
    const { data, status } = await api.post(`/capacitaciones/${props.cursoId}/dc3`, {
      curp: form.value.curp.trim().toUpperCase(),
      puesto: form.value.puesto.trim(),
      ocupacion_especifica: form.value.ocupacion_especifica.trim(),
      razon_social: form.value.razon_social.trim(),
      rfc: form.value.rfc.trim().toUpperCase(),
      nombre_patron: form.value.nombre_patron.trim(),
      representante_trabajadores: form.value.representante_trabajadores.trim(),
    })
    trabajadorCompleto.value = true
    // 202 significa "guardado, pero falta el instructor". No es un error y no
    // debe pintarse en rojo.
    if (status === 202) {
      empresaCompleta.value = false
      toast.success(data.mensaje || 'Guardamos tus datos')
      return
    }
    constanciaUrl.value = data.constancia_url || ''
    toast.success('Tu constancia está lista')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No se pudieron guardar tus datos')
  } finally {
    guardando.value = false
  }
}

onMounted(consultar)
watch(() => props.completado, (val) => { if (val) consultar() })
</script>

<template>
  <div v-if="completado && habilitado !== false && consultado" class="dc3-panel">
    <div class="dc3-head">
      <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" /><path d="M14 2v6h6" />
      </svg>
      <strong>Constancia DC-3</strong>
    </div>

    <!-- Lista -->
    <template v-if="estado === 'lista'">
      <p class="dc3-lead">Tu constancia de habilidades laborales está emitida.</p>
      <a :href="constanciaUrl" target="_blank" rel="noopener" class="btn btn-primary" download>
        Descargar constancia
      </a>
    </template>

    <!-- Faltan datos del alumno -->
    <template v-else-if="estado === 'faltan-mios'">
      <p class="dc3-lead">
        Completa estos datos y emitimos tu constancia. Solo se piden una vez:
        los reutilizamos en tus siguientes cursos.
      </p>
      <div class="dc3-form">
        <label class="dc3-field">
          <span>CURP <em>*</em></span>
          <input v-model="form.curp" maxlength="18" placeholder="18 caracteres"
                 class="field-input" :class="{ invalido: form.curp && !curpValida }" />
          <small v-if="form.curp && !curpValida" class="dc3-error">
            Llevas {{ form.curp.trim().length }} de 18 caracteres
          </small>
        </label>
        <label class="dc3-field">
          <span>Puesto <em>*</em></span>
          <input v-model="form.puesto" class="field-input" placeholder="Ej: Supervisor" />
        </label>
        <label class="dc3-field dc3-field-full">
          <span>Ocupación específica <em>*</em></span>
          <input v-model="form.ocupacion_especifica" class="field-input"
                 placeholder="Ej: 04.6 Supervisores en la construcción" />
          <small class="dc3-hint">Clave y nombre del Catálogo Nacional de Ocupaciones.</small>
        </label>
      </div>

      <!--
        Empresa del alumno: opcional.
        Legalmente el patrón es quien lo emplea, así que si trabaja para una
        empresa debe ir la suya. Quien se capacita por su cuenta lo deja vacío y
        recibe la constancia a nombre de quien la imparte.
      -->
      <div class="dc3-empresa-bloque">
        <div class="dc3-sub">
          <strong>¿Trabajas para una empresa?</strong>
          <p v-if="empresaRespaldo">
            Si lo dejas vacío, tu constancia saldrá a nombre de
            <b>{{ empresaRespaldo }}</b>.
          </p>
          <p v-else>
            Si lo dejas vacío, tu constancia saldrá a nombre de quien imparte la
            capacitación.
          </p>
        </div>
        <div class="dc3-form">
          <label class="dc3-field">
            <span>Razón social</span>
            <input v-model="form.razon_social" class="field-input" />
          </label>
          <label class="dc3-field">
            <span>RFC</span>
            <input v-model="form.rfc" maxlength="13" class="field-input" />
          </label>
          <label class="dc3-field">
            <span>Patrón o representante legal</span>
            <input v-model="form.nombre_patron" class="field-input" />
          </label>
          <label class="dc3-field">
            <span>Representante de los trabajadores</span>
            <input v-model="form.representante_trabajadores" class="field-input" />
          </label>
        </div>
        <small v-if="empresaAMedias" class="dc3-error">
          Completa los cuatro datos o déjalos todos vacíos: una constancia con
          media empresa no corresponde a ninguna entidad real.
        </small>
      </div>

      <button class="btn btn-primary"
              :disabled="guardando || !curpValida || empresaAMedias" @click="enviar">
        {{ guardando ? 'Emitiendo…' : 'Emitir mi constancia' }}
      </button>
    </template>

    <!-- Falta que el instructor configure la empresa -->
    <template v-else>
      <p class="dc3-lead">
        Tus datos están guardados. Falta que el instructor complete los datos de la
        empresa y los agentes capacitadores de esta capacitación; en cuanto lo haga,
        emitimos tu constancia automáticamente.
      </p>
    </template>
  </div>
</template>

<style scoped>
.dc3-panel {
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 20px;
  background: var(--surface);
  margin-top: 20px;
}

.dc3-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  color: var(--brand);
}

.dc3-head strong { color: var(--text); font-size: 1rem; }

.dc3-lead {
  margin: 0 0 16px;
  color: var(--muted);
  font-size: 0.9rem;
  line-height: 1.55;
}

.dc3-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr));
  gap: 14px;
  margin-bottom: 16px;
}

.dc3-field { display: flex; flex-direction: column; gap: 6px; }
.dc3-field-full { grid-column: 1 / -1; }

.dc3-field > span {
  font-size: 0.78rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--muted);
}

.dc3-field em { color: var(--danger); font-style: normal; }

.field-input.invalido { border-color: var(--danger); }

.dc3-hint, .dc3-error { font-size: 0.75rem; }
.dc3-hint { color: var(--muted); }
.dc3-error { color: var(--danger); }

</style>
