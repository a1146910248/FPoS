<template>
  <div id="app">
    <router-view v-if="isRouterAlive"></router-view>
  </div>
</template>
<script setup lang="ts">
import { nextTick, provide, ref } from 'vue'

// 无感刷新页面
const isRouterAlive = ref(true)
const reload = () => {
  isRouterAlive.value = false
  nextTick(()=>{
    isRouterAlive.value = true
  })
}
// 提供给所有子组件方法
provide('reload', reload)

</script>
