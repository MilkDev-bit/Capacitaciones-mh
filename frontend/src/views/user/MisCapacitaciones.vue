<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const cursosPublicos = ref<any[]>([])
const router = useRouter()
const activeTab = ref<'mis' | 'explorar'>('mis')
const inscribiendose = ref<string | null>(null)

// Código de acceso manual
const codigoInput = ref('')
const codigoLoading = ref(false)
const codigoError = ref('')
const codigoSuccess = ref('')

async function loadMis() {
  const res = await api.get('/mis-capacitaciones')
  capacitaciones.value = res.data || []
}

async function loadPublicos() {
  const res = await api.get('/cursos-publicos')
  cursosPublicos.value = res.data || []
}

onMounted(() => { loadMis(); loadPublicos() })

async function inscribirse(id: string) {
  inscribiendose.value = id
  try {
    await api.post(`/inscribirse/${id}`)
    await Promise.all([loadMis(), loadPublicos()])
    activeTab.value = 'mis'
  } finally {
    inscribiendose.value = null
  }
}

async function unirseConCodigo() {
  const code = codigoInput.value.trim().toUpperCase()
  if (!code) { codigoError.value = 'Ingresa un código'; return }
  codigoError.value = ''; codigoSuccess.value = ''
  codigoLoading.value = true
  try {
    const res = await api.post('/unirse-con-codigo', { codigo: code })
    codigoSuccess.value = `¡Te uniste a "${res.data.title}"!`
    codigoInput.value = ''
    await loadMis()
    setTimeout(() => { codigoSuccess.value = ''; activeTab.value = 'mis' }, 2000)
  } catch (e: any) {
    codigoError.value = e.response?.data?.error || 'Código inválido'
  } finally {
    codigoLoading.value = false
  }
}

function typeLabel(t: string) {
  return { video: '🎥 Video', document: '📄 Documento', text: '📝 Texto' }[t] || t
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h2>Capacitaciones</h2>
      <div class="tabs">
        <button :class="['tab', activeTab === 'mis' ? 'active' : '']" @click="activeTab = 'mis'">
          📚 Mis cursos
          <span class="count">{{ capacitaciones.length }}</span>
        </button>
        <button :class="['tab', activeTab === 'explorar' ? 'active' : '']" @click="activeTab = 'explorar'">
          🌐 Explorar cursos
          <span class="count">{{ cursosPublicos.length }}</span>
        </button>
      </div>
    </div>

    <!-- Mis capacitaciones asignadas/inscritas -->
    <div v-if="activeTab === 'mis'">
      <div v-if="capacitaciones.length" class="grid">
        <div
          v-for="c in capacitaciones"
          :key="c.id"
          class="card"
          @click="router.push('/usuario/capacitaciones/' + c.id)"
        >
          <div class="type-badge">{{ typeLabel(c.type) }}</div>
          <h3>{{ c.title }}</h3>
          <p>{{ c.description || 'Sin descripción' }}</p>
          <span class="see-more">Ver capacitación →</span>
        </div>
      </div>
      <div v-else class="empty">
        <p>No tienes capacitaciones asignadas aún.</p>
        <button class="btn-explore" @click="activeTab = 'explorar'">Explorar cursos disponibles</button>
      </div>
    </div>

    <!-- Explorar cursos públicos (Udemy-like) -->
    <div v-if="activeTab === 'explorar'">
      <!-- Unirse con código -->
      <div class="code-join-box">
        <div class="code-join-left">
          <span class="key-icon">🔑</span>
          <div>
            <strong>¿Tienes un código de acceso?</strong>
            <p>Ingresa el código que te compartió tu instructor para unirte a un curso privado.</p>
          </div>
        </div>
        <div class="code-join-right">
          <input
            v-model="codigoInput"
            placeholder="Ej: AB3X7K"
            maxlength="12"
            @keyup.enter="unirseConCodigo"
            class="code-input"
          />
          <button class="btn-join-code" :disabled="codigoLoading" @click="unirseConCodigo">
            {{ codigoLoading ? '…' : 'Unirme' }}
          </button>
        </div>
      </div>
      <p v-if="codigoError" class="msg error">{{ codigoError }}</p>
      <p v-if="codigoSuccess" class="msg success">{{ codigoSuccess }}</p>

      <p class="explore-hint">O explora los cursos disponibles para todos:</p>
      <div v-if="cursosPublicos.length" class="grid">
        <div v-for="c in cursosPublicos" :key="c.id" class="card public-card">
          <div class="type-badge">{{ typeLabel(c.type) }}</div>
          <h3>{{ c.title }}</h3>
          <p>{{ c.description || 'Sin descripción' }}</p>
          <div class="card-footer">
            <button
              v-if="!c.inscrito"
              class="btn-inscribir"
              :disabled="inscribiendose === c.id"
              @click.stop="inscribirse(c.id)"
            >
              {{ inscribiendose === c.id ? 'Inscribiendo...' : '+ Inscribirse' }}
            </button>
            <span v-else class="inscrito-badge">✓ Inscrito</span>
          </div>
        </div>
      </div>
      <div v-else class="empty">
        <p>No hay cursos públicos disponibles en este momento.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; flex-wrap: wrap; gap: 1rem; }
