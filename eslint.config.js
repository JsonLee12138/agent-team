import antfu from '@antfu/eslint-config'

export default antfu({
  typescript: true,
  formatters: true,
  rules: {
    'no-console': 'off',
    '@typescript-eslint/no-explicit-any': 'off',
    'node/prefer-global/process': 'off',
    'e18e/prefer-static-regex': 'off',
  },
  ignores: ['dist/', 'node_modules/', 'coverage/', '.agents/', '.claude/', 'bin/', '*.md', 'docs/', 'prompts/'],
})
