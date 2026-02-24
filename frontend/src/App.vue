<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import {
  NConfigProvider, NMessageProvider,
  darkTheme,
} from 'naive-ui'
import TopBar from '@/components/TopBar.vue'
import FileTable from '@/components/FileTable.vue'
import DropZone from '@/components/DropZone.vue'
import SettingsDrawer from '@/components/SettingsDrawer.vue'
import { useConfig } from '@/composables/useConfig'
import { useConvert } from '@/composables/useConvert'

const { load } = useConfig()
const { startConvert } = useConvert()
const showSettings = ref(false)

onMounted(() => load())
</script>

<template>
  <NConfigProvider :theme="darkTheme">
    <NMessageProvider>

      <!-- Plain CSS flex column — avoids NLayoutContent's broken height chain -->
      <div class="app-shell">
        <TopBar
          @open-settings="showSettings = true"
          @start-convert="startConvert"
        />

        <DropZone class="content-area">
          <FileTable />
        </DropZone>
      </div>

      <SettingsDrawer v-model:show="showSettings" />
    </NMessageProvider>
  </NConfigProvider>
</template>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #111;
  overflow: hidden;
}

.content-area {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 8px 12px;
}
</style>
