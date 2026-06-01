<script setup lang="ts">
import { ref, onMounted } from "vue";
import api from "../api";
import { toast } from "../utils/toast";
import { uploadToR2 } from "../utils/upload";
import DragDropUpload from "./DragDropUpload.vue";
import ContentTypeSelector from "./ContentTypeSelector.vue";

const props = defineProps<{ capId: string }>();
const emit = defineEmits<{ (e: "update"): void }>();

const lessons = ref<any[]>([]);
const loading = ref(false);

const editMode = ref<string | null>(null);
const editForm = ref<any>({});
const editFile = ref<File | null>(null);

const showCreate = ref(false);
const createForm = ref<any>({
  title: "",
  description: "",
  type: "video",
  content: "",
  duracion_min: 0,
});
const createFile = ref<File | null>(null);

async function fetchLessons() {
  loading.value = true;
  try {
    const res = await api.get(`/instructor/capacitaciones/${props.capId}/lecciones`);
    lessons.value = res.data || [];
  } catch (e: any) {
    toast.error("Error al cargar lecciones");
  } finally {
    loading.value = false;
  }
}

onMounted(fetchLessons);

function isValidUrl(value: string) {
  if (!value) return false;
  try {
    const u = new URL(value);
    return u.protocol === "http:" || u.protocol === "https:";
  } catch {
    return false;
  }
}

function startEdit(l: any) {
  editMode.value = l.id;
  editForm.value = { ...l };
  editFile.value = null;
}

function cancelEdit() {
  editMode.value = null;
}

async function saveEdit() {
  if (!editForm.value.title) return toast.error("Título requerido");
  if (editForm.value.type === "text" && !editForm.value.content?.trim()) {
    return toast.error("El contenido de texto es requerido");
  }
  if (editForm.value.type === "link" && !isValidUrl(editForm.value.content || "")) {
    return toast.error("Debes ingresar una URL válida (http/https)");
  }
  const payload: Record<string, any> = {
    title: editForm.value.title,
    description: editForm.value.description,
    type: editForm.value.type,
    content: editForm.value.content,
    duracion_min: Number(editForm.value.duracion_min || 0),
    file_path: editForm.value.file_path || '',
  };

  if (editFile.value) {
    const isMedia = editForm.value.type === "video" || editForm.value.type === "document";
    if (isMedia) {
      const uploadingToast = toast.loading("Subiendo archivo...");
      try {
        const prefix = editForm.value.type === "video" ? "videos" : "documents";
        payload.file_path = await uploadToR2(editFile.value, prefix);
      } finally {
        uploadingToast.close();
      }
    }
  }

  try {
    await api.put(`/instructor/capacitaciones/${props.capId}/lecciones/${editForm.value.id}`, payload);
    toast.success("Lección actualizada");
    editMode.value = null;
    fetchLessons();
  } catch (e: any) {
    toast.error("Error al actualizar");
  }
}

async function createLesson() {
  if (!createForm.value.title) return toast.error("Título requerido");
  if (createForm.value.type === "text" && !createForm.value.content?.trim()) {
    return toast.error("El contenido de texto es requerido");
  }
  if (createForm.value.type === "link" && !isValidUrl(createForm.value.content || "")) {
    return toast.error("Debes ingresar una URL válida (http/https)");
  }
  const payload: Record<string, any> = {
    title: createForm.value.title,
    description: createForm.value.description,
    type: createForm.value.type,
    content: createForm.value.content,
    duracion_min: Number(createForm.value.duracion_min || 0),
    orden: lessons.value.length,
    file_path: '',
  };

  if (createFile.value) {
    const isMedia = createForm.value.type === "video" || createForm.value.type === "document";
    if (isMedia) {
      const uploadingToast = toast.loading("Subiendo archivo...");
      try {
        const prefix = createForm.value.type === "video" ? "videos" : "documents";
        payload.file_path = await uploadToR2(createFile.value, prefix);
      } finally {
        uploadingToast.close();
      }
    }
  }

  const loadingToast = toast.loading("Creando lección...");
  try {
    await api.post(`/instructor/capacitaciones/${props.capId}/lecciones`, payload);
    loadingToast.close();
    toast.success("Lección creada");
    showCreate.value = false;
    createForm.value = { title: "", description: "", type: "video", content: "", duracion_min: 0 };
    createFile.value = null;
    fetchLessons();
  } catch (e: any) {
    loadingToast.close();
    toast.error("Error al crear");
  }
}

