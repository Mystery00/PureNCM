<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import {
  NConfigProvider, NMessageProvider, NLayout,
  NLayoutHeader, NLayoutContent,
  darkTheme,
} from 'naive-ui'
import TopBar from '@/components/TopBar.vue'
import FileTable from '@/components/FileTable.vue'
import DropZone from '@/components/DropZone.vue'
import SettingsDrawer from '@/components/SettingsDrawer.vue'
import { useConfig } from '@/composables/useConfig'

const { load } = useConfig()
const showSettings = ref(false)

onMounted(() => load())

function handleStartConvert() {
  // Step 5: IPC bridge will implement this
  console.log('start convert — to be implemented in Step 5')
}
</script>

<template>
  <NConfigProvider :theme="darkTheme">
    <NMessageProvider>
      <NLayout style="height: 100vh; background: #111;">

        <NLayoutHeader :bordered="false" style="padding: 0;">
          <TopBar
            @open-settings="showSettings = true"
            @start-convert="handleStartConvert"
          />
        </NLayoutHeader>

        <NLayoutContent style="overflow: hidden; display: flex; flex-direction: column;">
          <DropZone style="flex: 1; overflow: hidden; display: flex; flex-direction: column; padding: 8px 12px;">
            <FileTable />
          </DropZone>
        </NLayoutContent>

      </NLayout>

      <SettingsDrawer v-model:show="showSettings" />
    </NMessageProvider>
  </NConfigProvider>
</template>