h2 { font-size: 1.4rem; font-weight: 700; color: #1e293b; margin: 0; }
.tabs { display: flex; gap: 6px; background: #f1f5f9; padding: 4px; border-radius: 10px; }
.tab {
  padding: 8px 16px; border: none; border-radius: 8px; cursor: pointer;
  font-size: 0.88rem; font-weight: 600; color: #64748b; background: transparent;
  display: flex; align-items: center; gap: 6px; transition: all 0.15s;
}
.tab.active { background: white; color: #1e293b; box-shadow: 0 1px 4px rgba(0,0,0,0.1); }
.count { background: #e2e8f0; color: #475569; font-size: 0.75rem; padding: 1px 7px; border-radius: 12px; font-weight: 700; }
.tab.active .count { background: #10b981; color: white; }
.explore-hint { color: #64748b; font-size: 0.87rem; margin-bottom: 1.2rem; }
/* Código de acceso */
.code-join-box {
  display: flex; align-items: center; gap: 1.5rem; flex-wrap: wrap;
  background: white; border-radius: 12px; padding: 1.2rem 1.5rem;
  box-shadow: 0 1px 6px rgba(0,0,0,0.08); margin-bottom: 0.75rem;
  border-left: 4px solid #10b981;
}
.code-join-left { display: flex; align-items: center; gap: 12px; flex: 1; }
.key-icon { font-size: 1.5rem; }
.code-join-left strong { font-size: 0.95rem; color: #1e293b; }
.code-join-left p { font-size: 0.82rem; color: #64748b; margin-top: 2px; }
.code-join-right { display: flex; gap: 8px; align-items: center; }
.code-input {
  padding: 9px 14px; border: 2px solid #e2e8f0; border-radius: 8px;
  font-size: 1.1rem; font-weight: 800; letter-spacing: 0.15em; text-transform: uppercase;
  width: 130px; outline: none; font-family: 'Courier New', monospace;
}
.code-input:focus { border-color: #10b981; }
.btn-join-code {
  background: #10b981; color: white; border: none; border-radius: 8px;
  padding: 9px 16px; cursor: pointer; font-size: 0.9rem; font-weight: 700; white-space: nowrap;
}
.btn-join-code:hover:not(:disabled) { background: #059669; }
.btn-join-code:disabled { opacity: 0.6; cursor: not-allowed; }
.msg { font-size: 0.85rem; margin-bottom: 1rem; padding: 8px 12px; border-radius: 6px; }
.msg.error { background: #fee2e2; color: #dc2626; }
.msg.success { background: #d1fae5; color: #065f46; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 1.2rem; }
.card {
  background: white; border-radius: 12px; padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.07); cursor: pointer;
  transition: transform 0.15s, box-shadow 0.15s;
}
.card:hover { transform: translateY(-2px); box-shadow: 0 6px 18px rgba(0,0,0,0.1); }
.public-card { cursor: default; }
.type-badge { font-size: 0.8rem; color: #10b981; font-weight: 600; margin-bottom: 8px; }
h3 { font-size: 1rem; font-weight: 700; color: #1e293b; margin-bottom: 6px; }
p { font-size: 0.85rem; color: #64748b; line-height: 1.5; }
.see-more { display: inline-block; margin-top: 12px; color: #3b82f6; font-size: 0.85rem; font-weight: 600; }
.card-footer { margin-top: 1rem; }
.btn-inscribir {
  background: #10b981; color: white; border: none; border-radius: 8px;
  padding: 8px 16px; cursor: pointer; font-size: 0.85rem; font-weight: 600; width: 100%;
}
.btn-inscribir:hover:not(:disabled) { background: #059669; }
.btn-inscribir:disabled { opacity: 0.6; cursor: not-allowed; }
.inscrito-badge {
  display: block; text-align: center; color: #059669; font-size: 0.85rem;
  font-weight: 600; background: #d1fae5; padding: 8px; border-radius: 8px;
}
.empty { text-align: center; padding: 4rem; color: #94a3b8; }
.empty p { font-size: 1rem; margin-bottom: 12px; }
.btn-explore {
  background: #10b981; color: white; border: none; border-radius: 8px;
  padding: 10px 20px; cursor: pointer; font-size: 0.9rem; font-weight: 600;
}
.btn-explore:hover { background: #059669; }
</style>

