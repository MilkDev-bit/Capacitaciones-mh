<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'

const perfil = ref<any>(null)
const loading = ref(false)
const loadingSave = ref(false)
const error = ref('')
const success = ref('')

const form = ref({ name: '', bio: '', phone: '' })
const password = ref({ current: '', nueva: '', confirmar: '' })
const showPass = ref(false)

async function load() {
  loading.value = true
  const res = await api.get('/perfil')
  perfil.value = res.data
  form.value.name = res.data.name || ''
  form.value.bio = res.data.bio || ''
  form.value.phone = res.data.phone || ''
  loading.value = false
}

onMounted(load)

async function guardar() {
  error.value = ''; success.value = ''
  if (!form.value.name) { error.value = 'El nombre es requerido'; return }
  if (showPass.value) {
    if (!password.value.nueva) { error.value = 'Ingresa la nueva contrasena'; return }
    if (password.value.nueva.length < 6) { error.value = 'La contrasena debe tener al menos 6 caracteres'; return }
    if (password.value.nueva !== password.value.confirmar) { error.value = 'Las contrasenas no coinciden'; return }
  }
  loadingSave.value = true
  try {
    const payload: any = { name: form.value.name, bio: form.value.bio, phone: form.value.phone }
    if (showPass.value && password.value.nueva) {
      payload.password = password.value.nueva
    }
    await api.put('/perfil', payload)
    success.value = 'Perfil actualizado correctamente'
    password.value = { current: '', nueva: '', confirmar: '' }
    showPass.value = false
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loadingSave.value = false
  }
}

function initials(name: string) {
  return name ? name.split(' ').slice(0, 2).map(n => n[0]).join('').toUpperCase() : '?'
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 p-6">
    <div class="max-w-lg mx-auto">
      <h1 class="text-2xl font-bold text-gray-800 mb-6">Mi Perfil</h1>

      <div v-if="loading" class="text-center py-12 text-gray-400">Cargando...</div>
      <div v-else>
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-4">
          <div class="flex items-center gap-4 mb-6">
            <div class="w-16 h-16 rounded-full bg-blue-600 flex items-center justify-center text-white text-xl font-bold flex-shrink-0">
              {{ initials(form.name) }}
            </div>
            <div>
              <p class="font-semibold text-gray-800 text-lg">{{ perfil?.name }}</p>
              <p class="text-sm text-gray-500">{{ perfil?.email }}</p>
              <span class="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full capitalize mt-1 inline-block">{{ perfil?.role }}</span>
            </div>
          </div>

          <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">{{ error }}</div>
          <div v-if="success" class="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded-lg text-sm">{{ success }}</div>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Nombre *</label>
              <input v-model="form.name" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Biografia</label>
              <textarea v-model="form.bio" rows="3" placeholder="Cuentanos sobre ti..." class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Telefono</label>
              <input v-model="form.phone" type="tel" placeholder="+1 234 567 8900" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
          </div>
        </div>

        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-4">
          <div class="flex items-center justify-between mb-3">
            <h2 class="font-semibold text-gray-700">Cambiar contrasena</h2>
            <button @click="showPass = !showPass" class="text-sm text-blue-600 hover:underline">
              {{ showPass ? 'Cancelar' : 'Cambiar' }}
            </button>
          </div>
          <div v-if="showPass" class="space-y-3">
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Nueva contrasena</label>
              <input v-model="password.nueva" type="password" placeholder="Minimo 6 caracteres" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Confirmar contrasena</label>
              <input v-model="password.confirmar" type="password" placeholder="Repite la nueva contrasena" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
          </div>
        </div>

        <button @click="guardar" :disabled="loadingSave" class="w-full bg-blue-600 text-white py-2.5 rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 transition">
          {{ loadingSave ? 'Guardando...' : 'Guardar Cambios' }}
        </button>
      </div>
    </div>
  </div>
</template>
