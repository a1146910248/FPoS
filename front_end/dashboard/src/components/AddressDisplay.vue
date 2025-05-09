<!-- AddressDisplay.vue -->
<template>
  <div class="address-container">
    <span class="short-address">{{ shortAddress }}</span>
    <el-tooltip v-if="showFull" :content="address">
      <el-icon><View /></el-icon>
    </el-tooltip>
    <el-button
      v-if="link"
      type="primary"
      link
      @click="openExplorer"
    >
      <el-icon><Link /></el-icon>
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  address: {
    type: String,
    required: true
  },
  link: {
    type: Boolean,
    default: false
  }
})

const shortAddress = computed(() =>
  `${props.address.slice(0, 6)}...${props.address.slice(-4)}`
)

const openExplorer = () => {
  window.open(`https://explorer.example.com/address/${props.address}`, '_blank')
}
</script>
