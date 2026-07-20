import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
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
