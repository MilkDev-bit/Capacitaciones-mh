<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import api from '../../api'

const perfil = ref<any>(null)
const loading = ref(false)
const loadingSave = ref(false)
const error = ref('')
const success = ref('')

const form = ref({ name: '', bio: '', phone: '' })
const password = ref({ nueva: '', confirmar: '' })
const showPass = ref(false)

const roleLabel = computed(() => {
  const labels: Record<string, string> = {
    admin: 'Administrador',
    instructor: 'Instructor',
    user: 'Estudiante',
  }
  return labels[perfil.value?.role] || 'Usuario'
})

const completion = computed(() => {
  const fields = [form.value.name, form.value.bio, form.value.phone]
  const filled = fields.filter((value) => String(value || '').trim()).length
  return Math.round((filled / fields.length) * 100)
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.get('/perfil')
    perfil.value = res.data
    form.value.name = res.data.name || ''
    form.value.bio = res.data.bio || ''
    form.value.phone = res.data.phone || ''
  } catch (e: any) {
    error.value = e.response?.data?.error || 'No pudimos cargar tu perfil'
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function guardar() {
  error.value = ''
  success.value = ''

  if (!form.value.name.trim()) {
    error.value = 'El nombre es requerido'
    return
  }
  if (showPass.value) {
    if (!password.value.nueva) {
      error.value = 'Ingresa la nueva contraseña'
      return
    }
    if (password.value.nueva.length < 6) {
      error.value = 'La contraseña debe tener al menos 6 caracteres'
      return
    }
    if (password.value.nueva !== password.value.confirmar) {
      error.value = 'Las contraseñas no coinciden'
      return
    }
  }

  loadingSave.value = true
  try {
    const payload: any = {
      name: form.value.name,
      bio: form.value.bio,
      phone: form.value.phone,
    }
    if (showPass.value && password.value.nueva) payload.password = password.value.nueva

    await api.put('/perfil', payload)
    success.value = 'Perfil actualizado correctamente'
    password.value = { nueva: '', confirmar: '' }
    showPass.value = false
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al guardar'
  } finally {
    loadingSave.value = false
  }
}

function initials(name: string) {
  return name ? name.split(' ').slice(0, 2).map(n => n[0]).join('').toUpperCase() : '?'
}
</script>

<template>
  <div class="profile-page">
    <header class="ph profile-header">
      <div>
        <h1 class="ph-title">Mi perfil</h1>
        <p class="ph-sub">Actualiza tus datos visibles y los accesos de tu cuenta.</p>
      </div>
      <button class="btn btn-primary" :disabled="loadingSave || loading" @click="guardar">
        <span v-if="loadingSave" class="spinner profile-spinner"></span>
        {{ loadingSave ? 'Guardando...' : 'Guardar cambios' }}
      </button>
    </header>

    <div v-if="loading" class="profile-grid">
      <section class="profile-card profile-card-main">
        <div class="skeleton profile-avatar-skeleton"></div>
        <div class="profile-skeleton-lines">
          <div class="skeleton skel-title"></div>
          <div class="skeleton skel-text"></div>
          <div class="skeleton skel-text-sm"></div>
        </div>
      </section>
      <section class="profile-card">
        <div class="skeleton skel-title"></div>
        <div class="skeleton skel-line"></div>
        <div class="skeleton skel-line"></div>
      </section>
    </div>

    <div v-else class="profile-grid">
      <aside class="profile-card profile-summary">
        <div class="profile-avatar">{{ initials(form.name) }}</div>
        <div class="profile-summary-copy">
          <h2>{{ perfil?.name }}</h2>
          <p>{{ perfil?.email }}</p>
          <span class="profile-role">{{ roleLabel }}</span>
        </div>

        <div class="profile-completion">
          <div class="profile-completion-top">
            <span>Perfil completo</span>
            <strong>{{ completion }}%</strong>
          </div>
          <div class="progress-bar-bg">
            <div class="progress-bar-fill" :style="`width:${completion}%`"></div>
          </div>
        </div>
      </aside>

      <main class="profile-main">
        <div v-if="error" class="alert alert-error">{{ error }}</div>
        <div v-if="success" class="alert alert-success">{{ success }}</div>

        <section class="profile-card">
          <div class="profile-card-head">
            <div>
              <h2>Datos personales</h2>
              <p>Estos datos ayudan a identificarte dentro de cursos, examenes y foros.</p>
            </div>
          </div>

          <div class="profile-form">
            <label class="profile-field">
              <span>Nombre completo *</span>
              <input v-model="form.name" class="field-input" autocomplete="name" />
            </label>

            <label class="profile-field">
              <span>Telefono</span>
              <input v-model="form.phone" type="tel" class="field-input" placeholder="+52 55 0000 0000" autocomplete="tel" />
            </label>

            <label class="profile-field profile-field-full">
              <span>Biografia</span>
              <textarea
                v-model="form.bio"
                rows="4"
                placeholder="Cuenta brevemente tu rol, area o intereses de aprendizaje..."
                class="field-input profile-textarea"
              />
            </label>
          </div>
        </section>

        <section class="profile-card">
          <div class="profile-card-head security-head">
            <div>
              <h2>Seguridad</h2>
              <p>Cambia tu contraseña cuando necesites renovar el acceso.</p>
            </div>
            <button class="btn btn-secondary btn-sm" @click="showPass = !showPass">
              {{ showPass ? 'Cancelar' : 'Cambiar contraseña' }}
            </button>
          </div>

          <Transition name="slide-down">
            <div v-if="showPass" class="profile-form profile-password-form">
              <label class="profile-field">
                <span>Nueva contraseña</span>
                <input
                  v-model="password.nueva"
                  type="password"
                  placeholder="Minimo 6 caracteres"
                  class="field-input"
                  autocomplete="new-password"
                />
              </label>

              <label class="profile-field">
                <span>Confirmar contraseña</span>
                <input
                  v-model="password.confirmar"
                  type="password"
                  placeholder="Repite la nueva contraseña"
                  class="field-input"
                  autocomplete="new-password"
                />
              </label>
            </div>
          </Transition>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.profile-header {
  align-items: center;
  margin-bottom: 0;
}

.profile-grid {
  display: grid;
  grid-template-columns: minmax(240px, 320px) minmax(0, 1fr);
  gap: 22px;
  align-items: start;
}

.profile-main {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.profile-card {
  padding: 22px;
  border: 1px solid rgba(17, 24, 39, 0.08);
  border-radius: 8px;
  background: var(--surface);
  box-shadow: var(--shadow-sm);
}

.profile-card-main {
  display: flex;
  gap: 16px;
}

.profile-summary {
  position: sticky;
  top: calc(var(--topbar-h) + 24px);
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.profile-avatar,
.profile-avatar-skeleton {
  width: 76px;
  height: 76px;
  border-radius: 50%;
  flex-shrink: 0;
}

.profile-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--brand), #2563eb);
  color: #fff;
  font-size: 1.4rem;
  font-weight: 900;
}

.profile-summary-copy h2 {
  color: var(--dark);
  font-size: 1.15rem;
  font-weight: 900;
  line-height: 1.25;
}

.profile-summary-copy p {
  margin-top: 4px;
  color: var(--muted);
  font-size: 0.88rem;
  overflow-wrap: anywhere;
}

.profile-role {
  display: inline-flex;
  margin-top: 10px;
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--info-bg);
  color: var(--info);
  font-size: 0.75rem;
  font-weight: 800;
}

.profile-completion {
  padding-top: 16px;
  border-top: 1px solid var(--border-light);
}

.profile-completion-top {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 7px;
  color: var(--muted);
  font-size: 0.8rem;
}

.profile-completion-top strong {
  color: var(--brand);
}

.profile-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 18px;
}

.profile-card-head h2 {
  color: var(--dark);
  font-size: 1rem;
  font-weight: 900;
}

.profile-card-head p {
  margin-top: 4px;
  color: var(--muted);
  font-size: 0.84rem;
  line-height: 1.45;
}

.profile-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 15px;
}

.profile-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.profile-field-full {
  grid-column: 1 / -1;
}

.profile-field span {
  color: var(--muted);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.profile-textarea {
  min-height: 118px;
  resize: vertical;
}

.profile-password-form {
  padding-top: 16px;
  border-top: 1px solid var(--border-light);
}

.profile-spinner {
  width: 16px;
  height: 16px;
}

.profile-skeleton-lines {
  flex: 1;
}

@media (max-width: 820px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }

  .profile-summary {
    position: static;
  }
}

@media (max-width: 560px) {
  .profile-header,
  .profile-card-head,
  .security-head {
    align-items: stretch;
    flex-direction: column;
  }

  .profile-form {
    grid-template-columns: 1fr;
  }

  .profile-header .btn {
    width: 100%;
  }
}
</style>
