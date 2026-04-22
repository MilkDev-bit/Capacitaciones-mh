<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../../api'

const route = useRoute()
const cap = ref<any>(null)

onMounted(async () => {
  const res = await api.get(`/capacitaciones/${route.params.id}`)
  cap.value = res.data
})
</script>

<template>
  <div class="page">
    <router-link to="/usuario/capacitaciones" class="back">← Volver</router-link>
    <div v-if="cap" class="content-wrap">
      <h2>{{ cap.title }}</h2>
      <p class="desc">{{ cap.description }}</p>

      <div v-if="cap.type === 'video'" class="media-wrap">
        <video :src="cap.file_path" controls class="player"></video>
      </div>

      <div v-else-if="cap.type === 'document'" class="media-wrap">
        <iframe :src="cap.file_path" class="pdf-viewer"></iframe>
        <a :href="cap.file_path" target="_blank" class="btn-download">⬇ Descargar documento</a>
      </div>

      <div v-else-if="cap.type === 'text'" class="text-content" v-html="cap.content"></div>
    </div>
    <div v-else class="loading">Cargando capacitación...</div>
  </div>
</template>

<style scoped>
.page { padding: 2rem; max-width: 900px; }
.back { color: #3b82f6; text-decoration: none; font-size: 0.9rem; font-weight: 600; display: inline-block; margin-bottom: 1.5rem; }
.back:hover { text-decoration: underline; }
h2 { font-size: 1.5rem; font-weight: 700; color: #1e293b; margin-bottom: 0.5rem; }
.desc { color: #64748b; margin-bottom: 1.5rem; }
.media-wrap { background: white; border-radius: 12px; padding: 1rem; box-shadow: 0 2px 8px rgba(0,0,0,0.07); }
.player { width: 100%; border-radius: 8px; max-height: 480px; }
.pdf-viewer { width: 100%; height: 600px; border: none; border-radius: 8px; }
.btn-download { display: inline-block; margin-top: 1rem; background: #3b82f6; color: white; padding: 9px 18px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 0.9rem; }
.text-content { background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 2px 8px rgba(0,0,0,0.07); line-height: 1.7; }
.loading { padding: 3rem; text-align: center; color: #94a3b8; }
</style>
