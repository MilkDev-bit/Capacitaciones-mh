<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { useSuscripcionStore, precioDesdeCentavos } from '../stores/suscripcion'
import { useAuthStore } from '../stores/auth'
import { toast } from '../utils/toast'
import api from '../api'

const cart = useCartStore()
const susc = useSuscripcionStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const loading = ref(false)

const fileUrl = (path?: string) => (path ? `${import.meta.env.VITE_API_URL || ''}${path}` : '')

const money = (v: number) =>
  new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN', minimumFractionDigits: 2 }).format(v)

const totalLugares = computed(() => cart.items.reduce((a, i) => a + i.cantidad, 0))
const hayCorporativas = computed(() => cart.items.some((i) => i.type === 'b2b_direct'))

/** Bloquea el scroll del fondo mientras el panel está abierto. */
watch(
  () => cart.isDrawerOpen,
  (abierto) => {
    document.body.style.overflow = abierto ? 'hidden' : ''
  }
)
// Si el componente se destruye con el panel abierto, el body se quedaría bloqueado.
onUnmounted(() => { document.body.style.overflow = '' })

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') cart.closeDrawer()
}

async function checkout() {
  if (!cart.items.length) return

  if (!auth.isLoggedIn) {
    cart.closeDrawer()
    toast.info('Inicia sesión o regístrate para completar tu compra.')
    router.push({ path: '/login', query: { redirect: route.fullPath, opencart: '1' } })
    return
  }

  loading.value = true
  try {
    const res = await api.post('/checkout-session-cart', {
      items: cart.items.map((i) => ({
        curso_id: i.curso_id,
        cantidad: i.cantidad,
        type: i.type,
      })),
      // Pantalla de confirmación propia: decide a dónde mandar al comprador
      // según lo que compró (curso individual vs. licencias corporativas).
      success_url: window.location.origin + '/checkout/exito?session_id={CHECKOUT_SESSION_ID}',
      cancel_url: window.location.origin + '/tienda',
    })
    window.location.href = res.data.url
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'No pudimos iniciar el pago')
  } finally {
    loading.value = false
  }
}

function seguirComprando() {
  cart.closeDrawer()
  if (route.path !== '/tienda' && route.path !== '/') router.push('/tienda')
}

// ── Suscripción como alternativa ────────────────────────────────────────────

onMounted(() => susc.cargarPlanes())

/** Plan individual mensual más barato. */
const planMasBarato = computed(() => {
  const mensuales = susc.planesIndividuales.filter((p) => p.intervalo === 'mes')
  const base = mensuales.length ? mensuales : susc.planesIndividuales
  return [...base].sort((a, b) => a.precio_centavos - b.precio_centavos)[0] || null
})
const precioPlanTexto = computed(() =>
  planMasBarato.value
    ? precioDesdeCentavos(planMasBarato.value.precio_centavos, planMasBarato.value.moneda)
    : ''
)

/**
 * Se sugiere el plan solo si el carrito ya supera su costo mensual y el usuario
 * aún no está suscrito. Así el aviso siempre le ahorra dinero de verdad.
 */
const convienePlan = computed(() => {
  if (susc.tieneSuscripcion || !planMasBarato.value) return false
  if (cart.items.length < 2) return false
  return cart.totalPrice * 100 > planMasBarato.value.precio_centavos
})

function verPlanes() {
  cart.closeDrawer()
  router.push('/planes')
}
</script>

