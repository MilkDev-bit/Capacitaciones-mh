<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useTheme } from './composables/useTheme'
import CartDrawer from './components/CartDrawer.vue'
import DC3Modal from './components/DC3Modal.vue'
import BannerPrivacidad from './components/BannerPrivacidad.vue'
import ModalAviso from './components/ModalAviso.vue'
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
  <!-- Montado aquí y no en cada vista: los botones de constancia están
       repartidos por cinco pantallas y todas abren esta misma instancia. -->
  <DC3Modal />
  <!-- Aviso de privacidad. Dos piezas para dos situaciones distintas:
       el banner informa al visitante sin cuenta y no bloquea; el modal exige
       aceptación a quien ya tiene sesión, porque ahí sus datos ya se están
       tratando. -->
  <BannerPrivacidad />
  <ModalAviso />
</template>
