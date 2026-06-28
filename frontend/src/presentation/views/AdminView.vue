<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/app/stores/auth'
import { useRouter } from 'vue-router'
import { http } from '@/infrastructure/api/httpClient'

const auth = useAuthStore()
const router = useRouter()

interface ModuleStatus {
  ModuleID: string
  Descripcion: string
  Active: boolean
  Plan: string
  Costo: number
}

interface TenantInfo {
  tenant_id: string
  email: string
  is_admin: boolean
  is_suspended: boolean
  modules: ModuleStatus[]
}

const tenants = ref<TenantInfo[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const expandedTenants = ref<Set<string>>(new Set())

const filtered = computed(() => {
  const q = search.value.toLowerCase().trim()
  if (!q) return tenants.value
  return tenants.value.filter(t =>
    t.email.toLowerCase().includes(q) ||
    t.tenant_id.toLowerCase().includes(q)
  )
})

function toggleExpand(tenantId: string) {
  if (expandedTenants.value.has(tenantId)) {
    expandedTenants.value.delete(tenantId)
  } else {
    expandedTenants.value.add(tenantId)
  }
}

onMounted(async () => {
  try {
    const res = await http.get<any>('/admin/tenants')
    tenants.value = res.data?.items ?? []
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

async function toggleModule(tenantId: string, moduleId: string, active: boolean) {
  try {
    await http.post('/admin/capabilities', {
      tenant_id: tenantId,
      module_id: moduleId,
      active: !active,
    })
    // Recargar
    const res = await http.get<any>('/admin/tenants')
    tenants.value = res.data?.items ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function logout() {
  if (auth.refreshToken) {
    try {
      await fetch('/api/v1/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: auth.refreshToken }),
      })
    } catch {}
  }
  auth.clearSession()
  router.push('/login')
}
</script>

<template>
  <div class="admin-shell">
    <aside class="admin-sidebar">
      <div class="admin-brand">
        <span class="brand-icon">V</span>
        <span class="brand-name">vuhmik</span>
        <span class="admin-badge">admin</span>
      </div>
      <div class="admin-footer">
        <span class="admin-user">{{ auth.profile?.actor_id }}</span>
        <button class="btn-logout" @click="logout">Cerrar sesión</button>
      </div>
    </aside>

    <main class="admin-main">
      <div class="admin-page">
        <h2>Panel de control</h2>
        <p class="page-sub">Módulos activos por médico</p>

        <div v-if="loading" class="state-empty">Cargando...</div>
        <div v-else-if="error" class="alert-error">{{ error }}</div>

        <div v-else-if="tenants.length === 0" class="state-empty">
          No hay médicos registrados.
        </div>

        <div v-else>
          <input
            v-model="search"
            class="search-input"
            placeholder="Buscar por correo o tenant..."
          />
          <p v-if="filtered.length === 0" class="state-empty">Sin resultados.</p>
          <div v-else class="tenant-list">
            <div v-for="t in filtered" :key="t.tenant_id" class="tenant-card">
              <div class="tenant-header" @click="toggleExpand(t.tenant_id)" style="cursor:pointer;">
                <div class="tenant-header-left">
                  <span class="tenant-expand">{{ expandedTenants.has(t.tenant_id) ? '▾' : '▸' }}</span>
                  <span class="tenant-email">{{ t.email }}</span>
                  <span v-if="t.is_suspended" class="badge-suspended">Suspendido</span>
                  <span v-if="t.is_admin" class="badge-admin">Admin</span>
                </div>
                <span class="tenant-id">{{ t.tenant_id }}</span>
              </div>
              <div v-if="expandedTenants.has(t.tenant_id)" class="module-list">
                <div v-for="m in t.modules" :key="m.ModuleID" class="module-row">
                  <span class="module-name">{{ m.Descripcion || m.ModuleID }}</span>
                  <button
                    :class="['toggle-btn', m.Active ? 'active' : '']"
                    @click="toggleModule(t.tenant_id, m.ModuleID, m.Active)"
                  >
                    {{ m.Active ? 'Activo' : 'Inactivo' }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.admin-shell { display: flex; min-height: 100vh; background: var(--app-bg); }
.admin-sidebar { width: 240px; min-height: 100vh; background: var(--app-sidebar-bg); display: flex; flex-direction: column; padding: var(--space-4); position: fixed; top: 0; left: 0; }
.admin-brand { display: flex; align-items: center; gap: var(--space-2); padding: var(--space-4) 0; }
.brand-icon { width: 32px; height: 32px; background: var(--color-jade); color: var(--color-obsidian); border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 16px; }
.brand-name { font-family: var(--font-brand); font-weight: 700; font-size: 18px; color: var(--text-on-dark); }
.admin-badge { font-size: 10px; font-weight: 700; background: var(--color-warning, #FFB020); color: #000; border-radius: 4px; padding: 2px 6px; text-transform: uppercase; }
.admin-footer { margin-top: auto; display: flex; flex-direction: column; gap: var(--space-2); border-top: 1px solid rgba(255,255,255,0.08); padding-top: var(--space-4); }
.admin-user { font-size: 12px; color: var(--text-on-dark); opacity: 0.5; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.btn-logout { background: transparent; border: 1px solid rgba(255,255,255,0.15); color: var(--text-on-dark); padding: var(--space-2) var(--space-4); border-radius: var(--radius-sm); font-size: 13px; cursor: pointer; text-align: left; }
.btn-logout:hover { border-color: var(--color-error); color: var(--color-error); }
.admin-main { flex: 1; margin-left: 240px; padding: var(--space-8); }
.admin-page { max-width: 780px; }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: 2px; margin-bottom: var(--space-6); }
.tenant-list { display: flex; flex-direction: column; gap: var(--space-4); }
.tenant-card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-lg); overflow: hidden; }
.tenant-header { display: flex; justify-content: space-between; align-items: center; padding: var(--space-4) var(--space-6); background: #FAFBFC; border-bottom: 1px solid #E2E8F0; }
.tenant-email { font-weight: 600; font-size: 14px; }
.tenant-id { font-size: 12px; color: var(--text-secondary); font-family: monospace; }
.module-list { padding: var(--space-2) 0; }
.module-row { display: flex; align-items: center; justify-content: space-between; padding: var(--space-3) var(--space-6); border-bottom: 1px solid #F1F5F9; }
.module-row:last-child { border-bottom: none; }
.module-name { font-size: 14px; color: var(--text-primary); }
.toggle-btn { font-size: 12px; font-weight: 600; padding: 4px 12px; border-radius: var(--radius-sm); border: 1.5px solid #E2E8F0; background: var(--app-surface); color: var(--text-secondary); cursor: pointer; transition: all 0.15s; }
.toggle-btn.active { background: #DCFCE7; border-color: #86EFAC; color: #166534; }
.toggle-btn:not(.active):hover { border-color: var(--color-error); color: var(--color-error); }
.state-empty { color: var(--text-secondary); padding: var(--space-8); text-align: center; }
.search-input { width: 100%; font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-surface); outline: none; margin-bottom: var(--space-4); box-sizing: border-box; }
.search-input:focus { border-color: var(--color-turquoise); }
.tenant-header-left { display: flex; align-items: center; gap: var(--space-2); }
.tenant-expand { font-size: 12px; color: var(--text-secondary); width: 14px; }
.badge-suspended { font-size: 11px; font-weight: 700; background: #FEE2E2; color: #991B1B; border-radius: 4px; padding: 1px 6px; }
.badge-admin { font-size: 11px; font-weight: 700; background: #FEF9C3; color: #854D0E; border-radius: 4px; padding: 1px 6px; }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); }
</style>
