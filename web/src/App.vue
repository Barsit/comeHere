<script setup lang="ts">
import StatusBar from './components/StatusBar.vue'
</script>

<template>
  <div class="app">
    <StatusBar />
    <div class="container">
      <router-view />
    </div>
  </div>
</template>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; background: #f5f5f5; color: #333; }
.app { min-height: 100vh; }
.container { max-width: 960px; margin: 0 auto; padding: 20px; }
</style>
