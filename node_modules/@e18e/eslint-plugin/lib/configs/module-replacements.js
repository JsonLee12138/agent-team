export const moduleReplacements = (plugin) => ({
    plugins: {
        e18e: plugin
    },
    rules: {
        'e18e/ban-dependencies': 'error'
    }
});