<template>
  <Transition name="cart">
    <div
      v-if="cart.isDrawerOpen"
      class="overlay"
      role="dialog"
      aria-modal="true"
      aria-label="Carrito de compras"
      tabindex="-1"
      @click.self="cart.closeDrawer"
      @keydown="onKeydown"
    >
      <aside class="drawer">
        <!-- ── Encabezado ── -->
        <header class="head">
          <div>
            <h2 class="head__title">Tu carrito</h2>
            <p class="head__sub">
              <template v-if="cart.items.length">
                {{ totalLugares }} {{ totalLugares === 1 ? 'lugar' : 'lugares' }} ·
                {{ cart.items.length }} {{ cart.items.length === 1 ? 'curso' : 'cursos' }}
              </template>
              <template v-else>Todavía no has agregado nada</template>
            </p>
          </div>
          <button class="close" aria-label="Cerrar carrito" @click="cart.closeDrawer">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12" /></svg>
          </button>
        </header>

        <!-- ── Cuerpo ── -->
        <div class="body">
          <!-- Vacío -->
          <div v-if="!cart.items.length" class="empty">
            <div class="empty__icon">
              <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><circle cx="8" cy="21" r="1" /><circle cx="19" cy="21" r="1" /><path d="M2.05 2.05h2l2.66 12.42a2 2 0 0 0 2 1.58h9.78a2 2 0 0 0 1.95-1.57l1.65-7.43H5.12" /></svg>
            </div>
            <h3>Tu carrito está vacío</h3>
            <p>Explora el catálogo y agrega las capacitaciones que necesita tu equipo.</p>
            <button class="btn btn--solid" @click="seguirComprando">Ver catálogo</button>
          </div>

          <!-- Con artículos -->
          <TransitionGroup v-else tag="ul" name="row" class="items">
            <li v-for="(item, index) in cart.items" :key="item.curso_id + '-' + index" class="item">
              <div class="item__thumb">
                <img v-if="item.thumbnail" :src="fileUrl(item.thumbnail)" :alt="`Portada de ${item.title}`" />
                <span v-else class="item__thumb-letter">{{ item.title.charAt(0).toUpperCase() }}</span>
              </div>

              <div class="item__main">
                <div class="item__top">
                  <h3 class="item__title">{{ item.title }}</h3>
                  <button class="item__remove" :aria-label="`Quitar ${item.title}`" @click="cart.removeItem(index)">
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12" /></svg>
                  </button>
                </div>

                <!-- Tipo de compra: explicado, no solo etiquetado -->
                <div class="modo">
                  <div class="modo__switch" role="group" aria-label="Tipo de compra">
                    <button
                      :class="['modo__opt', item.type === 'b2c' && 'modo__opt--on']"
                      :aria-pressed="item.type === 'b2c'"
                      @click="cart.setTipo(index, 'b2c')"
                    >Para mí</button>
                    <button
                      :class="['modo__opt', item.type === 'b2b_direct' && 'modo__opt--on']"
                      :aria-pressed="item.type === 'b2b_direct'"
                      @click="cart.setTipo(index, 'b2b_direct')"
                    >Para mi equipo</button>
                  </div>
                  <p class="modo__hint">
                    <template v-if="item.type === 'b2b_direct'">
                      Compras {{ item.cantidad }} accesos y los repartes por correo.
                    </template>
                    <template v-else>Una inscripción a tu nombre.</template>
                  </p>
                </div>

                <!-- Cantidad -->
                <div v-if="item.type === 'b2b_direct'" class="qty">
                  <span class="qty__label">Lugares</span>
                  <div class="qty__ctrl">
                    <button
                      :disabled="item.cantidad <= 1"
                      aria-label="Quitar un lugar"
                      @click="cart.setCantidad(index, item.cantidad - 1)"
                    >−</button>
                    <input
                      type="number"
                      min="1"
                      max="500"
                      :value="item.cantidad"
                      :aria-label="`Lugares para ${item.title}`"
                      @change="cart.setCantidad(index, Number(($event.target as HTMLInputElement).value))"
                    />
                    <button
                      :disabled="item.cantidad >= 500"
                      aria-label="Agregar un lugar"
                      @click="cart.setCantidad(index, item.cantidad + 1)"
                    >+</button>
                  </div>
                </div>

                <!-- Precio desglosado: el total nunca aparece sin explicar de dónde sale -->
                <div class="precio">
                  <span v-if="item.cantidad > 1" class="precio__calc">
                    {{ money(item.precio) }} × {{ item.cantidad }}
                  </span>
                  <strong class="precio__total">{{ money(item.precio * item.cantidad) }}</strong>
                </div>
              </div>
            </li>
          </TransitionGroup>

          <p v-if="hayCorporativas" class="nota">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10" /><path d="M12 16v-4M12 8h.01" /></svg>
            Al terminar el pago recibirás los accesos por correo y podrás repartirlos
            entre tu equipo desde <strong>Mis Licencias</strong>.
          </p>
        </div>

        <!-- ── Pie ── -->
        <footer v-if="cart.items.length" class="foot">
          <!--
            Solo se sugiere el plan cuando el carrito ya cuesta más que un mes
            de suscripción. Ofrecerlo siempre sería ruido; ofrecerlo aquí es
            información útil justo antes de pagar.
          -->
          <div v-if="convienePlan" class="nudge">
            <div class="nudge__txt">
              Con una suscripción de <strong>{{ precioPlanTexto }}/mes</strong> tendrías
              estos {{ cart.items.length }} cursos y todo el catálogo.
            </div>
            <button class="nudge__btn" @click="verPlanes">Comparar planes</button>
          </div>
          <dl class="resumen">
            <div>
              <dt>Subtotal</dt>
              <dd>{{ money(cart.totalPrice) }}</dd>
            </div>
            <div class="resumen__nota">
              <dt>Impuestos y factura</dt>
              <dd>Se calculan en el pago</dd>
            </div>
            <div class="resumen__total">
              <dt>Total</dt>
              <dd>{{ money(cart.totalPrice) }}</dd>
            </div>
          </dl>

          <button class="btn btn--solid btn--block" :disabled="loading" @click="checkout">
            <template v-if="loading">Procesando…</template>
            <template v-else-if="!auth.isLoggedIn">Inicia sesión para pagar</template>
            <template v-else>
              Pagar {{ money(cart.totalPrice) }}
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M12 5l7 7-7 7" /></svg>
            </template>
          </button>

          <div class="foot__links">
            <button class="linkish" @click="seguirComprando">Seguir explorando</button>
            <span class="seguro">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><rect width="18" height="11" x="3" y="11" rx="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" /></svg>
              Pago seguro con Stripe
            </span>
          </div>
        </footer>
      </aside>
    </div>
  </Transition>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  justify-content: flex-end;
  background: rgba(0, 0, 0, 0.42);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
}

