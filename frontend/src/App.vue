<script setup lang="ts">
import { computed, watch, onMounted } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useTheme } from './composables/useTheme'
import { useCartStore } from './stores/cart'
import { useWSStore } from './stores/ws'
import { useLlamadas } from './composables/useLlamadas'

import CartDrawer from './components/CartDrawer.vue'
import DC3Modal from './components/DC3Modal.vue'
import BannerPrivacidad from './components/BannerPrivacidad.vue'
import ModalAviso from './components/ModalAviso.vue'
import LlamadaTimbrando from './components/LlamadaTimbrando.vue'
import VideoCallModal from './components/VideoCallModal.vue'

useTheme()
const route = useRoute()
const router = useRouter()
const cart = useCartStore()
const wsStore = useWSStore()
const {
  estado,
  llamada,
  credenciales,
  restantes,
  nombreOtro,
  aceptar,
  rechazar,
  colgar,
  manejarEvento
} = useLlamadas(wsStore.enviar)
const timbrando = computed(() => estado.value === 'entrante' || estado.value === 'saliente')
const enLlamada = computed(() => estado.value === 'en_llamada')

onMounted(() => {
  wsStore.conectar()
  wsStore.onMensaje((data) => {
    manejarEvento(data)
  })
})

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

  <DC3Modal />
  <BannerPrivacidad />
  <ModalAviso />

  <LlamadaTimbrando v-if="timbrando && llamada && (estado === 'entrante' || estado === 'saliente')" :modo="estado"
    :nombre="nombreOtro" :is-group="llamada.is_group" :restantes="restantes" @aceptar="aceptar" @rechazar="rechazar"
    @colgar="colgar" />

  <VideoCallModal v-if="enLlamada && credenciales" :room-name="credenciales.sala" :user-name="nombreOtro"
    :domain="credenciales.dominio" :jwt="credenciales.token" @close="colgar" />
</template>