import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AuthTokens, UserProfile } from '@/domain/types/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const profile = ref<UserProfile | null>(null)
  const termsAccepted = ref<boolean>(false)

  const isAuthenticated = computed(() => token.value !== null)
  const isAdmin = computed(() => profile.value?.is_admin ?? false)

  function setSession(tokens: AuthTokens) {
    token.value = tokens.token
    refreshToken.value = tokens.refresh_token ?? null
    termsAccepted.value = tokens.terms_accepted ?? false
    profile.value = {
      actor_id: tokens.actor_id,
      tenant_id: tokens.tenant_id,
      is_admin: tokens.is_admin ?? false,
    }
  }

  function markTermsAccepted() {
    termsAccepted.value = true
  }

  function clearSession() {
    token.value = null
    refreshToken.value = null
    profile.value = null
    termsAccepted.value = false
  }

  return { token, refreshToken, profile, termsAccepted, isAuthenticated, isAdmin, setSession, markTermsAccepted, clearSession }
})
