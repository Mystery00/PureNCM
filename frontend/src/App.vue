<script lang="ts" setup>
import { darkTheme, NConfigProvider, NMessageProvider, NLayout } from 'naive-ui'
</script>

<template>
  <NConfigProvider :theme="darkTheme">
    <NMessageProvider>
      <NLayout style="height: 100vh; background: #0f0f0f;" />
    </NMessageProvider>
  </NConfigProvider>
</template>
