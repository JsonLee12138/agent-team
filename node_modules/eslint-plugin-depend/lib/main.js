import { configFactory as configRecommended } from './configs/recommended.js';
import { config as configLegacyRecommended } from './configs/legacy-recommended.js';
import { rule as banDependencies } from './rules/ban-dependencies.js';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const __dirname = path.dirname(fileURLToPath(import.meta.url));
const { name, version } = JSON.parse(readFileSync(path.resolve(__dirname, '../package.json'), 'utf8'));
export const rules = {
    'ban-dependencies': banDependencies
};
const plugin = {
    meta: { name, version },
    rules
};
export const configs = {
    /**
     * @deprecated Use `flat/recommended` instead.
     * This legacy eslintrc format is not supported by ESLint 10.
     */
    recommended: configLegacyRecommended,
    'flat/recommended': configRecommended(plugin)
};
plugin.configs = configs;
export default plugin;
