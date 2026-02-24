<script lang="ts" setup>
import { ref } from 'vue'
import { useFiles } from '@/composables/useFiles'

const { addPaths } = useFiles()
const isDragging = ref(false)
let dragCounter = 0 // track nested dragenter/dragleave

function onDragEnter(e: DragEvent) {
  e.preventDefault()
  dragCounter++
  isDragging.value = true
}

function onDragLeave(e: DragEvent) {
  e.preventDefault()
  dragCounter--
  if (dragCounter <= 0) {
    dragCounter = 0
    isDragging.value = false
  }
}

function onDragOver(e: DragEvent) {
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy'
}

function onDrop(e: DragEvent) {
  e.preventDefault()
  isDragging.value = false
  dragCounter = 0
  const dt = e.dataTransfer
  if (!dt) return
  const paths: string[] = []
  for (const file of Array.from(dt.files)) {
    if (file.name.toLowerCase().endsWith('.ncm')) {
      // In Wails webview, file.path contains the real filesystem path
      const p = (file as any).path as string | undefined
      if (p) paths.push(p)
    }
  }
  if (paths.length > 0) addPaths(paths)
}
</script>

<template>
  <div
    class="drop-zone"
    :class="{ dragging: isDragging }"
    @dragenter="onDragEnter"
    @dragleave="onDragLeave"
    @dragover="onDragOver"
    @drop="onDrop"
  >
    <!-- Overlay shown during drag -->
    <Transition name="fade">
      <div v-if="isDragging" class="drop-overlay">
        <div class="drop-hint">松开鼠标，添加 NCM 文件</div>
      </div>
    </Transition>

    <slot />
  </div>
</template>

<style scoped>
.drop-zone {
  position: relative;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: background 0.15s;
}
.drop-zone.dragging {
  background: rgba(99, 226, 183, 0.05);
}

.drop-overlay {
  position: absolute;
  inset: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px dashed #63e2b7;
  border-radius: 8px;
  background: rgba(99, 226, 183, 0.06);
  pointer-events: none;
}
.drop-hint {
  font-size: 18px;
  font-weight: 600;
  color: #63e2b7;
  letter-spacing: 0.5px;
}

.fade-enter-active,
.fade-leave-active { transition: opacity 0.15s; }
.fade-enter-from,
.fade-leave-to { opacity: 0; }
</style>
