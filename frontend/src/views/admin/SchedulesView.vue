<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
import iziToast from 'izitoast'
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import timeGridPlugin from '@fullcalendar/timegrid'
import interactionPlugin from '@fullcalendar/interaction'

const loading = ref(true)
const users = ref<any[]>([])

const calendarOptions = ref({
  plugins: [dayGridPlugin, timeGridPlugin, interactionPlugin],
  initialView: 'timeGridWeek',
  locale: 'es',
  headerToolbar: {
    left: 'prev,next today',
    center: 'title',
    right: 'dayGridMonth,timeGridWeek,timeGridDay'
  },
  events: [] as any[],
  editable: true,
  selectable: true,
  selectMirror: true,
  dayMaxEvents: true,
  slotMinTime: '06:00:00',
  slotMaxTime: '23:00:00',
  allDaySlot: false,
  select: handleDateSelect,
  eventClick: handleEventClick,
  eventDrop: handleEventDrop,
  eventResize: handleEventResize
})

const showModal = ref(false)
const formData = ref({
  id: '',
  instructor_id: '',
  start_time: new Date(),
  end_time: new Date(),
  status: 'available',
  repeat: 'none', // none, daily, weekly
  repeat_until: new Date()
})

async function loadData() {
  loading.value = true
  try {
    const [schedRes, usersRes] = await Promise.all([
      api.get('/admin/schedules'),
      api.get('/admin/users')
    ])
    
    users.value = (usersRes.data || []).filter((u: any) => u.role === 'instructor' || u.role === 'admin')
    
    // Map schedules to FullCalendar events
    calendarOptions.value.events = (schedRes.data || []).map((s: any) => {
      const u = users.value.find(x => x.id === s.instructor_id)
      return {
        id: s.id,
        title: u ? u.name : 'Horario',
        start: s.start_time,
        end: s.end_time,
        extendedProps: {
          instructor_id: s.instructor_id,
          status: s.status
        },
        backgroundColor: s.status === 'booked' ? '#dc2626' : 'var(--primary)',
        borderColor: s.status === 'booked' ? '#b91c1c' : 'var(--primary)',
      }
    })
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function handleDateSelect(selectInfo: any) {
  formData.value = {
    id: '',
    instructor_id: '',
    start_time: selectInfo.start,
    end_time: selectInfo.end,
    status: 'available',
    repeat: 'none',
    repeat_until: selectInfo.start
  }
  showModal.value = true
  selectInfo.view.calendar.unselect()
}

function handleEventClick(clickInfo: any) {
  const ev = clickInfo.event
  formData.value = {
    id: ev.id,
    instructor_id: ev.extendedProps.instructor_id,
    start_time: ev.start,
    end_time: ev.end || ev.start,
    status: ev.extendedProps.status,
    repeat: 'none',
    repeat_until: ev.start
  }
  showModal.value = true
}

async function handleEventDrop(dropInfo: any) {
  await updateEventTimes(dropInfo.event)
}

async function handleEventResize(resizeInfo: any) {
  await updateEventTimes(resizeInfo.event)
}

async function updateEventTimes(ev: any) {
  try {
    await api.put(`/admin/schedules/${ev.id}`, {
      start_time: ev.start.toISOString(),
      end_time: ev.end ? ev.end.toISOString() : ev.start.toISOString(),
      status: ev.extendedProps.status
    })
    iziToast.success({ title: 'Actualizado', message: 'Horario movido' })
  } catch (e) {
    iziToast.error({ title: 'Error', message: 'No se pudo mover el horario' })
    loadData() // revert
  }
}

async function save() {
  if (!formData.value.instructor_id) {
    iziToast.error({ title: 'Error', message: 'Selecciona un instructor' })
    return
  }
  
  loading.value = true
  try {
    if (formData.value.id) {
      // Editar uno solo
      await api.put(`/admin/schedules/${formData.value.id}`, {
        start_time: formData.value.start_time.toISOString(),
        end_time: formData.value.end_time.toISOString(),
        status: formData.value.status
      })
    } else {
      // Crear nuevo (con posible repeticion)
      if (formData.value.repeat === 'none') {
        await api.post('/admin/schedules', {
          instructor_id: formData.value.instructor_id,
          start_time: formData.value.start_time.toISOString(),
          end_time: formData.value.end_time.toISOString(),
        })
      } else {
        // Crear multiples
        let currStart = new Date(formData.value.start_time)
        let currEnd = new Date(formData.value.end_time)
        let limit = new Date(formData.value.repeat_until)
        limit.setHours(23, 59, 59, 999) // end of the limit day
        
        const promises = []
        while (currStart <= limit) {
          promises.push(api.post('/admin/schedules', {
            instructor_id: formData.value.instructor_id,
            start_time: currStart.toISOString(),
            end_time: currEnd.toISOString(),
          }))
          
          if (formData.value.repeat === 'daily') {
            currStart.setDate(currStart.getDate() + 1)
            currEnd.setDate(currEnd.getDate() + 1)
          } else if (formData.value.repeat === 'weekly') {
            currStart.setDate(currStart.getDate() + 7)
            currEnd.setDate(currEnd.getDate() + 7)
          } else {
            break // fallback
          }
        }
        await Promise.all(promises)
        iziToast.success({ title: 'Éxito', message: `${promises.length} horarios creados` })
      }
    }
    showModal.value = false
    await loadData()
  } catch (e) {
    iziToast.error({ title: 'Error', message: 'Error al guardar horario' })
  } finally {
    loading.value = false
  }
}

async function remove() {
  if (!formData.value.id) return
  if (!confirm('¿Eliminar este horario?')) return
  try {
    await api.delete(`/admin/schedules/${formData.value.id}`)
    showModal.value = false
    await loadData()
  } catch (e) {
    iziToast.error({ title: 'Error', message: 'Error al eliminar' })
  }
}
</script>

<template>
  <div class="schedules-page">
    <div class="page-header">
      <h2>Calendario de Instructores</h2>
      <button class="btn btn-primary" @click="handleDateSelect({ start: new Date(), end: new Date(Date.now() + 3600000), view: { calendar: { unselect:()=>{} } } })">
        Crear Horario Manual
      </button>
    </div>

    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div> Cargando...
    </div>

    <div class="calendar-container">
      <FullCalendar :options="calendarOptions" />
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay">
      <div class="modal">
        <h3>{{ formData.id ? 'Editar' : 'Crear' }} Horario</h3>
        
        <div class="form-group">
          <label>Instructor</label>
          <select v-model="formData.instructor_id" class="input">
            <option value="" disabled>Seleccione un instructor</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ u.name }}</option>
          </select>
        </div>

        <div style="display:flex; gap:16px;">
          <div class="form-group" style="flex:1;">
            <label>Inicio</label>
            <VueDatePicker v-model="formData.start_time" :is-24="false" teleport="body" />
          </div>

          <div class="form-group" style="flex:1;">
            <label>Fin</label>
            <VueDatePicker v-model="formData.end_time" :is-24="false" teleport="body" />
          </div>
        </div>

        <div class="form-group" v-if="!formData.id">
          <label>Repetir (opcional)</label>
          <select v-model="formData.repeat" class="input">
            <option value="none">No repetir</option>
            <option value="daily">Diariamente</option>
            <option value="weekly">Semanalmente</option>
          </select>
        </div>
        
        <div class="form-group" v-if="!formData.id && formData.repeat !== 'none'">
          <label>Repetir hasta la fecha</label>
          <VueDatePicker v-model="formData.repeat_until" :enable-time-picker="false" teleport="body" />
        </div>

        <div class="form-group" v-if="formData.id">
          <label>Estado</label>
          <select v-model="formData.status" class="input">
            <option value="available">Disponible</option>
            <option value="booked">Reservado</option>
            <option value="cancelled">Cancelado</option>
          </select>
        </div>

        <div class="modal-actions" style="justify-content: space-between;">
          <button v-if="formData.id" class="btn btn-danger" @click="remove">Eliminar</button>
          <div v-else></div>
          
          <div style="display:flex; gap:12px;">
            <button class="btn btn-outline" @click="showModal = false">Cancelar</button>
            <button class="btn btn-primary" @click="save">Guardar</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.schedules-page {
  padding: 24px;
  position: relative;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.calendar-container {
  background: var(--surface);
  border-radius: var(--r-lg);
  padding: 20px;
  box-shadow: var(--shadow-sm);
  min-height: 700px;
}

/* Modal and Overlays */
.loading-overlay {
  position: absolute; inset: 0; background: rgba(0,0,0,0.3); z-index: 10;
  display: flex; align-items: center; justify-content: center; gap:10px; color:#fff;
  border-radius: var(--r-lg);
}
.spinner {
  width: 24px; height: 24px; border: 3px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%; animation: spin 1s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center; z-index: 9999;
}
.modal {
  background: var(--surface); padding: 24px; border-radius: var(--r-md); width: 500px; max-width: 95%;
  box-shadow: 0 10px 25px rgba(0,0,0,0.5);
}
.form-group {
  margin-bottom: 16px;
}
.form-group label {
  display: block; margin-bottom: 6px; font-weight: 600; font-size: 0.9rem;
}
.input {
  width: 100%; padding: 10px; border: 1px solid var(--border); border-radius: 6px;
  background: var(--bg-color); color: var(--text-color);
}
.modal-actions {
  display: flex; margin-top: 24px;
}
.btn-danger {
  background: #ef4444; color: #fff; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer;
}
.btn-danger:hover { background: #dc2626; }
</style>

<style>
/* FullCalendar Theme Overrides */
.fc {
  color: var(--text-color);
}
.fc-theme-standard td, .fc-theme-standard th {
  border-color: var(--border-light) !important;
}
.fc-theme-standard .fc-scrollgrid {
  border-color: var(--border) !important;
}
.fc .fc-button-primary {
  background-color: var(--primary) !important;
  border-color: var(--primary) !important;
}
.fc .fc-button-primary:not(:disabled):active, 
.fc .fc-button-primary:not(:disabled).fc-button-active {
  background-color: var(--primary-hover) !important;
  border-color: var(--primary-hover) !important;
}
</style>
