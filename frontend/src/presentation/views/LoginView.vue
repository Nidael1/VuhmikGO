<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/app/stores/auth'
import { authRepository } from '@/infrastructure/repositories/authRepository'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const mode = ref<'login' | 'register'>('login')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const tokens = mode.value === 'login'
      ? await authRepository.login({ email: email.value, password: password.value })
      : await authRepository.register({ email: email.value, password: password.value })
    auth.setSession(tokens)
    router.push('/patients')
  } catch (e: any) {
    error.value = e.message || 'Error al iniciar sesión'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-shell">
    <div class="login-card">
      <div class="login-brand">
        <div class="brand-icon">V</div>
        <h1 class="brand-name">vuhmik</h1>
        <p class="brand-tagline">Sistema clínico para médicos independientes</p>
      </div>

      <form class="login-form" @submit.prevent="submit">
        <h2 class="form-title">
          {{ mode === 'login' ? 'Iniciar sesión' : 'Crear cuenta' }}
        </h2>

        <div class="form-group">
          <label for="email">Correo electrónico</label>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="doctor@ejemplo.com"
            required
            autocomplete="email"
          />
        </div>

        <div class="form-group">
          <label for="password">Contraseña</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="Mínimo 8 caracteres"
            required
            autocomplete="current-password"
          />
        </div>

        <div class="form-error" v-if="error">
          {{ error }}
        </div>

        <button type="submit" class="btn-primary" :disabled="loading">
          {{ loading ? 'Cargando...' : mode === 'login' ? 'Entrar' : 'Registrarse' }}
        </button>

        <button type="button" class="btn-toggle" @click="mode = mode === 'login' ? 'register' : 'login'">
          {{ mode === 'login' ? '¿Sin cuenta? Regístrate' : '¿Ya tienes cuenta? Entra' }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-shell {
  min-height: 100vh;
  background: var(--color-obsidian);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
}

.login-card {
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
  width: 100%;
  max-width: 420px;
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.login-brand {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
}

.brand-icon {
  width: 48px;
  height: 48px;
  background: var(--color-jade);
  color: var(--color-obsidian);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-brand);
  font-weight: 700;
  font-size: 24px;
  margin-bottom: var(--space-2);
}

.brand-name {
  font-family: var(--font-brand);
  font-size: 28px;
  font-weight: 700;
  color: var(--color-obsidian);
  letter-spacing: -0.02em;
}

.brand-tagline {
  font-size: 14px;
  color: var(--text-secondary);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.form-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--space-2);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

input {
  font-family: var(--font-body);
  padding: var(--space-3) var(--space-4);
  border: 1.5px solid #E2E8F0;
  border-radius: var(--radius-md);
  font-size: 15px;
  color: var(--text-primary);
  background: var(--app-bg);
  transition: border-color 0.15s;
  outline: none;
}

input:focus {
  border-color: var(--color-turquoise);
}

.form-error {
  background: #FFF0F3;
  border: 1px solid var(--color-error);
  border-radius: var(--radius-sm);
  padding: var(--space-3) var(--space-4);
  font-size: 14px;
  color: var(--color-error);
}

.btn-primary {
  font-family: var(--font-brand);
  background: var(--action-primary-bg);
  color: var(--action-primary-text);
  border: none;
  padding: var(--space-3) var(--space-6);
  border-radius: var(--radius-md);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-toggle {
  background: transparent;
  border: none;
  color: var(--color-clinical-blue);
  font-size: 14px;
  cursor: pointer;
  padding: 0;
  text-align: center;
}

.btn-toggle:hover {
  text-decoration: underline;
}
</style>
