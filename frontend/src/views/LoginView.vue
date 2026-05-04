<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../api'
import router from '../router'
import { useRoute } from 'vue-router'

const auth = useAuthStore()
const route = useRoute()
const tab = ref<'login' | 'register'>('login')

const email = ref('')
const password = ref('')
const showPass = ref(false)
const error = ref('')
const loading = ref(false)

const regName = ref('')
const regEmail = ref('')
const regPassword = ref('')
const regRole = ref('user')
const regError = ref('')
const regSuccess = ref('')
const regLoading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res = await api.post('/login', { email: email.value, password: password.value })
    auth.token = res.data.token
    auth.user = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
    const redirect = route.query.redirect as string | undefined
    if (redirect) router.push(redirect)
    else if (res.data.user?.role === 'admin') router.push('/admin')
    else if (res.data.user?.role === 'instructor') router.push('/instructor')
    else router.push('/usuario')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Correo o contraseña incorrectos'
  } finally {
    loading.value = false
  }
}

async function register() {
  regError.value = ''; regSuccess.value = ''
  if (!regName.value || !regEmail.value || !regPassword.value) {
    regError.value = 'Todos los campos son requeridos'; return
  }
  regLoading.value = true
  try {
    await api.post('/register', { name: regName.value, email: regEmail.value, password: regPassword.value, role: regRole.value })
    regSuccess.value = 'Cuenta creada! Ya puedes iniciar sesion.'
    regName.value = ''; regEmail.value = ''; regPassword.value = ''; regRole.value = 'user'
    setTimeout(() => { tab.value = 'login'; regSuccess.value = '' }, 1500)
  } catch (e: any) {
    regError.value = e.response?.data?.error || 'Error al registrarse'
  } finally {
    regLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex">
    <!-- Left hero panel -->
    <div class="hidden lg:flex flex-col justify-center w-5/12 xl:w-2/5 bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 px-14 py-16 relative overflow-hidden flex-shrink-0">
      <div class="absolute -right-20 -bottom-20 w-80 h-80 rounded-full bg-orange-500/20 blur-3xl pointer-events-none" />
      <div class="absolute -left-10 top-10 w-48 h-48 rounded-full bg-orange-500/10 blur-2xl pointer-events-none" />

      <div class="relative z-10">
        <div class="flex items-center gap-3 mb-10">
          <div class="w-12 h-12 bg-orange-500/20 rounded-2xl flex items-center justify-center border border-orange-500/30">
            <svg width="24" height="24" viewBox="0 0 44 44" fill="none">
              <path d="M12 32L22 14L32 32H12Z" fill="#f97316"/>
            </svg>
          </div>
        </div>
        <h1 class="text-4xl font-black text-white leading-tight mb-4">
          Capacitaciones<br>
          <span class="text-orange-400">MH</span>
        </h1>
        <p class="text-gray-400 text-base leading-relaxed max-w-xs">
          Tu plataforma de aprendizaje corporativo. Accede a cursos, completa examenes y avanza a tu ritmo.
        </p>
        <div class="mt-10 space-y-4">
          <div class="flex items-center gap-3 text-gray-300 text-sm">
            <div class="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center text-base flex-shrink-0">&#128218;</div>
            Cursos en video, documentos y texto
          </div>
          <div class="flex items-center gap-3 text-gray-300 text-sm">
            <div class="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center text-base flex-shrink-0">&#128221;</div>
            Examenes y seguimiento de progreso
          </div>
          <div class="flex items-center gap-3 text-gray-300 text-sm">
            <div class="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center text-base flex-shrink-0">&#127891;</div>
            Certifica tus conocimientos
          </div>
        </div>
      </div>
    </div>

    <!-- Right form panel -->
    <div class="flex-1 flex items-center justify-center bg-gray-50 px-6 py-12">
      <div class="w-full max-w-md">
        <!-- Mobile logo -->
        <div class="flex lg:hidden items-center gap-2.5 mb-8">
          <div class="w-9 h-9 bg-orange-500 rounded-xl flex items-center justify-center">
            <svg width="18" height="18" viewBox="0 0 44 44" fill="none">
              <path d="M12 32L22 14L32 32H12Z" fill="white"/>
            </svg>
          </div>
          <span class="font-extrabold text-gray-900 text-lg">Capacitaciones MH</span>
        </div>

        <!-- Tabs -->
        <div class="flex bg-gray-200 rounded-xl p-1 gap-1 mb-8">
          <button
            :class="[
              'flex-1 py-2.5 rounded-lg text-sm font-semibold transition-all',
              tab === 'login' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
            ]"
            @click="tab = 'login'"
          >
            Iniciar sesion
          </button>
          <button
            :class="[
              'flex-1 py-2.5 rounded-lg text-sm font-semibold transition-all',
              tab === 'register' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
            ]"
            @click="tab = 'register'"
          >
            Registrarse
          </button>
        </div>

        <!-- Login form -->
        <form v-if="tab === 'login'" @submit.prevent="submit" class="space-y-5">
          <div class="space-y-1.5">
            <label class="block text-sm font-semibold text-gray-700">Correo electronico</label>
            <input
              v-model="email"
              type="email"
              placeholder="correo@empresa.com"
              autocomplete="email"
              required
              class="w-full px-4 py-3 rounded-xl border border-gray-200 bg-white text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-orange-300 focus:border-orange-400 transition"
            />
          </div>
          <div class="space-y-1.5">
            <label class="block text-sm font-semibold text-gray-700">Contrasena</label>
            <div class="relative">
              <input
                v-model="password"
                :type="showPass ? 'text' : 'password'"
                placeholder="password"
                autocomplete="current-password"
                required
                class="w-full px-4 py-3 pr-12 rounded-xl border border-gray-200 bg-white text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-orange-300 focus:border-orange-400 transition"
              />
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors text-sm font-medium"
                @click="showPass = !showPass"
              >
                {{ showPass ? 'Ocultar' : 'Ver' }}
              </button>
            </div>
          </div>

          <div v-if="error" class="flex items-center gap-2 bg-red-50 border border-red-200 text-red-700 text-sm rounded-xl px-4 py-3">
            {{ error }}
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-orange-500 hover:bg-orange-600 disabled:opacity-50 text-white font-bold py-3 rounded-xl text-sm transition-all flex items-center justify-center gap-2 shadow-sm"
          >
            <span v-if="loading" class="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin" />
            {{ loading ? 'Entrando...' : 'Entrar a la plataforma' }}
          </button>

          <p class="text-center text-sm text-gray-500">
            No tienes cuenta?
            <button type="button" class="text-orange-500 font-bold hover:underline" @click="tab = 'register'">
              Registrate gratis
            </button>
          </p>
        </form>

        <!-- Register form -->
        <form v-if="tab === 'register'" @submit.prevent="register" class="space-y-5">
          <div class="space-y-1.5">
            <label class="block text-sm font-semibold text-gray-700">Nombre completo</label>
            <input
              v-model="regName"
              type="text"
              placeholder="Tu nombre completo"
              autocomplete="name"
              required
              class="w-full px-4 py-3 rounded-xl border border-gray-200 bg-white text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-orange-300 focus:border-orange-400 transition"
            />
          </div>
          <div class="space-y-1.5">
            <label class="block text-sm font-semibold text-gray-700">Correo electronico</label>
            <input
              v-model="regEmail"
              type="email"
              placeholder="correo@empresa.com"
              autocomplete="email"
              required
              class="w-full px-4 py-3 rounded-xl border border-gray-200 bg-white text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-orange-300 focus:border-orange-400 transition"
            />
          </div>
          <div class="space-y-1.5">
            <label class="block text-sm font-semibold text-gray-700">Contrasena</label>
            <input
              v-model="regPassword"
              type="password"
              placeholder="Minimo 6 caracteres"
              autocomplete="new-password"
              required
              minlength="6"
              class="w-full px-4 py-3 rounded-xl border border-gray-200 bg-white text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-orange-300 focus:border-orange-400 transition"
            />
          </div>

          <div class="space-y-1.5">
            <label class="block text-sm font-semibold text-gray-700">Tipo de cuenta</label>
            <div class="grid grid-cols-2 gap-3">
              <button
                type="button"
                :class="[
                  'flex flex-col items-center gap-1.5 p-4 rounded-xl border-2 text-center transition-all',
                  regRole === 'user'
                    ? 'border-orange-500 bg-orange-50 text-gray-900'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-orange-300'
                ]"
                @click="regRole = 'user'"
              >
                <span class="text-2xl">&#127891;</span>
                <strong class="text-sm font-bold">Estudiante</strong>
                <small class="text-xs text-gray-500 leading-tight">Accede y aprende a tu ritmo</small>
              </button>
              <button
                type="button"
                :class="[
                  'flex flex-col items-center gap-1.5 p-4 rounded-xl border-2 text-center transition-all',
                  regRole === 'instructor'
                    ? 'border-orange-500 bg-orange-50 text-gray-900'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-orange-300'
                ]"
                @click="regRole = 'instructor'"
              >
                <span class="text-2xl">&#127979;</span>
                <strong class="text-sm font-bold">Instructor</strong>
                <small class="text-xs text-gray-500 leading-tight">Crea y gestiona cursos</small>
              </button>
            </div>
          </div>

          <div v-if="regError" class="flex items-center gap-2 bg-red-50 border border-red-200 text-red-700 text-sm rounded-xl px-4 py-3">
            {{ regError }}
          </div>
          <div v-if="regSuccess" class="flex items-center gap-2 bg-emerald-50 border border-emerald-200 text-emerald-700 text-sm rounded-xl px-4 py-3">
            {{ regSuccess }}
          </div>

          <button
            type="submit"
            :disabled="regLoading"
            class="w-full bg-orange-500 hover:bg-orange-600 disabled:opacity-50 text-white font-bold py-3 rounded-xl text-sm transition-all flex items-center justify-center gap-2 shadow-sm"
          >
            <span v-if="regLoading" class="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin" />
            {{ regLoading ? 'Creando cuenta...' : 'Crear cuenta gratis' }}
          </button>

          <p class="text-center text-sm text-gray-500">
            Ya tienes cuenta?
            <button type="button" class="text-orange-500 font-bold hover:underline" @click="tab = 'login'">
              Inicia sesion
            </button>
          </p>
        </form>
      </div>
    </div>
  </div>
</template>
