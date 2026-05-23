<script setup>
  import FormHorizontal from '@/templates/create-form-block/form-horizontal'
  import FieldText from '@aziontech/webkit/field-text'
  import FieldSwitchBlock from '@aziontech/webkit/field-switch-block'
  import { useField } from 'vee-validate'

  defineOptions({ name: 'form-fields-feature-flags-edit' })

  const props = defineProps({
    flagKey: { type: String, required: true }
  })

  const { value: description } = useField('description')
</script>

<template>
  <FormHorizontal
    title="Flag"
    description="Update the description or enabled state. The key cannot be changed."
  >
    <template #inputs>
      <div class="flex flex-col gap-1">
        <span class="text-sm font-medium text-color-secondary">Key</span>
        <span class="font-mono text-sm">{{ flagKey }}</span>
      </div>
      <FieldText
        label="Description"
        name="description"
        placeholder="Describe what this flag controls"
        :value="description"
        description="Optional. Max 200 characters."
      />
      <FieldSwitchBlock
        nameField="enabled"
        name="enabled"
        auto
        :isCard="false"
        title="Enabled"
      />
    </template>
  </FormHorizontal>
</template>
