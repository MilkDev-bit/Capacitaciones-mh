<script setup lang="ts">
/**
 * Franja informativa del aviso de privacidad para quien navega sin cuenta.
 *
 * NO es el consentimiento que cuenta. Vive en localStorage, así que se borra al
 * limpiar el navegador y no viaja entre dispositivos: no sirve como prueba de
 * nada. El consentimiento con valor se guarda en la cuenta al registrarse.
 *
 * Su función es informar antes de que alguien empiece a navegar, y por eso no
 * bloquea la página: a un visitante que solo mira el catálogo no se le puede
 * exigir aceptar nada para leerlo. A quien se registra sí, porque ahí empieza
 * el tratamiento de sus datos.
 */
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { bannerAceptado, aceptarBanner } from '../utils/aviso'

const auth = useAuthStore()
const visible = ref(false)

onMounted(() => {
  // A quien tiene sesión no se le enseña: su aceptación se gestiona contra la
  // base, y dos avisos a la vez compiten entre ellos.
  if (auth.isLoggedIn) return
  visible.value = !bannerAceptado()
})

function aceptar() {
  aceptarBanner()
  visible.value = false
}
</script>

<template>
  <Transition name="bp-fade">
    <div v-if="visible" class="bp" role="region" aria-label="Aviso de privacidad">
      <p class="bp-texto">
        Usamos tus datos para darte acceso a los cursos y emitir tu constancia
        DC-3. Consulta el
        <RouterLink to="/privacidad" class="bp-enlace">aviso de privacidad</RouterLink>.
      </p>
      <div class="bp-acciones">
        <RouterLink to="/privacidad" class="btn btn-secondary btn-sm">Leerlo</RouterLink>
        <button class="btn btn-primary btn-sm" @click="aceptar">Entendido</button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.bp {
  position: fixed;
  z-index: 1300;
  right: var(--gutter, 16px);
  bottom: var(--gutter, 16px);
  /* En móvil ocupa el ancho: una tarjeta flotante estrecha con dos botones
   * queda por debajo del objetivo táctil. */
  left: var(--gutter, 16px);
  max-width: 420px;
  margin-left: auto;
  padding: 16px 18px;
  border: 1px solid var(--border);
  border-radius: 14px;
  background: var(--surface);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.18);
  /* Respeta la barra de gestos del móvil. */
  margin-bottom: env(safe-area-inset-bottom);
}

.bp-texto {
  margin: 0 0 12px;
  font-size: 0.86rem;
  line-height: 1.55;
  color: var(--text);
}

.bp-enlace { color: var(--brand); text-decoration: underline; }

.bp-acciones {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.bp-acciones > * { min-height: var(--touch-min, 44px); display: inline-flex; align-items: center; }

@media (max-width: 479px) {
  .bp-acciones > * { flex: 1 1 100%; justify-content: center; }
}

.bp-fade-enter-active, .bp-fade-leave-active { transition: opacity 0.25s ease, transform 0.25s ease; }
.bp-fade-enter-from, .bp-fade-leave-to { opacity: 0; transform: translateY(10px); }

@media (prefers-reduced-motion: reduce) {
  .bp-fade-enter-active, .bp-fade-leave-active { transition: none; }
}
</style>
