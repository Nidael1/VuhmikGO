import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AuthTokens, UserProfile } from '@/domain/types/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const profile = ref<UserProfile | null>(null)

  const isAuthenticated = computed(() => token.value !== null)
  const isAdmin = computed(() => profile.value?.is_admin ?? false)

  function setSession(tokens: AuthTokens) {
    token.value = tokens.token
    refreshToken.value = tokens.refresh_token ?? null
    profile.value = {
      actor_id: tokens.actor_id,
      tenant_id: tokens.tenant_id,
      is_admin: tokens.is_admin ?? false,
    }
  }

  function clearSession() {
    token.value = null
    refreshToken.value = null
    profile.value = null
  }

  return { token, refreshToken, profile, isAuthenticated, isAdmin, setSession, clearSession }
})
