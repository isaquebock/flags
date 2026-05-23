<script setup>
  import { ref } from 'vue'
  import { useRouter } from 'vue-router'
  import ContentBlock from '@/templates/content-block'
  import PageHeadingBlock from '@/templates/page-heading-block'
  import InputText from '@aziontech/webkit/inputtext'
  import PrimeButton from '@aziontech/webkit/button'
  import { useFeatureFlagsAccountStore } from '@/stores/feature-flags-account'

  defineOptions({ name: 'feature-flags-account-view' })

  const store = useFeatureFlagsAccountStore()
  const router = useRouter()
  const inputValue = ref(store.clientId)
  const error = ref('')

  const handleSubmit = () => {
    const trimmed = inputValue.value.trim()
    if (!trimmed) {
      error.value = 'Client ID is required.'
      return
    }
    if (trimmed.length > 128) {
      error.value = 'Client ID must be 128 characters or fewer.'
      return
    }
    error.value = ''
    store.setClientId(trimmed)
    router.push({ name: 'list-feature-flags' })
  }
</script>

<template>
  <ContentBlock>
    <template #heading>
      <PageHeadingBlock
        pageTitle="Feature Flags"
        description="Enter your Client ID to manage the feature flags for that account."
      />
    </template>
    <template #content>
      <div class="flex flex-col gap-6 max-w-sm">
        <div class="flex flex-col gap-2">
          <label
            for="client-id-input"
            class="text-sm font-medium"
          >
            Client ID
          </label>
          <InputText
            id="client-id-input"
            v-model="inputValue"
            placeholder="e.g. my-app-prod"
            :class="{ 'p-invalid': error }"
            @keydown.enter="handleSubmit"
          />
          <small
            v-if="error"
            class="p-error"
          >{{ error }}</small>
          <small
            v-else
            class="text-color-secondary"
          >
            The identifier used to scope your feature flags.
          </small>
        </div>
        <PrimeButton
          label="Manage Flags"
          icon="pi pi-arrow-right"
          iconPos="right"
          @click="handleSubmit"
        />
      </div>
    </template>
  </ContentBlock>
</template>
