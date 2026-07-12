<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/app/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()
const menuAbierto = ref(false)

function cerrarMenu() { menuAbierto.value = false }

async function logout() {
  if (auth.refreshToken) {
    try {
      await fetch('/api/v1/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: auth.refreshToken }),
      })
    } catch { }
  }
  auth.clearSession()
  router.push('/login')
}
</script>

<template>
  <div class="app-shell">
    <!-- Topbar movil -->
    <header class="topbar">
      <button class="hamburger" @click="menuAbierto = !menuAbierto" aria-label="Menu">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line v-if="menuAbierto" x1="18" y1="6" x2="6" y2="18"/><line v-if="menuAbierto" x1="6" y1="6" x2="18" y2="18"/>
          <line v-if="!menuAbierto" x1="3" y1="6" x2="21" y2="6"/><line v-if="!menuAbierto" x1="3" y1="12" x2="21" y2="12"/><line v-if="!menuAbierto" x1="3" y1="18" x2="21" y2="18"/>
        </svg>
      </button>
      <span class="topbar-brand">vuhmik</span>
    </header>

    <!-- Overlay movil -->
    <div v-if="menuAbierto" class="overlay" @click="cerrarMenu" />

    <!-- Sidebar -->
    <aside class="sidebar" :class="{ 'sidebar-open': menuAbierto }">
      <div class="sidebar-brand">
        <span class="brand-icon">V</span>
        <span class="brand-name">vuhmik</span>
      </div>
      <nav class="sidebar-nav">
        <RouterLink to="/patients" class="nav-item" @click="cerrarMenu">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
          <span>Pacientes</span>
        </RouterLink>
        <RouterLink to="/consultations" class="nav-item" @click="cerrarMenu">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
          <span>Consultas</span>
        </RouterLink>
        <RouterLink to="/prescriptions" class="nav-item" @click="cerrarMenu">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
          <span>Recetas</span>
        </RouterLink>
      </nav>
      <div class="sidebar-footer">
        <RouterLink to="/profile" class="nav-item-profile" @click="cerrarMenu">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
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

/* Topbar — solo visible en movil */
.topbar {
  display: none;
  position: fixed; top: 0; left: 0; right: 0; z-index: 100;
  height: 52px; background: var(--app-sidebar-bg);
  align-items: center; padding: 0 16px; gap: 12px;
  box-shadow: 0 1px 0 rgba(255,255,255,0.06);
}
.topbar-brand { font-family: var(--font-brand); font-weight: 700; font-size: 17px; color: var(--text-on-dark); letter-spacing: -0.02em; }
.hamburger { background: none; border: none; color: var(--text-on-dark); cursor: pointer; padding: 4px; display: flex; align-items: center; justify-content: center; border-radius: 6px; }
.hamburger:hover { background: rgba(255,255,255,0.08); }

/* Overlay movil */
.overlay { display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 150; }

/* Sidebar */
.sidebar {
  width: 240px; min-height: 100vh;
  background: var(--app-sidebar-bg);
  display: flex; flex-direction: column;
  padding: var(--space-4); gap: var(--space-6);
  position: fixed; top: 0; left: 0; z-index: 200;
}
.sidebar-brand { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-4) 0; }
.brand-icon { width: 32px; height: 32px; background: var(--color-jade); color: var(--color-obsidian); border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; font-family: var(--font-brand); font-weight: 700; font-size: 16px; }
.brand-name { font-family: var(--font-brand); font-weight: 700; font-size: 18px; color: var(--text-on-dark); letter-spacing: -0.02em; }
.sidebar-nav { display: flex; flex-direction: column; gap: var(--space-2); flex: 1; }
.nav-item { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); color: var(--text-on-dark); opacity: 0.7; text-decoration: none; font-size: 14px; transition: all 0.15s; }
.nav-item:hover, .nav-item.router-link-active { background: rgba(255,255,255,0.08); opacity: 1; color: var(--color-turquoise); text-decoration: none; }
.sidebar-footer { display: flex; flex-direction: column; gap: var(--space-2); border-top: 1px solid rgba(255,255,255,0.08); padding-top: var(--space-4); }
.user-actor { font-size: 12px; color: var(--text-on-dark); opacity: 0.5; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.btn-logout { background: transparent; border: 1px solid rgba(255,255,255,0.15); color: var(--text-on-dark); padding: var(--space-2) var(--space-4); border-radius: var(--radius-sm); font-size: 13px; cursor: pointer; text-align: left; transition: all 0.15s; }
.btn-logout:hover { border-color: var(--color-error); color: var(--color-error); }
.nav-item-profile { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-2) var(--space-3); border-radius: var(--radius-md); color: var(--text-on-dark); opacity: 0.6; text-decoration: none; font-size: 13px; transition: all 0.15s; }
.nav-item-profile:hover, .nav-item-profile.router-link-active { opacity: 1; color: var(--color-turquoise); }

/* Contenido principal */
.main-content { flex: 1; margin-left: 240px; padding: var(--space-8); min-height: 100vh; }

/* Responsive movil */
@media (max-width: 768px) {
  .topbar { display: flex; }
  .overlay { display: block; }
  .sidebar {
    top: 0; left: -240px;
    transition: transform 0.25s ease;
    transform: translateX(0);
  }
  .sidebar-open { transform: translateX(240px); }
  .sidebar .sidebar-brand { padding-top: 60px; }
  .main-content { margin-left: 0; padding: 72px 16px 24px; }
}
</style>
