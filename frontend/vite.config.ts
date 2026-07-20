import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  clearScreen: false,
  build: {
    chunkSizeWarningLimit: 1200,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return undefined
          }

          const normalized = id.split('\\').join('/')
          if (
            normalized.includes('/node_modules/react/') ||
            normalized.includes('/node_modules/react-dom/') ||
            normalized.includes('/node_modules/scheduler/')
          ) {
            return 'vendor-react'
          }
          if (
            normalized.includes('/node_modules/antd/') ||
            normalized.includes('/node_modules/@ant-design/') ||
            normalized.includes('/node_modules/@rc-component/') ||
            normalized.includes('/node_modules/rc-') ||
            normalized.includes('/node_modules/dayjs/')
          ) {
            return 'vendor-antd'
          }

          return 'vendor-misc'
        }
      }
    }
  },
  server: {
    port: 36901,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:36900',
        changeOrigin: true
      }
    }
  }
})
