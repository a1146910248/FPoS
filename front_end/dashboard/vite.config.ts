import { fileURLToPath, URL } from 'node:url'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig(({mode}) =>{
  const env = loadEnv(mode, __dirname)
  return {
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    proxy: {
      "/dashboardApi": { // "/api" 以及前置字符串会被替换为真正域名
        target: env.VITE_BASE_URL, // 请求域名
        secure: false, // 请求是否为https
        changeOrigin: true, // 是否跨域
        rewrite: (path)=> path.replace(/^\/dashboardApi/,""),
      },
    }
  }
}
})
