<template>
  <div>
    <!-- Overlay -->
    <div 
      class="cart-overlay" 
      v-if="cart.isDrawerOpen" 
      @click="cart.closeDrawer"
    ></div>

    <!-- Drawer -->
    <div 
      class="cart-drawer" 
      :class="{ 'is-open': cart.isDrawerOpen }"
    >
      <div class="cart-header">
        <h3>Tu Carrito ({{ cart.totalItems }})</h3>
        <button class="btn-close" @click="cart.closeDrawer">✕</button>
      </div>

      <div class="cart-body">
        <div v-if="cart.items.length === 0" class="empty-cart">
          <p>Tu carrito está vacío.</p>
        </div>

        <div v-else class="cart-items">
          <div v-for="(item, index) in cart.items" :key="index" class="cart-item">
            <div class="item-img" :style="{ backgroundImage: 'url(' + (item.thumbnail || '/placeholder.png') + ')' }"></div>
            <div class="item-info">
              <h4>{{ item.title }}</h4>
              <span class="item-type">{{ item.type === 'b2b_direct' ? 'Licencias Corporativas (' + item.cantidad + ')' : 'Inscripción Individual' }}</span>
              <p class="item-price">{{ formatPrice(item.precio * item.cantidad) }}</p>
            </div>
            <button class="btn-remove" @click="cart.removeItem(index)">✕</button>
          </div>
        </div>
      </div>

      <div class="cart-footer" v-if="cart.items.length > 0">
        <div class="cart-total">
          <span>Total</span>
          <strong>{{ formatPrice(cart.totalPrice) }}</strong>
        </div>
        <button 
          class="btn-checkout" 
          @click="checkout" 
          :disabled="loading"
        >
          {{ loading ? 'Procesando...' : 'Proceder al Pago Seguro' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useCartStore } from '../stores/cart'
import api from '../api'

const cart = useCartStore()
const loading = ref(false)

function formatPrice(val: number) {
  return new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN' }).format(val)
}

async function checkout() {
  if (cart.items.length === 0) return
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
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0,0,0,0.5);
  backdrop-filter: blur(4px);
  z-index: 9998;
}

.cart-drawer {
  position: fixed;
  top: 0;
  right: -400px;
  width: 400px;
  height: 100vh;
  background: var(--surface);
  border-left: 1px solid rgba(255,255,255,0.1);
  box-shadow: -5px 0 25px rgba(0,0,0,0.5);
  z-index: 9999;
  transition: right 0.3s cubic-bezier(0.25, 1, 0.5, 1);
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
  padding: 1.5rem;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.cart-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #fff;
}

.btn-close {
  background: none;
  border: none;
  color: #888;
  font-size: 1.2rem;
  cursor: pointer;
}

.btn-close:hover {
  color: #fff;
}

.cart-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.empty-cart {
  text-align: center;
  color: #888;
  margin-top: 2rem;
}

.cart-items {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.cart-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: rgba(255,255,255,0.03);
  padding: 10px;
  border-radius: 8px;
  position: relative;
}

.item-img {
  width: 60px;
  height: 60px;
  border-radius: 6px;
  background-size: cover;
  background-position: center;
}

.item-info {
  flex: 1;
}

.item-info h4 {
  margin: 0 0 4px 0;
  font-size: 0.95rem;
  color: #fff;
}

.item-type {
  font-size: 0.8rem;
  color: var(--brand);
}

.item-price {
  margin: 4px 0 0 0;
  font-weight: 600;
  color: #fff;
}

.btn-remove {
  position: absolute;
  top: 10px;
  right: 10px;
  background: none;
  border: none;
  color: #888;
  cursor: pointer;
}
.btn-remove:hover {
  color: #ef4444;
}

.cart-footer {
  padding: 1.5rem;
  border-top: 1px solid rgba(255,255,255,0.1);
  background: var(--surface);
}

.cart-total {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1rem;
  font-size: 1.1rem;
  color: #fff;
}

.btn-checkout {
  width: 100%;
  padding: 0.8rem;
  background: var(--brand);
  color: #111;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-checkout:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-2px);
}
.btn-checkout:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
