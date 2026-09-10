import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    open: true,
    allowedHosts: ['fa4b-87-120-126-215.ngrok-free.app'], //ngrok: для dev режима - разрешает HTTP-запросы к указанному хосту
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