.drawer {
  width: 100%;
  max-width: 430px;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg);
  border-left: 1px solid var(--border);
  box-shadow: -12px 0 40px rgba(0, 0, 0, 0.16);
}

/* ── Transiciones ──────────────────────────────────────── */
.cart-enter-active, .cart-leave-active { transition: opacity 0.28s var(--ease-apple); }
.cart-enter-from, .cart-leave-to { opacity: 0; }
.cart-enter-active .drawer, .cart-leave-active .drawer { transition: transform 0.36s var(--ease-apple); }
.cart-enter-from .drawer, .cart-leave-to .drawer { transform: translateX(100%); }

.row-enter-active, .row-leave-active { transition: opacity 0.24s, transform 0.24s var(--ease-apple); }
.row-enter-from, .row-leave-to { opacity: 0; transform: translateX(22px); }
.row-leave-active { position: absolute; }

/* ── Encabezado ────────────────────────────────────────── */
.head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 20px 22px 16px;
  border-bottom: 1px solid var(--border);
  background: var(--surface);
}
.head__title { margin: 0; font-size: 1.15rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text); }
.head__sub { margin: 3px 0 0; font-size: 0.81rem; color: var(--muted); }

.close {
  flex-shrink: 0;
  width: 34px; height: 34px;
  display: grid; place-items: center;
  border: 1px solid var(--border); border-radius: 9px;
  background: var(--bg); color: var(--muted);
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
}
.close:hover { background: var(--surface-soft); color: var(--text); }

/* ── Cuerpo ────────────────────────────────────────────── */
.body { flex: 1; overflow-y: auto; padding: 18px 22px; }

.empty { text-align: center; padding: 52px 8px; }
.empty__icon {
  width: 66px; height: 66px; margin: 0 auto 18px;
  display: grid; place-items: center; border-radius: 50%;
  background: var(--brand-light); color: var(--brand);
}
.empty h3 { margin: 0 0 8px; font-size: 1.05rem; font-weight: 700; color: var(--text); }
.empty p { margin: 0 auto 22px; font-size: 0.87rem; line-height: 1.55; color: var(--muted); max-width: 30ch; }

