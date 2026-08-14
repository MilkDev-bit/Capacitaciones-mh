<script setup lang="ts">
/**
 * Modal único de la constancia DC-3.
 *
 * Se monta una sola vez en App.vue y lo abre cualquier vista a través del store.
 * Antes cada botón de "Tramitar constancia" abría una pestaña nueva contra un
 * formulario externo: el alumno salía de la plataforma, volvía a teclear datos
 * que ya teníamos y no había forma de devolverle el documento aquí dentro.
 *
 * El contenido es el mismo DC3Panel que vive al pie de la capacitación, en
 * variante plana. Compartir el componente evita que las dos entradas se
 * desincronicen, que es lo que pasó con las cinco copias del enlace externo.
 */
import { computed } from 'vue'
import { useDC3Store } from '../stores/dc3'
import { useScrollLock } from '../composables/useScrollLock'
import DC3Panel from './DC3Panel.vue'

const dc3 = useDC3Store()
const abierto = computed(() => dc3.abierto)

useScrollLock(abierto)
</script>

<template>
  <Teleport to="body">
    <Transition name="dc3-fade">
      <div v-if="dc3.abierto" class="dc3-overlay" @click.self="dc3.cerrar()">
        <!--
          role/aria-modal para que el lector de pantalla anuncie que el resto de
          la página queda fuera de alcance mientras el formulario está abierto.
        -->
        <div class="dc3-modal sheet-panel" role="dialog" aria-modal="true"
             aria-labelledby="dc3-modal-titulo">
          <header class="dc3-modal-head">
            <div>
              <h2 id="dc3-modal-titulo">Constancia DC-3</h2>
              <p v-if="dc3.titulo" class="dc3-modal-curso">{{ dc3.titulo }}</p>
            </div>
            <button class="dc3-cerrar" aria-label="Cerrar" @click="dc3.cerrar()">
              <svg width="20" height="20" fill="none" stroke="currentColor"
                   stroke-width="2" viewBox="0 0 24 24">
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
            </button>
          </header>

          <div class="dc3-modal-body sheet-body">
            <!--
              `key` fuerza el remontaje al cambiar de capacitación. Sin él, abrir
              el modal para un segundo curso conservaría en el formulario los
              datos de empresa del primero.
            -->
            <DC3Panel
              :key="dc3.cursoId"
              :curso-id="dc3.cursoId"
              :completado="true"
              variante="plano"
            />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.dc3-overlay {
  position: fixed;
  inset: 0;
  z-index: 1200;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--gutter, 16px);
}

.dc3-modal {
  width: 100%;
  max-width: 640px;
  max-height: 90dvh;
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
}

.dc3-modal-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 18px 20px;
  border-bottom: 1px solid var(--border);
}

.dc3-modal-head h2 {
  margin: 0;
  font-size: 1.05rem;
  color: var(--text);
}

.dc3-modal-curso {
  margin: 4px 0 0;
  font-size: 0.85rem;
  color: var(--muted);
}

.dc3-cerrar {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: var(--touch-min, 44px);
  height: var(--touch-min, 44px);
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
}

.dc3-cerrar:hover { background: var(--bg); color: var(--text); }

.dc3-modal-body {
  padding: 20px;
  flex: 1;
  min-height: 0;
}

.dc3-fade-enter-active, .dc3-fade-leave-active { transition: opacity 0.18s ease; }
.dc3-fade-enter-from, .dc3-fade-leave-to { opacity: 0; }

/* En móvil la hoja ocupa la pantalla: `.sheet-panel` ya lo resuelve, aquí solo
 * se retira el margen del overlay para que no quede un borde suelto. */
@media (max-width: 639px) {
  .dc3-overlay { padding: 0; align-items: stretch; }
}

@media (prefers-reduced-motion: reduce) {
  .dc3-fade-enter-active, .dc3-fade-leave-active { transition: none; }
}
</style>
