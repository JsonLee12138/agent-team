export const configFactory = (plugin) => ({
    plugins: {
        depend: plugin
    },
    rules: {
        'depend/ban-dependencies': 'error'
    }
});
