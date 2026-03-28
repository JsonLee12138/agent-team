import type { Rule } from 'eslint';
/**
 * Gets the node engine constraint from `package.json` if possible
 * @param {Record} packageJson Package to process
 * @return {string | null}
 */
export declare function getNodeConstraint(packageJson: Record<string, unknown>): string | null;
/**
 * Gets the closest `package.json` for a context
 * @param {Rule.RuleContext} context ESLint context
 * @return {Record | null}
 */
export declare function getClosestPackage(context: Rule.RuleContext): Record<string, unknown> | null;
/**
 * Tests whether the closest `package.json` satisfies the specified
 * node version.
 * This will be true if the package's `engines` field has a version range
 * which this version satisfies, or no field at all.
 * @param {Rule.RuleContext} context ESLint context
 * @param {string} version Version to test
 * @return {boolean}
 */
export declare function closestPackageSatisfiesNodeVersion(context: Rule.RuleContext, version: string): boolean;
