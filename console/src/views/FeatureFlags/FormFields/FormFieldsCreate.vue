<script setup>
  import FormHorizontal from '@/templates/create-form-block/form-horizontal'
  import FieldText from '@aziontech/webkit/field-text'
  import FieldSwitchBlock from '@aziontech/webkit/field-switch-block'
  import { useField } from 'vee-validate'

  defineOptions({ name: 'form-fields-feature-flags-create' })

  const { value: key } = useField('key')
  const { value: description } = useField('description')
</script>

<template>
  <FormHorizontal
    title="Flag"
    description="Define a feature flag for this client. The key is permanent and cannot be changed after creation."
  >
    <template #inputs>
      <FieldText
        label="Key"
        required
        name="key"
        placeholder="new-checkout"
        :value="key"
        description="Lowercase letters, digits, hyphens and underscores only. Max 64 characters."
      />
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
