<script setup lang="ts">
import { ref } from 'vue'
import { RouterView } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import AppSidebar from '../../components/AppSidebar.vue'

const auth = useAuthStore()
const sidebarOpen = ref(false)

function initials(name: string) {
  return (name || 'I').split(' ').map((w: string) => w[0]).join('').toUpperCase().slice(0, 2)
}
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-gray-50">
    <AppSidebar role="instructor" :open="sidebarOpen" @close="sidebarOpen = false" />

    <div class="flex-1 flex flex-col min-w-0 lg:ml-0">
      <header class="bg-white border-b border-gray-200 flex items-center gap-4 px-4 lg:px-6 flex-shrink-0 z-10" style="height:60px">
        <button
          class="lg:hidden p-2 rounded-lg hover:bg-gray-100 text-gray-600 transition-colors"
          @click="sidebarOpen = true"
          aria-label="Abrir menu"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16"/>
          </svg>
        </button>

        <span class="font-bold text-gray-800 text-base hidden sm:block">Panel Instructor</span>

        <div class="ml-auto flex items-center gap-3">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-full bg-orange-500 text-white flex items-center justify-center text-xs font-bold flex-shrink-0">
              {{ initials(auth.user?.name || '') }}
            </div>
            <span class="hidden sm:block text-sm font-semibold text-gray-700">{{ auth.user?.name }}</span>
            <span class="hidden sm:inline-flex items-center px-2 py-0.5 rounded-full text-xs font-bold bg-orange-100 text-orange-700">
              Instructor
            </span>
          </div>
        </div>
      </header>

      <main class="flex-1 overflow-y-auto p-4 lg:p-8">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped></style>
