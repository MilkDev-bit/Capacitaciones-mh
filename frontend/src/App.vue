<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useTheme } from './composables/useTheme'
import CartDrawer from './components/CartDrawer.vue'
import { watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCartStore } from './stores/cart'

useTheme()
const route = useRoute()
const cart = useCartStore()

const router = useRouter()

/**
 * `?opencart=1` reabre el carrito al volver de iniciar sesión.
 *
 * La marca se retira de la URL en cuanto se consume: si se quedara pegada, el
 * botón atrás del navegador y cualquier recarga volverían a abrir el panel, y
 * además acabaría copiada en enlaces compartidos.
 *
 * Se ignora dentro de /login: allí el panel solo taparía el formulario.
 */
watch(
  () => route.query.opencart,
  (val) => {
    if (val !== '1') return
    if (route.path.startsWith('/login')) return
    cart.openDrawer()
    const query = { ...route.query }
    delete query.opencart
    router.replace({ path: route.path, query, hash: route.hash })
  },
  { immediate: true }
)
</script>

<template>
  <RouterView />
  <CartDrawer />
</template>
