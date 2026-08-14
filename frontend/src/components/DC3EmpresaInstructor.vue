<script setup lang="ts">
/**
 * Empresa por defecto del instructor para las constancias DC-3.
 *
 * Se captura UNA vez y firma las constancias de todo alumno que no declare
 * patrón propio: típicamente el particular que se capacita por su cuenta. Si el
 * alumno sí trabaja para una empresa, la suya tiene precedencia — legalmente el
 * patrón es quien lo emplea.
 *
 * El nombre del capacitador NO se sustituye nunca: quien imparte no cambia
 * porque el trabajador tenga empleador propio.
 */
import { ref, computed, onMounted } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'
import { uploadToR2 } from '../utils/upload'

const cargando = ref(true)
const guardando = ref(false)
const subiendoLogo = ref(false)
const logoInput = ref<HTMLInputElement | null>(null)

const form = ref({
  razon_social: '',
  rfc: '',
  nombre_patron: '',
  representante_trabajadores: '',
  nombre_capacitador: '',
  // El logo es OPCIONAL: sin él la constancia sale con el de la plantilla, que
  // es un documento igual de válido. Por eso queda fuera de `completo`.
  logo_url: '',
})

const completo = computed(() =>
  ['razon_social', 'rfc', 'nombre_patron', 'representante_trabajadores', 'nombre_capacitador']
    .every(k => (form.value as any)[k].trim() !== ''),
)

/**
 * El logo se incrusta DENTRO del .docx, así que su peso es el peso del
 * documento que descarga cada alumno. Se acota antes de subirlo.
 */
const LOGO_MAX_MB = 4

async function subirLogo(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (file.size > LOGO_MAX_MB * 1024 * 1024) {
    toast.error(`El logo no puede superar ${LOGO_MAX_MB} MB`)
    return
  }
  subiendoLogo.value = true
  try {
    form.value.logo_url = await uploadToR2(file, 'dc3-logos')
    toast.success('Logo subido. Recuerda guardar los datos.')
  } catch {
    toast.error('No se pudo subir el logo')
  } finally {
    subiendoLogo.value = false
    if (logoInput.value) logoInput.value.value = ''
  }
}

function quitarLogo() {
  form.value.logo_url = ''
}

async function cargar() {
  try {
    const { data } = await api.get('/instructor/dc3-empresa')
    form.value = {
      razon_social: data.razon_social || '',
      rfc: data.rfc || '',
      nombre_patron: data.nombre_patron || '',
      representante_trabajadores: data.representante_trabajadores || '',
      nombre_capacitador: data.nombre_capacitador || '',
      logo_url: data.logo_url || '',
    }
  } catch {
    // Sin configurar todavía: se deja el formulario vacío.
  } finally {
    cargando.value = false
  }
}

async function guardar() {
  guardando.value = true
  try {
    await api.put('/instructor/dc3-empresa', {
      razon_social: form.value.razon_social.trim(),
      rfc: form.value.rfc.trim().toUpperCase(),
      nombre_patron: form.value.nombre_patron.trim(),
      representante_trabajadores: form.value.representante_trabajadores.trim(),
      nombre_capacitador: form.value.nombre_capacitador.trim(),
      logo_url: form.value.logo_url,
    })
    toast.success('Datos guardados')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No se pudieron guardar los datos')
  } finally {
    guardando.value = false
  }
}

onMounted(cargar)
</script>

