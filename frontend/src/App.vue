<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useTheme } from './composables/useTheme'
import CartDrawer from './components/CartDrawer.vue'
import { watch } from 'vue'
import { useRoute } from 'vue-router'
import { useCartStore } from './stores/cart'

useTheme()
const route = useRoute()
const cart = useCartStore()

watch(() => route.query.opencart, (val) => {
  if (val === '1') {
    cart.openDrawer()
  }
}, { immediate: true })
</script>

<template>
  <RouterView />
  <CartDrawer />
</template>
