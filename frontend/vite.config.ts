import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const srcDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), './src')

// The SPA is served same-origin with the API in production (karel.tilcer.cz):
// Traefik routes /api to the backend, everything else to the static Nginx image.
// In dev, proxy /api to the Go backend on :2002.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: { alias: { '@': srcDir } },
  server: {
    proxy: {
      '/api': { target: 'http://127.0.0.1:2002', changeOrigin: true },
    },
  },
  build: {
    // Keep the public bundle lean; the admin CMS is a lazy chunk.
    chunkSizeWarningLimit: 1200,
  },
})
