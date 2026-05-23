import { defineStore } from 'pinia'

export const useFeatureFlagsAccountStore = defineStore('feature-flags-account', {
  persist: true,
  state: () => ({ clientId: '' }),
  actions: {
    setClientId(id) {
      this.clientId = id.trim()
    },
    clear() {
      this.clientId = ''
    }
  }
})
