<script lang="ts" setup>
import {
  NButton, NSpace, NText, NTag, NTooltip,
  useMessage,
} from 'naive-ui'
import { Add, FolderOpen, Play, Trash, Settings } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { useFiles } from '@/composables/useFiles'
import { useConfig } from '@/composables/useConfig'
import { OpenFileDialog, OpenDirectoryDialog } from '../../wailsjs/go/main/App'

const emit = defineEmits<{
  openSettings: []
  startConvert: []
}>()

const message = useMessage()
const { addPaths, clearAll, hasFiles, isConverting, pendingCount } = useFiles()
const { config, updateOutputDir } = useConfig()

async function handleAddFiles() {
  try {
    const paths = await OpenFileDialog()
    if (paths && paths.length > 0) {
      addPaths(paths)
    }
  } catch (e) {
    message.error('打开文件対话框失败')
  }
}

async function handleSelectOutputDir() {
  try {
    const dir = await OpenDirectoryDialog()
    if (dir) {
      await updateOutputDir(dir)
      message.success(`输出目录已设置`)
    }
  } catch (e) {
    message.error('打开目录对话框失败')
  }
}
</script>

<template>
  <div class="topbar">
    <!-- Left: file actions -->
    <NSpace align="center" :size="8">
      <NButton type="primary" :disabled="isConverting" @click="handleAddFiles">
        <template #icon><NIcon><Add /></NIcon></template>
        添加文件
      </NButton>

      <NTooltip>
        <template #trigger>
          <NButton :disabled="isConverting" @click="handleSelectOutputDir">
            <template #icon><NIcon><FolderOpen /></NIcon></template>
            输出目录
          </NButton>
        </template>
        {{ config.outputDir || '未设置输出目录（将输出到源文件所在目录）' }}
      </NTooltip>

      <NButton
        v-if="hasFiles"
        quaternary
        :disabled="isConverting"
        @click="clearAll"
      >
        <template #icon><NIcon><Trash /></NIcon></template>
        清空列表
      </NButton>
    </NSpace>

    <!-- Right: status + convert -->
    <NSpace align="center" :size="12">
      <NText v-if="pendingCount > 0" depth="3" style="font-size:13px">
        待转换：<NTag type="info" size="small" round>{{ pendingCount }}</NTag>
      </NText>

      <NButton
        type="success"
        :disabled="pendingCount === 0 || isConverting"
        :loading="isConverting"
        @click="emit('startConvert')"
      >
        <template #icon><NIcon><Play /></NIcon></template>
        全部开始转换
      </NButton>

      <NButton quaternary circle @click="emit('openSettings')">
        <template #icon><NIcon><Settings /></NIcon></template>
      </NButton>
    </NSpace>
  </div>
</template>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.03);
  flex-shrink: 0;
}
</style>
