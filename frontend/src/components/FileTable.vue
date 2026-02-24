<script lang="ts" setup>
import { h, computed } from 'vue'
import {
  NDataTable, NTag, NButton, NIcon, NText, NEmpty,
  type DataTableColumns,
} from 'naive-ui'
import { CloseCircle, CheckmarkCircle, TimeOutline, SyncOutline } from '@vicons/ionicons5'
import { useFiles, type FileItem, type FileStatus } from '@/composables/useFiles'

const { files, removeFile, formatBytes } = useFiles()

const statusMap: Record<FileStatus, { label: string; type: 'default' | 'info' | 'success' | 'error' | 'warning' }> = {
  pending:    { label: '等待中',   type: 'default' },
  converting: { label: '转换中',   type: 'info'    },
  done:       { label: '已完成',   type: 'success' },
  error:      { label: '失败',     type: 'error'   },
}

const columns = computed<DataTableColumns<FileItem>>(() => [
  {
    title: '文件名',
    key: 'name',
    ellipsis: { tooltip: true },
    render(row) {
      return h(NText, { style: 'font-size:13px' }, { default: () => row.name })
    },
  },
  {
    title: '大小',
    key: 'size',
    width: 90,
    align: 'right',
    render(row) {
      return h(NText, { depth: 3, style: 'font-size:12px' }, {
        default: () => row.size > 0 ? formatBytes(row.size) : '—',
      })
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 110,
    align: 'center',
    render(row) {
      const s = statusMap[row.status]
      const icon = {
        pending:    TimeOutline,
        converting: SyncOutline,
        done:       CheckmarkCircle,
        error:      CloseCircle,
      }[row.status]
      return h(NTag, {
        type: s.type,
        size: 'small',
        round: true,
        style: row.status === 'converting' ? 'animation: spin 1s linear infinite' : '',
      }, {
        icon: () => h(NIcon, null, { default: () => h(icon) }),
        default: () => s.label,
      })
    },
  },
  {
    title: '错误信息',
    key: 'error',
    ellipsis: { tooltip: true },
    render(row) {
      return row.error
        ? h(NText, { type: 'error', style: 'font-size:12px' }, { default: () => row.error })
        : null
    },
  },
  {
    title: '',
    key: '_actions',
    width: 48,
    align: 'center',
    render(row) {
      return h(NButton, {
        quaternary: true,
        circle: true,
        size: 'small',
        disabled: row.status === 'converting',
        onClick: () => removeFile(row.id),
      }, { icon: () => h(NIcon, null, { default: () => h(CloseCircle) }) })
    },
  },
])
</script>

<template>
  <div class="file-table-wrap">
    <NDataTable
      v-if="files.length > 0"
      :data="files"
      :columns="columns"
      :row-key="(r: FileItem) => r.id"
      size="small"
      striped
    />
    <NEmpty
      v-else
      description="将 .ncm 文件拖拽到这里，或点击「添加文件」"
      style="margin: auto"
    />
  </div>
</template>

<style scoped>
.file-table-wrap {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  padding: 0 4px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
