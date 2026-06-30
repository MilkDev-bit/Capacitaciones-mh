<template>
  <div>
    <!-- Overlay -->
    <transition name="fade">
      <div 
        class="cart-overlay" 
        v-if="cart.isDrawerOpen" 
        @click="cart.closeDrawer"
      ></div>
    </transition>

    <!-- Drawer -->
    <div 
      class="cart-drawer" 
      :class="{ 'is-open': cart.isDrawerOpen }"
    >
      <div class="cart-header">
        <div class="header-title">
          <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z"/>
          </svg>
          <h3>Tu Carrito ({{ cart.totalItems }})</h3>
        </div>
        <button class="btn-close" @click="cart.closeDrawer" aria-label="Cerrar carrito">
          <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
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
          <div v-for="(item, index) in cart.items" :key="item.curso_id + item.type" class="cart-item">
            <div class="item-img" :style="{ backgroundImage: 'url(' + (item.thumbnail ? fileUrl(item.thumbnail) : '') + ')' }">
              <span v-if="!item.thumbnail" class="fallback-icon">
                <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
                </svg>
              </span>
            </div>
            <div class="item-info">
              <h4>{{ item.title }}</h4>
              <span class="item-type">{{ item.type === 'b2b_direct' ? 'Licencias Corporativas' : 'Inscripción Individual' }} <span v-if="item.cantidad > 1" class="qty-badge">x{{ item.cantidad }}</span></span>
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

function checkOpenCart() {
  if (route.query.open_cart === '1') {
    cart.openDrawer()
    const query = { ...route.query }
    delete query.open_cart
    router.replace({ query })
  }
}

onMounted(() => {
  checkOpenCart()
})

watch(() => route.query.open_cart, (val) => {
  if (val === '1') checkOpenCart()
})

