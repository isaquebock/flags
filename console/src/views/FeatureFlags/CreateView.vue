<script setup>
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import ContentBlock from '@/templates/content-block'
  import PageHeadingBlock from '@/templates/page-heading-block'
  import CreateFormBlock from '@/templates/create-form-block'
  import ActionBarTemplate from '@/templates/action-bar-block/action-bar-with-teleport'
  import FormFieldsCreate from './FormFields/FormFieldsCreate.vue'
  import * as yup from 'yup'
  import { useToast } from '@aziontech/webkit/use-toast'
  import { featureFlagsService } from '@/services/v2/feature-flags'
  import { useFeatureFlagsAccountStore } from '@/stores/feature-flags-account'

  defineOptions({ name: 'feature-flags-create-view' })

  const store = useFeatureFlagsAccountStore()
  const router = useRouter()
  const toast = useToast()

  if (!store.clientId) {
    router.replace({ name: 'feature-flags-account' })
  }

  const clientId = computed(() => store.clientId)

  const validationSchema = yup.object({
    key: yup
      .string()
      .matches(
        /^[a-z0-9][a-z0-9-_]{0,63}$/,
        'Must start with a lowercase letter or digit and contain only lowercase letters, digits, hyphens or underscores (max 64 chars)'
      )
      .required('Key is required'),
    description: yup.string().max(200, 'Max 200 characters').default(''),
    enabled: yup.boolean().required().default(false)
  })

  const createService = (payload) => featureFlagsService.create(clientId.value, payload)

  const handleResponse = (response) => {
    response.showToastWithActions({
      feedback: `Flag "${response.key}" created successfully`,
      actions: {
        link: {
          label: 'Back to list',
          callback: () => response.redirectToUrl('/feature-flags/list')
        }
      }
    })
  }

  const handleError = (error) => {
    const message = error?.message ?? 'Failed to create flag'
    toast.add({ severity: 'error', summary: 'error', detail: message, closable: true })
  }
</script>

<template>
  <ContentBlock>
    <template #heading>
      <PageHeadingBlock
        pageTitle="Create Flag"
        :description="`Creating flag for client: ${clientId}`"
      />
    </template>
    <template #content>
      <CreateFormBlock
        :createService="createService"
        :schema="validationSchema"
        disableToast
        @on-response="handleResponse"
        @on-response-fail="handleError"
      >
        <template #form>
          <FormFieldsCreate />
        </template>
        <template #action-bar="{ onSubmit, onCancel, loading }">
          <ActionBarTemplate
            @onSubmit="onSubmit"
            @onCancel="onCancel"
            :loading="loading"
          />
        </template>
      </CreateFormBlock>
    </template>
  </ContentBlock>
</template>