.items { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 12px; position: relative; }

.item {
  display: flex; gap: 13px;
  padding: 13px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r);
  transition: border-color 0.2s;
}
.item:hover { border-color: var(--brand-border); }

.item__thumb {
  flex-shrink: 0;
  width: 62px; height: 62px;
  display: grid; place-items: center;
  border-radius: 10px;
  overflow: hidden;
  background: linear-gradient(140deg, var(--brand), var(--brand-darker));
}
.item__thumb img { width: 100%; height: 100%; object-fit: cover; }
.item__thumb-letter { font-size: 1.5rem; font-weight: 800; color: rgba(255, 255, 255, 0.92); }

.item__main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 9px; }

.item__top { display: flex; align-items: flex-start; gap: 8px; }
.item__title {
  flex: 1; margin: 0;
  font-size: 0.92rem; font-weight: 700; line-height: 1.35; color: var(--text);
  display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2;
  -webkit-box-orient: vertical; overflow: hidden;
}
.item__remove {
  flex-shrink: 0;
  width: 24px; height: 24px;
  display: grid; place-items: center;
  border: none; border-radius: 7px;
  background: transparent; color: var(--subtle);
  cursor: pointer;
  transition: background 0.18s, color 0.18s;
}
.item__remove:hover { background: var(--danger-bg); color: var(--danger); }

/* ── Modo de compra ────────────────────────────────────── */
.modo { display: flex; flex-direction: column; gap: 5px; }
.modo__switch {
  display: inline-flex; align-self: flex-start;
  padding: 2px;
  background: var(--surface-soft);
  border-radius: 9px;
}
.modo__opt {
  border: none; background: transparent;
  padding: 5px 11px; border-radius: 7px;
  font-size: 0.75rem; font-weight: 700; color: var(--muted);
  cursor: pointer; font-family: inherit;
  transition: background 0.2s, color 0.2s, box-shadow 0.2s;
}
.modo__opt:hover { color: var(--text); }
.modo__opt--on { background: var(--surface); color: var(--brand); box-shadow: var(--shadow-xs); }
.modo__hint { margin: 0; font-size: 0.74rem; line-height: 1.45; color: var(--subtle); }

/* ── Cantidad ──────────────────────────────────────────── */
.qty { display: flex; align-items: center; gap: 10px; }
.qty__label { font-size: 0.75rem; font-weight: 600; color: var(--muted); }
.qty__ctrl {
  display: inline-flex; align-items: center;
  border: 1px solid var(--border); border-radius: 9px;
  overflow: hidden; background: var(--bg);
}
.qty__ctrl button {
  width: 28px; height: 28px;
  border: none; background: transparent;
  color: var(--brand); font-size: 1rem; font-weight: 700; line-height: 1;
  cursor: pointer; transition: background 0.18s;
}
.qty__ctrl button:hover:not(:disabled) { background: var(--brand-light); }
.qty__ctrl button:disabled { color: var(--subtle); cursor: not-allowed; }
.qty__ctrl input {
  width: 46px; height: 28px;
  border: none; border-left: 1px solid var(--border); border-right: 1px solid var(--border);
  background: transparent; text-align: center;
  font-size: 0.83rem; font-weight: 700; color: var(--text); font-family: inherit;
  -moz-appearance: textfield; appearance: textfield;
}
.qty__ctrl input::-webkit-outer-spin-button,
.qty__ctrl input::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
.qty__ctrl input:focus { outline: none; background: var(--brand-light); }

/* ── Precio ────────────────────────────────────────────── */
.precio { display: flex; align-items: baseline; justify-content: space-between; gap: 8px; }
.precio__calc { font-size: 0.75rem; color: var(--subtle); }
.precio__total { margin-left: auto; font-size: 1rem; font-weight: 800; color: var(--text); letter-spacing: -0.02em; }

