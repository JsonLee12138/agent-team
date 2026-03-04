import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import UnoCSS from 'unocss/vite'
import { resolve } from 'node:path'

export default defineConfig({
  plugins: [UnoCSS(), react()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
})
