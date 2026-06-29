<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { useAuthStore } from '../stores/auth'
import { toast } from '../utils/toast'
import logoSrc from '../assets/logo-capacitaciones.png'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const id = route.params.id as string
const curso = ref<any>(null)
const loading = ref(true)
const buying = ref(false)
const enrolling = ref(false)
const codigoForm = ref('')
const joiningCodigo = ref(false)

const showB2BModal = ref(false)
const b2bCantidad = ref(5)
const buyingB2B = ref(false)

const typeLabel: Record<string, string> = {
  video: 'Video',
  document: 'Documento',
  text: 'Lectura',
  link: 'Enlace',
}

const typeIcon: Record<string, string> = {
  video: '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>',
  document: '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>',
  text: '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>',
  link: '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>',
}
const defaultIcon = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20"/></svg>'

const typeGradient: Record<string, string> = {
  video:    'linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)',
  document: 'linear-gradient(135deg, #0d0d0d 0%, #1a1a1a 50%, #2d1b69 100%)',
  text:     'linear-gradient(135deg, #0f2027 0%, #203a43 50%, #2c5364 100%)',
  link:     'linear-gradient(135deg, #1a0533 0%, #2d1b69 50%, #11998e 100%)',
}

const formattedPrice = computed(() => {
  if (!curso.value?.precio) return null
  return new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN', maximumFractionDigits: 0 }).format(curso.value.precio)
})

function fileUrl(path: string) {
  return path ? `${import.meta.env.VITE_API_URL || ''}${path}` : ''
}

onMounted(async () => {
  try {
    const res = await api.get(`/cursos-publicos/${id}`)
    curso.value = res.data
  } catch (e: any) {
    if (e.response?.status === 403 || e.response?.status === 404) {
      toast.error('Este curso no está disponible públicamente')
      router.push('/tienda')
    } else {
      toast.error('Error al cargar el curso')
    }
  } finally {
    loading.value = false
  }
})

async function enrollFree() {
  if (!auth.isLoggedIn) {
    toast.info('Crea una cuenta o inicia sesión para guardar tu progreso')
    router.push(`/login?redirect=/curso/${id}`)
    return
  }
  enrolling.value = true
  try {
    await api.post(`/cursos/${id}/inscripciones`)
    toast.success('¡Te has inscrito correctamente!')
    router.push('/usuario/capacitaciones')
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Error al inscribirse')
  } finally {
    enrolling.value = false
  }
}

async function buyCourse() {
  if (!auth.isLoggedIn) {
    toast.info('Crea una cuenta o inicia sesión para continuar con la compra')
    router.push(`/login?redirect=/curso/${id}`)
    return
  }
  if (!curso.value) return
  buying.value = true
  try {
    const res = await api.post('/checkout-session', {
      curso_id: curso.value.id,
      success_url: window.location.origin + '/usuario/capacitaciones?session_id={CHECKOUT_SESSION_ID}&tipo=curso',
      cancel_url: window.location.href,
    })
    if (res.data?.url) {
      window.location.href = res.data.url
    } else {
      toast.error('No se pudo obtener el enlace de pago')
    }
  } catch (e: any) {
    const msg = e.response?.data?.error || ''
    if (e.response?.status === 500 || !msg) {
      toast.error('El servicio de pagos no está disponible en este momento. Intenta más tarde o contacta al soporte.')
    } else {
      toast.error(msg)
    }
  } finally {
    buying.value = false
  }
}

async function unirseConCodigo() {
  if (!auth.isLoggedIn) {
    toast.info('Inicia sesión para canjear tu código')
    router.push(`/login?redirect=/curso/${id}`)
    return
  }
  if (!codigoForm.value.trim()) return
  joiningCodigo.value = true
  try {
    await api.post('/inscripciones-licencia', {
      capacitacion_id: id,
      codigo_acceso: codigoForm.value.trim(),
    })
    toast.success('¡Inscrito correctamente a la cohorte!')
    router.push('/usuario/capacitaciones')
  } catch {
    try {
      await api.post('/inscripciones', { codigo: codigoForm.value.trim() })
      toast.success('¡Inscrito correctamente!')
      router.push('/usuario/capacitaciones')
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Código inválido o expirado')
    }
  } finally {
    joiningCodigo.value = false
  }
}

