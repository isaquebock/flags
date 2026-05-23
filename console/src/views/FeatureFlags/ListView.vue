<script setup>
  import { computed, h, reactive } from 'vue'
  import { useRouter } from 'vue-router'
  import ContentBlock from '@/templates/content-block'
  import PageHeadingBlock from '@/templates/page-heading-block'
  import ListTable from '@/components/list-table'
  import { DataTableActionsButtons } from '@/components/list-table'
  import InputSwitch from '@aziontech/webkit/inputswitch'
  import { useToast } from '@aziontech/webkit/use-toast'
  import { featureFlagsService } from '@/services/v2/feature-flags'
  import { useFeatureFlagsAccountStore } from '@/stores/feature-flags-account'

  defineOptions({ name: 'feature-flags-list-view' })

  const store = useFeatureFlagsAccountStore()
  const router = useRouter()
  const toast = useToast()

  if (!store.clientId) {
    router.replace({ name: 'feature-flags-account' })
  }

  const clientId = computed(() => store.clientId)

  // local overrides for optimistic toggle: { [key]: boolean }
  const toggleOverrides = reactive({})

  const listService = () => featureFlagsService.list(clientId.value)

  const handleToggle = async (key, newValue) => {
    const previous = toggleOverrides[key] ?? !newValue
    toggleOverrides[key] = newValue

    try {
      await featureFlagsService.patch(clientId.value, key, { enabled: newValue })
      delete toggleOverrides[key]
    } catch {
      toggleOverrides[key] = previous
      toast.add({
        severity: 'error',
        summary: 'error',
        detail: `Failed to update flag "${key}". Please try again.`,
        closable: true
      })
    }
  }

  const columns = computed(() => [
    {
      field: 'key',
      header: 'Key'
    },
    {
      field: 'description',
      header: 'Description',
      style: 'max-width: 300px'
    },
    {
      field: 'enabled',
      header: 'Enabled',
      type: 'component',
      component: (columnData) =>
        h(InputSwitch, {
          modelValue: toggleOverrides[columnData.key] ?? columnData.enabled,
          'onUpdate:modelValue': (newValue) => handleToggle(columnData.key, newValue)
        })
    },
    {
      field: 'updatedAt',
      header: 'Last Modified'
    }
  ])

  const actions = computed(() => [
    {
      label: 'Delete',
      type: 'delete',
      title: 'flag',
      icon: 'pi pi-trash',
      service: (key) => featureFlagsService.remove(clientId.value, key)
    }
  ])
</script>

<template>
  <ContentBlock>
    <template #heading>
      <PageHeadingBlock
        pageTitle="Feature Flags"
        :description="`Managing flags for client: ${clientId}`"
      >
        <template #default>
          <div class="flex items-center gap-3">
            <span
              class="text-sm text-color-secondary cursor-pointer underline"
              @click="router.push({ name: 'feature-flags-account' })"
            >
              Change account
            </span>
            <DataTableActionsButtons
              size="small"
              label="Flag"
              createPagePath="/feature-flags/create"
            />
          </div>
        </template>
      </PageHeadingBlock>
    </template>
    <template #content>
      <ListTable
        :listService="listService"
        :columns="columns"
        :actions="actions"
        :frozenColumns="['key']"
        editPagePath="/feature-flags/edit"
        exportFileName="FeatureFlags"
        :lazy="false"
        emptyListMessage="No flags found for this client."
        :empty-block="{
          title: 'No flags yet',
          description: 'Create your first feature flag to start controlling feature rollouts.',
          createButtonLabel: 'Flag'
        }"
      />
    </template>
  </ContentBlock>
</template>
