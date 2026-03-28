import type { TSNode, TSToken, TSESTree, ParserServicesWithTypeInformation } from '@typescript-eslint/typescript-estree';
import type { TSESLint } from '@typescript-eslint/utils';
import type ts from 'typescript';
export interface ParserServices {
    emitDecoratorMetadata: boolean | undefined;
    experimentalDecorators: boolean | undefined;
    isolatedDeclarations: boolean | undefined;
    esTreeNodeToTSNodeMap: WeakMap<TSESTree.Node, TSNode | TSToken>;
    tsNodeToESTreeNodeMap: WeakMap<TSNode | TSToken, TSESTree.Node>;
    getSymbolAtLocation: (node: TSESTree.Node) => ts.Symbol | undefined;
    getTypeAtLocation: (node: TSESTree.Node) => ts.Type;
    program: ts.Program;
}
export declare function tryGetTypedParserServices(context: Readonly<TSESLint.RuleContext<string, unknown[]>>): ParserServicesWithTypeInformation | null;
export declare function getTypedParserServices(context: Readonly<TSESLint.RuleContext<string, unknown[]>>): ParserServicesWithTypeInformation;
/**
 * Checks if a node's type is an Array type (Array, tuple, or typed array)
 * Returns true if types are unavailable (to avoid false negatives)
 */
export declare function isArrayType(node: TSESTree.Node, context: Readonly<TSESLint.RuleContext<string, unknown[]>>): boolean;
/**
 * Checks if a node's type is a Set
 * Returns true if types are unavailable (to avoid false negatives)
 */
export declare function isSetType(node: TSESTree.Node, context: Readonly<TSESLint.RuleContext<string, unknown[]>>): boolean;
/**
 * Checks if a node's type is a string
 * Returns false if types are unavailable
 */
export declare function isStringType(node: TSESTree.Node, context: Readonly<TSESLint.RuleContext<string, unknown[]>>): boolean;