function openB2BModal() {
  if (!auth.isLoggedIn) {
    toast.info('Inicia sesión para comprar licencias corporativas')
    router.push(`/login?redirect=/curso/${id}`)
    return
  }
  b2bCantidad.value = 5 // Default
  showB2BModal.value = true
}

async function buyB2B() {
  if (!curso.value) return
  buyingB2B.value = true
  try {
    const res = await api.post('/checkout-session-b2b-direct', {
      curso_id: curso.value.id,
      cantidad: b2bCantidad.value,
      success_url: window.location.origin + '/usuario/licencias?session_id={CHECKOUT_SESSION_ID}&tipo=b2b',
      cancel_url: window.location.href,
    })
    if (res.data?.url) {
      window.location.href = res.data.url
    } else {
      toast.error('No se pudo obtener el enlace de pago corporativo')
    }
  } catch (e: any) {
    const msg = e.response?.data?.error || ''
    if (e.response?.status === 500 || !msg) {
      toast.error('El servicio de pagos no está disponible en este momento.')
    } else {
      toast.error(msg)
    }
  } finally {
    buyingB2B.value = false
  }
}
</script>

<template>
  <div class="cpv-root">
    <!-- Particles -->
    <div class="cpv-particles" aria-hidden="true">
      <span v-for="n in 12" :key="n" :class="`pp pp-${n}`"></span>
    </div>

    <!-- NAV -->
    <nav class="cpv-nav">
      <div class="cpv-nav-inner">
        <button class="cpv-brand" @click="router.push('/tienda')">
          <img :src="logoSrc" alt="MH Logo" class="cpv-logo" />
          <span>MH <span class="brand-orange">Capacitaciones</span></span>
        </button>
        <div class="cpv-nav-right">
          <button class="cpv-nav-btn ghost" @click="router.push('/tienda')">
            ← Catálogo
          </button>
          <template v-if="auth.isLoggedIn">
            <button class="cpv-nav-btn primary" @click="router.push(auth.user?.role === 'instructor' ? '/instructor' : '/usuario')">
              Mi Panel
            </button>
          </template>
          <template v-else>
            <button class="cpv-nav-btn ghost" @click="router.push('/login')">Iniciar sesión</button>
            <button class="cpv-nav-btn primary" @click="router.push('/login')">Registrarse</button>
          </template>
        </div>
      </div>
    </nav>

    <!-- LOADING -->
    <div v-if="loading" class="cpv-loading">
      <div class="cpv-spinner"></div>
      <p>Cargando curso…</p>
    </div>

    <!-- CONTENT -->
    <main v-else-if="curso" class="cpv-main">

      <!-- ── HERO ── -->
      <section class="cpv-hero">
        <!-- Background image or gradient -->
        <div class="cpv-hero-bg">
          <img
            v-if="curso.thumbnail_url"
            :src="fileUrl(curso.thumbnail_url)"
            alt="Portada del curso"
            class="cpv-hero-img"
          />
          <div
            v-else
            class="cpv-hero-placeholder"
            :style="{ background: typeGradient[curso.type] || typeGradient['video'] }"
          ></div>
          <div class="cpv-hero-overlay"></div>
        </div>

        <!-- Hero content -->
        <div class="cpv-hero-content">
          <div class="cpv-hero-chips">
            <span class="cpv-chip type">
              <span class="inline-svg" v-html="typeIcon[curso.type] || defaultIcon"></span>
              {{ typeLabel[curso.type] || 'Curso' }}
            </span>
            <span v-if="!curso.precio || curso.precio === 0" class="cpv-chip free">Gratis</span>
            <span v-else class="cpv-chip paid">Premium</span>
          </div>
          <h1 class="cpv-title">{{ curso.title }}</h1>
          <p class="cpv-subtitle">{{ curso.description || 'Desarrolla tus habilidades profesionales con este curso de expertos.' }}</p>

          <!-- Mobile CTA -->
          <div class="cpv-mobile-cta">
            <button v-if="curso.precio > 0" class="cpv-btn-buy" @click="buyCourse" :disabled="buying">
              <span v-if="buying" class="cpv-btn-spinner"></span>
              <span v-else>{{ formattedPrice }} — Comprar ahora</span>
            </button>
            <button v-else class="cpv-btn-free" @click="enrollFree" :disabled="enrolling">
              <span v-if="enrolling" class="cpv-btn-spinner"></span>
              <span v-else>Inscribirse gratis</span>
            </button>
          </div>
        </div>
      </section>

      <!-- ── BODY GRID ── -->
      <div class="cpv-body">

        <!-- LEFT COL -->
        <div class="cpv-left">

          <!-- About card -->
          <div class="glass-card">
            <h2 class="glass-card-title">Sobre este curso</h2>
            <p class="glass-card-text">{{ curso.welcome_message || curso.description || 'Aprende y desarrolla nuevas habilidades profesionales con este curso diseñado por expertos en la materia.' }}</p>

            <!-- What you'll learn -->
            <div class="learn-grid">
              <div class="learn-item">
                <span class="learn-icon">
                  <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M20 6L9 17l-5-5"/></svg>
                </span>
                <span>Contenido actualizado al {{ new Date().getFullYear() }}</span>
              </div>
              <div class="learn-item">
                <span class="learn-icon">
                  <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><path d="M12 18h.01"/></svg>
                </span>
                <span>Acceso desde cualquier dispositivo</span>
              </div>
              <div class="learn-item">
                <span class="learn-icon">
                  <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="8" r="7"/><polyline points="8.21 13.89 7 23 12 20 17 23 15.79 13.88"/></svg>
                </span>
                <span>Certificado al completar</span>
              </div>
            </div>
          </div>

          <!-- Code section -->
          <div class="glass-card mt-4">
            <h2 class="glass-card-title">
              <svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="flex-shrink:0; color: #fb923c;"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg>
              ¿Tienes un código de acceso?
            </h2>
            <p class="glass-card-text">
              Si tu empresa adquirió una licencia corporativa o tienes una invitación, canjéala aquí para acceder al curso.
            </p>
            <div class="code-row">
              <div class="code-input-wrap">
                <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" class="code-icon">
                  <path d="M21 2H3a1 1 0 00-1 1v11a1 1 0 001 1h3.07l-1.03 5.48L12 17h9a1 1 0 001-1V3a1 1 0 00-1-1z"/>
                </svg>
                <input
                  v-model="codigoForm"
                  class="code-input"
                  placeholder="CORP-XXXX-XXXX"
                  @keydown.enter="unirseConCodigo"
                />
              </div>
              <button class="cpv-btn-code" @click="unirseConCodigo" :disabled="joiningCodigo || !codigoForm.trim()">
                <span v-if="joiningCodigo" class="cpv-btn-spinner small"></span>
                <span v-else>Canjear</span>
              </button>
            </div>
          </div>

        </div>

        <!-- RIGHT COL -->
        <aside class="cpv-right">
          <div class="purchase-glass" :class="{ paid: curso.precio > 0 }">

            <!-- Glow orb -->
            <div class="purchase-orb" :class="{ 'orb-green': !curso.precio || curso.precio === 0 }"></div>

            <div v-if="curso.precio > 0">
              <div class="purchase-label">Precio del curso</div>
              <div class="purchase-price">{{ formattedPrice }}</div>
              <div class="purchase-period">Pago único</div>

              <button class="cpv-btn-buy w-full mt-4" @click="buyCourse" :disabled="buying">
                <span v-if="buying" class="cpv-btn-spinner"></span>
                <template v-else>
                  <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                    <path d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"/>
                  </svg>
                  Comprar ahora — {{ formattedPrice }}
                </template>
              </button>

              <div class="purchase-badges">
                <span class="badge"><svg class="inline-svg" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg> Pago seguro</span>
                <span class="badge"><svg class="inline-svg" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"></polyline><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"></path></svg> Garantía 30 días</span>
              </div>

              <button class="cpv-btn-b2b w-full mt-3" @click="openB2BModal">
                <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24" style="margin-right:8px;">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
                Comprar Licencias Corporativas
              </button>
            </div>

            <div v-else>
              <div class="purchase-label free-label">Acceso Gratuito</div>
              <div class="purchase-price free-price">GRATIS</div>
              <div class="purchase-period">Sin costo</div>

              <button class="cpv-btn-free w-full mt-4" @click="enrollFree" :disabled="enrolling">
                <span v-if="enrolling" class="cpv-btn-spinner"></span>
                <template v-else>
                  <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                    <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"/>
                    <path d="M9 12l2 2 4-4"/>
                  </svg>
                  Inscribirse gratis
                </template>
              </button>

              <div class="purchase-badges">
                <span class="badge"><svg class="inline-svg" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 12H5M12 19l-7-7 7-7"></path></svg> Sin tarjeta</span>
                <span class="badge"><svg class="inline-svg" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg> Acceso inmediato</span>
              </div>
            </div>

            <!-- Modal B2B -->
            <div v-if="showB2BModal" class="b2b-modal-overlay" @click.self="showB2BModal = false">
              <div class="b2b-modal-content glass-card">
                <h3 class="glass-card-title">Comprar Licencias Corporativas</h3>
                <p class="glass-card-text" style="margin-bottom:1rem;">
                  ¿Cuántas licencias necesitas para tu equipo? El precio por licencia es de {{ formattedPrice }}.
                </p>
                
                <div class="b2b-input-group">
                  <label>Cantidad de licencias</label>
                  <input type="number" v-model.number="b2bCantidad" min="2" max="1000" class="code-input" />
                </div>
                
                <div class="b2b-total" v-if="b2bCantidad >= 1 && curso.precio">
                  Total a pagar: <strong>{{ new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN', maximumFractionDigits: 0 }).format(curso.precio * b2bCantidad) }}</strong>
                </div>

                <div class="b2b-actions" style="margin-top:1.5rem; display:flex; gap:10px;">
                  <button class="cpv-btn-cancel" @click="showB2BModal = false" style="flex:1;">Cancelar</button>
                  <button class="cpv-btn-buy" @click="buyB2B" :disabled="buyingB2B || b2bCantidad < 1" style="flex:1;">
                    <span v-if="buyingB2B" class="cpv-btn-spinner"></span>
                    <span v-else>Comprar {{ b2bCantidad }} licencias</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- Includes list -->
            <div class="purchase-includes">
              <div class="includes-title">Este curso incluye:</div>
              <div class="include-item">
                <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="flex-shrink:0"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
                <span>Contenido multimedia</span>
              </div>
              <div class="include-item">
                <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="flex-shrink:0"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><path d="M12 18h.01"/></svg>
                <span>Acceso móvil y escritorio</span>
              </div>
              <div class="include-item">
                <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="flex-shrink:0"><circle cx="12" cy="8" r="7"/><polyline points="8.21 13.89 7 23 12 20 17 23 15.79 13.88"/></svg>
                <span>Certificado de finalización</span>
              </div>
            </div>
          </div>
        </aside>

      </div>
    </main>

    <!-- 404 state -->
    <div v-else class="cpv-loading">
      <div class="empty-icon glass-icon-xl">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
      </div>
      <h2 style="color: #fff;">Curso no encontrado</h2>
      <p style="color: rgba(255,255,255,0.5);">El curso que buscas no existe o no está disponible.</p>
      <button class="cpv-nav-btn primary mt-4" style="margin-top: 24px;" @click="router.push('/tienda')">Ir al catálogo</button>
    </div>

  </div>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800;900&display=swap');

/* ── ROOT ──────────────────────────────────────────────── */
.cpv-root {
  min-height: 100vh;
  background: #08090a;
  color: #f1f5f9;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  -webkit-font-smoothing: antialiased;
  position: relative;
  overflow-x: hidden;
}

/* ── PARTICLES ─────────────────────────────────────────── */
.cpv-particles {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}

.pp {
  position: absolute;
  border-radius: 50%;
  opacity: 0;
  animation: pp-rise linear infinite;
}

.pp-1  { width: 3px;  height: 3px;  background: rgba(249,115,22,.8); left:5%;   animation-duration:12s; animation-delay:0s;  }
.pp-2  { width: 2px;  height: 2px;  background: rgba(255,255,255,.5);left:15%;  animation-duration:16s; animation-delay:3s;  }
.pp-3  { width: 4px;  height: 4px;  background: rgba(167,139,250,.6);left:25%;  animation-duration:10s; animation-delay:6s;  }
.pp-4  { width: 2px;  height: 2px;  background: rgba(249,115,22,.6); left:35%;  animation-duration:18s; animation-delay:1s;  }
.pp-5  { width: 3px;  height: 3px;  background: rgba(52,211,153,.7); left:48%;  animation-duration:14s; animation-delay:8s;  }
.pp-6  { width: 2px;  height: 2px;  background: rgba(255,255,255,.4);left:58%;  animation-duration:11s; animation-delay:4s;  }
.pp-7  { width: 4px;  height: 4px;  background: rgba(249,115,22,.5); left:68%;  animation-duration:19s; animation-delay:7s;  }
.pp-8  { width: 3px;  height: 3px;  background: rgba(167,139,250,.5);left:78%;  animation-duration:13s; animation-delay:2s;  }
.pp-9  { width: 2px;  height: 2px;  background: rgba(52,211,153,.6); left:88%;  animation-duration:15s; animation-delay:9s;  }
.pp-10 { width: 5px;  height: 5px;  background: rgba(249,115,22,.3); left:10%;  animation-duration:20s; animation-delay:11s; }
.pp-11 { width: 2px;  height: 2px;  background: rgba(255,255,255,.3);left:50%;  animation-duration:8s;  animation-delay:5s;  }
.pp-12 { width: 3px;  height: 3px;  background: rgba(167,139,250,.7);left:92%;  animation-duration:17s; animation-delay:13s; }

@keyframes pp-rise {
  0%   { transform: translateY(100vh) scale(0); opacity: 0; }
  10%  { opacity: 1; }
  90%  { opacity: 0.5; }
  100% { transform: translateY(-100px) scale(1.3) rotate(180deg); opacity: 0; }
}

/* ── NAV ───────────────────────────────────────────────── */
.cpv-nav {
  position: fixed;
  top: 0; left: 0; right: 0;
  z-index: 100;
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  background: rgba(8,9,10,0.75);
  border-bottom: 1px solid rgba(255,255,255,0.07);
}

.cpv-nav-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 32px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.cpv-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  background: none;
  border: none;
  color: #f1f5f9;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  letter-spacing: -0.01em;
  transition: opacity 0.2s;
}
.cpv-brand:hover { opacity: 0.8; }

.cpv-logo {
  width: 32px; height: 32px;
  object-fit: contain;
  filter: drop-shadow(0 0 8px rgba(249,115,22,0.6));
  animation: logo-glow 3s ease-in-out infinite;
}

@keyframes logo-glow {
  0%,100% { filter: drop-shadow(0 0 6px rgba(249,115,22,.5)); }
  50%      { filter: drop-shadow(0 0 14px rgba(249,115,22,.9)); }
}

.brand-orange { color: #fb923c; }

.cpv-nav-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cpv-nav-btn {
  padding: 7px 16px;
  border-radius: 9999px;
  font-size: 0.84rem;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.2s;
}

.cpv-nav-btn.ghost {
  background: transparent;
  border: 1px solid rgba(255,255,255,0.14);
  color: rgba(255,255,255,0.75);
}
.cpv-nav-btn.ghost:hover {
  border-color: rgba(255,255,255,0.3);
  color: #fff;
}

.cpv-nav-btn.primary {
  background: #f97316;
  border: 1px solid rgba(249,115,22,0.5);
  color: #fff;
  box-shadow: 0 0 14px rgba(249,115,22,0.3);
}
.cpv-nav-btn.primary:hover {
  background: #ea580c;
  box-shadow: 0 0 22px rgba(249,115,22,0.5);
}

/* ── LOADING ───────────────────────────────────────────── */
.cpv-loading {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: rgba(255,255,255,0.5);
  font-size: 1rem;
  position: relative;
  z-index: 1;
}

.cpv-spinner {
  width: 44px; height: 44px;
  border-radius: 50%;
  border: 3px solid rgba(249,115,22,.2);
  border-top-color: #f97316;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── HERO ──────────────────────────────────────────────── */
.cpv-hero {
  position: relative;
  min-height: 520px;
  display: flex;
  align-items: flex-end;
  overflow: hidden;
  padding-top: 60px;
}

.cpv-hero-bg {
  position: absolute;
  inset: 0;
}

.cpv-hero-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  animation: hero-zoom 12s ease-in-out infinite alternate;
}

@keyframes hero-zoom {
  from { transform: scale(1); }
  to   { transform: scale(1.06); }
}

.cpv-hero-placeholder {
  width: 100%;
  height: 100%;
  animation: bg-shift 8s ease-in-out infinite alternate;
}

@keyframes bg-shift {
  from { filter: brightness(0.9); }
  to   { filter: brightness(1.1); }
}

.cpv-hero-overlay {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(to top, rgba(8,9,10,1) 0%, rgba(8,9,10,0.7) 40%, rgba(8,9,10,0.25) 100%),
    linear-gradient(to right, rgba(8,9,10,0.5) 0%, transparent 60%);
}

.cpv-hero-content {
  position: relative;
  z-index: 2;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  padding: 0 32px 56px;
  animation: cpv-fade-up 0.8s ease both;
}

@keyframes cpv-fade-up {
  from { opacity: 0; transform: translateY(32px); }
  to   { opacity: 1; transform: translateY(0); }
}

.cpv-hero-chips {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.cpv-chip {
  display: inline-block;
  padding: 4px 13px;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 700;
  backdrop-filter: blur(8px);
  letter-spacing: 0.02em;
}

.cpv-chip.type {
  background: rgba(255,255,255,0.1);
  border: 1px solid rgba(255,255,255,0.18);
  color: rgba(255,255,255,0.85);
}

.cpv-chip.free {
  background: rgba(16,185,129,0.2);
  border: 1px solid rgba(16,185,129,0.4);
  color: #34d399;
}

.cpv-chip.paid {
  background: rgba(249,115,22,0.2);
  border: 1px solid rgba(249,115,22,0.4);
  color: #fb923c;
}

.cpv-title {
  font-size: clamp(2rem, 5vw, 3.5rem);
  font-weight: 900;
  line-height: 1.1;
  letter-spacing: -0.03em;
  color: #fff;
  margin: 0 0 16px 0;
  max-width: 700px;
  text-shadow: 0 2px 20px rgba(0,0,0,0.5);
}

.cpv-subtitle {
  font-size: 1.05rem;
  color: rgba(255,255,255,0.7);
  line-height: 1.6;
  max-width: 600px;
  margin: 0 0 24px 0;
}

/* Mobile CTA (hidden on desktop) */
.cpv-mobile-cta { display: none; }

/* ── BODY ──────────────────────────────────────────────── */
.cpv-body {
  max-width: 1200px;
  margin: 0 auto;
  padding: 40px 32px 80px;
  display: grid;
  grid-template-columns: 1fr 380px;
  gap: 32px;
  align-items: start;
  position: relative;
  z-index: 2;
  animation: cpv-fade-up 0.8s ease both 0.2s;
}

/* ── GLASS CARD ────────────────────────────────────────── */
.glass-card {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.09);
  border-radius: 20px;
  padding: 28px;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  transition: border-color 0.3s;
}

.glass-card:hover {
  border-color: rgba(255,255,255,0.14);
}

.glass-card-title {
  font-size: 1.2rem;
  font-weight: 800;
  color: rgba(255,255,255,0.95);
  margin: 0 0 12px 0;
  letter-spacing: -0.01em;
  display: flex;
  align-items: center;
  gap: 8px;
}

.glass-card-text {
  font-size: 0.95rem;
  color: rgba(255,255,255,0.55);
  line-height: 1.7;
  margin: 0 0 20px 0;
}

.mt-4 { margin-top: 20px; }

/* Learn grid */
.learn-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.learn-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 10px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.07);
  font-size: 0.85rem;
  color: rgba(255,255,255,0.7);
}

