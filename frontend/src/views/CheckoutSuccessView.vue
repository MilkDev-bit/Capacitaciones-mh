<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { useCartStore } from '../stores/cart'
import { toast } from '../utils/toast'

interface ItemCompra {
  tipo: 'b2c' | 'b2b_direct'
  curso_id: string
  titulo: string
  curso_type: string
  cantidad: number
  licencia_id?: string
}

const route = useRoute()
const router = useRouter()
const cart = useCartStore()

// 'pendiente' es el pago en tienda (OXXO): la sesión se completó y la ficha
// existe, pero el dinero no ha entrado. No es un error y no debe pintarse como
// tal, o el comprador tira el voucher creyendo que la compra falló.
const estado = ref<'procesando' | 'listo' | 'pendiente' | 'error'>('procesando')
const items = ref<ItemCompra[]>([])
const total = ref(0)
const redirect = ref('/usuario/dashboard')
const etiqueta = ref('Ir a mi panel')
const mensajeError = ref('')

const hayLicencias = computed(() => items.value.some((i) => i.tipo === 'b2b_direct'))
const formatPrice = (p: number) => `$${p.toLocaleString('es-MX', { minimumFractionDigits: 2 })} MXN`

onMounted(async () => {
  const sessionId = route.query.session_id as string | undefined
  if (!sessionId) {
    estado.value = 'error'
    mensajeError.value = 'No recibimos la referencia del pago.'
    return
  }

  try {
    const res = await api.post('/verify-checkout-session', { session_id: sessionId })
    items.value = res.data.items || []
    total.value = res.data.total || 0
    redirect.value = res.data.redirect || '/usuario/dashboard'
    etiqueta.value = res.data.etiqueta || 'Continuar'
    estado.value = res.data.pendiente ? 'pendiente' : 'listo'
    // El carrito se vacía también con la ficha pendiente: los renglones ya
    // quedaron congelados en la orden, y dejarlos ahí llevaría a pagarlos dos veces.
    cart.clearCart()
  } catch (e: any) {
    // Si el webhook de Stripe ya procesó la compra, verify puede fallar por
    // duplicado: eso NO es un error para el usuario, su compra está hecha.
    const msg = e.response?.data?.error || ''
    if (e.response?.status === 402) {
      estado.value = 'error'
      mensajeError.value = 'El pago no se completó. No se realizó ningún cargo.'
      return
    }
    console.warn('verify-checkout-session:', msg)
    estado.value = 'listo'
    cart.clearCart()
  }
})

function continuar() {
  router.push(redirect.value)
}

async function copiarComprobante() {
  const detalle = items.value
    .map((i) => `${i.titulo} — ${i.tipo === 'b2b_direct' ? `${i.cantidad} licencias` : 'inscripción individual'}`)
    .join('\n')
  try {
    await navigator.clipboard.writeText(`${detalle}\nTotal: ${formatPrice(total.value)}`)
    toast.success('Detalle copiado')
  } catch {
    toast.error('No se pudo copiar')
  }
}
</script>

<template>
  <div class="success-page">
    <div class="success-card">
      <!-- Procesando -->
      <template v-if="estado === 'procesando'">
        <div class="spinner" />
        <h1>Confirmando tu pago…</h1>
        <p class="lead">Estamos activando tus accesos. No cierres esta ventana.</p>
      </template>

      <!-- Error -->
      <template v-else-if="estado === 'error'">
        <div class="icon-badge error">
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <circle cx="12" cy="12" r="10" /><path d="m15 9-6 6m0-6 6 6" />
          </svg>
        </div>
        <h1>No pudimos confirmar el pago</h1>
        <p class="lead">{{ mensajeError }}</p>
        <router-link to="/tienda" class="btn-primary">Volver a la tienda</router-link>
      </template>

      <!-- Pago en tienda pendiente (OXXO) -->
      <template v-else-if="estado === 'pendiente'">
        <div class="icon-badge pendiente">
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10" /><path d="M12 7v5l3 2" />
          </svg>
        </div>
        <h1>Tu ficha de pago está lista</h1>
        <p class="lead">
          Te la enviamos por correo con las instrucciones y el número de referencia.
          <br />Paga en cualquier tienda OXXO y activamos tu acceso automáticamente.
          La confirmación puede tardar hasta un día hábil después de pagar.
        </p>
        <p class="lead aviso">
          Guarda tu comprobante de la tienda. Los pagos en efectivo no admiten devolución.
        </p>
        <button class="btn-primary" @click="continuar">{{ etiqueta }}</button>
      </template>

      <!-- Éxito -->
      <template v-else>
        <div class="icon-badge">
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 6 9 17l-5-5" />
          </svg>
        </div>

        <h1>¡Pago confirmado!</h1>
        <p class="lead">
          Te enviamos el comprobante por correo.
          <template v-if="hayLicencias"><br />Los accesos de tus licencias también van en camino.</template>
        </p>

        <div v-if="items.length" class="detalle">
          <div v-for="(item, i) in items" :key="i" class="detalle-row">
            <div class="detalle-info">
              <span class="detalle-titulo">{{ item.titulo || 'Capacitación' }}</span>
              <span class="detalle-tipo">
                {{ item.tipo === 'b2b_direct' ? 'Licencias corporativas' : 'Inscripción individual' }}
                <template v-if="item.cantidad > 1"> · {{ item.cantidad }} lugares</template>
              </span>
            </div>
          </div>

          <div v-if="total > 0" class="detalle-total">
            <span>Total pagado</span>
            <strong>{{ formatPrice(total) }}</strong>
          </div>
        </div>

        <!-- Siguiente paso explícito: era lo que faltaba en el flujo anterior -->
        <div v-if="hayLicencias" class="siguiente-paso">
          <strong>Siguiente paso</strong>
          <p>
            Captura los correos de tu equipo y cada persona recibirá su acceso
            individual. También puedes repartir los códigos a mano.
          </p>
        </div>

        <button class="btn-primary" @click="continuar">{{ etiqueta }} →</button>

        <div class="acciones-secundarias">
          <button class="btn-link" @click="copiarComprobante">Copiar detalle</button>
          <router-link to="/tienda" class="btn-link">Seguir explorando</router-link>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.success-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: linear-gradient(160deg, #ecfdf5 0%, #f4f4f5 60%);
}

