export const performanceImprovements = (plugin) => ({
    plugins: {
        e18e: plugin
    },
    rules: {
        'e18e/prefer-array-from-map': 'error',
        'e18e/prefer-timer-args': 'error',
        'e18e/prefer-date-now': 'error',
        'e18e/prefer-regex-test': 'error',
        'e18e/prefer-array-some': 'error',
        'e18e/prefer-static-regex': 'error'
    }
});