.learn-icon { font-size: 1rem; }

/* Code section */
.code-row {
  display: flex;
  gap: 10px;
}

.code-input-wrap {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.code-icon {
  position: absolute;
  left: 14px;
  color: rgba(255,255,255,0.3);
  flex-shrink: 0;
}

.code-input {
  width: 100%;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 12px;
  padding: 12px 14px 12px 38px;
  font-size: 0.95rem;
  font-family: inherit;
  color: #fff;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  caret-color: #f97316;
}

.code-input::placeholder { color: rgba(255,255,255,0.3); }

.code-input:focus {
  border-color: rgba(249,115,22,0.5);
  box-shadow: 0 0 0 3px rgba(249,115,22,0.12);
}

/* ── PURCHASE GLASS ────────────────────────────────────── */
.purchase-glass {
  position: sticky;
  top: 80px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.09);
  border-radius: 24px;
  padding: 28px;
  backdrop-filter: blur(32px);
  -webkit-backdrop-filter: blur(32px);
  overflow: hidden;
  animation: cpv-fade-up 0.8s ease both 0.35s;
}

.purchase-glass.paid {
  border-color: rgba(249,115,22,0.18);
  background: rgba(249,115,22,0.04);
}

.purchase-orb {
  position: absolute;
  top: -60px; right: -60px;
  width: 180px; height: 180px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(249,115,22,0.25), transparent 70%);
  filter: blur(30px);
  pointer-events: none;
  animation: orb-float 5s ease-in-out infinite;
}