.success-card {
  width: 100%;
  max-width: 520px;
  background: #fff;
  border-radius: 24px;
  padding: 44px 40px 34px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.08);
  text-align: center;
}

.icon-badge {
  width: 62px; height: 62px; margin: 0 auto 22px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: rgba(16, 185, 129, 0.12); color: #10b981;
}
.icon-badge.error { background: rgba(239, 68, 68, 0.1); color: #ef4444; }
/* Ámbar y no rojo: la ficha pendiente es un paso normal del pago en tienda,
   no un fallo. El color es la primera señal que lee el comprador. */
.icon-badge.pendiente { background: rgba(245, 158, 11, 0.12); color: #f59e0b; }
.lead.aviso { font-weight: 600; color: #b45309; }

.spinner {
  width: 46px; height: 46px; margin: 0 auto 22px;
  border: 4px solid #e5e7eb; border-top-color: #f97316; border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

h1 { margin: 0 0 10px; font-size: 1.6rem; font-weight: 800; color: #111827; }
.lead { margin: 0 0 26px; color: #6b7280; font-size: 0.95rem; line-height: 1.6; }

.detalle {
  background: #f9fafb; border-radius: 16px; padding: 6px 18px 14px;
  text-align: left; margin-bottom: 22px;
}
.detalle-row {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
  padding: 14px 0; border-bottom: 1px solid #f3f4f6;
}
.detalle-row:last-of-type { border-bottom: none; }
.detalle-info { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.detalle-titulo { font-weight: 700; color: #111827; font-size: 0.95rem; }
.detalle-tipo { font-size: 0.8rem; color: #6b7280; }
.pill {
  background: rgba(99, 102, 241, 0.1); color: #6366f1; font-size: 0.7rem;
  font-weight: 700; padding: 4px 9px; border-radius: 6px; white-space: nowrap;
}
.detalle-total {
  display: flex; justify-content: space-between; align-items: center;
  padding-top: 14px; margin-top: 6px; border-top: 1px solid #e5e7eb;
}
.detalle-total span { color: #6b7280; font-size: 0.9rem; }
.detalle-total strong { font-size: 1.25rem; color: #111827; font-weight: 800; }

.siguiente-paso {
  background: #fff7ed; border-left: 4px solid #f97316;
  border-radius: 0 12px 12px 0; padding: 14px 18px;
  text-align: left; margin-bottom: 22px;
}
.siguiente-paso strong { display: block; color: #ea580c; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 5px; }
.siguiente-paso p { margin: 0; font-size: 0.87rem; color: #7c2d12; line-height: 1.55; }

.btn-primary {
  display: inline-block; width: 100%; padding: 15px; border: none; border-radius: 12px;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: #fff; font-size: 1rem; font-weight: 700; cursor: pointer; text-decoration: none;
  box-shadow: 0 8px 20px rgba(16, 185, 129, 0.25); transition: transform 0.2s, box-shadow 0.2s;
}
.btn-primary:hover { transform: translateY(-2px); box-shadow: 0 12px 26px rgba(16, 185, 129, 0.32); }

.acciones-secundarias { display: flex; gap: 20px; justify-content: center; margin-top: 18px; }
.btn-link {
  background: none; border: none; color: #6b7280; font-size: 0.85rem;
  cursor: pointer; text-decoration: none; padding: 0;
}
.btn-link:hover { color: #111827; text-decoration: underline; }

@media (max-width: 480px) {
  .success-card { padding: 34px 22px 26px; }
}
</style>
