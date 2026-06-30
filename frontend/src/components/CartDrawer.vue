<template>
  <div class="cart-wrapper">
    <Transition name="cart">
      <div v-if="cart.isDrawerOpen" class="cart-overlay" @click.self="cart.closeDrawer">
        <div class="cart-drawer">
          <div class="cart-header">
            <h3>Tu Carrito</h3>
            <button class="btn-close" @click="cart.closeDrawer" aria-label="Cerrar carrito">
              <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="cart-body">
            <div v-if="cart.items.length === 0" class="empty-cart">
              <div class="empty-icon">
                <svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"/>
                </svg>
              </div>
              <p>Tu carrito está vacío</p>
              <span>Agrega cursos o licencias para comenzar.</span>
            </div>

            <TransitionGroup name="list" tag="div" v-else class="cart-items">
              <div v-for="(item, index) in cart.items" :key="item.curso_id + index" class="cart-item">
                <div class="item-img" :style="{ backgroundImage: 'url(' + (item.thumbnail ? fileUrl(item.thumbnail) : '') + ')' }">
                  <span v-if="!item.thumbnail" class="fallback-icon">
                    <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
                    </svg>
                  </span>
                </div>
                <div class="item-info">
                  <h4>{{ item.title }}</h4>
                  <span class="item-type">{{ item.type === 'b2b_direct' ? 'Licencias Corporativas' : 'Inscripción Individual' }} <span v-if="item.cantidad > 1" class="qty-badge">x{{ item.cantidad }} lugares</span></span>
                  <p class="item-price">{{ formatPrice(item.precio * item.cantidad) }}</p>
                </div>
                <button class="btn-remove" @click="cart.removeItem(index)" aria-label="Eliminar item">
                  <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                  </svg>
                </button>
              </div>
            </TransitionGroup>
          </div>

          <div class="cart-footer" v-if="cart.items.length > 0">
            <div class="cart-total">
              <span>Total a pagar</span>
              <strong>{{ formatPrice(cart.totalPrice) }}</strong>
            </div>
            <button 
              class="btn-checkout" 
              @click="checkout" 
              :disabled="loading"
            >
              <span class="btn-content">
                {{ loading ? 'Procesando...' : 'Proceder al Pago Seguro' }}
                <svg v-if="!loading" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M14 5l7 7m0 0l-7 7m7-7H3"/>
                </svg>
              </span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { useAuthStore } from '../stores/auth'
import { toast } from '../utils/toast'
import api from '../api'

const cart = useCartStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const loading = ref(false)

const fileUrl = (path: string) => `http://100.64.0.2:8080${path}`
const formatPrice = (p: number) => `$${p.toLocaleString('es-MX', { minimumFractionDigits: 2 })} MXN`

