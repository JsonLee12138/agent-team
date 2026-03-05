import { defineConfig, presetUno, presetIcons } from 'unocss'

export default defineConfig({
  presets: [
    presetUno(),
    presetIcons({
      scale: 1.2,
      cdn: 'https://esm.sh/',
    }),
  ],
  theme: {
    colors: {
      bg: '#09090B',
      surface: '#18181B',
      border: '#27272A',
      primary: '#A8D8C8',
      'accent-purple': '#B8A9C9',
      'accent-yellow': '#FFF3C4',
      'text-main': '#FAFAFA',
      'text-sub': '#A1A1AA',
      verified: '#10B981',
      invalid: '#F43F5E',
      warning: '#F59E0B',
    },
    fontFamily: {
      ui: ['Inter', 'system-ui', 'sans-serif'],
      code: ['JetBrains Mono', 'Fira Code', 'monospace'],
    },
  },
  shortcuts: {
    'btn-primary': 'bg-primary text-bg font-semibold px-6 py-3 rounded-lg cursor-pointer hover:opacity-90 transition-opacity',
    'btn-secondary': 'bg-transparent text-text-main border border-border px-6 py-3 rounded-lg cursor-pointer hover:bg-surface transition-colors',
    card: 'bg-surface border border-border rounded-xl p-6 hover:border-primary/50 transition-colors',
    tag: 'border border-border rounded-full px-3 py-1 text-xs text-text-main',
    'input-field': 'bg-surface border border-border rounded-lg px-4 py-3 text-sm text-text-main placeholder-text-sub font-ui w-full outline-none focus:border-primary/50 transition-colors',
  },
})
