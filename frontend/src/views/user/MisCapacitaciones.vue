<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'

const capacitaciones = ref<any[]>([])
const router = useRouter()

onMounted(async () => {
  const res = await api.get('/mis-capacitaciones')
  capacitaciones.value = res.data || []
})

function typeLabel(t: string) {
  return { video: '🎥 Video', document: '📄 Documento', text: '📝 Texto' }[t] || t
}
</script>

<template>
  <div class="page">
    <h2>Mis Capacitaciones</h2>
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
      <p>Contacta a tu administrador.</p>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; }
h2 { font-size: 1.4rem; font-weight: 700; color: #1e293b; margin-bottom: 1.5rem; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 1.2rem; }
.card {
  background: white; border-radius: 12px; padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0,0,0,0.07); cursor: pointer;
  transition: transform 0.15s, box-shadow 0.15s;
}
.card:hover { transform: translateY(-2px); box-shadow: 0 6px 18px rgba(0,0,0,0.1); }
.type-badge { font-size: 0.8rem; color: #10b981; font-weight: 600; margin-bottom: 8px; }
h3 { font-size: 1rem; font-weight: 700; color: #1e293b; margin-bottom: 6px; }
p { font-size: 0.85rem; color: #64748b; line-height: 1.5; }
.see-more { display: inline-block; margin-top: 12px; color: #3b82f6; font-size: 0.85rem; font-weight: 600; }
.empty { text-align: center; padding: 4rem; color: #94a3b8; }
.empty p { font-size: 1rem; margin-bottom: 4px; }
</style>
