import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 36901,
    proxy: {
      '/api': {
        target: 'http://localhost:36900',
        changeOrigin: true
      }
    }
  }
})
