import { defineConfig } from 'tsdown'

export default defineConfig({
  entry: ['src/index.ts', 'src/commands/**/*.ts'],
  outDir: 'dist',
  unbundle: true,
  format: ['esm'],
  dts: true,
  clean: true,
  shims: true,
  platform: 'node',
})
