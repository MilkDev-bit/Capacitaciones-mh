<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const props = defineProps<{
  role: 'user' | 'instructor' | 'admin'
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const auth = useAuthStore()

const brandLabel = computed(() => {
  if (props.role === 'instructor') return 'MH Instructor'
  if (props.role === 'admin') return 'MH Admin'
  return 'MH Aprende'
})

const navItems = computed(() => {
  if (props.role === 'instructor') {
    return [
      { to: '/instructor/capacitaciones', label: 'Mis cursos', icon: 'book' },
      { to: '/instructor/examenes', label: 'Exámenes', icon: 'clipboard' },
      { to: '/instructor/estudiantes', label: 'Estudiantes', icon: 'users' },
    ]
  }
  if (props.role === 'admin') {
    return [
      { to: '/admin/capacitaciones', label: 'Capacitaciones', icon: 'book' },
      { to: '/admin/examenes', label: 'Exámenes', icon: 'clipboard' },
      { to: '/admin/usuarios', label: 'Usuarios', icon: 'users' },
    ]
  }
  // user
  return [
    { to: '/usuario/dashboard', label: 'Dashboard', icon: 'home' },
    { to: '/usuario/capacitaciones', label: 'Mis cursos', icon: 'book' },
    { to: '/usuario/examenes', label: 'Mis exámenes', icon: 'clipboard' },
    { to: '/usuario/perfil', label: 'Mi perfil', icon: 'user' },
  ]
})

function close() {
  emit('close')
}
</script>

<template>
  <!-- Mobile overlay -->
  <div
    v-if="open"
    class="fixed inset-0 bg-black/50 z-20 lg:hidden"
    @click="close"
    aria-hidden="true"
  />

  <!-- Sidebar -->
  <aside
    :class="[
      'fixed top-0 left-0 h-full w-60 bg-gray-900 flex flex-col z-30 transition-transform duration-300',
      open ? 'translate-x-0' : '-translate-x-full',
      'lg:translate-x-0 lg:static lg:flex',
    ]"
    aria-label="Sidebar"
  >
    <!-- Brand -->
    <div class="flex items-center gap-3 px-5 py-5 border-b border-white/10">
      <div class="w-9 h-9 bg-brand rounded-xl flex items-center justify-center flex-shrink-0">
        <svg width="18" height="18" viewBox="0 0 44 44" fill="none">
          <path d="M10 34L22 12L34 34H10Z" fill="white"/>
        </svg>
      </div>
      <span class="font-extrabold text-white text-base tracking-tight">{{ brandLabel }}</span>
    </div>

    <!-- Nav -->
    <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        @click="close"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium text-gray-300 hover:bg-white/10 hover:text-white transition-all duration-150"
        active-class="!bg-brand/20 !text-brand border border-brand/30"
      >
        <!-- Icons -->
        <svg v-if="item.icon === 'home'" class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
        </svg>
        <svg v-if="item.icon === 'book'" class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
        </svg>
        <svg v-if="item.icon === 'clipboard'" class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
        </svg>
        <svg v-if="item.icon === 'users'" class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/>
        </svg>
        <svg v-if="item.icon === 'user'" class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
        </svg>
        {{ item.label }}
      </RouterLink>
    </nav>

    <!-- Footer: logout -->
    <div class="px-3 py-4 border-t border-white/10">
      <button
        @click="auth.logout()"
        class="flex items-center gap-3 w-full px-3 py-2.5 rounded-xl text-sm font-medium text-gray-400 hover:bg-red-500/10 hover:text-red-400 transition-all duration-150"
      >
        <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
        </svg>
        Cerrar sesión
      </button>
    </div>
  </aside>
</template>