<template>
  <div v-if="!cargando" class="dc3-card">
    <div class="dc3-card-head">
      <h2>Datos para constancias DC-3</h2>
      <p>
        Se usan en las constancias de los alumnos que no declaran una empresa
        propia. Se capturan una vez y valen para todas tus capacitaciones.
      </p>
    </div>

    <!--
      Aviso, no error: no tenerlo configurado es un estado normal al empezar.
      Pero conviene que el instructor sepa qué se rompe si lo deja así, porque
      el síntoma le llega indirecto —alumnos que no reciben su constancia—.
    -->
    <div v-if="!completo" class="dc3-aviso">
      Faltan datos. Hasta completarlos, los alumnos sin empresa propia no podrán
      recibir su constancia al terminar tus cursos.
    </div>

    <div class="dc3-grid">
      <label class="dc3-field">
        <span>Razón social <em>*</em></span>
        <input v-model="form.razon_social" class="field-input"
               placeholder="Nombre o razón social" />
      </label>
      <label class="dc3-field">
        <span>RFC <em>*</em></span>
        <input v-model="form.rfc" maxlength="13" class="field-input"
               placeholder="12 o 13 caracteres" />
      </label>
      <label class="dc3-field">
        <span>Patrón o representante legal <em>*</em></span>
        <input v-model="form.nombre_patron" class="field-input" />
      </label>
      <label class="dc3-field">
        <span>Representante de los trabajadores <em>*</em></span>
        <input v-model="form.representante_trabajadores" class="field-input" />
      </label>
      <label class="dc3-field dc3-field-full">
        <span>Nombre del capacitador <em>*</em></span>
        <input v-model="form.nombre_capacitador" class="field-input" />
        <small class="dc3-hint">
          Capacitador acreditado que imparte. Va en todas las constancias, incluso
          en las de alumnos con empresa propia.
        </small>
      </label>
    </div>

    <!--
      Logotipo de la empresa.
      La cabecera de la constancia lleva DOS: este y un sello del formato
      oficial que no se sustituye nunca. Es opcional — sin él sale el de la
      plantilla, que es un documento igual de válido.
    -->
    <div class="dc3-logo">
      <div class="dc3-logo-info">
        <strong>Logotipo de la empresa</strong>
        <p>
          Aparece en la cabecera de la constancia, junto al sello del formato
          oficial. Opcional: si no lo subes se usa el de la plantilla.
        </p>
      </div>

      <div class="dc3-logo-zona">
        <img v-if="form.logo_url" :src="form.logo_url" alt="Logotipo" class="dc3-logo-prev" />
        <div v-else class="dc3-logo-vacio">Sin logotipo</div>

        <div class="dc3-logo-btns">
          <button class="btn btn-secondary btn-sm" :disabled="subiendoLogo"
                  @click="logoInput?.click()">
            {{ subiendoLogo ? 'Subiendo…' : form.logo_url ? 'Cambiar' : 'Subir logotipo' }}
          </button>
          <button v-if="form.logo_url" class="btn btn-secondary btn-sm" @click="quitarLogo">
            Quitar
          </button>
        </div>
        <input ref="logoInput" type="file" accept="image/jpeg,image/png,image/webp"
               style="display:none" @change="subirLogo" />
      </div>
    </div>

    <div class="dc3-actions">
      <button class="btn btn-primary" :disabled="guardando" @click="guardar">
        {{ guardando ? 'Guardando…' : 'Guardar datos' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.dc3-card {
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 22px;
  background: var(--surface);
}

.dc3-card-head { margin-bottom: 16px; }
.dc3-card-head h2 { margin: 0 0 6px; font-size: 1.05rem; color: var(--text); }
.dc3-card-head p { margin: 0; font-size: 0.87rem; color: var(--muted); line-height: 1.55; }

.dc3-aviso {
  padding: 10px 14px;
  border-radius: 8px;
  background: rgba(245, 158, 11, 0.12);
  color: #b45309;
  font-size: 0.83rem;
  margin-bottom: 16px;
}

.dc3-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr));
  gap: 14px;
  margin-bottom: 18px;
}

.dc3-field { display: flex; flex-direction: column; gap: 6px; }
.dc3-field-full { grid-column: 1 / -1; }

.dc3-field > span {
  font-size: 0.76rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--muted);
}

.dc3-field em { color: var(--danger); font-style: normal; }
.dc3-hint { font-size: 0.75rem; color: var(--muted); }

.dc3-logo {
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 18px;
  background: var(--bg);
}

.dc3-logo-info strong { font-size: 0.9rem; color: var(--text); }
.dc3-logo-info p { margin: 4px 0 14px; font-size: 0.8rem; color: var(--muted); line-height: 1.5; }

/* Sin media query: se apila solo cuando no cabe. */
.dc3-logo-zona {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 14px;
}

.dc3-logo-prev,
.dc3-logo-vacio {
  width: 88px; height: 88px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface);
  flex-shrink: 0;
}

.dc3-logo-prev { object-fit: contain; padding: 6px; }

.dc3-logo-vacio {
  display: grid; place-items: center;
  font-size: 0.72rem; color: var(--muted); text-align: center;
}

.dc3-logo-btns { display: flex; flex-wrap: wrap; gap: 8px; }

.dc3-actions { text-align: right; }

</style>