.purchase-orb.orb-green {
  background: radial-gradient(circle, rgba(52,211,153,0.25), transparent 70%);
}

@keyframes orb-float {
  0%,100% { transform: translate(0,0); }
  50%      { transform: translate(-10px, 10px); }
}

.purchase-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: rgba(255,255,255,0.4);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 8px;
}

.free-label { color: rgba(52,211,153,0.7); }

.purchase-price {
  font-size: 2.8rem;
  font-weight: 900;
  color: #fff;
  letter-spacing: -0.04em;
  line-height: 1;
  margin-bottom: 6px;
}

.free-price {
  background: linear-gradient(135deg, #34d399, #10b981);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.purchase-period {
  font-size: 0.8rem;
  color: rgba(255,255,255,0.35);
  margin-bottom: 4px;
}

/* Buttons */
.cpv-btn-buy, .cpv-btn-free, .cpv-btn-code {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border-radius: 12px;
  font-family: inherit;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.25s;
  padding: 14px 20px;
  font-size: 0.95rem;
  border: none;
  position: relative;
  overflow: hidden;
}

.cpv-btn-buy {
  background: linear-gradient(135deg, #f97316, #ea580c);
  color: #fff;
  box-shadow: 0 4px 20px rgba(249,115,22,0.35);
}
.cpv-btn-buy:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 28px rgba(249,115,22,0.5);
}
.cpv-btn-buy:active:not(:disabled) { transform: translateY(0); }
.cpv-btn-buy:disabled { opacity: 0.6; cursor: not-allowed; }

.cpv-btn-free {
  background: linear-gradient(135deg, #10b981, #059669);
  color: #fff;
  box-shadow: 0 4px 20px rgba(16,185,129,0.35);
}
.cpv-btn-free:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 28px rgba(16,185,129,0.5);
}
.cpv-btn-free:disabled { opacity: 0.6; cursor: not-allowed; }

.cpv-btn-code:hover:not(:disabled) {
  opacity: 0.9;
  box-shadow: 0 4px 14px rgba(251, 146, 60, 0.4);
}

.cpv-btn-b2b {
  background: rgba(255, 255, 255, 0.05);
  color: #f1f5f9;
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 12px 24px;
  border-radius: 12px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.cpv-btn-b2b:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}

/* Modals */
.b2b-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 999;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.b2b-modal-content {
  width: 100%;
  max-width: 400px;
  background: rgba(15, 23, 42, 0.85);
  border: 1px solid rgba(255,255,255,0.1);
}

.b2b-input-group label {
  display: block;
  font-size: 0.9rem;
  color: #94a3b8;
  margin-bottom: 5px;
}

.b2b-total {
  margin-top: 15px;
  font-size: 1.1rem;
  color: #fb923c;
  text-align: right;
}

.cpv-btn-cancel {
  background: transparent;
  color: #94a3b8;
  border: 1px solid rgba(255,255,255,0.1);
  padding: 12px;
  border-radius: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.cpv-btn-cancel:hover {
  background: rgba(255,255,255,0.05);
  color: #fff;
}

.cpv-btn-spinner {
  display: inline-block;
}

.cpv-btn-code {
  background: rgba(249,115,22,0.15);
  border: 1px solid rgba(249,115,22,0.35) !important;
  color: #fb923c;
  padding: 12px 18px;
  font-size: 0.88rem;
  border-radius: 12px;
  white-space: nowrap;
}
.cpv-btn-code:hover:not(:disabled) {
  background: rgba(249,115,22,0.25);
  border-color: rgba(249,115,22,0.6) !important;
}
.cpv-btn-code:disabled { opacity: 0.4; cursor: not-allowed; }

.w-full { width: 100%; }
.mt-4 { margin-top: 16px; }

/* Shimmer on buy button */
.cpv-btn-buy::after, .cpv-btn-free::after {
  content: '';
  position: absolute;
  top: -50%;
  left: -60%;
  width: 60%;
  height: 200%;
  background: linear-gradient(105deg, transparent 40%, rgba(255,255,255,0.2) 50%, transparent 60%);
  animation: btn-shimmer 3s ease-in-out infinite;
}

@keyframes btn-shimmer {
  0%   { left: -60%; }
  50%,100% { left: 160%; }
}

.cpv-btn-spinner {
  width: 18px; height: 18px;
  border: 2.5px solid rgba(255,255,255,0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
.cpv-btn-spinner.small { width: 14px; height: 14px; }

/* Badges */
.purchase-badges {
  display: flex;
  gap: 8px;
  margin-top: 14px;
  flex-wrap: wrap;
}

.badge {
  font-size: 0.72rem;
  color: rgba(255,255,255,0.45);
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 6px;
  padding: 4px 10px;
}

/* Includes */
.purchase-includes {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(255,255,255,0.07);
}

.includes-title {
  font-size: 0.78rem;
  font-weight: 700;
  color: rgba(255,255,255,0.4);
  text-transform: uppercase;
  letter-spacing: 0.07em;
  margin-bottom: 12px;
}

.include-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.88rem;
  color: rgba(255,255,255,0.6);
  margin-bottom: 8px;
}

/* ── RESPONSIVE ────────────────────────────────────────── */
@media (max-width: 900px) {
  .cpv-body { grid-template-columns: 1fr; }
  .cpv-right { order: -1; }
  .purchase-glass { position: static; }
}

@media (max-width: 600px) {
  .cpv-nav-inner { padding: 0 16px; }
  .cpv-hero-content { padding: 0 16px 40px; }
  .cpv-body { padding: 24px 16px 60px; }
  .cpv-title { font-size: 2rem; }
  .cpv-mobile-cta { display: block; }
  .learn-grid { grid-template-columns: 1fr; }
  .code-row { flex-direction: column; }
}
</style>
