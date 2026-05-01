<template>
  <el-drawer :model-value="visible" size="50%" :title="title" class="editor-drawer" @close="$emit('close')">
    <div class="editor-drawer__meta">
      <p class="eyebrow">{{ copy.editor.eyebrow }}</p>
      <p class="editor-drawer__hint">{{ alertTitle }}</p>
    </div>

    <div class="editor-surface">
      <el-input
        v-model="draft"
        type="textarea"
        :rows="22"
        class="editor-textarea"
        :placeholder="placeholderText"
      />
    </div>

    <template #footer>
      <div class="drawer-footer">
        <el-button @click="$emit('close')">{{ copy.common.cancel }}</el-button>
        <el-button type="primary" @click="submit">{{ copy.common.save }}</el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { copy } from '../copy'

const props = defineProps<{
  visible: boolean
  title: string
  initialValue: string
}>()

const emit = defineEmits<{
  close: []
  save: [value: Record<string, unknown>]
}>()

const alertTitle = copy.editor.hint
const placeholderText = copy.editor.placeholder
const draft = ref(props.initialValue)

watch(
  () => props.initialValue,
  (value) => {
    draft.value = value
  },
  { immediate: true },
)

watch(
  () => props.visible,
  (value) => {
    if (value) {
      draft.value = props.initialValue
    }
  },
)

function submit() {
  try {
    const parsed = JSON.parse(draft.value) as Record<string, unknown>
    emit('save', parsed)
  } catch (error) {
    ElMessage.error((error as Error).message || copy.editor.invalidJson)
  }
}
</script>
