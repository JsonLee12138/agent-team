import type { Rule } from 'eslint';
import type { ModuleReplacement } from 'module-replacements';
export type ImportListenerCallback = (context: Rule.RuleContext, node: Rule.Node, source: string) => void;
/**
 * Creates a rule listener which listens for import/require calls and
 * calls a callback when one is found
 * @param {Rule.RuleContext} context ESLint context
 * @param {Function} callback Callback to call when an import/require is found
 * @return {Rule.RuleListener}
 */
export declare function createImportListener(context: Rule.RuleContext, callback: ImportListenerCallback): Rule.RuleListener;
/**
 * Creates a rule listener for detecting dependencies in a `package.json`
 * file
 * @param {Rule.RuleContext} context ESLint context
 * @param {ImportListenerCallback} callback Listener callback
 * @return {Rule.RuleListener}
 */
export declare function createPackageJsonListener(context: Rule.RuleContext, callback: ImportListenerCallback): Rule.RuleListener;
/**
 * Creates a rule listener which finds replacements in imports/requires
 * @param {Rule.RuleContext} context ESLint context
 * @param {ModuleReplacement[]} replacements List of replacements
 * @return {Rule.RuleListener}
 */
export declare function createReplacementListener(context: Rule.RuleContext, replacements: ModuleReplacement[]): Rule.RuleListener;
