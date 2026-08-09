<script setup lang="ts">
import { TERMS_HTML, TERMS_VERSION } from '@/domain/legal/terms-mx-medicine-v1'

defineProps<{ open: boolean; required?: boolean }>()

// accept y close son eventos distintos por diseno. Un cierre accidental
// nunca debe registrarse como aceptacion: el consentimiento tiene valor
// probatorio y requiere accion afirmativa explicita del Medico.
const emit = defineEmits<{ close: []; accept: [version: string] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="terms-overlay" @click.self="!required && emit('close')">
      <div class="terms-dialog" role="dialog" aria-modal="true" aria-labelledby="terms-title">
        <div class="terms-header">
          <h2 id="terms-title">Términos y Condiciones de Uso</h2>
          <button v-if="!required" class="btn-close" @click="emit('close')" aria-label="Cerrar">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <div class="terms-body">
          <div v-html="TERMS_HTML"></div>
        </div>

        <div class="terms-footer">
          <button
            class="btn-accept"
            @click="required ? emit('accept', TERMS_VERSION) : emit('close')"
          >
            {{ required ? 'Acepto los Términos y Condiciones' : 'Entendido' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.terms-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
}

.terms-dialog {
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 680px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.terms-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-6) var(--space-6) var(--space-4);
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
  flex-shrink: 0;
}

.terms-header h2 {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-brand);
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  transition: color 0.15s;
}

.btn-close:hover { color: var(--text-primary); }

.terms-body {
  overflow-y: auto;
  padding: var(--space-6);
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.terms-body :deep(.terms-updated) {
  font-size: 12px;
  color: var(--text-secondary);
}

.terms-body :deep(section) {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.terms-body :deep(section h3) {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.terms-body :deep(section p),
.terms-body :deep(section ul) {
  font-size: 14px;
  line-height: 1.65;
  color: var(--text-secondary);
}

.terms-body :deep(section ul) {
  padding-left: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.terms-body :deep(section strong) { color: var(--text-primary); font-weight: 600; }

.terms-body :deep(.terms-contact) {
  font-size: 13px;
  color: var(--text-secondary);
  padding-top: var(--space-2);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.terms-footer {
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid rgba(0, 0, 0, 0.08);
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}

.btn-accept {
  font-family: var(--font-brand);
  background: var(--action-primary-bg);
  color: var(--action-primary-text);
  border: none;
  padding: var(--space-3) var(--space-6);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn-accept:hover { opacity: 0.9; }
</style>
