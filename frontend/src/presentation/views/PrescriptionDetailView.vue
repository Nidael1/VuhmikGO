<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/presentation/layouts/AppLayout.vue'
import { prescriptionRepository } from '@/infrastructure/repositories/prescriptionRepository'
import { patientRepository } from '@/infrastructure/repositories/patientRepository'
import { useAuthStore } from '@/app/stores/auth'
import type { Prescription } from '@/domain/types/prescription'
import type { Patient } from '@/domain/types/patient'

const route = useRoute()
const id = route.params.id as string
const auth = useAuthStore()

const prescription = ref<Prescription | null>(null)
const patient = ref<Patient | null>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const rx = await prescriptionRepository.get(id)
    prescription.value = rx
    patient.value = await patientRepository.get(rx.patient_id)
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('es-MX', {
    year: 'numeric', month: 'long', day: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}

function reimprimir() {
  if (!prescription.value || !auth.token) return
  window.open(`/api/v1/prescriptions/${prescription.value.id}/print?token=${auth.token}`, '_blank')
}
</script>

<template>
  <AppLayout>
    <div class="page">
      <div v-if="loading" class="state-empty">Cargando...</div>
      <div v-else-if="error" class="alert-error">{{ error }}</div>

      <template v-else-if="prescription">
        <div class="page-header">
          <div>
            <h2>Receta electrónica</h2>
            <p class="page-sub">{{ prescription.medicamento_generico }}</p>
          </div>
          <RouterLink to="/prescriptions" class="btn-back">← Recetas</RouterLink>
        </div>

        <div class="seccion seccion--recetas">
          <div class="seccion-header">
            <div class="seccion-titulo">
              <svg class="seccion-icono" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M8 2v4"/><path d="M16 2v4"/>
                <rect x="3" y="6" width="18" height="16" rx="2"/>
                <line x1="9" y1="13" x2="15" y2="13"/>
                <line x1="9" y1="17" x2="15" y2="17"/>
              </svg>
              <h3>Detalle de receta</h3>
            </div>
            <div class="seccion-acciones">
              <span class="rx-estado">{{ prescription.state === 'issued' ? 'emitida' : prescription.state }}</span>
              <button class="btn-reimprimir" @click="reimprimir">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="6 9 6 2 18 2 18 9"/>
                  <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/>
                  <rect x="6" y="14" width="12" height="8"/>
                </svg>
                Reimprimir
              </button>
            </div>
          </div>

          <div class="detalle-grid">
            <div class="detalle-item">
              <span class="detalle-label">Paciente</span>
              <RouterLink v-if="patient" :to="`/patients/${patient.id}`" class="detalle-link">
                {{ patient.nombre }}
              </RouterLink>
            </div>
            <div class="detalle-item">
              <span class="detalle-label">Medicamento genérico</span>
              <span class="detalle-valor">{{ prescription.medicamento_generico }}</span>
            </div>
            <div class="detalle-item">
              <span class="detalle-label">Dosis</span>
              <span class="detalle-valor">{{ prescription.dosis }}</span>
            </div>
            <div v-if="prescription.diagnostico" class="detalle-item">
              <span class="detalle-label">Diagnóstico</span>
              <span class="detalle-valor">{{ prescription.diagnostico }}</span>
            </div>
            <div v-if="prescription.indicaciones" class="detalle-item">
              <span class="detalle-label">Indicaciones</span>
              <span class="detalle-valor">{{ prescription.indicaciones }}</span>
            </div>
            <div v-if="prescription.seguimiento" class="detalle-item">
              <span class="detalle-label">Seguimiento</span>
              <span class="detalle-valor">{{ prescription.seguimiento }}</span>
            </div>
            <div v-if="prescription.issued_at" class="detalle-item">
              <span class="detalle-label">Fecha de emisión</span>
              <span class="detalle-valor">{{ formatDate(prescription.issued_at) }}</span>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<style scoped>
.page { max-width: 780px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: var(--space-6); }
.page-sub { color: var(--text-secondary); font-size: 13px; margin-top: 2px; }
.btn-back { color: var(--color-clinical-blue); font-size: 14px; text-decoration: none; white-space: nowrap; }

.seccion {
  background: var(--app-surface);
  border: 1px solid #E2E8F0;
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.seccion-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid #E2E8F0;
}

.seccion-titulo { display: flex; align-items: center; gap: var(--space-2); }
.seccion-titulo h3 { margin: 0; font-size: 14px; font-weight: 700; color: var(--text-primary); }
.seccion-icono { display: flex; align-items: center; color: var(--text-secondary); flex-shrink: 0; }

.seccion-acciones {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.seccion--recetas .seccion-header {
  background: #F5F8FF;
  border-left: 3px solid var(--color-clinical-blue, #3B82F6);
}

.rx-estado {
  font-size: 11px;
  font-weight: 600;
  background: #DCFCE7;
  color: #166534;
  border-radius: 999px;
  padding: 2px 10px;
}

.btn-reimprimir {
  display: flex;
  align-items: center;
  gap: 5px;
  font-family: var(--font-body);
  font-size: 13px;
  font-weight: 600;
  color: var(--color-clinical-blue, #3B82F6);
  background: transparent;
  border: 1.5px solid var(--color-clinical-blue, #3B82F6);
  border-radius: var(--radius-md);
  padding: 4px 12px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.btn-reimprimir:hover {
  background: var(--color-clinical-blue, #3B82F6);
  color: #fff;
}
.btn-reimprimir svg { flex-shrink: 0; }

.detalle-grid {
  display: flex;
  flex-direction: column;
}

.detalle-item {
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid #F1F5F9;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.detalle-item:last-child { border-bottom: none; }

.detalle-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.detalle-valor {
  font-size: 15px;
  color: var(--text-primary);
}

.detalle-link {
  font-size: 15px;
  color: var(--color-clinical-blue);
  text-decoration: none;
  font-weight: 600;
  width: fit-content;
}
.detalle-link:hover { text-decoration: underline; }

.state-empty { color: var(--text-secondary); text-align: center; padding: var(--space-8); }
.alert-error {
  background: #FFF0F3; border: 1px solid var(--color-error);
  border-radius: var(--radius-sm); padding: var(--space-3);
  font-size: 14px; color: var(--color-error);
}
</style>