// Deshabilitar scroll del body cuando el carrito está abierto
watch(() => cart.isDrawerOpen, (isOpen) => {
  if (isOpen) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

// Cleanup
onMounted(() => {
  document.body.style.overflow = ''
})

async function checkout() {
  if (!auth.isLoggedIn) {
    cart.closeDrawer()
    toast.info('Inicia sesión o regístrate para continuar con la compra.')
    router.push('/login')
    return
  }

  loading.value = true
  try {
    const res = await api.post('/checkout-session-cart', {
      items: cart.items.map(i => ({
        curso_id: i.curso_id,
        cantidad: i.cantidad,
        type: i.type,
        schedule_id: i.schedule_id || ''
      })),
      success_url: window.location.origin + '/usuario/capacitaciones?session_id={CHECKOUT_SESSION_ID}',
      cancel_url: window.location.origin + '/tienda'
    })
    window.location.href = res.data.url
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al iniciar pago')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.cart-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4); backdrop-filter: blur(8px);
  z-index: 1000; display: flex; justify-content: flex-end;
}
.cart-overlay.cart-enter-active, .cart-overlay.cart-leave-active { transition: opacity 0.3s; }
.cart-overlay.cart-enter-from, .cart-overlay.cart-leave-to { opacity: 0; }

.cart-drawer {
  width: 100%; max-width: 420px;
  background: rgba(255, 255, 255, 0.85); backdrop-filter: blur(24px);
  box-shadow: -8px 0 32px rgba(0, 0, 0, 0.15);
  display: flex; flex-direction: column;
  transform: translateX(0); border-left: 1px solid rgba(255, 255, 255, 0.4);
}
.cart-enter-active .cart-drawer, .cart-leave-active .cart-drawer { transition: transform 0.4s cubic-bezier(0.16, 1, 0.3, 1); }
.cart-enter-from .cart-drawer, .cart-leave-to .cart-drawer { transform: translateX(100%); }

.cart-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 24px; border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}
.cart-header h3 { margin: 0; font-size: 1.25rem; font-weight: 700; color: #1f2937; }
.btn-close {
  background: rgba(0, 0, 0, 0.04); border: none; width: 36px; height: 36px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center; cursor: pointer; color: #4b5563;
  transition: all 0.2s ease;
}
.btn-close:hover { background: rgba(0, 0, 0, 0.08); transform: rotate(90deg); }

.cart-body {
  flex: 1; overflow-y: auto; padding: 24px;
}
.empty-cart {
  display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%;
  color: #6b7280; text-align: center; padding: 20px;
}
.empty-icon {
  background: rgba(249, 115, 22, 0.1); color: #f97316; width: 80px; height: 80px;
  border-radius: 50%; display: flex; align-items: center; justify-content: center; margin-bottom: 16px;
}
.empty-cart p { margin: 0; font-weight: 600; font-size: 1.1rem; color: #374151; }
.empty-cart span { font-size: 0.9rem; margin-top: 4px; }

.cart-items { display: flex; flex-direction: column; gap: 16px; }
.cart-item {
  display: flex; gap: 16px; align-items: center; background: rgba(255, 255, 255, 0.6);
  padding: 16px; border-radius: 16px; border: 1px solid rgba(255, 255, 255, 0.5);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.03); transition: transform 0.2s, box-shadow 0.2s; position: relative;
}
.cart-item:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06); }
.item-img {
  width: 72px; height: 72px; border-radius: 12px; background-size: cover; background-position: center;
  background-color: rgba(0, 0, 0, 0.03); display: flex; align-items: center; justify-content: center;
  flex-shrink: 0; box-shadow: inset 0 2px 8px rgba(0,0,0,0.05);
}
.fallback-icon { color: #9ca3af; opacity: 0.5; }

.item-info { flex: 1; display: flex; flex-direction: column; justify-content: center; }
.item-info h4 { margin: 0 0 4px 0; font-size: 1rem; color: #1f2937; line-height: 1.3; }
.item-type { font-size: 0.8rem; color: #f97316; font-weight: 600; margin-bottom: 6px; display: inline-flex; align-items: center; gap: 6px;}
.qty-badge { background: #f97316; color: #fff; padding: 2px 6px; border-radius: 4px; font-size: 0.7rem; font-weight: bold; }
.item-price { margin: 0; font-weight: 700; color: #111827; font-size: 1.05rem; }

.btn-remove {
  background: rgba(239, 68, 68, 0.1); border: none; color: #ef4444; width: 32px; height: 32px;
  border-radius: 50%; display: flex; align-items: center; justify-content: center; cursor: pointer;
  transition: all 0.2s; flex-shrink: 0;
}
.btn-remove:hover { background: #ef4444; color: #fff; transform: scale(1.1); }

.list-enter-active, .list-leave-active { transition: all 0.3s ease; }
.list-enter-from, .list-leave-to { opacity: 0; transform: translateX(30px); }

.cart-footer {
  padding: 24px; border-top: 1px solid rgba(0, 0, 0, 0.05);
  background: rgba(255, 255, 255, 0.5); backdrop-filter: blur(8px);
}
.cart-total { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.cart-total span { color: #6b7280; font-size: 1.1rem; }
.cart-total strong { font-size: 1.6rem; color: #111827; font-weight: 800; }

.btn-checkout {
  width: 100%; padding: 16px; border: none; border-radius: 12px;
  background: linear-gradient(135deg, #f97316 0%, #ea580c 100%);
  color: white; font-weight: 700; font-size: 1.1rem; cursor: pointer;
  transition: all 0.3s; position: relative; overflow: hidden;
  box-shadow: 0 8px 24px rgba(249, 115, 22, 0.25);
}
.btn-checkout:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 12px 28px rgba(249, 115, 22, 0.35); }
.btn-checkout:disabled { opacity: 0.7; cursor: not-allowed; transform: none; box-shadow: none; }
.btn-content { display: flex; align-items: center; justify-content: center; gap: 8px; position: relative; z-index: 2; }
</style>
