<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../api'
import { toast } from '../utils/toast'

const props = defineProps<{
  cursoId: string
}>()

const licencias = ref<any[]>([])
const loading = ref(false)

const showModal = ref(false)
const editMode = ref(false)
const form = ref<any>({ nombre: '', precio: 0, capacidad_maxima: 50 })

async function fetchLicencias() {
  loading.value = true
  try {
    const res = await api.get(`/capacitaciones/${props.cursoId}/licencias`)
    licencias.value = res.data || []
  } catch (e: any) {
    toast.error('Error al cargar licencias')
  } finally {
    loading.value = false
  }
}

onMounted(fetchLicencias)

function openCreate() {
  editMode.value = false
  form.value = { nombre: '', precio: 0, capacidad_maxima: 50 }
  showModal.value = true
}

function openEdit(lic: any) {
  editMode.value = true
  form.value = { ...lic }
  showModal.value = true
}

async function save() {
  if (!form.value.nombre) return toast.error('Nombre es requerido')
  if (form.value.precio < 0) return toast.error('Precio inválido')
  
  loading.value = true
  try {
    if (editMode.value) {
      await api.put(`/instructor/licencias/${form.value.id}`, {
        nombre: form.value.nombre,
        precio: Number(form.value.precio),
        capacidad_maxima: Number(form.value.capacidad_maxima)
      })
      toast.success('Licencia actualizada')
    } else {
      await api.post(`/instructor/licencias`, {
        capacitacion_id: props.cursoId,
        nombre: form.value.nombre,
        precio: Number(form.value.precio),
        capacidad_maxima: Number(form.value.capacidad_maxima)
      })
      toast.success('Licencia creada')
    }
    showModal.value = false
    fetchLicencias()
  } catch(e:any) {
    toast.error(e?.response?.data?.error || 'Error al guardar')
  } finally {
    loading.value = false
  }
}

async function remove(id: string) {
  if (!await toast.confirm('¿Seguro que deseas eliminar esta licencia?')) return
  try {
    await api.delete(`/instructor/licencias/${id}`)
    toast.success('Licencia eliminada')
    fetchLicencias()
  } catch (e) {
    toast.error('Error al eliminar')
  }
}
</script>

<template>
  <div class="licencia-list">
    <div class="header">
      <h3>Licencias y Cohortes</h3>
      <button class="btn btn-primary small" @click="openCreate">Nueva Licencia</button>
    </div>
    
    <p class="desc">
      Las licencias permiten vender acceso a tu curso a grupos de usuarios.
      Cada licencia genera automáticamente un <strong>grupo de chat aislado</strong> para esa cohorte.
    </p>

    <div v-if="loading && licencias.length === 0" class="loading">Cargando...</div>
    
    <div v-else-if="licencias.length === 0" class="empty">
      No hay licencias creadas para este curso.
    </div>

    <div v-else class="list">
      <div v-for="lic in licencias" :key="lic.id" class="lic-card">
        <div class="lic-info">
          <h4>{{ lic.nombre }}</h4>
          <div class="lic-meta">
            <span class="price">${{ lic.precio }}</span>
            <span class="capacity">Máx: {{ lic.capacidad_maxima || 'Ilimitado' }} alumnos</span>
            <span class="badge" v-if="lic.is_active !== false">Activa</span>
          </div>
          <div class="code-share" v-if="lic.codigo_acceso">
            <small>Código para alumnos:</small>
            <code class="code-box">{{ lic.codigo_acceso }}</code>
          </div>
        </div>
        <div class="lic-actions">
          <button class="btn-text" @click="openEdit(lic)">Editar</button>
          <button class="btn-text text-danger" @click="remove(lic.id)">Eliminar</button>
        </div>
      </div>
    </div>

    <!-- Modal Form -->
    <div v-if="showModal" class="modal-overlay" @mousedown.self="showModal = false">
      <div class="modal-content">
        <h4>{{ editMode ? 'Editar Licencia' : 'Nueva Licencia' }}</h4>
        
        <div class="field">
          <label>Nombre de la Cohorte</label>
          <input v-model="form.nombre" class="field-input" placeholder="Ej: Generación 2026 - A" />
        </div>
        
        <div class="row">
          <div class="field">
            <label>Precio (MXN)</label>
            <input type="number" v-model="form.precio" class="field-input" min="0" step="10" />
          </div>
          <div class="field">
            <label>Capacidad Máxima</label>
            <input type="number" v-model="form.capacidad_maxima" class="field-input" min="0" />
            <small>0 para ilimitado</small>
          </div>
        </div>

        <div class="actions">
          <button class="btn btn-secondary" @click="showModal = false">Cancelar</button>
          <button class="btn btn-primary" @click="save" :disabled="loading">Guardar</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.licencia-list { display: flex; flex-direction: column; gap: 16px; }
.header { display: flex; align-items: center; justify-content: space-between; }
.header h3 { margin: 0; font-size: 1.1rem; color: var(--dark); }
.desc { margin: 0; color: var(--muted); font-size: 0.9rem; }

.list { display: flex; flex-direction: column; gap: 12px; }
.lic-card { display: flex; align-items: center; justify-content: space-between; padding: 16px; background: var(--surface-soft); border-radius: var(--r-md); border: 1px solid var(--border-light); }
.lic-info h4 { margin: 0 0 6px 0; font-size: 1rem; color: var(--dark); }
.lic-meta { display: flex; gap: 12px; font-size: 0.85rem; color: var(--muted); align-items: center; }
.price { font-weight: 700; color: var(--brand); }
.capacity { background: rgba(0,0,0,0.05); padding: 2px 6px; border-radius: 4px; }
.lic-actions { display: flex; gap: 8px; }
.code-share { margin-top: 10px; display: flex; align-items: center; gap: 8px; }
.code-share small { color: var(--muted); font-size: 0.8rem; }
.code-box { background: rgba(249,115,22,0.1); color: var(--brand); padding: 4px 8px; border-radius: 6px; font-weight: bold; font-size: 0.9rem; user-select: all; }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 2000; }
.modal-content { background: var(--surface); padding: 24px; border-radius: var(--r-lg); width: 100%; max-width: 400px; display: flex; flex-direction: column; gap: 16px; }
.modal-content h4 { margin: 0; font-size: 1.25rem; color: var(--dark); }
.row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: 0.85rem; font-weight: 600; color: var(--dark); }
.field-input { width: 100%; padding: 10px 14px; border: 1px solid var(--border); border-radius: var(--r-md); outline: none; transition: all 0.2s; font-family: inherit; font-size: 0.95rem; }
.field-input:focus { border-color: var(--brand); box-shadow: 0 0 0 3px var(--brand-light); }
small { font-size: 0.75rem; color: var(--muted); }
.actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 8px; }
</style>
