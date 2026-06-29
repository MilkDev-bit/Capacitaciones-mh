<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const schedules = ref<any[]>([])
const loading = ref(true)

const users = ref<any[]>([])

const showModal = ref(false)
const formData = ref({
  id: '',
  instructor_id: '',
  start_time: '',
  end_time: '',
  status: 'available'
})

async function loadData() {
  loading.value = true
  try {
    const [schedRes, usersRes] = await Promise.all([
      api.get('/admin/schedules'),
      api.get('/admin/users') // To pick an instructor
    ])
    schedules.value = schedRes.data.schedules || []
    users.value = (usersRes.data || []).filter((u: any) => u.role === 'instructor')
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function formatTime(iso: string) {
  if (!iso) return ''
  return new Date(iso).toLocaleString()
}

function getInstructorName(id: string) {
  const u = users.value.find(x => x.id === id)
  return u ? u.name : 'Desconocido'
}

function openCreate() {
  formData.value = {
    id: '',
    instructor_id: '',
    start_time: '',
    end_time: '',
    status: 'available'
  }
  showModal.value = true
}

function openEdit(s: any) {
  // Convert timestamps to datetime-local format YYYY-MM-DDThh:mm
  const toLocal = (iso: string) => {
    if (!iso) return ''
    const d = new Date(iso)
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
    return d.toISOString().slice(0,16)
  }

  formData.value = {
    id: s.id,
    instructor_id: s.instructor_id,
    start_time: toLocal(s.start_time),
    end_time: toLocal(s.end_time),
    status: s.status
  }
  showModal.value = true
}

async function save() {
  try {
    // Convert back to ISO
    const start = new Date(formData.value.start_time).toISOString()
    const end = new Date(formData.value.end_time).toISOString()

    if (formData.value.id) {
      await api.put(`/admin/schedules/${formData.value.id}`, {
        start_time: start,
        end_time: end,
        status: formData.value.status
      })
    } else {
      await api.post('/admin/schedules', {
        instructor_id: formData.value.instructor_id,
        start_time: start,
        end_time: end,
      })
    }
    showModal.value = false
    loadData()
  } catch (e) {
    alert('Error al guardar horario')
  }
}

async function remove(id: string) {
  if (!confirm('¿Eliminar horario?')) return
  try {
    await api.delete(`/admin/schedules/${id}`)
    loadData()
  } catch (e) {
    alert('Error al eliminar')
  }
}

</script>

<template>
  <div class="schedules-page">
    <div class="page-header">
      <h2>Disponibilidad de Instructores</h2>
      <button class="btn btn-primary" @click="openCreate">Crear Horario</button>
    </div>

    <div v-if="loading" class="loading">Cargando...</div>
    <div v-else>
      <table class="table">
        <thead>
          <tr>
            <th>Instructor</th>
            <th>Inicio</th>
            <th>Fin</th>
            <th>Estado</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in schedules" :key="s.id">
            <td>{{ getInstructorName(s.instructor_id) }}</td>
            <td>{{ formatTime(s.start_time) }}</td>
            <td>{{ formatTime(s.end_time) }}</td>
            <td>{{ s.status }}</td>
            <td>
              <button class="btn btn-outline btn-sm" @click="openEdit(s)">Editar</button>
              <button class="btn btn-danger btn-sm" @click="remove(s.id)">Eliminar</button>
            </td>
          </tr>
          <tr v-if="schedules.length === 0">
            <td colspan="5" style="text-align:center">No hay horarios registrados.</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay">
      <div class="modal">
        <h3>{{ formData.id ? 'Editar' : 'Crear' }} Horario</h3>
        
        <div class="form-group" v-if="!formData.id">
          <label>Instructor</label>
          <select v-model="formData.instructor_id" class="input">
            <option value="" disabled>Seleccione un instructor</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ u.name }}</option>
          </select>
        </div>

        <div class="form-group">
          <label>Inicio</label>
          <input type="datetime-local" v-model="formData.start_time" class="input" />
        </div>

        <div class="form-group">
          <label>Fin</label>
          <input type="datetime-local" v-model="formData.end_time" class="input" />
        </div>

        <div class="form-group" v-if="formData.id">
          <label>Estado</label>
          <select v-model="formData.status" class="input">
            <option value="available">Disponible</option>
            <option value="booked">Reservado</option>
            <option value="cancelled">Cancelado</option>
          </select>
        </div>

        <div class="modal-actions">
          <button class="btn btn-outline" @click="showModal = false">Cancelar</button>
          <button class="btn btn-primary" @click="save">Guardar</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.schedules-page {
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.table {
  width: 100%;
  border-collapse: collapse;
  background: var(--surface);
  border-radius: var(--r-md);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}
.table th, .table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid var(--border-light);
}
.btn-sm {
  padding: 4px 8px;
  font-size: 0.8rem;
  margin-right: 8px;
}
.btn-danger {
  background: #ef4444; color: #fff; border: none;
}
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center; z-index: 99;
}
.modal {
  background: var(--surface); padding: 24px; border-radius: var(--r-md); width: 400px; max-width: 90%;
}
.form-group {
  margin-bottom: 16px;
}
.form-group label {
  display: block; margin-bottom: 4px; font-weight: 600;
}
.input {
  width: 100%; padding: 8px; border: 1px solid var(--border); border-radius: 4px;
}
.modal-actions {
  display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px;
}
</style>
