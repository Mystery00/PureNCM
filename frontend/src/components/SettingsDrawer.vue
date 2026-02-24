<script lang="ts" setup>
import { ref, watch } from 'vue'
import {
  NDrawer, NDrawerContent, NForm, NFormItem,
  NInput, NButton, NText, NIcon, NSpace, NDivider,
} from 'naive-ui'
import { FolderOpen } from '@vicons/ionicons5'
import { useConfig } from '@/composables/useConfig'
import { OpenDirectoryDialog } from '../../wailsjs/go/main/App'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ 'update:show': [boolean] }>()

const { config, updateOutputDir, updateFilenamePattern } = useConfig()

// Local editable copy of filename pattern (committed on blur/enter)
const patternDraft = ref(config.value.filenamePattern)
watch(() => config.value.filenamePattern, v => { patternDraft.value = v })

async function selectDir() {
  const dir = await OpenDirectoryDialog()
  if (dir) await updateOutputDir(dir)
}

async function savePattern() {
  await updateFilenamePattern(patternDraft.value.trim() || '{title}')
}

// Preview the pattern with dummy data
const previewName = computed(() => {
  return patternDraft.value
    .replace('{title}', '两个你')
    .replace('{artist}', 'G.E.M.邓紫棋')
    .replace('{album}', '两个你')
})

import { computed } from 'vue'
</script>

<template>
  <NDrawer
    :show="props.show"
    :width="320"
    placement="right"
    @update:show="emit('update:show', $event)"
  >
    <NDrawerContent title="设置" closable>
      <NForm label-placement="top" :label-style="{ fontWeight: '500' }">

        <NFormItem label="输出目录">
          <NSpace vertical :size="6" style="width:100%">
            <NInput
              :value="config.outputDir || '（未设置，将输出到源文件目录）'"
              readonly
              size="small"
            />
            <NButton size="small" @click="selectDir">
              <template #icon><NIcon><FolderOpen /></NIcon></template>
              选择目录
            </NButton>
          </NSpace>
        </NFormItem>

        <NDivider />

        <NFormItem label="文件名格式">
          <NSpace vertical :size="6" style="width:100%">
            <NInput
              v-model:value="patternDraft"
              size="small"
              placeholder="{title}"
              @blur="savePattern"
              @keydown.enter="savePattern"
            />
            <NText depth="3" style="font-size:12px">
              可用占位符：{title}、{artist}、{album}
            </NText>
            <NText style="font-size:12px; color:#63e2b7">
              预览：{{ previewName }}
            </NText>
          </NSpace>
        </NFormItem>

      </NForm>
    </NDrawerContent>
  </NDrawer>
</template>
