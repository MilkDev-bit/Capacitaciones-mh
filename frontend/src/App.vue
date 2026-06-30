<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useTheme } from './composables/useTheme'
import CartDrawer from './components/CartDrawer.vue'

useTheme()
</script>

<template>
  <RouterView />
  <CartDrawer />
</template>