function fileUrl(path: string) {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${import.meta.env.VITE_API_URL || ''}${path}`
}

function formatPrice(val: number) {
  return new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN' }).format(val)
}

async function checkout() {
  if (cart.items.length === 0) return
  if (!auth.isLoggedIn) {
    cart.closeDrawer()
    toast.info('Por favor, inicia sesión o crea una cuenta para continuar con tu compra.')
    router.push({
      path: '/login',
      query: { redirect: route.fullPath.includes('?') ? `${route.fullPath}&open_cart=1` : `${route.fullPath}?open_cart=1` }
    })
    return
  }
  loading.value = true
  try {
    const payload = {
      items: cart.items.map(i => ({
        curso_id: i.curso_id,
        cantidad: i.cantidad,
        type: i.type,
        schedule_id: i.schedule_id || ''
      })),
      success_url: window.location.origin + '/usuario/capacitaciones?session_id={CHECKOUT_SESSION_ID}',
      cancel_url: window.location.href
    }
    const res = await api.post('/checkout-session-cart', payload)
    if (res.data?.url) {
      window.location.href = res.data.url
    } else {
      alert('No se pudo generar la sesión de pago')
    }
  } catch (e: any) {
    console.error('Error al hacer checkout:', e)
    alert(e.response?.data?.error || 'Error al procesar el pago.')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.cart-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  backdrop-filter: blur(8px);
  z-index: 9998;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.cart-drawer {
  position: fixed;
  top: 0;
  right: -450px;
  width: 100%;
  max-width: 420px;
  height: 100vh;
  background: rgba(23, 23, 23, 0.85);
  backdrop-filter: blur(20px);
  border-left: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: -10px 0 40px rgba(0, 0, 0, 0.5);
  z-index: 9999;
  transition: right 0.4s cubic-bezier(0.16, 1, 0.3, 1);
  display: flex;
  flex-direction: column;
}

.cart-drawer.is-open {
  right: 0;
}

.cart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.8rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: linear-gradient(180deg, rgba(255,255,255,0.03) 0%, transparent 100%);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fff;
}

.header-title h3 {
  margin: 0;
  font-size: 1.3rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.header-title svg {
  color: var(--brand);
}

.btn-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.05);
  border-radius: 50%;
  color: #a3a3a3;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-close:hover {
  background: rgba(255,255,255,0.1);
  color: #fff;
  transform: rotate(90deg);
}

.cart-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
}

.cart-body::-webkit-scrollbar {
  width: 6px;
}
.cart-body::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.1);
  border-radius: 10px;
}

.empty-cart {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: #a3a3a3;
  padding: 2rem;
  animation: fadeIn 0.5s ease;
}

.empty-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: rgba(255,255,255,0.03);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
  color: rgba(255,255,255,0.15);
}

.empty-cart p {
  font-size: 1.2rem;
  font-weight: 600;
  color: #e5e5e5;
  margin: 0 0 8px 0;
}

.empty-cart span {
  font-size: 0.9rem;
}

.cart-items {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
}

/* Animations for list */
.list-enter-active,
.list-leave-active {
  transition: all 0.4s ease;
}
.list-enter-from {
  opacity: 0;
  transform: translateX(30px);
}
.list-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}

.cart-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.05);
  padding: 12px;
  border-radius: 12px;
  position: relative;
  transition: all 0.25s ease;
}

.cart-item:hover {
  background: rgba(255,255,255,0.06);
  border-color: rgba(255,255,255,0.1);
  transform: translateY(-2px);
  box-shadow: 0 10px 20px -10px rgba(0,0,0,0.5);
}

.item-img {
  width: 72px;
  height: 72px;
  border-radius: 8px;
  background-color: rgba(249, 115, 22, 0.15);
  background-size: cover;
  background-position: center;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.item-img::after {
  content: '';
  position: absolute;
  inset: 0;
  box-shadow: inset 0 0 0 1px rgba(255,255,255,0.1);
  border-radius: 8px;
}

.fallback-icon {
  color: var(--brand);
  opacity: 0.8;
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-info h4 {
  margin: 0 0 4px 0;
  font-size: 1rem;
  color: #fff;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding-right: 24px;
}

.item-type {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  color: #a3a3a3;
  margin-bottom: 6px;
}

.qty-badge {
  background: rgba(249, 115, 22, 0.15);
  color: var(--brand);
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 700;
  font-size: 0.75rem;
}

.item-price {
  margin: 0;
  font-weight: 700;
  color: #fff;
  font-size: 1.1rem;
}

.btn-remove {
  position: absolute;
  top: 12px;
  right: 12px;
  background: rgba(0,0,0,0.2);
  border: none;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  color: #888;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  opacity: 0;
  transform: scale(0.9);
}

.cart-item:hover .btn-remove {
  opacity: 1;
  transform: scale(1);
}

.btn-remove:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.cart-footer {
  padding: 1.5rem 1.8rem 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0,0,0,0.2);
  backdrop-filter: blur(10px);
}

.cart-total {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 1.5rem;
  color: #a3a3a3;
}

.cart-total span {
  font-size: 0.95rem;
}

.cart-total strong {
  font-size: 1.8rem;
  color: #fff;
  line-height: 1;
}

.btn-checkout {
  width: 100%;
  padding: 1rem;
  background: linear-gradient(135deg, var(--brand) 0%, #fb923c 100%);
  color: #111;
  border: none;
  border-radius: 12px;
  font-weight: 700;
  font-size: 1.05rem;
  cursor: pointer;
  transition: all 0.25s ease;
  box-shadow: 0 4px 15px rgba(249, 115, 22, 0.25);
  position: relative;
  overflow: hidden;
}

.btn-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  position: relative;
  z-index: 2;
}

.btn-checkout::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: linear-gradient(135deg, #fb923c 0%, var(--brand) 100%);
  opacity: 0;
  transition: opacity 0.3s;
  z-index: 1;
}

.btn-checkout:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(249, 115, 22, 0.4);
}

.btn-checkout:hover:not(:disabled)::before {
  opacity: 1;
}

.btn-checkout:disabled {
  background: #333;
  color: #888;
  box-shadow: none;
  cursor: not-allowed;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
