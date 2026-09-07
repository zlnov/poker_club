import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    open: true,
    allowedHosts: ['aa46-87-120-126-160.ngrok-free.app'],
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
