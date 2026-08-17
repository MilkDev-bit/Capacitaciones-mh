<script setup lang="ts">
/**
 * Panel de la constancia DC-3 dentro de una capacitación.
 *
 * Cada estado responde a una pregunta distinta del alumno:
 *
 *   lista         → "ya está, descárgala"
 *   faltan-mios   → "faltan TUS datos, captúralos" (el único accionable por él)
 *   falta-empresa → "faltan los del instructor, no puedes hacer nada"
 *   sin-emitir    → sus datos y los del instructor están, pero no hay documento
 *   no-aplica     → el curso no emite constancia, no se muestra nada
 *
 * Distinguir los dos "faltan" importa: mandar al alumno a rellenar un formulario
 * que no va a desbloquear nada es peor que decirle a quién reclamar.
 *
 * `sin-emitir` estaba fundido con `falta-empresa` y ese era el bug: el alumno
 * veía "falta que el instructor complete la empresa" cuando la empresa ya
 * estaba completa y lo que había fallado era el armado del documento. Le
 * echaba la culpa a un tercero por algo que un reintento resuelve.
 */
import { ref, computed, onMounted, watch } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'
import { estadoDC3 } from '../utils/dc3'

const props = withDefaults(defineProps<{
  cursoId: string
  /** El panel solo aparece cuando el curso está terminado. */
  completado: boolean
  /**
   * dc3_enabled de la capacitación.
   *
   * El default explícito NO es decorativo. Vue castea las props declaradas como
   * `boolean` cuando no se pasan: `habilitado?: boolean` omitido llegaba como
   * `false`, no como `undefined`, así que el guard `habilitado === false` daba
   * positivo y el panel se cortaba antes de consultar nada. Resultado: el modal
   * se abría en blanco y el alumno no podía capturar su CURP, con lo que el
   * backend nunca llegaba a emitir ninguna constancia.
   */
  habilitado?: boolean
  /**
   * `plano` quita el marco propio. Dentro del modal el panel ya está sobre una
   * tarjeta, y anidar dos bordes con el mismo radio se lee como un error.
   */
  variante?: 'tarjeta' | 'plano'
}>(), { variante: 'tarjeta', habilitado: true })

const emit = defineEmits<{ (e: 'emitida', url: string): void }>()

const cargando = ref(false)
const guardando = ref(false)
const constanciaUrl = ref('')
const trabajadorCompleto = ref(false)
const empresaCompleta = ref(false)
const consultado = ref(false)
const errorCarga = ref('')
/** Campos que el servidor reportó como vacíos en el último intento de emisión. */
const faltantes = ref<string[]>([])

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

/**
 * El alumno pidió corregir datos que ya había guardado.
 *
 * Sin esto, una CURP mal tecleada quedaba congelada para siempre: el formulario
 * solo se mostraba mientras `trabajador_completo` fuera falso, y en cuanto
 * guardaba una vez ya no había forma de volver a él.
 */
const editando = ref(false)

// La decisión vive en utils/dc3 para poder cubrirla con tests: aquí solo se le
// pasa el estado actual.
const estado = computed(() => estadoDC3({
  consultado: consultado.value,
  constanciaUrl: constanciaUrl.value,
  trabajadorCompleto: trabajadorCompleto.value,
  empresaCompleta: empresaCompleta.value,
  editando: editando.value,
}))

const curpValida = computed(() => form.value.curp.trim().length === 18)

async function consultar() {
  if (!props.completado || !props.habilitado) return
  cargando.value = true
  errorCarga.value = ''
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
  } catch (e: any) {
    // Antes se tragaba en silencio. El precio de eso fue un modal en blanco que
    // parecía una pantalla rota: sin mensaje, el alumno no sabe si esperar,
    // reintentar o avisar a alguien.
    errorCarga.value = e.response?.data?.error || 'No pudimos consultar el estado de tu constancia.'
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
    editando.value = false
    // 202 significa "guardado, pero el documento no salió".
    //
    // Aquí se ponía `empresaCompleta = false` a pelo, y era una invención: el
    // cliente no había comprobado nada de la empresa. Bastaba con que faltara
    // la duración del curso para que el panel acusara al instructor de no haber
    // capturado unos datos que sí tenía. Ahora se vuelve a preguntar al
    // servidor, que es el único que sabe qué falta.
    if (status === 202) {
      faltantes.value = Array.isArray(data.faltan) ? data.faltan : []
      toast.success(data.mensaje || 'Guardamos tus datos')
      await consultar()
      return
    }
    faltantes.value = []
    constanciaUrl.value = data.constancia_url || ''
    if (constanciaUrl.value) {
      emit('emitida', constanciaUrl.value)
      toast.success('Tu constancia está lista')
    } else {
      // El backend aceptó los datos pero no devolvió documento. Decirle
      // "está lista" y no darle nada que descargar es la peor combinación.
      toast.error('Guardamos tus datos, pero no pudimos armar el documento')
    }
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No se pudieron guardar tus datos')
  } finally {
    guardando.value = false
  }
}

/**
 * Reintenta la emisión reenviando los datos ya guardados.
 *
 * El endpoint es idempotente: si el documento existe lo devuelve, y si no,
 * vuelve a intentar armarlo. Por eso reutiliza `enviar` en vez de duplicar la
 * llamada con otra ruta.
 */
