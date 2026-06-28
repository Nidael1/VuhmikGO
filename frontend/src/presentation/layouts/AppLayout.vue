<script setup lang="ts">
import { useAuthStore } from '@/app/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

async function logout() {
  // Revocar refresh token en servidor antes de limpiar sesion local
  if (auth.refreshToken) {
    try {
      await fetch('/api/v1/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: auth.refreshToken }),
      })
    } catch { /* si falla la revocacion, igual limpiamos local */ }
  }
  auth.clearSession()
  router.push('/login')
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="sidebar-brand">
        <span class="brand-icon">V</span>
        <span class="brand-name">vuhmik</span>
      </div>
      <nav class="sidebar-nav">
        <RouterLink to="/patients" class="nav-item">
          <span class="nav-icon">👥</span>
          <span>Pacientes</span>
        </RouterLink>
        <RouterLink to="/consultations" class="nav-item">
          <span class="nav-icon">🩺</span>
          <span>Consultas</span>
        </RouterLink>
        <RouterLink to="/prescriptions" class="nav-item">
          <span class="nav-icon">📋</span>
          <span>Recetas</span>
        </RouterLink>

      </nav>
      <div class="sidebar-footer">
        <RouterLink to="/profile" class="nav-item-profile">
          <span class="nav-icon">👤</span>
          <span>Mi perfil</span>
        </RouterLink>
        <div class="user-info" v-if="auth.profile">
          <span class="user-actor">{{ auth.profile.actor_id }}</span>
        </div>
        <button class="btn-logout" @click="logout">Cerrar sesión</button>
      </div>
    </aside>
    <main class="main-content">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.app-shell { display: flex; min-height: 100vh; background: var(--app-bg); }
.sidebar { width: 240px; min-height: 100vh; background: var(--app-sidebar-bg); display: flex; flex-direction: column; padding: var(--space-4); gap: var(--space-6); position: fixed; top: 0; left: 0; }
.sidebar-brand { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-4) 0; }
.brand-icon { width: 32px; height: 32px; background: var(--color-jade); color: var(--color-obsidian); border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; font-family: var(--font-brand); font-weight: 700; font-size: 16px; }
.brand-name { font-family: var(--font-brand); font-weight: 700; font-size: 18px; color: var(--text-on-dark); letter-spacing: -0.02em; }
.sidebar-nav { display: flex; flex-direction: column; gap: var(--space-2); flex: 1; }
.nav-item { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); color: var(--text-on-dark); opacity: 0.7; text-decoration: none; font-size: 14px; transition: all 0.15s; }
.nav-item:hover, .nav-item.router-link-active { background: rgba(255,255,255,0.08); opacity: 1; color: var(--color-turquoise); text-decoration: none; }
.nav-icon { font-size: 16px; }
.sidebar-footer { display: flex; flex-direction: column; gap: var(--space-2); border-top: 1px solid rgba(255,255,255,0.08); padding-top: var(--space-4); }
.user-actor { font-size: 12px; color: var(--text-on-dark); opacity: 0.5; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.btn-logout { background: transparent; border: 1px solid rgba(255,255,255,0.15); color: var(--text-on-dark); padding: var(--space-2) var(--space-4); border-radius: var(--radius-sm); font-size: 13px; cursor: pointer; text-align: left; transition: all 0.15s; }
.btn-logout:hover { border-color: var(--color-error); color: var(--color-error); }
.nav-item-profile { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-2) var(--space-3); border-radius: var(--radius-md); color: var(--text-on-dark); opacity: 0.6; text-decoration: none; font-size: 13px; transition: all 0.15s; }
.nav-item-profile:hover, .nav-item-profile.router-link-active { opacity: 1; color: var(--color-turquoise); }
.main-content { flex: 1; margin-left: 240px; padding: var(--space-8); min-height: 100vh; }
</style>
