<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/app/stores/auth'
import { useRouter } from 'vue-router'
import { http } from '@/infrastructure/api/httpClient'

const auth = useAuthStore()
const router = useRouter()

type Section = 'operaciones' | 'metricas' | 'actividad'
const activeSection = ref<Section>('operaciones')

interface ModuleStatus { ModuleID: string; Descripcion: string; Active: boolean; Plan: string; Costo: number }
interface TenantInfo { tenant_id: string; email: string; is_admin: boolean; is_suspended: boolean; modules: ModuleStatus[] }
const tenants = ref<TenantInfo[]>([])
const loadingTenants = ref(true)
const errorTenants = ref('')
const search = ref('')
const expandedTenants = ref<Set<string>>(new Set())
const showCreateForm = ref(false)
const createLoading = ref(false)
const createError = ref('')
const createSuccess = ref('')
const createForm = ref({ email: '', password: '', nombre_completo: '', cedula_profesional: '', especialidad: '', universidad: '', direccion: '', telefono: '', curp: '' })
const filtered = computed(() => { const q = search.value.toLowerCase().trim(); if (!q) return tenants.value; return tenants.value.filter(t => t.email.toLowerCase().includes(q) || t.tenant_id.toLowerCase().includes(q)) })
function toggleExpand(id: string) { expandedTenants.value.has(id) ? expandedTenants.value.delete(id) : expandedTenants.value.add(id) }
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
async function loadMetrics() { loadingMetrics.value = true; errorMetrics.value = ''; try { const [snapRes, accRes, modRes] = await Promise.all([http.get<any>('/admin/metrics'), http.get<any>('/admin/metrics/accounts'), http.get<any>('/admin/metrics/modules')]); metrics.value = snapRes.data ?? null; const rawAcc = accRes.data?.accounts; metricsAccounts.value = typeof rawAcc === 'string' ? JSON.parse(rawAcc) : (rawAcc ?? []); const rawMod = modRes.data?.modules; metricsModules.value = typeof rawMod === 'string' ? JSON.parse(rawMod) : (rawMod ?? {}) } catch (e: any) { errorMetrics.value = e.message?.includes('NO_SNAPSHOT') ? 'El worker aún no ha calculado métricas. Estará disponible en las próximas horas.' : e.message } finally { loadingMetrics.value = false } }
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
async function switchSection(s: Section) { activeSection.value = s; if (s === 'metricas' && !metrics.value && !loadingMetrics.value) await loadMetrics(); if (s === 'actividad' && activity.value.length === 0 && !loadingActivity.value) await loadActivity() }
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
            <div class="accounts-table-wrap">
              <table class="accounts-table">
                <thead><tr><th>Cuenta</th><th>Estado</th><th>MRR</th><th>Pacientes</th><th>Último registro</th></tr></thead>
                <tbody>
                  <tr v-for="a in metricsAccounts" :key="a.tenant_id">
                    <td class="td-email">{{ a.email }}</td>
                    <td><span :class="['badge-state', a.state === 'active' ? 'badge-active' : 'badge-suspended']">{{ a.state === 'active' ? 'Activa' : 'Suspendida' }}</span></td>
                    <td>{{ fmtMXN(a.mrr) }}</td>
                    <td>{{ a.patients }}</td>
                    <td class="td-date">{{ fmtDate(a.last_record) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
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
</style>
VUEEOFcat > /Volumes/D/vuhmikGO/frontend/src/presentation/views/AdminView.vue << 'VUEEOF'
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/app/stores/auth'
import { useRouter } from 'vue-router'
import { http } from '@/infrastructure/api/httpClient'

const auth = useAuthStore()
const router = useRouter()

type Section = 'operaciones' | 'metricas' | 'actividad'
const activeSection = ref<Section>('operaciones')

interface ModuleStatus { ModuleID: string; Descripcion: string; Active: boolean; Plan: string; Costo: number }
interface TenantInfo { tenant_id: string; email: string; is_admin: boolean; is_suspended: boolean; modules: ModuleStatus[] }
const tenants = ref<TenantInfo[]>([])
const loadingTenants = ref(true)
const errorTenants = ref('')
const search = ref('')
const expandedTenants = ref<Set<string>>(new Set())
const showCreateForm = ref(false)
const createLoading = ref(false)
const createError = ref('')
const createSuccess = ref('')
const createForm = ref({ email: '', password: '', nombre_completo: '', cedula_profesional: '', especialidad: '', universidad: '', direccion: '', telefono: '', curp: '' })
const filtered = computed(() => { const q = search.value.toLowerCase().trim(); if (!q) return tenants.value; return tenants.value.filter(t => t.email.toLowerCase().includes(q) || t.tenant_id.toLowerCase().includes(q)) })
function toggleExpand(id: string) { expandedTenants.value.has(id) ? expandedTenants.value.delete(id) : expandedTenants.value.add(id) }
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
async function loadMetrics() { loadingMetrics.value = true; errorMetrics.value = ''; try { const [snapRes, accRes, modRes] = await Promise.all([http.get<any>('/admin/metrics'), http.get<any>('/admin/metrics/accounts'), http.get<any>('/admin/metrics/modules')]); metrics.value = snapRes.data ?? null; const rawAcc = accRes.data?.accounts; metricsAccounts.value = typeof rawAcc === 'string' ? JSON.parse(rawAcc) : (rawAcc ?? []); const rawMod = modRes.data?.modules; metricsModules.value = typeof rawMod === 'string' ? JSON.parse(rawMod) : (rawMod ?? {}) } catch (e: any) { errorMetrics.value = e.message?.includes('NO_SNAPSHOT') ? 'El worker aún no ha calculado métricas. Estará disponible en las próximas horas.' : e.message } finally { loadingMetrics.value = false } }
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
async function switchSection(s: Section) { activeSection.value = s; if (s === 'metricas' && !metrics.value && !loadingMetrics.value) await loadMetrics(); if (s === 'actividad' && activity.value.length === 0 && !loadingActivity.value) await loadActivity() }
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
            <div class="accounts-table-wrap">
              <table class="accounts-table">
                <thead><tr><th>Cuenta</th><th>Estado</th><th>MRR</th><th>Pacientes</th><th>Último registro</th></tr></thead>
                <tbody>
                  <tr v-for="a in metricsAccounts" :key="a.tenant_id">
                    <td class="td-email">{{ a.email }}</td>
                    <td><span :class="['badge-state', a.state === 'active' ? 'badge-active' : 'badge-suspended']">{{ a.state === 'active' ? 'Activa' : 'Suspendida' }}</span></td>
                    <td>{{ fmtMXN(a.mrr) }}</td>
                    <td>{{ a.patients }}</td>
                    <td class="td-date">{{ fmtDate(a.last_record) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
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
</style>
