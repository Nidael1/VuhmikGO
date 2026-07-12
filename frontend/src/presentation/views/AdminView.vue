<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/app/stores/auth'
import { useRouter } from 'vue-router'
import { http } from '@/infrastructure/api/httpClient'

const auth = useAuthStore()
const router = useRouter()

type Section = 'operaciones' | 'metricas' | 'actividad' | 'salud' | 'sistema'
const activeSection = ref<Section>('operaciones')
const health = ref<any[]>([])
const healthSummary = ref({ active: 0, at_risk: 0, inactive: 0 })
const healthFilter = ref('')
const loadingHealth = ref(false)
const errorHealth = ref('')
const systemSnap = ref<any>(null)
const failedLogins = ref<any[]>([])
const loadingSystem = ref(false)
const errorSystem = ref('')

interface ModuleStatus { ModuleID: string; Descripcion: string; Active: boolean; Plan: string; Costo: number }
interface TenantInfo { tenant_id: string; email: string; is_admin: boolean; is_suspended: boolean; modules: ModuleStatus[] }
const tenants = ref<TenantInfo[]>([])
const loadingTenants = ref(true)
const errorTenants = ref('')
const search = ref('')
const expandedTenants = ref<Set<string>>(new Set())
const editingTenant = ref<string | null>(null)
const editMode = ref<'profile' | 'billing' | 'password' | null>(null)
const editProfileForm = ref({ nombre_completo: '', cedula_profesional: '', especialidad: '', universidad: '', direccion: '', telefono: '' })
const editBillingForm = ref({ billing_mode: 'per_module', monthly_fee: 0 })
const editPasswordForm = ref({ new_password: '' })
const editLoading = ref(false)
const editError = ref('')
const editSuccess = ref('')
const showCreateForm = ref(false)
const createLoading = ref(false)
const createError = ref('')
const createSuccess = ref('')
const createForm = ref({ email: '', password: '', nombre_completo: '', cedula_profesional: '', especialidad: '', universidad: '', direccion: '', telefono: '', curp: '' })
const filtered = computed(() => { const q = search.value.toLowerCase().trim(); if (!q) return tenants.value; return tenants.value.filter(t => t.email.toLowerCase().includes(q) || t.tenant_id.toLowerCase().includes(q)) })
function toggleExpand(id: string) { expandedTenants.value.has(id) ? expandedTenants.value.delete(id) : expandedTenants.value.add(id) }

