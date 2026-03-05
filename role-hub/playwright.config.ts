import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  retries: 0,
  timeout: 30_000,
  use: {
    baseURL: 'http://localhost:3000',
  },
})
