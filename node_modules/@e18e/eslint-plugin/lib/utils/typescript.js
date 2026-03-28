export function tryGetTypedParserServices(context) {
    if (context.sourceCode.parserServices?.program == null) {
        return null;
    }
    return context.sourceCode.parserServices;
}
export function getTypedParserServices(context) {
    const services = tryGetTypedParserServices(context);
    if (services === null) {
        throw new Error(`You have used a rule which requires type information. Please ensure you have typescript-eslint setup alongside this plugin and configured to enable type-aware linting. See https://typescript-eslint.io/getting-started/typed-linting for more information.`);
    }
    return services;
}
const typedArrayTypes = [
    'Int8Array',
    'Uint8Array',
    'Uint8ClampedArray',
    'Int16Array',
    'Uint16Array',
    'Int32Array',
    'Uint32Array',
    'Float32Array',
    'Float64Array',
    'BigInt64Array',
    'BigUint64Array'
];
/**
 * Checks if a node's type is an Array type (Array, tuple, or typed array)
 * Returns true if types are unavailable (to avoid false negatives)
 */
export function isArrayType(node, context) {
    const services = tryGetTypedParserServices(context);
    if (!services) {
        return true;
    }
    const type = services.getTypeAtLocation(node);
    if (!type) {
        return true;
    }
    const checker = services.program.getTypeChecker();
    if (checker.isArrayType(type)) {
        return true;
    }
    if (checker.isTupleType(type)) {
        return true;
    }
    const typeString = checker.typeToString(type);
    if (typedArrayTypes.some((t) => typeString.startsWith(t))) {
        return true;
    }
    return false;
}
const setTypePattern = /^(Readonly)?Set</;
/**
 * Checks if a node's type is a Set
 * Returns true if types are unavailable (to avoid false negatives)
 */
export function isSetType(node, context) {
    const services = tryGetTypedParserServices(context);
    if (!services) {
        return true;
    }
    const type = services.getTypeAtLocation(node);
    if (!type) {
        return true;
    }
    const checker = services.program.getTypeChecker();
    const typeString = checker.typeToString(type);
    return setTypePattern.test(typeString);
}
/**
 * Checks if a node's type is a string
 * Returns false if types are unavailable
 */
export function isStringType(node, context) {
    const services = tryGetTypedParserServices(context);
    if (!services) {
        return false;
    }
    const type = services.getTypeAtLocation(node);
    if (!type) {
        return false;
    }
    const checker = services.program.getTypeChecker();
    const stringType = checker.getStringType();
    return checker.isTypeAssignableTo(type, stringType);
}
