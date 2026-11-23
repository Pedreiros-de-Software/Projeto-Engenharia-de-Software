import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxy = {
    '/api': env.VITE_PROXY_TARGET || 'http://localhost:8080',
    '/health': env.VITE_PROXY_TARGET || 'http://localhost:8080',
  }
  return {
    plugins: [react()],
    server: { proxy },
    preview: { proxy },
  }
})