function startEdit(tenantId: string, mode: 'profile' | 'billing' | 'password', t: any) {
  editingTenant.value = tenantId; editMode.value = mode; editError.value = ''; editSuccess.value = ''
  if (mode === 'profile') { editProfileForm.value = { nombre_completo: t.profile?.nombre_completo || '', cedula_profesional: t.profile?.cedula_profesional || '', especialidad: t.profile?.especialidad || '', universidad: t.profile?.universidad || '', direccion: t.profile?.direccion || '', telefono: t.profile?.telefono || '' } }
  else if (mode === 'billing') { editBillingForm.value = { billing_mode: t.billing_mode || 'per_module', monthly_fee: t.monthly_fee || 0 } }
  else { editPasswordForm.value = { new_password: '' } }
}
function cancelEdit() { editingTenant.value = null; editMode.value = null; editError.value = ''; editSuccess.value = '' }
async function saveEdit(tenantId: string) {
  editLoading.value = true; editError.value = ''; editSuccess.value = ''
  try {
    const token = auth.token
    const headers = { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` }
    let url = ''; let body: any = {}
    if (editMode.value === 'profile') { url = `/api/v1/admin/users/${tenantId}/profile`; body = editProfileForm.value }
    else if (editMode.value === 'billing') { url = `/api/v1/admin/users/${tenantId}/billing`; body = editBillingForm.value }
    else { url = `/api/v1/admin/users/${tenantId}/password`; body = editPasswordForm.value }
    const res = await fetch(url, { method: 'PUT', headers, body: JSON.stringify(body) })
    const data = await res.json()
    if (data.error) throw new Error(data.error.message)
    editSuccess.value = 'Guardado correctamente'
    if (editMode.value === 'billing') { try { await http.post('/admin/metrics/recalculate', {}) } catch (_) {} }
    setTimeout(() => { cancelEdit(); loadTenants() }, 1200)
  } catch (e: any) { editError.value = e.message } finally { editLoading.value = false }
}
function resetCreateForm() { createForm.value = { email: '', password: '', nombre_completo: '', cedula_profesional: '', especialidad: '', universidad: '', direccion: '', telefono: '', curp: '' }; createError.value = ''; createSuccess.value = '' }
async function loadTenants() { const res = await http.get<any>('/admin/tenants'); tenants.value = res.data?.items ?? [] }
async function toggleModule(tenantId: string, moduleId: string, active: boolean) { try { await http.post('/admin/capabilities', { tenant_id: tenantId, module_id: moduleId, active: !active }); await loadTenants() } catch (e: any) { errorTenants.value = e.message } }
async function createMedico() { createError.value = ''; createSuccess.value = ''; createLoading.value = true; try { const payload: Record<string, string> = { email: createForm.value.email.trim(), password: createForm.value.password, nombre_completo: createForm.value.nombre_completo.trim(), cedula_profesional: createForm.value.cedula_profesional.trim(), especialidad: createForm.value.especialidad.trim(), universidad: createForm.value.universidad.trim(), direccion: createForm.value.direccion.trim(), telefono: createForm.value.telefono.trim() }; if (createForm.value.curp.trim()) payload.curp = createForm.value.curp.trim(); const res = await http.post<any>('/admin/users', payload); createSuccess.value = `Médico creado: ${res.data?.email} — Módulos: ${res.data?.modulos_activos?.join(', ')}`; await loadTenants(); resetCreateForm(); showCreateForm.value = false } catch (e: any) { createError.value = e.message } finally { createLoading.value = false } }

interface MetricsSnapshot { calculated_at: string; total_accounts: number; active_accounts: number; suspended_accounts: number; mrr: number; total_patients: number; total_notes: number; total_allergies: number; total_prescriptions: number }
interface AccountDetail { tenant_id: string; email: string; state: string; mrr: number; patients: number; last_record: string }
const metrics = ref<MetricsSnapshot | null>(null)
const metricsAccounts = ref<AccountDetail[]>([])
const metricsModules = ref<Record<string, number>>({})
const loadingMetrics = ref(false)
const errorMetrics = ref('')
async function loadMetrics() { loadingMetrics.value = true; errorMetrics.value = ''; metrics.value = null; metricsAccounts.value = []; metricsModules.value = {}; try { try { await http.post('/admin/metrics/recalculate', {}) } catch (_) {}; const [snapRes, accRes, modRes] = await Promise.all([http.get<any>('/admin/metrics'), http.get<any>('/admin/metrics/accounts'), http.get<any>('/admin/metrics/modules')]); metrics.value = snapRes.data ?? null; const rawAcc = accRes.data?.accounts; metricsAccounts.value = typeof rawAcc === 'string' ? JSON.parse(rawAcc) : (rawAcc ?? []); const rawMod = modRes.data?.modules; metricsModules.value = typeof rawMod === 'string' ? JSON.parse(rawMod) : (rawMod ?? {}) } catch (e: any) { errorMetrics.value = e.message?.includes('NO_SNAPSHOT') ? 'El worker aún no ha calculado métricas. Estará disponible en las próximas horas.' : e.message } finally { loadingMetrics.value = false } }
function fmtMXN(v: number) { return new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN', maximumFractionDigits: 0 }).format(v) }
function fmtDate(s: string) { if (!s) return '—'; return new Date(s).toLocaleString('es-MX', { dateStyle: 'short', timeStyle: 'short' }) }

interface ActivityItem { tenant_id: string; sessions_total: number; notes_total: number; allergies_total: number; prescriptions_total: number; exports_total: number; patients_total: number; last_period: string }
interface PeriodItem { period: string; sessions_count: number; notes_count: number; allergies_count: number; prescriptions_count: number; exports_count: number; patients_count: number }
const activity = ref<ActivityItem[]>([])
const activityDetail = ref<PeriodItem[]>([])
const selectedTenant = ref('')
const loadingActivity = ref(false)
const errorActivity = ref('')
const loadingDetail = ref(false)
async function loadActivity() { loadingActivity.value = true; errorActivity.value = ''; try { const res = await http.get<any>('/admin/activity'); activity.value = res.data?.items ?? [] } catch (e: any) { errorActivity.value = e.message } finally { loadingActivity.value = false } }
async function loadActivityDetail(tenantId: string) { selectedTenant.value = tenantId; loadingDetail.value = true; try { const res = await http.get<any>(`/admin/activity/${tenantId}`); activityDetail.value = res.data?.periods ?? [] } catch { activityDetail.value = [] } finally { loadingDetail.value = false } }

onMounted(async () => { try { await loadTenants() } catch (e: any) { errorTenants.value = e.message } finally { loadingTenants.value = false } })
async function switchSection(s: Section) {
  activeSection.value = s
  if (s === 'metricas' && !metrics.value && !loadingMetrics.value) await loadMetrics()
  if (s === 'actividad' && activity.value.length === 0 && !loadingActivity.value) await loadActivity()
  if (s === 'salud' && health.value.length === 0 && !loadingHealth.value) await loadHealth()
  if (s === 'sistema' && !systemSnap.value && !loadingSystem.value) await loadSystem()
}
async function loadHealth() {
  loadingHealth.value = true; errorHealth.value = ''; health.value = []
  try {
    const url = healthFilter.value ? `/admin/health/accounts?status=${healthFilter.value}` : '/admin/health/accounts'
    const res = await http.get<any>(url) as any
    health.value = res.data?.items ?? []
    healthSummary.value = res.data?.summary ?? { active: 0, at_risk: 0, inactive: 0 }
  } catch (e: any) { errorHealth.value = e.message } finally { loadingHealth.value = false }
}
async function loadSystem() {
  loadingSystem.value = true; errorSystem.value = ''; systemSnap.value = null
  try {
    await http.post('/admin/system/recalculate', {})
    const res = await http.get<any>('/admin/system') as any
    systemSnap.value = res.data?.system ?? null
    failedLogins.value = res.data?.failed_logins ?? []
  } catch (e: any) { errorSystem.value = e.message } finally { loadingSystem.value = false }
}
function healthLabel(status: string) {
  if (status === 'active') return 'Activo'
  if (status === 'at_risk') return 'En riesgo'
  return 'Inactivo'
}
function fmtDaysAgo(days: number) {
  if (days === 0) return 'Hoy'
  if (days === 1) return 'Ayer'
  if (days > 900) return 'Nunca'
  return `Hace ${days} dias`
}
async function logout() { if (auth.refreshToken) { try { await fetch('/api/v1/auth/logout', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ refresh_token: auth.refreshToken }) }) } catch {} } auth.clearSession(); router.push('/login') }
</script>

<template>
  <div class="admin-shell">
    <aside class="admin-sidebar">
      <div class="admin-brand">
        <span class="brand-icon">V</span>
        <span class="brand-name">vuhmik</span>
        <span class="admin-badge">admin</span>
      </div>
      <nav class="admin-nav">
        <button :class="['nav-item', activeSection === 'operaciones' ? 'active' : '']" @click="switchSection('operaciones')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
          Operaciones
        </button>
        <button :class="['nav-item', activeSection === 'metricas' ? 'active' : '']" @click="switchSection('metricas')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
          Métricas
        </button>
        <button :class="['nav-item', activeSection === 'actividad' ? 'active' : '']" @click="switchSection('actividad')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
          Actividad
        </button>
        <button :class="['nav-item', activeSection === 'salud' ? 'active' : '']" @click="switchSection('salud')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/></svg>
          Salud
        </button>
        <button :class="['nav-item', activeSection === 'sistema' ? 'active' : '']" @click="switchSection('sistema')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
          Sistema
        </button>
      </nav>
      <div class="admin-footer">
        <span class="admin-user">{{ auth.profile?.actor_id }}</span>
        <button class="btn-logout" @click="logout">Cerrar sesión</button>
      </div>
    </aside>

    <main class="admin-main">

      <!-- OPERACIONES -->
      <div v-if="activeSection === 'operaciones'" class="admin-page">
        <div class="page-header">
          <div><h2>Operaciones</h2><p class="page-sub">Gestión de médicos y módulos activos</p></div>
          <button class="btn-primary" @click="showCreateForm = !showCreateForm">{{ showCreateForm ? 'Cancelar' : '+ Nuevo médico' }}</button>
        </div>
        <div v-if="createSuccess" class="alert-success">{{ createSuccess }}</div>
        <div v-if="showCreateForm" class="create-form-card">
          <h3 class="form-title">Alta de médico</h3>
          <p class="form-subtitle">Todos los campos son obligatorios para cumplir NOM-024-SSA3-2012, excepto CURP.</p>
          <div class="alert-error" v-if="createError">{{ createError }}</div>
          <div class="form-grid">
            <div class="form-group form-group--full"><label>Nombre completo *</label><input v-model="createForm.nombre_completo" class="input-field" placeholder="DR. JUAN PÉREZ GARCÍA" /></div>
            <div class="form-group"><label>Correo electrónico *</label><input v-model="createForm.email" class="input-field" type="email" /></div>
            <div class="form-group"><label>Contraseña inicial *</label><input v-model="createForm.password" class="input-field" type="password" /></div>
            <div class="form-group"><label>Cédula profesional *</label><input v-model="createForm.cedula_profesional" class="input-field" /></div>
            <div class="form-group"><label>Especialidad *</label><input v-model="createForm.especialidad" class="input-field" /></div>
            <div class="form-group"><label>Universidad *</label><input v-model="createForm.universidad" class="input-field" /></div>
            <div class="form-group"><label>Teléfono *</label><input v-model="createForm.telefono" class="input-field" /></div>
            <div class="form-group form-group--full"><label>Dirección del consultorio *</label><input v-model="createForm.direccion" class="input-field" /></div>
            <div class="form-group"><label>CURP <span class="optional">(opcional)</span></label><input v-model="createForm.curp" class="input-field" maxlength="18" style="text-transform:uppercase" /></div>
          </div>
          <div class="form-modules-note">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            Se activarán automáticamente: Alergias, Recetas y Notas clínicas.
          </div>
          <div class="form-actions">
            <button class="btn-secondary" @click="showCreateForm = false; resetCreateForm()">Cancelar</button>
            <button class="btn-primary" @click="createMedico" :disabled="createLoading">{{ createLoading ? 'Creando...' : 'Crear médico' }}</button>
          </div>
        </div>
        <div v-if="loadingTenants" class="state-empty">Cargando...</div>
        <div v-else-if="errorTenants" class="alert-error">{{ errorTenants }}</div>
        <div v-else-if="tenants.length === 0" class="state-empty">No hay médicos registrados.</div>
        <div v-else>
          <input v-model="search" class="search-input" placeholder="Buscar por correo o tenant..." />
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
                  <button :class="['toggle-btn', m.Active ? 'active' : '']" @click="toggleModule(t.tenant_id, m.ModuleID, m.Active)">{{ m.Active ? 'Activo' : 'Inactivo' }}</button>
                </div>
                <div class="tenant-actions">
                  <button class="btn-accion" @click="startEdit(t.tenant_id, 'profile', t)">Editar perfil</button>
                  <button class="btn-accion" @click="startEdit(t.tenant_id, 'billing', t)">Facturación</button>
                  <button class="btn-accion" @click="startEdit(t.tenant_id, 'password', t)">Resetear contraseña</button>
                </div>
                <div v-if="editingTenant === t.tenant_id" class="edit-panel">
                  <div class="alert-error" v-if="editError">{{ editError }}</div>
                  <div class="alert-success" v-if="editSuccess">{{ editSuccess }}</div>
                  <template v-if="editMode === 'profile'">
                    <div class="edit-title">Editar perfil profesional</div>
                    <div class="edit-grid">
                      <div class="form-row"><label>Nombre completo</label><input v-model="editProfileForm.nombre_completo" class="input" placeholder="DR. JUAN PÉREZ" /></div>
                      <div class="form-row"><label>Cédula profesional</label><input v-model="editProfileForm.cedula_profesional" class="input" placeholder="1234567" /></div>
                      <div class="form-row"><label>Especialidad</label><input v-model="editProfileForm.especialidad" class="input" placeholder="Medicina General" /></div>
                      <div class="form-row"><label>Universidad</label><input v-model="editProfileForm.universidad" class="input" placeholder="UNAM" /></div>
                      <div class="form-row"><label>Dirección</label><input v-model="editProfileForm.direccion" class="input" placeholder="Consultorio..." /></div>
                      <div class="form-row"><label>Teléfono</label><input v-model="editProfileForm.telefono" class="input" placeholder="55 1234 5678" /></div>
                    </div>
                  </template>
                  <template v-else-if="editMode === 'billing'">
                    <div class="edit-title">Modo de facturación</div>
                    <div class="billing-options">
                      <button :class="['billing-btn', editBillingForm.billing_mode === 'per_module' ? 'active' : '']" @click="editBillingForm.billing_mode = 'per_module'">Por módulo</button>
                      <button :class="['billing-btn', editBillingForm.billing_mode === 'monthly' ? 'active' : '']" @click="editBillingForm.billing_mode = 'monthly'">Plan mensual</button>
                    </div>
                    <div v-if="editBillingForm.billing_mode === 'monthly'" class="form-row" style="margin-top:0.75rem">
                      <label>Cuota mensual (MXN)</label>
                      <input v-model.number="editBillingForm.monthly_fee" type="number" min="0" step="50" class="input" placeholder="499" />
                    </div>
                    <div v-else class="billing-info">El MRR se calcula sumando el costo de cada módulo activo.</div>
                  </template>
                  <template v-else-if="editMode === 'password'">
                    <div class="edit-title">Nueva contraseña</div>
                    <div class="form-row"><label>Contraseña nueva (mín. 8 caracteres)</label><input v-model="editPasswordForm.new_password" type="password" class="input" placeholder="••••••••" /></div>
                  </template>
                  <div class="edit-actions">
                    <button class="btn-primary" @click="saveEdit(t.tenant_id)" :disabled="editLoading">{{ editLoading ? 'Guardando...' : 'Guardar' }}</button>
                    <button class="btn-secondary" @click="cancelEdit">Cancelar</button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- MÉTRICAS -->
      <div v-else-if="activeSection === 'metricas'" class="admin-page">
        <div class="page-header">
          <div><h2>Métricas de negocio</h2><p class="page-sub">Snapshot precalculado.<span v-if="metrics"> Actualizado: {{ fmtDate(metrics.calculated_at) }}</span></p></div>
          <button class="btn-secondary" @click="loadMetrics" :disabled="loadingMetrics">{{ loadingMetrics ? 'Cargando...' : 'Actualizar' }}</button>
        </div>
        <div v-if="loadingMetrics" class="state-empty">Calculando métricas...</div>
        <div v-else-if="errorMetrics" class="alert-info">{{ errorMetrics }}</div>
        <div v-else-if="!metrics" class="state-empty">Sin datos disponibles.</div>
        <div v-else>
          <div class="kpi-grid">
            <div class="kpi-card"><span class="kpi-label">MRR</span><span class="kpi-value kpi-highlight">{{ fmtMXN(metrics.mrr) }}</span></div>
            <div class="kpi-card"><span class="kpi-label">Cuentas activas</span><span class="kpi-value">{{ metrics.active_accounts }}</span></div>
            <div class="kpi-card"><span class="kpi-label">Suspendidas</span><span class="kpi-value">{{ metrics.suspended_accounts }}</span></div>
            <div class="kpi-card"><span class="kpi-label">Total pacientes</span><span class="kpi-value">{{ metrics.total_patients }}</span></div>
            <div class="kpi-card"><span class="kpi-label">Notas emitidas</span><span class="kpi-value">{{ metrics.total_notes }}</span></div>
            <div class="kpi-card"><span class="kpi-label">Recetas emitidas</span><span class="kpi-value">{{ metrics.total_prescriptions }}</span></div>
          </div>
          <div v-if="Object.keys(metricsModules).length > 0" class="section-block">
            <h3 class="section-title">Módulos activos por cuenta</h3>
            <div class="module-dist-list">
              <div v-for="(count, key) in metricsModules" :key="key" class="module-dist-row">
                <span class="module-dist-key">{{ key }}</span>
                <span class="module-dist-bar-wrap"><span class="module-dist-bar" :style="{ width: Math.max(4, (count / metrics!.active_accounts) * 100) + '%' }"></span></span>
                <span class="module-dist-count">{{ count }}</span>
              </div>
            </div>
          </div>
          <div v-if="metricsAccounts.length > 0" class="section-block">
            <h3 class="section-title">Detalle por cuenta</h3>
            <!-- Grupo: Plan mensual -->
            <template v-if="metricsAccounts.filter(a => a.billing_mode === 'monthly').length > 0">
              <div class="plan-group-header">Plan mensual</div>
              <template v-for="price in [...new Set(metricsAccounts.filter(a => a.billing_mode === 'monthly').map(a => a.monthly_fee))].sort((a,b) => b - a)" :key="'m-' + price">
                <div class="plan-price-header">{{ fmtMXN(price) }}/mes</div>
                <table class="accounts-table plan-table">
                  <thead><tr><th>Cuenta</th><th>Estado</th><th>Pacientes</th><th>Último registro</th></tr></thead>
                  <tbody>
                    <tr v-for="a in metricsAccounts.filter(a => a.billing_mode === 'monthly' && a.monthly_fee === price)" :key="a.tenant_id">
                      <td class="td-email">{{ a.email }}</td>
                      <td><span :class="['badge-state', a.state === 'active' ? 'badge-active' : 'badge-suspended']">{{ a.state === 'active' ? 'Activa' : 'Suspendida' }}</span></td>
                      <td>{{ a.patients }}</td>
                      <td class="td-date">{{ fmtDate(a.last_record) }}</td>
                    </tr>
                  </tbody>
                </table>
              </template>
            </template>
            <!-- Grupo: Por modulo -->
            <template v-if="metricsAccounts.filter(a => a.billing_mode !== 'monthly').length > 0">
              <div class="plan-group-header" :style="metricsAccounts.filter(a => a.billing_mode === 'monthly').length > 0 ? 'margin-top:1.25rem' : ''">Por módulo</div>
              <table class="accounts-table plan-table">
                <thead><tr><th>Cuenta</th><th>Estado</th><th>MRR módulos</th><th>Pacientes</th><th>Último registro</th></tr></thead>
                <tbody>
                  <tr v-for="a in metricsAccounts.filter(a => a.billing_mode !== 'monthly')" :key="a.tenant_id">
                    <td class="td-email">{{ a.email }}</td>
                    <td><span :class="['badge-state', a.state === 'active' ? 'badge-active' : 'badge-suspended']">{{ a.state === 'active' ? 'Activa' : 'Suspendida' }}</span></td>
                    <td>{{ fmtMXN(a.mrr) }}</td>
                    <td>{{ a.patients }}</td>
                    <td class="td-date">{{ fmtDate(a.last_record) }}</td>
                  </tr>
                </tbody>
              </table>
            </template>
          </div>
        </div>
      </div>

      <!-- ACTIVIDAD -->
      <div v-else-if="activeSection === 'actividad'" class="admin-page">
        <div class="page-header">
          <div><h2>Actividad y uso</h2><p class="page-sub">Conteos por médico. Sin datos clínicos.</p></div>
          <button class="btn-secondary" @click="loadActivity" :disabled="loadingActivity">{{ loadingActivity ? 'Cargando...' : 'Actualizar' }}</button>
        </div>
        <div v-if="loadingActivity" class="state-empty">Cargando actividad...</div>
        <div v-else-if="errorActivity" class="alert-error">{{ errorActivity }}</div>
        <div v-else-if="activity.length === 0" class="state-empty">Sin datos de actividad aún. El sistema registra sesiones conforme los médicos usen la plataforma.</div>
        <div v-else class="activity-layout">
          <div class="activity-list">
            <div v-for="a in activity" :key="a.tenant_id" :class="['activity-row', selectedTenant === a.tenant_id ? 'selected' : '']" @click="loadActivityDetail(a.tenant_id)">
              <div class="activity-row-email">{{ a.tenant_id }}</div>
              <div class="activity-row-stats">
                <span class="stat-pill">{{ a.sessions_total }} sesiones</span>
                <span class="stat-pill">{{ a.patients_total }} pacientes</span>
                <span class="stat-pill">{{ a.prescriptions_total }} recetas</span>
              </div>
            </div>
          </div>
          <div class="activity-detail">
            <div v-if="!selectedTenant" class="state-empty">Selecciona una cuenta para ver el detalle.</div>
            <div v-else-if="loadingDetail" class="state-empty">Cargando...</div>
            <div v-else-if="activityDetail.length === 0" class="state-empty">Sin registros mensuales para esta cuenta.</div>
            <div v-else>
              <h3 class="section-title">Últimos 12 meses</h3>
              <div class="accounts-table-wrap">
                <table class="accounts-table">
                  <thead><tr><th>Periodo</th><th>Sesiones</th><th>Notas</th><th>Recetas</th><th>Alergias</th><th>Pacientes</th></tr></thead>
                  <tbody>
                    <tr v-for="p in activityDetail" :key="p.period">
                      <td>{{ p.period }}</td><td>{{ p.sessions_count }}</td><td>{{ p.notes_count }}</td>
                      <td>{{ p.prescriptions_count }}</td><td>{{ p.allergies_count }}</td><td>{{ p.patients_count }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- SALUD -->
      <div v-else-if="activeSection === 'salud'" class="admin-page">
        <div class="page-header">
          <div><h2>Salud de cuentas</h2><p class="page-sub">Adopcion y riesgo por medico. Sin datos clinicos.</p></div>
          <button class="btn-secondary" @click="loadHealth" :disabled="loadingHealth">{{ loadingHealth ? 'Cargando...' : 'Actualizar' }}</button>
        </div>
        <div v-if="loadingHealth" class="state-empty">Calculando...</div>
        <div v-else-if="errorHealth" class="alert-error">{{ errorHealth }}</div>
        <div v-else>
          <div class="health-summary">
            <button :class="['health-filter-btn', healthFilter === '' ? 'selected' : '']" @click="healthFilter = ''; loadHealth()">Todos ({{ healthSummary.active + healthSummary.at_risk + healthSummary.inactive }})</button>
            <button :class="['health-filter-btn green', healthFilter === 'active' ? 'selected' : '']" @click="healthFilter = 'active'; loadHealth()">Activos ({{ healthSummary.active }})</button>
            <button :class="['health-filter-btn yellow', healthFilter === 'at_risk' ? 'selected' : '']" @click="healthFilter = 'at_risk'; loadHealth()">En riesgo ({{ healthSummary.at_risk }})</button>
            <button :class="['health-filter-btn red', healthFilter === 'inactive' ? 'selected' : '']" @click="healthFilter = 'inactive'; loadHealth()">Inactivos ({{ healthSummary.inactive }})</button>
          </div>
          <div v-if="health.length === 0" class="state-empty">Sin datos aun. El worker calcula cada hora.</div>
          <div v-else class="accounts-table-wrap">
            <table class="accounts-table">
              <thead><tr><th>Medico</th><th>Estado</th><th>Antiguedad</th><th>Ultimo login</th><th>Sesiones/mes</th><th>Notas/mes</th><th>Recetas/mes</th><th>Pacientes</th><th>Modulos</th></tr></thead>
              <tbody>
                <tr v-for="a in health" :key="a.tenant_id" :class="['health-row', a.health_status]">
                  <td class="td-email">{{ a.email }}</td>
                  <td><span :class="['badge-state', a.health_status === 'active' ? 'badge-active' : a.health_status === 'at_risk' ? 'badge-risk' : 'badge-suspended']">{{ healthLabel(a.health_status) }}</span></td>
                  <td>{{ a.account_age_days }} dias</td>
                  <td>{{ fmtDaysAgo(a.days_since_login) }}</td>
                  <td>{{ a.sessions_this_month }}<span v-if="a.sessions_last_month > 0" class="trend"> ({{ a.sessions_last_month }} ant.)</span></td>
                  <td>{{ a.notes_this_month }}</td>
                  <td>{{ a.prescriptions_this_month }}</td>
                  <td>{{ a.total_patients }}</td>
                  <td>{{ a.modules_used }}/{{ a.modules_active }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- SISTEMA -->
      <div v-else-if="activeSection === 'sistema'" class="admin-page">
        <div class="page-header">
          <div><h2>Estado del sistema</h2><p class="page-sub">Actualizado cada hora automaticamente.</p></div>
          <button class="btn-secondary" @click="loadSystem" :disabled="loadingSystem">{{ loadingSystem ? 'Revisando...' : 'Revisar ahora' }}</button>
        </div>
        <div v-if="loadingSystem" class="state-empty">Revisando sistema...</div>
        <div v-else-if="errorSystem" class="alert-error">{{ errorSystem }}</div>
        <div v-else-if="!systemSnap" class="state-empty">Sin datos de sistema aun.</div>
        <div v-else>
          <div :class="['system-banner', systemSnap.overall_ok ? 'ok' : 'error']">
            <span>{{ systemSnap.overall_ok ? 'Todo funciona correctamente' : systemSnap.issues }}</span>
            <span class="system-banner-time">Revisado: {{ fmtDate(systemSnap.calculated_at) }}</span>
          </div>
          <div class="system-grid">
            <div :class="['system-card', systemSnap.db_ok ? 'ok' : 'error']">
              <div class="system-card-title">Base de datos</div>
              <div class="system-card-desc">{{ systemSnap.db_ok ? 'Conectada y respondiendo' : 'No responde — revisar urgente' }}</div>
            </div>
            <div :class="['system-card', systemSnap.backup_ok ? 'ok' : 'error']">
              <div class="system-card-title">Backups</div>
              <div class="system-card-desc">
                <span v-if="systemSnap.last_backup_at">Ultimo: {{ fmtDate(systemSnap.last_backup_at) }} ({{ systemSnap.last_backup_size_kb }} KB)</span>
                <span v-else>Sin backups registrados</span>
              </div>
            </div>
            <div :class="['system-card', systemSnap.metrics_ok ? 'ok' : 'error']">
              <div class="system-card-title">Worker de metricas</div>
              <div class="system-card-desc">
                <span v-if="systemSnap.metrics_last_run_at">Ultimo calculo: {{ fmtDate(systemSnap.metrics_last_run_at) }}</span>
                <span v-else>Sin calculos registrados</span>
              </div>
            </div>
            <div :class="['system-card', systemSnap.disk_ok ? 'ok' : 'error']">
              <div class="system-card-title">Espacio en disco</div>
              <div class="system-card-desc">{{ systemSnap.disk_used_pct }}% utilizado{{ !systemSnap.disk_ok ? ' — revisar pronto' : '' }}</div>
            </div>
          </div>
          <div style="margin-top:1.5rem">
            <h3 style="font-size:14px;font-weight:700;margin-bottom:0.75rem;color:var(--color-text-secondary)">Ultimos accesos fallidos</h3>
            <div v-if="failedLogins.length === 0" class="state-empty" style="padding:0.5rem 0">Sin intentos fallidos recientes.</div>
            <div v-else class="accounts-table-wrap">
              <table class="accounts-table">
                <thead><tr><th>Correo</th><th>Fecha y hora</th></tr></thead>
                <tbody>
                  <tr v-for="f in failedLogins" :key="f.occurred_at">
                    <td class="td-email">{{ f.email }}</td>
                    <td>{{ fmtDate(f.occurred_at) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

    </main>
  </div>
</template>

<style scoped>
.admin-shell { display: flex; min-height: 100vh; background: var(--app-bg); }
.admin-sidebar { width: 220px; min-height: 100vh; background: var(--app-sidebar-bg); display: flex; flex-direction: column; padding: var(--space-4); position: fixed; top: 0; left: 0; }
.admin-brand { display: flex; align-items: center; gap: var(--space-2); padding: var(--space-4) 0 var(--space-6) 0; }
.brand-icon { width: 32px; height: 32px; background: var(--color-jade); color: var(--color-obsidian); border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 16px; }
.brand-name { font-family: var(--font-brand); font-weight: 700; font-size: 18px; color: var(--text-on-dark); }
.admin-badge { font-size: 10px; font-weight: 700; background: var(--color-warning, #FFB020); color: #000; border-radius: 4px; padding: 2px 6px; text-transform: uppercase; }
.admin-nav { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.nav-item { display: flex; align-items: center; gap: var(--space-2); background: transparent; border: none; color: rgba(240,246,252,0.55); font-family: var(--font-body); font-size: 14px; font-weight: 500; padding: var(--space-2) var(--space-3); border-radius: var(--radius-sm); cursor: pointer; text-align: left; transition: all 0.15s; width: 100%; }
.nav-item:hover { background: rgba(255,255,255,0.06); color: var(--text-on-dark); }
.nav-item.active { background: rgba(0,223,162,0.12); color: var(--color-jade); font-weight: 700; }
.admin-footer { margin-top: auto; display: flex; flex-direction: column; gap: var(--space-2); border-top: 1px solid rgba(255,255,255,0.08); padding-top: var(--space-4); }
.admin-user { font-size: 12px; color: var(--text-on-dark); opacity: 0.5; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.btn-logout { background: transparent; border: 1px solid rgba(255,255,255,0.15); color: var(--text-on-dark); padding: var(--space-2) var(--space-4); border-radius: var(--radius-sm); font-size: 13px; cursor: pointer; text-align: left; }
.btn-logout:hover { border-color: var(--color-error); color: var(--color-error); }
.admin-main { flex: 1; margin-left: 220px; padding: var(--space-8); }
.admin-page { max-width: 900px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: 2px; }
.kpi-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: var(--space-4); margin-bottom: var(--space-6); }
.kpi-card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-lg); padding: var(--space-4) var(--space-5); display: flex; flex-direction: column; gap: 4px; }
.kpi-label { font-size: 12px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.04em; }
.kpi-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.kpi-highlight { color: var(--color-jade); }
.section-block { margin-bottom: var(--space-6); }
.section-title { font-size: 14px; font-weight: 700; color: var(--text-primary); margin: 0 0 var(--space-3) 0; }
.module-dist-list { display: flex; flex-direction: column; gap: var(--space-2); }
.module-dist-row { display: flex; align-items: center; gap: var(--space-3); }
.module-dist-key { font-size: 13px; color: var(--text-primary); width: 140px; flex-shrink: 0; }
.module-dist-bar-wrap { flex: 1; height: 8px; background: #F1F5F9; border-radius: 4px; overflow: hidden; }
.module-dist-bar { height: 100%; background: var(--color-jade); border-radius: 4px; }
.module-dist-count { font-size: 13px; font-weight: 600; color: var(--text-secondary); width: 32px; text-align: right; }
.accounts-table-wrap { overflow-x: auto; }
.accounts-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.accounts-table th { text-align: left; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; color: var(--text-secondary); padding: var(--space-2) var(--space-3); border-bottom: 1px solid #E2E8F0; }
.accounts-table td { padding: var(--space-3); border-bottom: 1px solid #F1F5F9; color: var(--text-primary); }
.accounts-table tr:last-child td { border-bottom: none; }
.td-email { font-size: 13px; }
.td-date { color: var(--text-secondary); }
.badge-state { font-size: 11px; font-weight: 700; border-radius: 4px; padding: 2px 7px; }
.badge-active { background: #DCFCE7; color: #166534; }
.badge-suspended { background: #FEE2E2; color: #991B1B; }
.activity-layout { display: grid; grid-template-columns: 300px 1fr; gap: var(--space-6); }
.activity-list { display: flex; flex-direction: column; gap: var(--space-2); }
.activity-row { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-md); padding: var(--space-3) var(--space-4); cursor: pointer; transition: border-color 0.15s; }
.activity-row:hover { border-color: var(--color-turquoise); }
.activity-row.selected { border-color: var(--color-jade); background: #F0FDF4; }
.activity-row-email { font-size: 13px; font-weight: 600; color: var(--text-primary); margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.activity-row-stats { display: flex; gap: 6px; flex-wrap: wrap; }
.stat-pill { font-size: 11px; background: #F1F5F9; color: var(--text-secondary); border-radius: 10px; padding: 2px 8px; }
.activity-detail { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-lg); padding: var(--space-5); min-height: 200px; }
.alert-info { background: #EFF6FF; border: 1px solid #BFDBFE; border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: #1D4ED8; margin-bottom: var(--space-4); }
.plan-group-header { font-size: 12px; font-weight: 700; color: var(--color-text-secondary); text-transform: uppercase; letter-spacing: 0.05em; padding: 0.75rem 0 0.25rem; border-top: 1px solid var(--color-border); margin-top: 0.5rem; }
.plan-price-header { font-size: 17px; font-weight: 700; color: var(--color-jade, #00DFA2); padding: 0.4rem 0 0.2rem; }
.plan-table { margin-bottom: 0.75rem; }
.health-summary { display: flex; gap: 0.5rem; margin-bottom: 1rem; flex-wrap: wrap; }
.health-filter-btn { padding: 0.35rem 0.9rem; border-radius: 6px; border: 1px solid var(--color-border); background: transparent; color: var(--color-text-secondary); cursor: pointer; font-size: 12px; }
.health-filter-btn.selected { background: var(--color-jade, #00DFA2); color: #090C10; border-color: var(--color-jade); font-weight: 700; }
.health-filter-btn.green.selected { background: #00DFA2; border-color: #00DFA2; }
.health-filter-btn.yellow.selected { background: #F59E0B; border-color: #F59E0B; color: #fff; }
.health-filter-btn.red.selected { background: #EF4444; border-color: #EF4444; color: #fff; }
.health-row.at_risk td { background: #FFFBEB; }
.health-row.inactive td { background: #FEF2F2; }
.badge-risk { background: #FEF3C7; color: #92400E; border-radius: 4px; padding: 2px 8px; font-size: 11px; font-weight: 700; }
.trend { color: var(--color-text-secondary); font-size: 11px; }
.system-banner { display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem 1rem; border-radius: 8px; margin-bottom: 1.25rem; font-size: 14px; }
.system-banner.ok { background: #ECFDF5; color: #065F46; border: 1px solid #00DFA2; }
.system-banner.error { background: #FEF2F2; color: #991B1B; border: 1px solid #FCA5A5; }
.system-banner-time { margin-left: auto; font-size: 12px; opacity: 0.7; }
.system-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 0.75rem; }
.system-card { padding: 1rem; border-radius: 8px; border: 1px solid var(--color-border); }
.system-card.ok { border-color: #00DFA2; background: #F0FDF4; }
.system-card.error { border-color: #FCA5A5; background: #FEF2F2; }
.system-card-title { font-weight: 700; font-size: 13px; color: #111827; margin-bottom: 0.25rem; }
.system-card-desc { font-size: 12px; color: #6B7280; }
.alert-success { background: #F0FDF4; border: 1px solid #86EFAC; border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: #166534; margin-bottom: var(--space-4); }
.alert-error { background: #FFF0F3; border: 1px solid var(--color-error); border-radius: var(--radius-sm); padding: var(--space-3); font-size: 14px; color: var(--color-error); margin-bottom: var(--space-4); }
.state-empty { color: var(--text-secondary); padding: var(--space-8); text-align: center; font-size: 14px; }
.btn-primary { font-family: var(--font-brand); background: var(--action-primary-bg); color: var(--action-primary-text); border: none; padding: var(--space-2) var(--space-5); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { font-family: var(--font-brand); background: transparent; color: var(--text-primary); border: 1.5px solid #E2E8F0; padding: var(--space-2) var(--space-5); border-radius: var(--radius-md); font-size: 14px; font-weight: 600; cursor: pointer; }
.btn-secondary:hover { border-color: var(--color-turquoise); }
.btn-secondary:disabled { opacity: 0.5; cursor: not-allowed; }
.search-input { width: 100%; font-family: var(--font-body); padding: var(--space-3) var(--space-4); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 15px; color: var(--text-primary); background: var(--app-surface); outline: none; margin-bottom: var(--space-4); box-sizing: border-box; }
.search-input:focus { border-color: var(--color-turquoise); }
.create-form-card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-lg); padding: var(--space-6); margin-bottom: var(--space-6); }
.form-title { font-size: 15px; font-weight: 700; color: var(--text-primary); margin: 0 0 var(--space-1) 0; }
.form-subtitle { font-size: 13px; color: var(--text-secondary); margin: 0 0 var(--space-4) 0; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-4); margin-bottom: var(--space-4); }
.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-group--full { grid-column: 1 / -1; }
.form-group label { font-size: 12px; font-weight: 600; color: var(--text-secondary); }
.optional { font-weight: 400; }
.input-field { font-family: var(--font-body); padding: var(--space-2) var(--space-3); border: 1.5px solid #E2E8F0; border-radius: var(--radius-md); font-size: 14px; color: var(--text-primary); background: var(--app-bg); outline: none; }
.input-field:focus { border-color: var(--color-turquoise); }
.form-modules-note { display: flex; align-items: center; gap: var(--space-2); font-size: 12px; color: var(--text-secondary); background: #F8FAFC; border: 1px solid #E2E8F0; border-radius: var(--radius-sm); padding: var(--space-3) var(--space-4); margin-bottom: var(--space-4); }
.form-actions { display: flex; gap: var(--space-3); justify-content: flex-end; }
.tenant-list { display: flex; flex-direction: column; gap: var(--space-4); }
.tenant-card { background: var(--app-surface); border: 1px solid #E2E8F0; border-radius: var(--radius-lg); overflow: hidden; }
.tenant-header { display: flex; justify-content: space-between; align-items: center; padding: var(--space-4) var(--space-6); background: #FAFBFC; border-bottom: 1px solid #E2E8F0; }
.tenant-header-left { display: flex; align-items: center; gap: var(--space-2); }
.tenant-expand { font-size: 12px; color: var(--text-secondary); width: 14px; }
.tenant-email { font-weight: 600; font-size: 14px; }
.tenant-id { font-size: 12px; color: var(--text-secondary); font-family: monospace; }
.module-list { padding: var(--space-2) 0; }
.module-row { display: flex; align-items: center; justify-content: space-between; padding: var(--space-3) var(--space-6); border-bottom: 1px solid #F1F5F9; }
.module-row:last-child { border-bottom: none; }
.module-name { font-size: 14px; color: var(--text-primary); }
.toggle-btn { font-size: 12px; font-weight: 600; padding: 4px 12px; border-radius: var(--radius-sm); border: 1.5px solid #E2E8F0; background: var(--app-surface); color: var(--text-secondary); cursor: pointer; transition: all 0.15s; }
.toggle-btn.active { background: #DCFCE7; border-color: #86EFAC; color: #166534; }
.toggle-btn:not(.active):hover { border-color: var(--color-error); color: var(--color-error); }
.badge-admin { font-size: 11px; font-weight: 700; background: #FEF9C3; color: #854D0E; border-radius: 4px; padding: 1px 6px; }
.tenant-actions { display: flex; gap: 0.5rem; padding: 0.5rem 0 0.25rem; border-top: 1px solid var(--color-border); margin-top: 0.5rem; }
.edit-panel { background: #F8FAFB; border: 1px solid #D1D5DB; border-radius: 8px; padding: 1rem; margin-top: 0.75rem; }
.edit-title { font-size: 13px; font-weight: 700; color: #374151; margin-bottom: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; }
.edit-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem 1rem; }
.edit-grid .form-row label { color: #374151; font-size: 12px; }
.edit-grid .form-row .input { background: #fff; color: #111827; border: 1px solid #D1D5DB; border-radius: 6px; padding: 0.4rem 0.6rem; font-size: 13px; width: 100%; }
.edit-actions { display: flex; gap: 0.5rem; margin-top: 1rem; }
.billing-options { display: flex; gap: 0.5rem; }
.billing-btn { padding: 0.4rem 1rem; border-radius: 6px; border: 1px solid #D1D5DB; background: #fff; color: #374151; cursor: pointer; font-size: 13px; }
.billing-btn.active { background: #00DFA2; color: #090C10; border-color: #00DFA2; font-weight: 700; }
.billing-info { font-size: 12px; color: #6B7280; margin-top: 0.5rem; }
.alert-success { background: #ECFDF5; border: 1px solid #00DFA2; color: #065F46; border-radius: 6px; padding: 0.5rem 0.75rem; font-size: 13px; margin-bottom: 0.5rem; }
.edit-panel .alert-error { background: #FEF2F2; border: 1px solid #FCA5A5; color: #991B1B; border-radius: 6px; padding: 0.5rem 0.75rem; font-size: 13px; margin-bottom: 0.5rem; }
.edit-panel .form-row { display: flex; flex-direction: column; gap: 0.25rem; }
.edit-panel .btn-primary { background: #00DFA2; color: #090C10; border: none; border-radius: 6px; padding: 0.5rem 1.2rem; font-size: 13px; font-weight: 700; cursor: pointer; }
.edit-panel .btn-secondary { background: #fff; color: #374151; border: 1px solid #D1D5DB; border-radius: 6px; padding: 0.5rem 1.2rem; font-size: 13px; cursor: pointer; }
.edit-panel input[type="password"], .edit-panel input[type="number"] { background: #fff; color: #111827; border: 1px solid #D1D5DB; border-radius: 6px; padding: 0.4rem 0.6rem; font-size: 13px; width: 100%; }
</style>
