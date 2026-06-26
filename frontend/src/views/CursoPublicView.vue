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

const typeLabel: Record<string, string> = {
  video: '🎬 Video',
  document: '📄 Documento',
  text: '📖 Lectura',
  link: '🔗 Enlace',
}

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
    toast.success('¡Te has inscrito correctamente! 🎉')
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
      success_url: window.location.origin + '/usuario/capacitaciones?success=true',
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
    toast.success('¡Inscrito correctamente a la cohorte! 🎉')
    router.push('/usuario/capacitaciones')
  } catch {
    try {
      await api.post('/inscripciones', { codigo: codigoForm.value.trim() })
      toast.success('¡Inscrito correctamente! 🎉')
      router.push('/usuario/capacitaciones')
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Código inválido o expirado')
    }
  } finally {
    joiningCodigo.value = false
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
            <span class="cpv-chip type">{{ typeLabel[curso.type] || '📚 Curso' }}</span>
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
                <span class="learn-icon">✅</span>
                <span>Contenido actualizado al {{ new Date().getFullYear() }}</span>
              </div>
              <div class="learn-item">
                <span class="learn-icon">📱</span>
                <span>Acceso desde cualquier dispositivo</span>
              </div>
              <div class="learn-item">
                <span class="learn-icon">🏆</span>
                <span>Certificado al completar</span>
              </div>
              <div class="learn-item">
                <span class="learn-icon">♾️</span>
                <span>Acceso de por vida</span>
              </div>
            </div>
          </div>

          <!-- Code section -->
          <div class="glass-card mt-4">
            <h2 class="glass-card-title">
              <span>🔑</span> ¿Tienes un código de acceso?
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
              <div class="purchase-period">Pago único · Acceso permanente</div>

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
                <span class="badge">🔒 Pago seguro</span>
                <span class="badge">↩️ Garantía 30 días</span>
              </div>
            </div>

            <div v-else>
              <div class="purchase-label free-label">Acceso Gratuito</div>
              <div class="purchase-price free-price">GRATIS</div>
              <div class="purchase-period">Sin costo · Para siempre</div>

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
                <span class="badge">🆓 Sin tarjeta</span>
                <span class="badge">⚡ Acceso inmediato</span>
              </div>
            </div>

            <!-- Includes list -->
            <div class="purchase-includes">
              <div class="includes-title">Este curso incluye:</div>
              <div class="include-item"><span>🎬</span><span>Contenido multimedia</span></div>
              <div class="include-item"><span>📱</span><span>Acceso móvil y escritorio</span></div>
              <div class="include-item"><span>🏅</span><span>Certificado de finalización</span></div>
              <div class="include-item"><span>♾️</span><span>Acceso de por vida</span></div>
            </div>
          </div>
        </aside>

      </div>
    </main>

    <!-- 404 state -->
    <div v-else class="cpv-loading">
      <div style="font-size: 3rem; margin-bottom: 16px;">😕</div>
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