.nota {
  display: flex; align-items: flex-start; gap: 9px;
  margin: 16px 0 0; padding: 12px 14px;
  background: var(--info-bg); border-radius: var(--r-sm);
  font-size: 0.78rem; line-height: 1.5; color: var(--text);
}
.nota svg { flex-shrink: 0; margin-top: 1px; color: var(--info); }

/* ── Pie ───────────────────────────────────────────────── */
.foot {
  padding: 18px 22px 20px;
  border-top: 1px solid var(--border);
  background: var(--surface);
}

.resumen { margin: 0 0 16px; padding: 0; display: flex; flex-direction: column; gap: 8px; }
.resumen > div { display: flex; align-items: baseline; justify-content: space-between; gap: 10px; }
.resumen dt { margin: 0; font-size: 0.85rem; color: var(--muted); }
.resumen dd { margin: 0; font-size: 0.88rem; font-weight: 600; color: var(--text); }
.resumen__nota dt, .resumen__nota dd { font-size: 0.76rem; color: var(--subtle); font-weight: 500; }
.resumen__total {
  padding-top: 10px; margin-top: 2px;
  border-top: 1px solid var(--border);
}
.resumen__total dt { font-size: 0.95rem; font-weight: 700; color: var(--text); }
.resumen__total dd { font-size: 1.3rem; font-weight: 800; color: var(--text); letter-spacing: -0.03em; }

.btn {
  display: inline-flex; align-items: center; justify-content: center; gap: 8px;
  padding: 11px 20px; border: none; border-radius: 11px;
  font-size: 0.9rem; font-weight: 700; cursor: pointer; font-family: inherit;
  transition: background 0.2s, transform 0.2s var(--ease-apple), box-shadow 0.2s, opacity 0.2s;
}
.btn--solid { background: var(--brand); color: #fff; box-shadow: 0 6px 18px rgba(249, 115, 22, 0.28); }
.btn--solid:hover:not(:disabled) { background: var(--brand-dark); transform: translateY(-1px); }
.btn--solid:disabled { opacity: 0.6; cursor: not-allowed; transform: none; box-shadow: none; }
.btn--block { width: 100%; padding: 14px 20px; font-size: 0.97rem; }

.foot__links {
  display: flex; align-items: center; justify-content: space-between;
  gap: 12px; margin-top: 13px; flex-wrap: wrap;
}
.linkish {
  background: none; border: none; padding: 0;
  font-size: 0.81rem; font-weight: 600; color: var(--muted);
  cursor: pointer; font-family: inherit;
}
.linkish:hover { color: var(--text); text-decoration: underline; }
.seguro { display: inline-flex; align-items: center; gap: 5px; font-size: 0.76rem; color: var(--subtle); }

/* ── Responsivo ────────────────────────────────────────── */
@media (max-width: 480px) {
  .drawer { max-width: 100%; }
  .head, .body, .foot { padding-left: 16px; padding-right: 16px; }
  .item__thumb { width: 52px; height: 52px; }
}

@media (prefers-reduced-motion: reduce) {
  .cart-enter-active, .cart-leave-active,
  .cart-enter-active .drawer, .cart-leave-active .drawer,
  .row-enter-active, .row-leave-active,
  .btn, .item { transition: none; }
  .btn--solid:hover:not(:disabled) { transform: none; }
}

/* ── Sugerencia de suscripción ──────────────────────────── */
.nudge {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  flex-wrap: wrap;
  padding: 0.85rem 1rem;
  margin-bottom: 1rem;
  border-radius: var(--r-sm);
  background: var(--brand-light);
  border: 1px solid var(--brand-border);
}
.nudge__txt { flex: 1 1 190px; font-size: 0.86rem; line-height: 1.45; color: var(--text); }
.nudge__txt strong { font-weight: 750; }
.nudge__btn {
  flex: 0 0 auto;
  border: 0;
  border-radius: 999px;
  padding: 0.45rem 0.9rem;
  background: var(--brand);
  color: #fff;
  font-size: 0.82rem;
  font-weight: 650;
  cursor: pointer;
}
.nudge__btn:hover { filter: brightness(1.08); }
</style>
