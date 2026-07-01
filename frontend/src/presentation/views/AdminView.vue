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

// Formulario nuevo médico
const showCreateForm = ref(false)
const createLoading = ref(false)
const createError = ref('')
const createSuccess = ref('')
const createForm = ref({
  email: '',
  password: '',
  nombre_completo: '',
  cedula_profesional: '',
  especialidad: '',
  universidad: '',
  direccion: '',
  telefono: '',
  curp: '',
})

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

function resetCreateForm() {
  createForm.value = {
    email: '',
    password: '',
    nombre_completo: '',
    cedula_profesional: '',
    especialidad: '',
    universidad: '',
    direccion: '',
    telefono: '',
    curp: '',
  }
  createError.value = ''
  createSuccess.value = ''
}

async function loadTenants() {
  const res = await http.get<any>('/admin/tenants')
  tenants.value = res.data?.items ?? []
}

onMounted(async () => {
  try {
    await loadTenants()
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
    await loadTenants()
  } catch (e: any) {
    error.value = e.message
  }
}

async function createMedico() {
  createError.value = ''
  createSuccess.value = ''
  createLoading.value = true
  try {
    const payload: Record<string, string> = {
      email: createForm.value.email.trim(),
      password: createForm.value.password,
      nombre_completo: createForm.value.nombre_completo.trim(),
      cedula_profesional: createForm.value.cedula_profesional.trim(),
      especialidad: createForm.value.especialidad.trim(),
      universidad: createForm.value.universidad.trim(),
      direccion: createForm.value.direccion.trim(),
      telefono: createForm.value.telefono.trim(),
    }
    if (createForm.value.curp.trim()) {
      payload.curp = createForm.value.curp.trim()
    }
    const res = await http.post<any>('/admin/users', payload)
    createSuccess.value = `Médico creado: ${res.data?.email} — Módulos activos: ${res.data?.modulos_activos?.join(', ')}`
    await loadTenants()
    resetCreateForm()
    showCreateForm.value = false
  } catch (e: any) {
    createError.value = e.message
  } finally {
    createLoading.value = false
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
        <div class="page-header">
          <div>
            <h2>Panel de control</h2>
            <p class="page-sub">Gestión de médicos y módulos activos</p>
          </div>
          <button class="btn-primary" @click="showCreateForm = !showCreateForm">
            {{ showCreateForm ? 'Cancelar' : '+ Nuevo médico' }}
          </button>
        </div>

        <!-- Mensaje de éxito -->
        <div v-if="createSuccess" class="alert-success">{{ createSuccess }}</div>

        <!-- Formulario alta de médico -->
        <div v-if="showCreateForm" class="create-form-card">
          <h3 class="form-title">Alta de médico</h3>
          <p class="form-subtitle">Todos los campos son obligatorios para cumplir NOM-024-SSA3-2012, excepto CURP.</p>

          <div class="alert-error" v-if="createError">{{ createError }}</div>

          <div class="form-grid">
            <div class="form-group form-group--full">
              <label>Nombre completo *</label>
              <input v-model="createForm.nombre_completo" class="input-field" placeholder="DR. JUAN PÉREZ GARCÍA" />
            </div>
            <div class="form-group">
              <label>Correo electrónico *</label>
              <input v-model="createForm.email" class="input-field" type="email" placeholder="dr.juan@ejemplo.com" />
            </div>
            <div class="form-group">
              <label>Contraseña inicial *</label>
              <input v-model="createForm.password" class="input-field" type="password" placeholder="Mínimo 8 caracteres" />
            </div>
            <div class="form-group">
              <label>Cédula profesional *</label>
              <input v-model="createForm.cedula_profesional" class="input-field" placeholder="12345678" />
            </div>
            <div class="form-group">
              <label>Especialidad *</label>
              <input v-model="createForm.especialidad" class="input-field" placeholder="Medicina General" />
            </div>
            <div class="form-group">
              <label>Universidad *</label>
              <input v-model="createForm.universidad" class="input-field" placeholder="UNAM" />
            </div>
            <div class="form-group">
              <label>Teléfono *</label>
              <input v-model="createForm.telefono" class="input-field" placeholder="5512345678" />
            </div>
            <div class="form-group form-group--full">
              <label>Dirección del consultorio *</label>
              <input v-model="createForm.direccion" class="input-field" placeholder="Av. Insurgentes 123, Col. Roma, CDMX" />
            </div>
            <div class="form-group">
              <label>CURP <span class="optional">(opcional)</span></label>
              <input v-model="createForm.curp" class="input-field" placeholder="XXXXX000000XXXXXX00" maxlength="18" style="text-transform:uppercase" />
            </div>
          </div>

          <div class="form-modules-note">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="8" x2="12" y2="12"/>
              <line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            Se activarán automáticamente los módulos: Alergias, Recetas y Notas clínicas.
          </div>

          <div class="form-actions">
            <button class="btn-secondary" @click="showCreateForm = false; resetCreateForm()">Cancelar</button>
            <button class="btn-primary" @click="createMedico" :disabled="createLoading">
              {{ createLoading ? 'Creando...' : 'Crear médico' }}
            </button>
          </div>
        </div>

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

.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-4); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: 2px; }

/* Formulario alta */
.create-form-card {
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  margin-bottom: var(--space-6);
}
.form-title { font-size: 15px; font-weight: 700; color: var(--text-primary); margin: 0 0 var(--space-1) 0; }
.form-subtitle { font-size: 13px; color: var(--text-secondary); margin: 0 0 var(--space-4) 0; }
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}
.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-group--full { grid-column: 1 / -1; }
.form-group label { font-size: 12px; font-weight: 600; color: var(--text-secondary); }
.optional { font-weight: 400; }
.input-field {
  font-family: var(--font-body);
  padding: var(--space-2) var(--space-3);
  border: 1.5px solid #E2E8F0;
  border-radius: var(--radius-md);
  font-size: 14px;
  color: var(--text-primary);
  background: var(--app-bg);
  outline: none;
}
.input-field:focus { border-color: var(--color-turquoise); }

.form-modules-note {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 12px;
  color: var(--text-secondary);
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-sm);
  padding: var(--space-3) var(--space-4);
  margin-bottom: var(--space-4);
}

.form-actions { display: flex; gap: var(--space-3); justify-content: flex-end; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-2) var(--space-5); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { font-family: var(--font-brand); background: transparent; color: var(--text-primary); border: 1.5px solid #E2E8F0; padding: var(--space-2) var(--space-5); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-secondary:hover { border-color: var(--color-turquoise); }

.alert-success { background: #F0FDF4; border: 1px solid #86EFAC; border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: #166534; margin-bottom: var(--space-4); }

/* Lista tenants */
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
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); margin-bottom: var(--space-4); }
</style>
