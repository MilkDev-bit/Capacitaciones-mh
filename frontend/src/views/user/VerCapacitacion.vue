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
  <div>
    <router-link to="/usuario/capacitaciones" class="back-link">
      <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M19 12H5M12 5l-7 7 7 7"/></svg>
      Volver a mis cursos
    </router-link>

    <div v-if="cap" class="viewer">
      <!-- Header -->
      <div class="viewer-header">
        <div :class="['viewer-type-bar', { 'bar-video': cap.type==='video', 'bar-document': cap.type==='document', 'bar-text': cap.type==='text' }]"></div>
        <div class="viewer-meta">
          <span class="viewer-type-badge">
            {{ ({ video: '🎥 Video', document: '📎 Documento', text: '📝 Texto' } as Record<string,string>)[cap.type] || cap.type }}
          </span>
          <h2 class="viewer-title">{{ cap.title }}</h2>
          <p v-if="cap.description" class="viewer-desc">{{ cap.description }}</p>
        </div>
      </div>

      <!-- Content -->
      <div class="viewer-content">
        <div v-if="cap.type === 'video'" class="media-box">
          <video :src="cap.file_path" controls class="video-player" preload="metadata"></video>
        </div>
        <div v-else-if="cap.type === 'document'" class="media-box">
          <iframe :src="cap.file_path" class="pdf-frame" title="Documento"></iframe>
          <a :href="cap.file_path" target="_blank" class="btn btn-primary" style="margin-top:16px;display:inline-flex">
            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>
            Descargar documento
          </a>
        </div>
        <div v-else-if="cap.type === 'text'" class="prose-box" v-html="cap.content"></div>
      </div>
    </div>

    <div v-else class="loading">
      <div class="spinner"></div>
      <p>Cargando contenido…</p>
    </div>
  </div>
</template>

<style scoped>
.back-link {
  display: inline-flex; align-items: center; gap: 6px;
  color: var(--muted); font-size: 0.87rem; font-weight: 600;
  margin-bottom: 24px; transition: color 0.15s;
}
.back-link:hover { color: var(--brand); }

.viewer { max-width: 860px; }
.viewer-header { background: var(--surface); border-radius: var(--r-lg); overflow: hidden; box-shadow: var(--shadow-sm); margin-bottom: 20px; }
.viewer-type-bar { height: 5px; }
.bar-video    { background: linear-gradient(90deg, #f97316, #dc2626); }
.bar-document { background: linear-gradient(90deg, #3b82f6, #6366f1); }
.bar-text     { background: linear-gradient(90deg, #10b981, #059669); }
.viewer-meta { padding: 20px 24px; }
.viewer-type-badge {
  font-size: 0.75rem; font-weight: 700; text-transform: uppercase; letter-spacing: .06em;
  color: var(--brand-dark); background: var(--brand-light); padding: 3px 10px; border-radius: 4px; display: inline-block;
}
.viewer-title { font-size: 1.4rem; font-weight: 800; color: var(--dark); margin: 10px 0 6px; }
.viewer-desc { color: var(--muted); font-size: 0.92rem; line-height: 1.6; }

.viewer-content {}
.media-box { background: var(--surface); border-radius: var(--r-lg); padding: 20px; box-shadow: var(--shadow-sm); }
.video-player { width: 100%; border-radius: var(--r); max-height: 500px; background: #000; }
.pdf-frame { width: 100%; height: 620px; border: none; border-radius: var(--r); }
.prose-box {
  background: var(--surface); border-radius: var(--r-lg); padding: 32px;
  box-shadow: var(--shadow-sm); line-height: 1.75; color: var(--text); font-size: 1rem;
}
.loading { display: flex; flex-direction: column; align-items: center; gap: 16px; padding: 60px; color: var(--muted); }

@media (max-width: 600px) {
  .pdf-frame { height: 400px; }
  .viewer-title { font-size: 1.1rem; }
}
</style>
