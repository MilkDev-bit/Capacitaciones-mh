<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Error al iniciar sesión'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-bg">
    <div class="login-card">
      <div class="login-logo">
        <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="10" fill="#3b82f6"/>
          <path d="M10 28L20 12L30 28H10Z" fill="white"/>
        </svg>
        <h1>Capacitaciones MH</h1>
      </div>
      <h2>Iniciar sesión</h2>
      <form @submit.prevent="submit">
        <label>Correo electrónico</label>
        <input v-model="email" type="email" placeholder="correo@empresa.com" required />
        <label>Contraseña</label>
        <input v-model="password" type="password" placeholder="••••••••" required />
        <p v-if="error" class="error">{{ error }}</p>
        <button type="submit" :disabled="loading">
          {{ loading ? 'Entrando...' : 'Entrar' }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-bg {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1e3a5f 0%, #3b82f6 100%);
}
.login-card {
  background: white;
  border-radius: 16px;
  padding: 2.5rem 2rem;
  width: 100%;
  max-width: 380px;
  box-shadow: 0 20px 40px rgba(0,0,0,0.18);
}
.login-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 1.5rem;
}
.login-logo h1 { font-size: 1.1rem; font-weight: 700; color: #1e3a5f; }
h2 { font-size: 1.3rem; font-weight: 600; margin-bottom: 1.5rem; color: #334155; }
label { display: block; font-size: 0.8rem; font-weight: 600; color: #64748b; margin-bottom: 4px; margin-top: 1rem; }
input {
  width: 100%; padding: 10px 12px; border: 1.5px solid #e2e8f0;
  border-radius: 8px; font-size: 0.95rem; outline: none; transition: border 0.2s;
}
input:focus { border-color: #3b82f6; }
button {
  margin-top: 1.5rem; width: 100%; padding: 11px;
  background: #3b82f6; color: white; border: none;
  border-radius: 8px; font-size: 1rem; font-weight: 600; cursor: pointer; transition: background 0.2s;
}
button:hover:not(:disabled) { background: #2563eb; }
button:disabled { opacity: 0.6; cursor: not-allowed; }
.error { color: #ef4444; font-size: 0.85rem; margin-top: 8px; }
</style>