async function deleteLesson(id: string) {
  if (!confirm("¿Eliminar lección?")) return;
  try {
    await api.delete(`/instructor/capacitaciones/${props.capId}/lecciones/${id}`);
    toast.success("Lección eliminada");
    fetchLessons();
  } catch (e: any) {
    toast.error("Error al eliminar");
  }
}

async function move(index: number, direction: -1 | 1) {
  const newIndex = index + direction;
  if (newIndex < 0 || newIndex >= lessons.value.length) return;

  const temp = lessons.value[index];
  lessons.value[index] = lessons.value[newIndex];
  lessons.value[newIndex] = temp;

  const payload = lessons.value.map((l, i) => ({ id: l.id, orden: i }));
  try {
    await api.put(`/instructor/capacitaciones/${props.capId}/lecciones/reorder`, payload);
  } catch (e) {
    toast.error("Error al reordenar");
    fetchLessons();
  }
}
</script>

<template>
  <div class="lesson-list">
    <div v-if="loading" class="text-center p-4">Cargando...</div>
    <div v-else>
      <div v-for="(l, index) in lessons" :key="l.id" class="lesson-item">
        <div v-if="editMode === l.id" class="lesson-edit-pane slide-down-enter-active">
          <h4>Editar Lección</h4>
          <div class="field">
            <label>Título</label>
            <input v-model="editForm.title" class="field-input" />
          </div>
          <div class="field mt-3">
            <label>Descripción</label>
            <textarea v-model="editForm.description" class="field-input" rows="2"></textarea>
          </div>
          <div class="field mt-3">
            <label>Tipo</label>
            <ContentTypeSelector v-model="editForm.type" />
          </div>
          <div class="field mt-3" v-if="editForm.type === 'video' || editForm.type === 'document'">
            <label>Archivo nuevo (opcional)</label>
            <DragDropUpload v-model="editFile" />
          </div>
          <div class="field mt-3" v-if="editForm.type === 'text'">
            <label>Contenido</label>
            <textarea v-model="editForm.content" class="field-input" rows="4"></textarea>
          </div>
          <div class="field mt-3" v-if="editForm.type === 'link'">
            <label>URL del enlace</label>
            <input
              v-model="editForm.content"
              type="url"
              class="field-input"
              placeholder="https://..."
            />
          </div>
          <div class="field mt-3">
            <label>Duración (minutos)</label>
            <input
              type="number"
              v-model="editForm.duracion_min"
              class="field-input"
              style="width: 100px"
            />
          </div>
          <div class="edit-actions mt-4">
            <button class="btn btn-secondary" @click="cancelEdit">Cancelar</button>
            <button class="btn btn-primary" @click="saveEdit">Guardar</button>
          </div>
        </div>

        <div v-else class="lesson-row">
          <div class="lesson-drag">
            <button class="icon-btn" :disabled="index === 0" @click="move(index, -1)">
              <svg
                width="20"
                height="20"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                viewBox="0 0 24 24"
              >
                <path d="M5 15l7-7 7 7" />
              </svg>
            </button>
            <button
              class="icon-btn"
              :disabled="index === lessons.length - 1"
              @click="move(index, 1)"
            >
              <svg
                width="20"
                height="20"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                viewBox="0 0 24 24"
              >
                <path d="M19 9l-7 7-7-7" />
              </svg>
            </button>
          </div>
          <div class="lesson-info">
            <span class="lesson-title">{{ l.title }}</span>
            <span class="lesson-meta">{{ l.type }} • {{ l.duracion_min }} min</span>
          </div>
          <div class="lesson-actions">
            <button class="icon-btn" @click="startEdit(l)" title="Editar">
              <svg
                width="18"
                height="18"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                viewBox="0 0 24 24"
              >
                <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" />
                <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
              </svg>
            </button>
            <button class="icon-btn text-danger" @click="deleteLesson(l.id)" title="Eliminar">
              <svg
                width="18"
                height="18"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                viewBox="0 0 24 24"
              >
                <path
                  d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"
                />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <div v-if="showCreate" class="lesson-create-pane mt-4 slide-down-enter-active">
        <h4>Nueva Lección</h4>
        <div class="field">
          <label>Título</label>
          <input v-model="createForm.title" class="field-input" autofocus />
        </div>
        <div class="field mt-3">
          <label>Descripción</label>
          <textarea v-model="createForm.description" class="field-input" rows="2"></textarea>
        </div>
        <div class="field mt-3">
          <label>Tipo</label>
          <ContentTypeSelector v-model="createForm.type" />
        </div>
        <div
          class="field mt-3"
          v-if="createForm.type === 'video' || createForm.type === 'document'"
        >
          <label>Archivo</label>
          <DragDropUpload v-model="createFile" />
        </div>
        <div class="field mt-3" v-if="createForm.type === 'text'">
          <label>Contenido</label>
          <textarea v-model="createForm.content" class="field-input" rows="4"></textarea>
        </div>
        <div class="field mt-3" v-if="createForm.type === 'link'">
          <label>URL del enlace</label>
          <input
            v-model="createForm.content"
            type="url"
            class="field-input"
            placeholder="https://..."
          />
        </div>
        <div class="field mt-3">
          <label>Duración (minutos)</label>
          <input
            type="number"
            v-model="createForm.duracion_min"
            class="field-input"
            style="width: 100px"
          />
        </div>
        <div class="edit-actions mt-4">
          <button class="btn btn-secondary" @click="showCreate = false">Cancelar</button>
          <button class="btn btn-primary" @click="createLesson">Crear</button>
        </div>
      </div>

      <button v-if="!showCreate" class="btn-dashed mt-4" @click="showCreate = true">
        <svg
          width="20"
          height="20"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          viewBox="0 0 24 24"
        >
          <path d="M12 5v14M5 12h14" />
        </svg>
        Agregar Lección
      </button>
    </div>
  </div>