function reintentar() {
  enviar()
}

onMounted(consultar)
watch(() => props.completado, (val) => { if (val) consultar() })
// El modal reutiliza la misma instancia al abrirlo para otra capacitación.
watch(() => props.cursoId, () => {
  consultado.value = false
  constanciaUrl.value = ''
  editando.value = false
  consultar()
})
</script>

<template>
  <!--
    `consultado` salió del v-if a propósito.
    Mientras estaba aquí, cualquier fallo antes de que la consulta terminara
    dejaba el componente sin pintar ni un carácter, y un componente invisible no
    se puede diagnosticar mirando la pantalla. Ahora la carga y el error tienen
    su propio estado visible.
  -->
  <div v-if="completado && habilitado"
       class="dc3-panel" :class="{ 'dc3-plano': variante === 'plano' }">
    <div v-if="variante !== 'plano'" class="dc3-head">
      <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" /><path d="M14 2v6h6" />
      </svg>
      <strong>Constancia DC-3</strong>
    </div>

    <!-- Cargando -->
    <p v-if="estado === 'cargando'" class="dc3-lead">Consultando el estado de tu constancia…</p>

    <!-- Error de consulta -->
    <template v-else-if="errorCarga">
      <p class="dc3-lead">{{ errorCarga }}</p>
      <button class="btn btn-secondary" :disabled="cargando" @click="consultar">Reintentar</button>
    </template>

    <!-- Lista -->
    <template v-else-if="estado === 'lista'">
      <p class="dc3-lead">Tu constancia de habilidades laborales está emitida.</p>
      <div class="dc3-acciones">
        <a :href="constanciaUrl" target="_blank" rel="noopener" class="btn btn-primary" download>
          Descargar constancia
        </a>
        <button class="btn btn-ghost" @click="editando = true">Corregir mis datos</button>
      </div>
      <small class="dc3-hint">
        La tienes siempre disponible en <b>Mis constancias</b>, dentro de tu perfil.
      </small>
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

      <div class="dc3-acciones">
        <button class="btn btn-primary"
                :disabled="guardando || !curpValida || empresaAMedias" @click="enviar">
          {{ guardando ? 'Emitiendo…' : 'Emitir mi constancia' }}
        </button>
        <button v-if="editando" class="btn btn-ghost" :disabled="guardando"
                @click="editando = false">
          Cancelar
        </button>
      </div>
    </template>

    <!-- Falta que el instructor configure la empresa -->
    <template v-else-if="estado === 'falta-empresa'">
      <!--
        Se nombra el área temática además de la empresa porque el backend mete
        las dos cosas en la misma bandera: un curso con la empresa perfectamente
        capturada pero sin área marcada cae aquí, y el instructor se pondría a
        revisar su perfil sin encontrar nada mal.
      -->
      <p class="dc3-lead">
        Tus datos están guardados. Falta que el instructor complete los datos de la
        empresa y el capacitador en su perfil, y el área temática de esta
        capacitación; en cuanto lo haga, emitimos tu constancia automáticamente y la
        verás en <b>Mis constancias</b>.
      </p>
      <button class="btn btn-ghost" @click="editando = true">Revisar mis datos</button>
    </template>

    <!--
      Sin emitir: ya no hay nadie a quien esperar.
      Se ofrece el reintento antes que el aviso al soporte porque la causa
      habitual —un dato del curso que el instructor acaba de corregir— se
      resuelve sola en cuanto se vuelve a pedir el documento.
    -->
    <template v-else>
      <p class="dc3-lead">
        Tus datos están guardados, pero el documento todavía no se puede generar.
      </p>
      <!--
        La lista viene del servidor, que es quien intentó armar el documento.
        Sin ella el mensaje tenía que elegir un culpable a ciegas y solía elegir
        mal; con ella el alumno sabe exactamente qué pedirle a su instructor.
      -->
      <ul v-if="faltantes.length" class="dc3-faltantes">
        <li v-for="f in faltantes" :key="f">{{ f }}</li>
      </ul>
      <p v-else class="dc3-lead">
        Avisa a tu instructor de que revise la duración y el área temática del curso.
      </p>
      <div class="dc3-acciones">
        <button class="btn btn-primary" :disabled="guardando" @click="reintentar">
          {{ guardando ? 'Reintentando…' : 'Reintentar emisión' }}
        </button>
        <button class="btn btn-ghost" :disabled="guardando" @click="editando = true">
          Corregir mis datos
        </button>
      </div>
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

.dc3-faltantes {
  margin: 0 0 16px;
  padding-left: 20px;
  color: var(--text);
  font-size: 0.88rem;
  line-height: 1.6;
}

/* Dentro del modal el marco lo pone el propio modal. */
.dc3-plano {
  border: 0;
  padding: 0;
  background: none;
  margin-top: 0;
}

/* En móvil los botones se apilan a ancho completo: dos acciones en fila
 * quedan por debajo del objetivo táctil de 44px en pantallas estrechas. */
.dc3-acciones {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 8px;
}

.dc3-acciones > * { min-height: var(--touch-min); }

@media (max-width: 479px) {
  .dc3-acciones { flex-direction: column; align-items: stretch; }
  .dc3-acciones > * { width: 100%; text-align: center; }
}

</style>
