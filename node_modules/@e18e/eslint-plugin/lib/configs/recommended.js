import { modernization } from './modernization.js';
import { moduleReplacements } from './module-replacements.js';
import { performanceImprovements } from './performance-improvements.js';
export const recommended = (plugin) => {
    const modernizationConfig = modernization(plugin);
    const moduleReplacementsConfig = moduleReplacements(plugin);
    const performanceImprovementsConfig = performanceImprovements(plugin);
    return {
        plugins: {
            e18e: plugin
        },
        rules: {
            ...modernizationConfig.rules,
            ...moduleReplacementsConfig.rules,
            ...performanceImprovementsConfig.rules
        }
    };
};