</template>

<style scoped>
.lesson-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.lesson-item {
  background: var(--surface);
  border: 1px solid var(--border-light);
  border-radius: var(--r-md);
  overflow: hidden;
}
.lesson-row {
  display: flex;
  align-items: center;
  padding: 12px;
  gap: 12px;
}
.lesson-drag {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.icon-btn {
  background: none;
  border: none;
  color: var(--muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.icon-btn:hover:not(:disabled) {
  background: var(--surface-soft);
  color: var(--dark);
}
.icon-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
.icon-btn.text-danger:hover {
  color: var(--danger);
  background: var(--danger-bg);
}

.lesson-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.lesson-title {
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--dark);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.lesson-meta {
  font-size: 0.75rem;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.lesson-actions {
  display: flex;
  gap: 4px;
}

.lesson-edit-pane,
.lesson-create-pane {
  padding: 20px;
  background: var(--surface-soft);
  border-top: 1px solid var(--border-light);
  border-radius: var(--r-md);
}
.lesson-edit-pane h4,
.lesson-create-pane h4 {
  margin: 0 0 16px 0;
  font-size: 1.1rem;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.field label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--dark);
}
.mt-3 {
  margin-top: 12px;
}
.mt-4 {
  margin-top: 16px;
}
.edit-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn-dashed {
  width: 100%;
  padding: 16px;
  border: 2px dashed var(--border);
  background: transparent;
  color: var(--muted);
  border-radius: var(--r-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-weight: 600;
  transition: all 0.2s;
}
.btn-dashed:hover {
  border-color: var(--brand);
  color: var(--brand);
  background: var(--brand-light);
}
</style>
