<script setup lang="ts">
/**
 * Aviso de privacidad para quien ya tiene cuenta y no lo ha aceptado.
 *
 * Aparece cuando la versión guardada en su cuenta no coincide con la vigente:
 * o se registró antes de que esto existiera, o el texto cambió desde entonces.
 *
 * A diferencia del banner de visitantes, este SÍ bloquea. La razón no es
 * cosmética: con sesión iniciada la plataforma ya está tratando sus datos
 * —CURP, RFC, avance de cursos—, así que seguir usándola sin haber aceptado es
 * justo lo que el aviso debe evitar. Al visitante anónimo que solo mira el
 * catálogo no se le exige nada.
 *
 * Su aceptación se guarda en la BASE, no en el navegador: es la que sirve como
 * constancia si alguien reclama.
 */
import { ref, computed, watch } from 'vue'
import api from '../api'
import { useAuthStore } from '../stores/auth'
import { toast } from '../utils/toast'
import { AVISO_VERSION, usuarioDebeAceptar } from '../utils/aviso'
import { useScrollLock } from '../composables/useScrollLock'

const auth = useAuthStore()
const guardando = ref(false)
const marcado = ref(false)

/** Versión que el backend dice que aceptó este usuario. */
const versionUsuario = ref<string | null>(null)
const consultado = ref(false)

const visible = computed(() =>
  auth.isLoggedIn && consultado.value && usuarioDebeAceptar(versionUsuario.value)
)

useScrollLock(visible)

async function consultar() {
  if (!auth.isLoggedIn) {
    consultado.value = false
    return
  }
  try {
    const { data } = await api.get('/perfil')
    versionUsuario.value = data?.aviso_version ?? ''
  } catch {
    // Si no se puede consultar, NO se bloquea la aplicación. Un fallo de red no
    // debe dejar a nadie sin poder usar su cuenta; se volverá a preguntar en la
    // siguiente carga.
    versionUsuario.value = AVISO_VERSION
  } finally {
    consultado.value = true
  }
}

watch(() => auth.isLoggedIn, consultar, { immediate: true })

async function aceptar() {
  if (!marcado.value) return
  guardando.value = true
  try {
    await api.post('/perfil/aviso', { version: AVISO_VERSION })
    versionUsuario.value = AVISO_VERSION
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos registrar tu aceptación')
  } finally {
    guardando.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="ma-fade">
      <!-- Sin cierre por clic fuera ni tecla Escape: aceptar o salir son las
           dos únicas salidas, y un cierre accidental dejaría a la persona
           usando la plataforma sin haber aceptado. -->
      <div v-if="visible" class="ma-overlay">
        <div class="ma-modal sheet-panel" role="dialog" aria-modal="true"
             aria-labelledby="ma-titulo">
          <h2 id="ma-titulo">Aviso de privacidad</h2>
          <p class="ma-texto">
            Actualizamos cómo tratamos tus datos. Para seguir usando la
            plataforma necesitamos que lo revises y lo aceptes.
          </p>
          <p class="ma-texto">
            Tratamos tu nombre, correo y —para emitir tu constancia DC-3— tu
            CURP, puesto y ocupación. Puedes leer el detalle completo:
          </p>

          <RouterLink to="/privacidad" target="_blank" class="ma-enlace">
            Leer el aviso de privacidad
          </RouterLink>

          <label class="ma-check">
            <input v-model="marcado" type="checkbox" />
            <span>He leído y acepto el aviso de privacidad.</span>
          </label>

          <div class="ma-acciones">
            <button class="btn btn-primary" :disabled="!marcado || guardando" @click="aceptar">
              {{ guardando ? 'Guardando…' : 'Aceptar y continuar' }}
            </button>
            <button class="btn btn-ghost" :disabled="guardando" @click="auth.logout()">
              Cerrar sesión
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.ma-overlay {
  position: fixed;
  inset: 0;
  z-index: 1400;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--gutter, 16px);
}

.ma-modal {
  width: 100%;
  max-width: 480px;
  padding: 24px;
  border-radius: 14px;
  background: var(--surface);
  border: 1px solid var(--border);
}

.ma-modal h2 { margin: 0 0 12px; font-size: 1.1rem; color: var(--text); }

.ma-texto {
  margin: 0 0 12px;
  font-size: 0.9rem;
  line-height: 1.6;
  color: var(--muted);
}

.ma-enlace {
  display: inline-block;
  margin-bottom: 18px;
  color: var(--brand);
  font-weight: 700;
  font-size: 0.9rem;
  text-decoration: underline;
}

.ma-check {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 20px;
  font-size: 0.9rem;
  color: var(--text);
  cursor: pointer;
}

.ma-check input { margin-top: 3px; flex-shrink: 0; }

.ma-acciones { display: flex; gap: 10px; flex-wrap: wrap; }
.ma-acciones > * { min-height: var(--touch-min, 44px); }

@media (max-width: 479px) {
  .ma-acciones { flex-direction: column; }
  .ma-acciones > * { width: 100%; }
}

.ma-fade-enter-active, .ma-fade-leave-active { transition: opacity 0.2s ease; }
.ma-fade-enter-from, .ma-fade-leave-to { opacity: 0; }

@media (prefers-reduced-motion: reduce) {
  .ma-fade-enter-active, .ma-fade-leave-active { transition: none; }
}
</style>
