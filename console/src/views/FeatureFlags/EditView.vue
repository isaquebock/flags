<script setup>
  import { computed, ref } from 'vue'
  import { useRouter, useRoute } from 'vue-router'
  import ContentBlock from '@/templates/content-block'
  import PageHeadingBlock from '@/templates/page-heading-block'
  import EditFormBlock from '@/templates/edit-form-block'
  import ActionBarTemplate from '@/templates/action-bar-block/action-bar-with-teleport'
  import FormFieldsEdit from './FormFields/FormFieldsEdit.vue'
  import * as yup from 'yup'
  import { featureFlagsService } from '@/services/v2/feature-flags'
  import { useFeatureFlagsAccountStore } from '@/stores/feature-flags-account'

  defineOptions({ name: 'feature-flags-edit-view' })

  const store = useFeatureFlagsAccountStore()
  const router = useRouter()
  const route = useRoute()

  if (!store.clientId) {
    router.replace({ name: 'feature-flags-account' })
  }

  const clientId = computed(() => store.clientId)
  const flagKey = ref(route.params.id)

  const validationSchema = yup.object({
    description: yup.string().max(200, 'Max 200 characters').default(''),
    enabled: yup.boolean().required()
  })

  // EditFormBlock calls loadService({ id }) where id = route.params.id
  const loadService = ({ id }) => featureFlagsService.load(clientId.value, id)

  // EditFormBlock calls editService(formValues); we need the key from the route
  const editService = (payload) =>
    featureFlagsService.patch(clientId.value, flagKey.value, {
      enabled: payload.enabled,
      description: payload.description
    })
</script>

<template>
  <ContentBlock>
    <template #heading>
      <PageHeadingBlock
        :pageTitle="flagKey"
        :description="`Editing flag for client: ${clientId}`"
      />
    </template>
    <template #content>
      <EditFormBlock
        :editService="editService"
        :loadService="loadService"
        updatedRedirect="list-feature-flags"
        :schema="validationSchema"
      >
        <template #form>
          <FormFieldsEdit :flagKey="flagKey" />
        </template>
        <template #action-bar="{ onSubmit, onCancel, loading }">
          <ActionBarTemplate
            @onSubmit="onSubmit"
            @onCancel="onCancel"
            :loading="loading"
          />
        </template>
      </EditFormBlock>
    </template>
  </ContentBlock>
</template>
