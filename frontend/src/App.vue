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
import { useConvert } from '@/composables/useConvert'

const { load } = useConfig()
const { startConvert } = useConvert()
const showSettings = ref(false)

onMounted(() => load())
</script>

<template>
  <NConfigProvider :theme="darkTheme">
    <NMessageProvider>
      <NLayout style="height: 100vh; background: #111;">

        <NLayoutHeader :bordered="false" style="padding: 0;">
          <TopBar
            @open-settings="showSettings = true"
            @start-convert="startConvert"
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
