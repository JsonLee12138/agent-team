import type { Rule, ESLint } from 'eslint';
export declare const rules: Record<string, Rule.RuleModule>;
declare const plugin: ESLint.Plugin;
export declare const configs: {
    /**
     * @deprecated Use `flat/recommended` instead.
     * This legacy eslintrc format is not supported by ESLint 10.
     */
    recommended: ESLint.ConfigData<import("@eslint/core", { with: { "resolution-mode": "require" } }).RulesConfig>;
    'flat/recommended': import("eslint").Linter.Config<import("@eslint/core", { with: { "resolution-mode": "require" } }).RulesConfig>;
};
export default plugin;
