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

const cargando = ref(true)
const guardando = ref(false)

const form = ref({
  razon_social: '',
  rfc: '',
  nombre_patron: '',
  representante_trabajadores: '',
  nombre_capacitador: '',
})

const completo = computed(() =>
  Object.values(form.value).every(v => v.trim() !== ''),
)

async function cargar() {
  try {
    const { data } = await api.get('/instructor/dc3-empresa')
    form.value = {
      razon_social: data.razon_social || '',
      rfc: data.rfc || '',
      nombre_patron: data.nombre_patron || '',
      representante_trabajadores: data.representante_trabajadores || '',
      nombre_capacitador: data.nombre_capacitador || '',
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

.dc3-actions { text-align: right; }

</style>
