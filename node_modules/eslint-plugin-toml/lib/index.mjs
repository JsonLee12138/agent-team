import { t as __exportAll } from "./chunk-DQk6qfdC.mjs";
import * as tomlESLintParser from "toml-eslint-parser";
import { VisitorKeys, getStaticTOMLValue, parseTOML, traverseNodes } from "toml-eslint-parser";
import path from "path";
import { CallMethodStep, ConfigCommentParser, Directive, TextSourceCodeBase, VisitNodeStep } from "@eslint/plugin-kit";
import { TokenStore } from "@ota-meshi/ast-token-store";

//#region src/utils/index.ts
/**
* Define the rule.
* @param ruleName ruleName
* @param rule rule module
*/
function createRule(ruleName, rule) {
	return {
		meta: {
			...rule.meta,
			docs: {
				...rule.meta.docs,
				url: `https://ota-meshi.github.io/eslint-plugin-toml/rules/${ruleName}.html`,
				ruleId: `toml/${ruleName}`,
				ruleName
			}
		},
		create(context) {
			const parserServices = context.sourceCode.parserServices || {};
			if (typeof parserServices.defineCustomBlocksVisitor === "function" && path.extname(context.filename) === ".vue") return parserServices.defineCustomBlocksVisitor(context, tomlESLintParser, {
				target: ["toml", "toml"],
				create(blockContext) {
					return rule.create(blockContext, { customBlock: true });
				}
			});
			return rule.create(context, { customBlock: false });
		}
	};
}

//#endregion
//#region src/utils/ast-utils.ts
const LINEBREAK_MATCHER = /\r\n|[\n\r\u2028\u2029]/u;
/**
* Checks if the given token is a comment token or not.
* @param {Token} token The token to check.
* @returns {boolean} `true` if the token is a comment token.
*/
function isCommentToken(token) {
	return Boolean(token && token.type === "Block");
}
/**
* Checks if the given token is a comma token or not.
* @param token The token to check.
* @returns `true` if the token is a comma token.
*/
function isCommaToken(token) {
	return token != null && token.value === "," && token.type === "Punctuator";
}
/**
* Check whether the given token is a equal sign.
* @param token The token to check.
* @returns `true` if the token is a equal sign.
*/
function isEqualSign(token) {
	return token != null && token.type === "Punctuator" && token.value === "=";
}
/**
* Checks if the given token is a closing parenthesis token or not.
* @param token The token to check.
* @returns `true` if the token is a closing parenthesis token.
*/
function isClosingParenToken(token) {
	return token != null && token.value === ")" && token.type === "Punctuator";
}
const isNotClosingParenToken = negate(isClosingParenToken);
/**
* Checks if the given token is a closing square bracket token or not.
* @param token The token to check.
* @returns `true` if the token is a closing square bracket token.
*/
function isClosingBracketToken(token) {
	return token != null && token.value === "]" && token.type === "Punctuator";
}
/**
* Checks if the given token is a closing brace token or not.
* @param token The token to check.
* @returns `true` if the token is a closing brace token.
*/
function isClosingBraceToken(token) {
	return token != null && token.value === "}" && token.type === "Punctuator";
}
/**
* Determines whether two adjacent tokens are on the same line.
* @param left The left token object.
* @param right The right token object.
* @returns Whether or not the tokens are on the same line.
* @public
*/
function isTokenOnSameLine(left, right) {
	return left?.loc?.end.line === right?.loc?.start.line;
}
/**
* Creates the negate function of the given function.
* @param f The function to negate.
* @returns Negated function.
*/
function negate(f) {
	return ((token) => !f(token));
}

//#endregion
//#region src/rules/array-bracket-newline.ts
var array_bracket_newline_default = createRule("array-bracket-newline", {
	meta: {
		docs: {
			description: "enforce linebreaks after opening and before closing array brackets",
			categories: ["standard"],
			extensionRule: "array-bracket-newline"
		},
		type: "layout",
		fixable: "whitespace",
		schema: [{ oneOf: [{
			type: "string",
			enum: [
				"always",
				"never",
				"consistent"
			]
		}, {
			type: "object",
			properties: {
				multiline: { type: "boolean" },
				minItems: {
					type: ["integer", "null"],
					minimum: 0
				}
			},
			additionalProperties: false
		}] }],
		messages: {
			unexpectedOpeningLinebreak: "There should be no linebreak after '['.",
			unexpectedClosingLinebreak: "There should be no linebreak before ']'.",
			missingOpeningLinebreak: "A linebreak is required after '['.",
			missingClosingLinebreak: "A linebreak is required before ']'."
		}
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		/**
		* Normalizes a given option value.
		* @param option An option value to parse.
		* @returns Normalized option object.
		*/
		function normalizeOptionValue(option) {
			let consistent = false;
			let multiline = false;
			let minItems = 0;
			if (option) if (option === "consistent") {
				consistent = true;
				minItems = Number.POSITIVE_INFINITY;
			} else if (option === "always" || typeof option !== "string" && option.minItems === 0) minItems = 0;
			else if (option === "never") minItems = Number.POSITIVE_INFINITY;
			else {
				multiline = Boolean(option.multiline);
				minItems = option.minItems || Number.POSITIVE_INFINITY;
			}
			else {
				consistent = false;
				multiline = true;
				minItems = Number.POSITIVE_INFINITY;
			}
			return {
				consistent,
				multiline,
				minItems
			};
		}
		/**
		* Normalizes a given option value.
		* @param options An option value to parse.
		* @returns Normalized option object.
		*/
		function normalizeOptions(options) {
			return { TOMLArray: normalizeOptionValue(options) };
		}
		/**
		* Reports that there shouldn't be a linebreak after the first token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportNoBeginningLinebreak(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "unexpectedOpeningLinebreak",
				fix(fixer) {
					const nextToken = sourceCode.getTokenAfter(token, { includeComments: true });
					if (!nextToken || isCommentToken(nextToken)) return null;
					return fixer.removeRange([token.range[1], nextToken.range[0]]);
				}
			});
		}
		/**
		* Reports that there shouldn't be a linebreak before the last token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportNoEndingLinebreak(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "unexpectedClosingLinebreak",
				fix(fixer) {
					const previousToken = sourceCode.getTokenBefore(token, { includeComments: true });
					if (!previousToken || isCommentToken(previousToken)) return null;
					return fixer.removeRange([previousToken.range[1], token.range[0]]);
				}
			});
		}
		/**
		* Reports that there should be a linebreak after the first token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportRequiredBeginningLinebreak(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "missingOpeningLinebreak",
				fix(fixer) {
					return fixer.insertTextAfter(token, "\n");
				}
			});
		}
		/**
		* Reports that there should be a linebreak before the last token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportRequiredEndingLinebreak(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "missingClosingLinebreak",
				fix(fixer) {
					return fixer.insertTextBefore(token, "\n");
				}
			});
		}
		/**
		* Reports a given node if it violated this rule.
		* @param node A node to check. This is an ArrayExpression node or an ArrayPattern node.
		*/
		function check(node) {
			const elements = node.elements;
			const options = normalizeOptions(context.options[0])[node.type];
			const openBracket = sourceCode.getFirstToken(node);
			const closeBracket = sourceCode.getLastToken(node);
			const firstIncComment = sourceCode.getTokenAfter(openBracket, { includeComments: true });
			const lastIncComment = sourceCode.getTokenBefore(closeBracket, { includeComments: true });
			const first = sourceCode.getTokenAfter(openBracket);
			const last = sourceCode.getTokenBefore(closeBracket);
			/**
			* Use tokens or comments to check multiline or not.
			* But use only tokens to check whether linebreaks are needed.
			* This allows:
			*     var arr = [ // eslint-disable-line foo
			*         'a'
			*     ]
			*/
			if (elements.length >= options.minItems || options.multiline && elements.length > 0 && firstIncComment.loc.start.line !== lastIncComment.loc.end.line || elements.length === 0 && firstIncComment.type === "Block" && firstIncComment.loc.start.line !== lastIncComment.loc.end.line && firstIncComment === lastIncComment || options.consistent && openBracket.loc.end.line !== first.loc.start.line) {
				if (isTokenOnSameLine(openBracket, first)) reportRequiredBeginningLinebreak(node, openBracket);
				if (isTokenOnSameLine(last, closeBracket)) reportRequiredEndingLinebreak(node, closeBracket);
			} else {
				if (!isTokenOnSameLine(openBracket, first)) reportNoBeginningLinebreak(node, openBracket);
				if (!isTokenOnSameLine(last, closeBracket)) reportNoEndingLinebreak(node, closeBracket);
			}
		}
		return { TOMLArray: check };
	}
});

//#endregion
//#region src/rules/array-bracket-spacing.ts
var array_bracket_spacing_default = createRule("array-bracket-spacing", {
	meta: {
		docs: {
			description: "enforce consistent spacing inside array brackets",
			categories: ["standard"],
			extensionRule: "array-bracket-spacing"
		},
		type: "layout",
		fixable: "whitespace",
		schema: [{
			type: "string",
			enum: ["always", "never"]
		}, {
			type: "object",
			properties: {
				singleValue: { type: "boolean" },
				objectsInArrays: { type: "boolean" },
				arraysInArrays: { type: "boolean" }
			},
			additionalProperties: false
		}],
		messages: {
			unexpectedSpaceAfter: "There should be no space after '{{tokenValue}}'.",
			unexpectedSpaceBefore: "There should be no space before '{{tokenValue}}'.",
			missingSpaceAfter: "A space is required after '{{tokenValue}}'.",
			missingSpaceBefore: "A space is required before '{{tokenValue}}'."
		}
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		const spaced = (context.options[0] || "always") === "always";
		/**
		* Determines whether an option is set, relative to the spacing option.
		* If spaced is "always", then check whether option is set to false.
		* If spaced is "never", then check whether option is set to true.
		* @param option The option to exclude.
		* @returns Whether or not the property is excluded.
		*/
		function isOptionSet(option) {
			return context.options[1] ? context.options[1][option] === !spaced : false;
		}
		const options = {
			spaced,
			singleElementException: isOptionSet("singleValue"),
			objectsInArraysException: isOptionSet("objectsInArrays"),
			arraysInArraysException: isOptionSet("arraysInArrays"),
			isOpeningBracketMustBeSpaced(node) {
				if (options.singleElementException && node.elements.length === 1) return !options.spaced;
				const firstElement = node.elements[0];
				return firstElement && (options.objectsInArraysException && isObjectType(firstElement) || options.arraysInArraysException && isArrayType(firstElement)) ? !options.spaced : options.spaced;
			},
			isClosingBracketMustBeSpaced(node) {
				if (options.singleElementException && node.elements.length === 1) return !options.spaced;
				const lastElement = node.elements[node.elements.length - 1];
				return lastElement && (options.objectsInArraysException && isObjectType(lastElement) || options.arraysInArraysException && isArrayType(lastElement)) ? !options.spaced : options.spaced;
			}
		};
		/**
		* Reports that there shouldn't be a space after the first token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportNoBeginningSpace(node, token) {
			const nextToken = sourceCode.getTokenAfter(token);
			context.report({
				node,
				loc: {
					start: token.loc.end,
					end: nextToken.loc.start
				},
				messageId: "unexpectedSpaceAfter",
				data: { tokenValue: token.value },
				fix(fixer) {
					return fixer.removeRange([token.range[1], nextToken.range[0]]);
				}
			});
		}
		/**
		* Reports that there shouldn't be a space before the last token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportNoEndingSpace(node, token) {
			const previousToken = sourceCode.getTokenBefore(token);
			context.report({
				node,
				loc: {
					start: previousToken.loc.end,
					end: token.loc.start
				},
				messageId: "unexpectedSpaceBefore",
				data: { tokenValue: token.value },
				fix(fixer) {
					return fixer.removeRange([previousToken.range[1], token.range[0]]);
				}
			});
		}
		/**
		* Reports that there should be a space after the first token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportRequiredBeginningSpace(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "missingSpaceAfter",
				data: { tokenValue: token.value },
				fix(fixer) {
					return fixer.insertTextAfter(token, " ");
				}
			});
		}
		/**
		* Reports that there should be a space before the last token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportRequiredEndingSpace(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "missingSpaceBefore",
				data: { tokenValue: token.value },
				fix(fixer) {
					return fixer.insertTextBefore(token, " ");
				}
			});
		}
		/**
		* Determines if a node is an object type
		* @param node The node to check.
		* @returns Whether or not the node is an object type.
		*/
		function isObjectType(node) {
			return node && node.type === "TOMLInlineTable";
		}
		/**
		* Determines if a node is an array type
		* @param node The node to check.
		* @returns Whether or not the node is an array type.
		*/
		function isArrayType(node) {
			return node && node.type === "TOMLArray";
		}
		/**
		* Validates the spacing around array brackets
		* @param node The node we're checking for spacing
		*/
		function validateArraySpacing(node) {
			if (options.spaced && node.elements.length === 0) return;
			const first = sourceCode.getFirstToken(node);
			const last = sourceCode.getLastToken(node);
			const second = sourceCode.getTokenAfter(first, { includeComments: true });
			const penultimate = sourceCode.getTokenBefore(last, { includeComments: true });
			if (isTokenOnSameLine(first, second)) {
				if (options.isOpeningBracketMustBeSpaced(node)) {
					if (!sourceCode.isSpaceBetween(first, second)) reportRequiredBeginningSpace(node, first);
				} else if (sourceCode.isSpaceBetween(first, second)) reportNoBeginningSpace(node, first);
			}
			if (first !== penultimate && isTokenOnSameLine(penultimate, last)) {
				if (options.isClosingBracketMustBeSpaced(node)) {
					if (!sourceCode.isSpaceBetween(penultimate, last)) reportRequiredEndingSpace(node, last);
				} else if (sourceCode.isSpaceBetween(penultimate, last)) reportNoEndingSpace(node, last);
			}
		}
		return { TOMLArray: validateArraySpacing };
	}
});

//#endregion
//#region src/rules/array-element-newline.ts
var array_element_newline_default = createRule("array-element-newline", {
	meta: {
		docs: {
			description: "enforce line breaks between array elements",
			categories: ["standard"],
			extensionRule: "array-element-newline"
		},
		type: "layout",
		fixable: "whitespace",
		schema: {
			definitions: { basicConfig: { oneOf: [{
				type: "string",
				enum: [
					"always",
					"never",
					"consistent"
				]
			}, {
				type: "object",
				properties: {
					multiline: { type: "boolean" },
					minItems: {
						type: ["integer", "null"],
						minimum: 0
					}
				},
				additionalProperties: false
			}] } },
			type: "array",
			items: [{ oneOf: [{ $ref: "#/definitions/basicConfig" }, {
				type: "object",
				properties: {
					ArrayExpression: { $ref: "#/definitions/basicConfig" },
					ArrayPattern: { $ref: "#/definitions/basicConfig" },
					TOMLArray: { $ref: "#/definitions/basicConfig" }
				},
				additionalProperties: false,
				minProperties: 1
			}] }]
		},
		messages: {
			unexpectedLineBreak: "There should be no linebreak here.",
			missingLineBreak: "There should be a linebreak after this element."
		}
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		/**
		* Normalizes a given option value.
		* @param providedOption An option value to parse.
		* @returns Normalized option object.
		*/
		function normalizeOptionValue(providedOption) {
			let consistent = false;
			let multiline = false;
			let minItems;
			const option = providedOption || "always";
			if (!option || option === "always" || typeof option === "object" && option.minItems === 0) minItems = 0;
			else if (option === "never") minItems = Number.POSITIVE_INFINITY;
			else if (option === "consistent") {
				consistent = true;
				minItems = Number.POSITIVE_INFINITY;
			} else {
				multiline = Boolean(option.multiline);
				minItems = option.minItems || Number.POSITIVE_INFINITY;
			}
			return {
				consistent,
				multiline,
				minItems
			};
		}
		/**
		* Normalizes a given option value.
		* @param options An option value to parse.
		* @returns Normalized option object.
		*/
		function normalizeOptions(options) {
			if (options && (options.ArrayExpression || options.TOMLArray)) {
				let expressionOptions;
				if (options.ArrayExpression) expressionOptions = normalizeOptionValue(options.ArrayExpression);
				if (options.TOMLArray) expressionOptions = normalizeOptionValue(options.TOMLArray);
				return { TOMLArray: expressionOptions };
			}
			return { TOMLArray: normalizeOptionValue(options) };
		}
		/**
		* Reports that there shouldn't be a line break after the first token
		* @param token The token to use for the report.
		*/
		function reportNoLineBreak(token) {
			const tokenBefore = sourceCode.getTokenBefore(token, { includeComments: true });
			context.report({
				loc: {
					start: tokenBefore.loc.end,
					end: token.loc.start
				},
				messageId: "unexpectedLineBreak",
				fix(fixer) {
					if (isCommentToken(tokenBefore)) return null;
					if (!isTokenOnSameLine(tokenBefore, token)) return fixer.replaceTextRange([tokenBefore.range[1], token.range[0]], " ");
					/**
					* This will check if the comma is on the same line as the next element
					* Following array:
					* [
					*     1
					*     , 2
					*     , 3
					* ]
					*
					* will be fixed to:
					* [
					*     1, 2, 3
					* ]
					*/
					const twoTokensBefore = sourceCode.getTokenBefore(tokenBefore, { includeComments: true });
					if (isCommentToken(twoTokensBefore)) return null;
					return fixer.replaceTextRange([twoTokensBefore.range[1], tokenBefore.range[0]], "");
				}
			});
		}
		/**
		* Reports that there should be a line break after the first token
		* @param token The token to use for the report.
		*/
		function reportRequiredLineBreak(token) {
			const tokenBefore = sourceCode.getTokenBefore(token, { includeComments: true });
			context.report({
				loc: {
					start: tokenBefore.loc.end,
					end: token.loc.start
				},
				messageId: "missingLineBreak",
				fix(fixer) {
					return fixer.replaceTextRange([tokenBefore.range[1], token.range[0]], "\n");
				}
			});
		}
		/**
		* Reports a given node if it violated this rule.
		* @param node A node to check. This is an ObjectExpression node or an ObjectPattern node.
		*/
		function check(node) {
			const elements = node.elements;
			const options = normalizeOptions(context.options[0])[node.type];
			if (!options) return;
			let elementBreak = false;
			/**
			* MULTILINE: true
			* loop through every element and check
			* if at least one element has linebreaks inside
			* this ensures that following is not valid (due to elements are on the same line):
			*
			* [
			*      1,
			*      2,
			*      3
			* ]
			*/
			if (options.multiline) elementBreak = elements.filter((element) => element !== null).some((element) => element.loc.start.line !== element.loc.end.line);
			let linebreaksCount = 0;
			for (let i = 0; i < node.elements.length; i++) {
				const element = node.elements[i];
				const previousElement = elements[i - 1];
				if (i === 0 || element === null || previousElement === null) continue;
				const commaToken = sourceCode.getFirstTokenBetween(previousElement, element, isCommaToken);
				if (!isTokenOnSameLine(sourceCode.getTokenBefore(commaToken), sourceCode.getTokenAfter(commaToken))) linebreaksCount++;
			}
			const needsLinebreaks = elements.length >= options.minItems || options.multiline && elementBreak || options.consistent && linebreaksCount > 0 && linebreaksCount < node.elements.length;
			elements.forEach((element, i) => {
				const previousElement = elements[i - 1];
				if (i === 0 || element === null || previousElement === null) return;
				const commaToken = sourceCode.getFirstTokenBetween(previousElement, element, isCommaToken);
				const lastTokenOfPreviousElement = sourceCode.getTokenBefore(commaToken);
				const firstTokenOfCurrentElement = sourceCode.getTokenAfter(commaToken);
				if (needsLinebreaks) {
					if (isTokenOnSameLine(lastTokenOfPreviousElement, firstTokenOfCurrentElement)) reportRequiredLineBreak(firstTokenOfCurrentElement);
				} else if (!isTokenOnSameLine(lastTokenOfPreviousElement, firstTokenOfCurrentElement)) reportNoLineBreak(firstTokenOfCurrentElement);
			});
		}
		return { TOMLArray: check };
	}
});

//#endregion
//#region src/rules/comma-style.ts
var comma_style_default = createRule("comma-style", {
	meta: {
		docs: {
			description: "enforce consistent comma style in array",
			categories: ["standard"],
			extensionRule: "comma-style"
		},
		type: "layout",
		fixable: "code",
		schema: [{
			type: "string",
			enum: ["first", "last"]
		}, {
			type: "object",
			properties: { exceptions: {
				type: "object",
				additionalProperties: { type: "boolean" }
			} },
			additionalProperties: false
		}],
		messages: {
			unexpectedLineBeforeAndAfterComma: "Bad line breaking before and after ','.",
			expectedCommaFirst: "',' should be placed first.",
			expectedCommaLast: "',' should be placed last."
		}
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		const style = context.options[0] || "last";
		const exceptions = {};
		if (context.options.length === 2 && Object.prototype.hasOwnProperty.call(context.options[1], "exceptions")) {
			context.options[1] ??= { exceptions: {} };
			const rawExceptions = context.options[1].exceptions;
			const keys = Object.keys(rawExceptions);
			for (let i = 0; i < keys.length; i++) exceptions[keys[i]] = rawExceptions[keys[i]];
		}
		/**
		* Modified text based on the style
		* @param styleType Style type
		* @param text Source code text
		* @returns modified text
		* @private
		*/
		function getReplacedText(styleType, text) {
			switch (styleType) {
				case "between": return `,${text.replace(LINEBREAK_MATCHER, "")}`;
				case "first": return `${text},`;
				case "last": return `,${text}`;
				default: return "";
			}
		}
		/**
		* Determines the fixer function for a given style.
		* @param styleType comma style
		* @param previousItemToken The token to check.
		* @param commaToken The token to check.
		* @param currentItemToken The token to check.
		* @returns Fixer function
		* @private
		*/
		function getFixerFunction(styleType, previousItemToken, commaToken, currentItemToken) {
			const text = sourceCode.text.slice(previousItemToken.range[1], commaToken.range[0]) + sourceCode.text.slice(commaToken.range[1], currentItemToken.range[0]);
			const range = [previousItemToken.range[1], currentItemToken.range[0]];
			return function(fixer) {
				return fixer.replaceTextRange(range, getReplacedText(styleType, text));
			};
		}
		/**
		* Validates the spacing around single items in lists.
		* @param previousItemToken The last token from the previous item.
		* @param commaToken The token representing the comma.
		* @param currentItemToken The first token of the current item.
		* @param reportItem The item to use when reporting an error.
		* @private
		*/
		function validateCommaItemSpacing(previousItemToken, commaToken, currentItemToken, reportItem) {
			if (isTokenOnSameLine(commaToken, currentItemToken) && isTokenOnSameLine(previousItemToken, commaToken)) {} else if (!isTokenOnSameLine(commaToken, currentItemToken) && !isTokenOnSameLine(previousItemToken, commaToken)) {
				const comment = sourceCode.getCommentsAfter(commaToken)[0];
				const styleType = comment && comment.type === "Block" && isTokenOnSameLine(commaToken, comment) ? style : "between";
				context.report({
					node: reportItem,
					loc: commaToken.loc,
					messageId: "unexpectedLineBeforeAndAfterComma",
					fix: getFixerFunction(styleType, previousItemToken, commaToken, currentItemToken)
				});
			} else if (style === "first" && !isTokenOnSameLine(commaToken, currentItemToken)) context.report({
				node: reportItem,
				loc: commaToken.loc,
				messageId: "expectedCommaFirst",
				fix: getFixerFunction(style, previousItemToken, commaToken, currentItemToken)
			});
			else if (style === "last" && isTokenOnSameLine(commaToken, currentItemToken)) context.report({
				node: reportItem,
				loc: commaToken.loc,
				messageId: "expectedCommaLast",
				fix: getFixerFunction(style, previousItemToken, commaToken, currentItemToken)
			});
		}
		/**
		* Checks the comma placement with regards to a declaration/property/element
		* @param node The binary expression node to check
		* @param property The property of the node containing child nodes.
		* @private
		*/
		function validateComma(node, property) {
			const items = node[property];
			const arrayLiteral = node.type === "TOMLArray";
			if (items.length > 1 || arrayLiteral) {
				let previousItemToken = sourceCode.getFirstToken(node);
				items.forEach((item) => {
					const commaToken = item ? sourceCode.getTokenBefore(item) : previousItemToken;
					const currentItemToken = item ? sourceCode.getFirstToken(item) : sourceCode.getTokenAfter(commaToken);
					const reportItem = item || currentItemToken;
					/**
					* This works by comparing three token locations:
					* - previousItemToken is the last token of the previous item
					* - commaToken is the location of the comma before the current item
					* - currentItemToken is the first token of the current item
					*
					* These values get switched around if item is undefined.
					* previousItemToken will refer to the last token not belonging
					* to the current item, which could be a comma or an opening
					* square bracket. currentItemToken could be a comma.
					*
					* All comparisons are done based on these tokens directly, so
					* they are always valid regardless of an undefined item.
					*/
					if (isCommaToken(commaToken)) validateCommaItemSpacing(previousItemToken, commaToken, currentItemToken, reportItem);
					if (item) {
						const tokenAfterItem = sourceCode.getTokenAfter(item, isNotClosingParenToken);
						previousItemToken = tokenAfterItem ? sourceCode.getTokenBefore(tokenAfterItem) : sourceCode.ast.tokens[sourceCode.ast.tokens.length - 1];
					} else previousItemToken = currentItemToken;
				});
				/**
				* Special case for array literals that have empty last items, such
				* as [ 1, 2, ]. These arrays only have two items show up in the
				* AST, so we need to look at the token to verify that there's no
				* dangling comma.
				*/
				if (arrayLiteral) {
					const lastToken = sourceCode.getLastToken(node);
					const nextToLastToken = sourceCode.getTokenBefore(lastToken);
					if (isCommaToken(nextToLastToken)) validateCommaItemSpacing(sourceCode.getTokenBefore(nextToLastToken), nextToLastToken, lastToken, lastToken);
				}
			}
		}
		const nodes = {};
		if (!exceptions.ObjectExpression && !exceptions.TOMLInlineTable) nodes.TOMLInlineTable = function(node) {
			validateComma(node, "body");
		};
		if (!exceptions.ArrayExpression && !exceptions.TOMLArray) nodes.TOMLArray = function(node) {
			validateComma(node, "elements");
		};
		return nodes;
	}
});

//#endregion
//#region src/rules/indent.ts
const ITERATION_OPTS = Object.freeze({ includeComments: true });
/**
* Create get indentation function
*/
function buildIndentUtility(optionValue) {
	const indent = optionValue ?? 2;
	const textIndent = typeof indent === "number" ? " ".repeat(indent) : "	";
	return {
		getIndentText: (offset) => offset === 1 ? textIndent : textIndent.repeat(offset),
		outdent(indent) {
			return indent.slice(0, -textIndent.length);
		}
	};
}
var indent_default = createRule("indent", {
	meta: {
		docs: {
			description: "enforce consistent indentation",
			categories: ["standard"],
			extensionRule: false
		},
		fixable: "whitespace",
		schema: [{ oneOf: [{ enum: ["tab"] }, {
			type: "integer",
			minimum: 0
		}] }, {
			type: "object",
			properties: {
				subTables: {
					type: "integer",
					minimum: 0
				},
				keyValuePairs: {
					type: "integer",
					minimum: 0
				}
			},
			additionalProperties: false
		}],
		messages: { wrongIndentation: "Expected indentation of {{expected}} but found {{actual}}." },
		type: "layout"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		const { getIndentText, outdent } = buildIndentUtility(context.options[0]);
		const subTablesOffset = context.options[1]?.subTables ?? 0;
		const keyValuePairsOffset = context.options[1]?.keyValuePairs ?? 0;
		const offsets = /* @__PURE__ */ new Map();
		/**
		* Set offset to the given tokens.
		* @param token The token to set.
		* @param offset The offset of the tokens.
		* @param baseToken The token of the base offset.
		*/
		function setOffset(token, offset, baseToken) {
			if (token == null) return;
			if (Array.isArray(token)) for (const t of token) setOffset(t, offset, baseToken);
			else offsets.set(token, {
				baseToken,
				offset
			});
		}
		/**
		* Process the given node list.
		* @param {(AST.YAMLNode|null)[]} nodeList The node to process.
		* @param {number} offset The offset to set.
		* @returns {void}
		*/
		function processNodeList(nodeList, left, right, offset) {
			let lastToken = left;
			const alignTokens = /* @__PURE__ */ new Set();
			for (const node of nodeList) {
				if (node == null) continue;
				const elementTokens = {
					firstToken: sourceCode.getFirstToken(node),
					lastToken: sourceCode.getLastToken(node)
				};
				let t = lastToken;
				while ((t = sourceCode.getTokenAfter(t, ITERATION_OPTS)) != null && t.range[1] <= elementTokens.firstToken.range[0]) alignTokens.add(t);
				alignTokens.add(elementTokens.firstToken);
				lastToken = elementTokens.lastToken;
			}
			if (right != null) {
				let t = lastToken;
				while ((t = sourceCode.getTokenAfter(t, ITERATION_OPTS)) != null && t.range[1] <= right.range[0]) alignTokens.add(t);
			}
			alignTokens.delete(left);
			setOffset([...alignTokens], offset, left);
			if (right != null) setOffset(right, 0, left);
		}
		return {
			TOMLTopLevelTable(node) {
				const first = sourceCode.getFirstToken(node, ITERATION_OPTS);
				if (!first) return;
				const beforeTokens = sourceCode.getTokensBefore(first, ITERATION_OPTS);
				if (beforeTokens.length) {
					const firstOfAllTokens = beforeTokens[0];
					offsets.set(firstOfAllTokens, {
						baseToken: null,
						offset: 0,
						expectedIndent: ""
					});
					setOffset(beforeTokens.slice(1), 0, firstOfAllTokens);
					setOffset(first, 0, firstOfAllTokens);
				} else offsets.set(first, {
					baseToken: null,
					offset: 0,
					expectedIndent: ""
				});
				let tableKeyStack = [];
				/** Get offset from given table keys */
				function getTableOffset(keys) {
					let last = tableKeyStack.pop();
					while (last) {
						if (last.keys.length && last.keys.length <= keys.length && last.keys.every((k, i) => k === keys[i])) {
							if (last.keys.length < keys.length) {
								tableKeyStack.push(last);
								return last.offset + subTablesOffset;
							}
							return last.offset;
						}
						last = tableKeyStack.pop();
					}
					return 0;
				}
				for (const body of node.body) {
					const bodyFirstToken = sourceCode.getFirstToken(body);
					if (body.type === "TOMLKeyValue") {
						if (bodyFirstToken !== first) setOffset(bodyFirstToken, 0, first);
					}
					if (body.type === "TOMLTable") {
						const keys = getStaticTOMLValue(body.key);
						const offset = getTableOffset(keys);
						tableKeyStack.push({
							keys,
							offset
						});
						if (bodyFirstToken !== first) setOffset(bodyFirstToken, offset, first);
					} else tableKeyStack = [];
				}
			},
			TOMLTable(node) {
				const openBracket = sourceCode.getFirstToken(node);
				if (node.kind === "array") setOffset(sourceCode.getTokenAfter(openBracket), 0, openBracket);
				setOffset(sourceCode.getFirstToken(node.key), 1, openBracket);
				const closeBracket = sourceCode.getTokenAfter(node.key);
				setOffset(closeBracket, 0, openBracket);
				if (node.kind === "array") setOffset(sourceCode.getTokenAfter(closeBracket), 0, closeBracket);
				processNodeList(node.body, openBracket, null, keyValuePairsOffset);
			},
			TOMLKeyValue(node) {
				const keyToken = sourceCode.getFirstToken(node.key);
				const valueToken = sourceCode.getFirstToken(node.value);
				const eqToken = sourceCode.getTokenBefore(node.value, isEqualSign);
				setOffset(eqToken, 1, keyToken);
				setOffset(valueToken, 1, eqToken);
			},
			TOMLKey(node) {
				const first = sourceCode.getFirstToken(node, ITERATION_OPTS);
				processNodeList(node.keys, first, null, 1);
			},
			TOMLValue() {},
			TOMLBare() {},
			TOMLQuoted() {},
			TOMLArray(node) {
				const openBracket = sourceCode.getFirstToken(node);
				const closeBracket = sourceCode.getLastToken(node);
				processNodeList(node.elements, openBracket, closeBracket, 1);
			},
			TOMLInlineTable(node) {
				const openBrace = sourceCode.getFirstToken(node);
				const closeBrace = sourceCode.getLastToken(node);
				processNodeList(node.body, openBrace, closeBrace, 1);
			},
			"Program:exit"(node) {
				const lineIndentsStep1 = [];
				let tokensOnSameLine = [];
				for (const token of sourceCode.getTokens(node, ITERATION_OPTS)) if (tokensOnSameLine.length === 0 || tokensOnSameLine[0].loc.start.line === token.loc.start.line) tokensOnSameLine.push(token);
				else {
					const lineIndent = processExpectedIndent(tokensOnSameLine);
					lineIndentsStep1[lineIndent.line] = lineIndent;
					tokensOnSameLine = [token];
				}
				if (tokensOnSameLine.length >= 1) {
					const lineIndent = processExpectedIndent(tokensOnSameLine);
					lineIndentsStep1[lineIndent.line] = lineIndent;
				}
				validateLines(processMissingLines(lineIndentsStep1));
			}
		};
		/**
		* Process the expected indent for given line tokens
		*/
		function processExpectedIndent(lineTokens) {
			const firstToken = lineTokens.shift();
			let token = firstToken;
			const expectedIndent = getExpectedIndent(token);
			let lineExpectedIndent = expectedIndent;
			if (lineExpectedIndent == null) while ((token = lineTokens.shift()) != null) {
				lineExpectedIndent = getExpectedIndent(token);
				if (lineExpectedIndent != null) break;
			}
			if (expectedIndent != null) while ((token = lineTokens.shift()) != null) {
				const offset = offsets.get(token);
				if (offset) offset.expectedIndent = expectedIndent;
			}
			const { line, column } = firstToken.loc.start;
			return {
				expectedIndent: lineExpectedIndent,
				actualIndent: sourceCode.lines[line - 1].slice(0, column),
				firstToken,
				line
			};
		}
		/**
		* Get the expected indent from given token
		*/
		function getExpectedIndent(token) {
			const offset = offsets.get(token);
			if (!offset) return null;
			if (offset.expectedIndent != null) return offset.expectedIndent;
			if (offset.baseToken == null) return null;
			const baseIndent = getExpectedIndent(offset.baseToken);
			if (baseIndent == null) return null;
			const offsetIndent = offset.offset;
			return offset.expectedIndent = baseIndent + getIndentText(offsetIndent);
		}
		/**
		* Calculates the indent for lines with missing indent information.
		*/
		function processMissingLines(lineIndents) {
			const results = [];
			const commentLines = [];
			for (const lineIndent of lineIndents) {
				if (!lineIndent) continue;
				const line = lineIndent.line;
				if (isCommentToken(lineIndent.firstToken)) {
					const last = commentLines[commentLines.length - 1];
					if (last && last.range[1] === line - 1) {
						last.range[1] = line;
						last.commentLineIndents.push(lineIndent);
					} else commentLines.push({
						range: [line, line],
						commentLineIndents: [lineIndent]
					});
				} else if (lineIndent.expectedIndent != null) {
					const indent = {
						line,
						expectedIndent: lineIndent.expectedIndent,
						actualIndent: lineIndent.actualIndent,
						firstToken: lineIndent.firstToken
					};
					if (!results[line]) results[line] = indent;
				}
			}
			processComments(commentLines);
			return results;
			/**
			* Process comments.
			*/
			function processComments(commentLines) {
				for (const { range, commentLineIndents } of commentLines) {
					const prev = results.slice(0, range[0]).filter((data) => data).pop();
					const next = results.slice(range[1] + 1).filter((data) => data).shift();
					const expectedIndents = [];
					let either;
					if (prev && next) {
						expectedIndents.unshift(next.expectedIndent);
						if (next.expectedIndent < prev.expectedIndent) {
							let indent = next.expectedIndent + getIndentText(1);
							while (indent.length <= prev.expectedIndent.length) {
								expectedIndents.unshift(indent);
								indent += getIndentText(1);
							}
						}
					} else if (either = prev || next) {
						expectedIndents.unshift(either.expectedIndent);
						if (!next) {
							let indent = outdent(either.expectedIndent);
							while (indent.length > 0) {
								expectedIndents.push(indent);
								indent = outdent(indent);
								if (indent.length <= 0) {
									expectedIndents.push(indent);
									break;
								}
							}
						}
					}
					if (!expectedIndents.length) continue;
					let expectedIndent = expectedIndents[0];
					for (const commentLineIndent of commentLineIndents) {
						if (results[commentLineIndent.line]) continue;
						const indentCandidate = expectedIndents.find((indent, index) => {
							if (indent.length <= commentLineIndent.actualIndent.length) return true;
							return (expectedIndents[index + 1]?.length ?? -1) < commentLineIndent.actualIndent.length && commentLineIndent.actualIndent.length < indent.length;
						});
						if (indentCandidate != null && indentCandidate.length < expectedIndent.length) expectedIndent = indentCandidate;
						results[commentLineIndent.line] = {
							line: commentLineIndent.line,
							expectedIndent,
							actualIndent: commentLineIndent.actualIndent,
							firstToken: commentLineIndent.firstToken
						};
					}
				}
			}
		}
		/**
		* Validate lines
		*/
		function validateLines(lineIndents) {
			for (const lineIndent of lineIndents) {
				if (!lineIndent) continue;
				if (lineIndent.actualIndent !== lineIndent.expectedIndent) {
					const startLoc = {
						line: lineIndent.line,
						column: 0
					};
					context.report({
						loc: {
							start: startLoc,
							end: lineIndent.firstToken.loc.start
						},
						messageId: "wrongIndentation",
						data: getIndentData(lineIndent),
						fix(fixer) {
							return fixer.replaceTextRange([sourceCode.getIndexFromLoc(startLoc), lineIndent.firstToken.range[0]], lineIndent.expectedIndent);
						}
					});
				}
			}
		}
		/**
		* Gets the indentation information to display.
		*/
		function getIndentData(lineIndent) {
			return {
				expected: toDisplayText(lineIndent.expectedIndent),
				actual: toDisplayText(lineIndent.actualIndent)
			};
			/**
			* indentation to display text
			*/
			function toDisplayText(indent) {
				if (indent.length === 0) return "0 spaces";
				const char = indent[0];
				if (char === " " || char === "	") {
					let uni = true;
					for (const c of indent) if (c !== char) uni = false;
					if (uni) {
						const unit = char === " " ? "spaces" : "tabs";
						return `${indent.length} ${unit}`;
					}
				}
				return `"${replaceToDisplay(indent)}"`;
			}
			/**
			* Space replacer
			*/
			function replaceToDisplay(indent) {
				return indent.replace(/\s/gu, (c) => {
					if (c === "	") return "\\t";
					if (c === " ") return " ";
					return `\\u${`000${c.codePointAt(0).toString(16)}`.slice(-4)}`;
				});
			}
		}
	}
});

//#endregion
//#region src/rules/inline-table-curly-newline.ts
var inline_table_curly_newline_default = createRule("inline-table-curly-newline", {
	meta: {
		docs: {
			description: "enforce linebreaks after opening and before closing braces",
			categories: ["standard"],
			extensionRule: "object-curly-newline"
		},
		type: "layout",
		fixable: "whitespace",
		schema: [{ oneOf: [{
			type: "string",
			enum: ["always", "never"]
		}, {
			type: "object",
			properties: {
				multiline: { type: "boolean" },
				minProperties: {
					type: "integer",
					minimum: 0
				},
				consistent: { type: "boolean" }
			},
			additionalProperties: false,
			minProperties: 1
		}] }],
		messages: {
			unexpectedLinebreakBeforeClosingBrace: "Unexpected line break before this closing brace.",
			unexpectedLinebreakAfterOpeningBrace: "Unexpected line break after this opening brace.",
			expectedLinebreakBeforeClosingBrace: "Expected a line break before this closing brace.",
			expectedLinebreakAfterOpeningBrace: "Expected a line break after this opening brace."
		}
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		if (context.languageOptions.parserOptions?.tomlVersion) {
			const tomlVersion = context.languageOptions.parserOptions.tomlVersion.includes(".") && context.languageOptions.parserOptions.tomlVersion.split(".");
			if (tomlVersion && tomlVersion[0] === "1" && tomlVersion[1] === "0") return {};
		}
		/**
		* Normalizes a given option.
		* @param value An option value to parse.
		* @returns Normalized option object.
		*/
		function normalizeOptions(value) {
			let multiline = false;
			let minProperties = Number.POSITIVE_INFINITY;
			let consistent = false;
			if (value) if (value === "always") minProperties = 0;
			else if (value === "never") minProperties = Number.POSITIVE_INFINITY;
			else {
				multiline = Boolean(value.multiline);
				minProperties = value.minProperties || Number.POSITIVE_INFINITY;
				consistent = Boolean(value.consistent);
			}
			else consistent = true;
			return {
				multiline,
				minProperties,
				consistent
			};
		}
		const options = normalizeOptions(context.options[0]);
		/**
		* Determines if ObjectExpression, ObjectPattern, ImportDeclaration, ExportNamedDeclaration, TSTypeLiteral or TSInterfaceBody
		* node needs to be checked for missing line breaks
		* @param node Node under inspection
		* @param options option specific to node type
		* @param first First object property
		* @param last Last object property
		* @returns `true` if node needs to be checked for missing line breaks
		*/
		function areLineBreaksRequired(node, options, first, last) {
			const objectProperties = node.body;
			return objectProperties.length >= options.minProperties || options.multiline && objectProperties.length > 0 && !isTokenOnSameLine(last, first);
		}
		/**
		* Reports a given node if it violated this rule.
		* @param node A node to check. This is an ObjectExpression, ObjectPattern, ImportDeclaration, ExportNamedDeclaration, TSTypeLiteral or TSInterfaceBody node.
		*/
		function check(node) {
			const openBrace = sourceCode.getFirstToken(node, (token) => token.value === "{");
			const closeBrace = sourceCode.getLastToken(node, (token) => token.value === "}");
			const firstTokenOrComment = sourceCode.getTokenAfter(openBrace, { includeComments: true });
			const lastTokenOrComment = sourceCode.getTokenBefore(closeBrace, { includeComments: true });
			const needsLineBreaks = areLineBreaksRequired(node, options, firstTokenOrComment, lastTokenOrComment);
			const hasCommentsFirstToken = isCommentToken(firstTokenOrComment);
			const hasCommentsLastToken = isCommentToken(lastTokenOrComment);
			/**
			* Use tokens or comments to check multiline or not.
			* But use only tokens to check whether line breaks are needed.
			* This allows:
			*     obj = { # eslint-disable-line foo
			*         a: 1
			*     }
			*/
			const firstToken = sourceCode.getTokenAfter(openBrace);
			const lastToken = sourceCode.getTokenBefore(closeBrace);
			if (needsLineBreaks) {
				if (isTokenOnSameLine(openBrace, firstToken)) context.report({
					messageId: "expectedLinebreakAfterOpeningBrace",
					node,
					loc: openBrace.loc,
					fix(fixer) {
						if (hasCommentsFirstToken) return null;
						return fixer.insertTextAfter(openBrace, "\n");
					}
				});
				if (isTokenOnSameLine(lastToken, closeBrace)) context.report({
					messageId: "expectedLinebreakBeforeClosingBrace",
					node,
					loc: closeBrace.loc,
					fix(fixer) {
						if (hasCommentsLastToken) return null;
						return fixer.insertTextBefore(closeBrace, "\n");
					}
				});
			} else {
				const consistent = options.consistent;
				const hasLineBreakBetweenOpenBraceAndFirst = !isTokenOnSameLine(openBrace, firstToken);
				const hasLineBreakBetweenCloseBraceAndLast = !isTokenOnSameLine(lastToken, closeBrace);
				if (!consistent && hasLineBreakBetweenOpenBraceAndFirst || consistent && hasLineBreakBetweenOpenBraceAndFirst && !hasLineBreakBetweenCloseBraceAndLast) context.report({
					messageId: "unexpectedLinebreakAfterOpeningBrace",
					node,
					loc: openBrace.loc,
					fix(fixer) {
						if (hasCommentsFirstToken) return null;
						return fixer.removeRange([openBrace.range[1], firstToken.range[0]]);
					}
				});
				if (!consistent && hasLineBreakBetweenCloseBraceAndLast || consistent && !hasLineBreakBetweenOpenBraceAndFirst && hasLineBreakBetweenCloseBraceAndLast) context.report({
					messageId: "unexpectedLinebreakBeforeClosingBrace",
					node,
					loc: closeBrace.loc,
					fix(fixer) {
						if (hasCommentsLastToken) return null;
						return fixer.removeRange([lastToken.range[1], closeBrace.range[0]]);
					}
				});
			}
		}
		return { TOMLInlineTable: check };
	}
});

//#endregion
//#region src/rules/inline-table-curly-spacing.ts
/**
* Parses the options for this rule and returns an object containing the spacing option,
* the emptyObjects option, and two functions to determine if there should be spaces
* after the opening curly brace and before the closing curly brace, based on the options and the surrounding tokens.
* @param options The options passed to the rule.
* @param sourceCode The source code object, used to get nodes by range index.
* @returns An object containing the spacing option, the emptyObjects option, and two functions to determine if there should be spaces after the opening curly brace and before the closing curly brace.
*/
function parseOptions$1(options, sourceCode) {
	const spaced = options[0] ?? "always";
	/**
	* Determines whether an option is set, relative to the spacing option.
	* If spaced is "always", then check whether option is set to false.
	* If spaced is "never", then check whether option is set to true.
	* @param option The option to exclude.
	* @returns Whether or not the property is excluded.
	*/
	function isOptionSet(option) {
		return options[1] ? options[1][option] === (spaced === "never") : false;
	}
	const arraysInObjectsException = isOptionSet("arraysInObjects");
	const objectsInObjectsException = isOptionSet("objectsInObjects");
	const emptyObjects = options[1]?.emptyObjects ?? "ignore";
	/**
	* Determines if there should be a space after the opening curly brace,
	* based on the spacing option and the second token.
	* @param spaced The spacing option ("always" or "never").
	* @param second The second token after the opening curly brace.
	* @returns Whether or not there should be a space after the opening curly brace.
	*/
	function isOpeningCurlyBraceMustBeSpaced(spaced, _second) {
		return spaced === "always";
	}
	/**
	* Determines if there should be a space before the closing curly brace,
	* based on the spacing option and the penultimate token.
	* @param spaced The spacing option ("always" or "never").
	* @param penultimate The penultimate token before the closing curly brace.
	* @returns Whether or not there should be a space before the closing curly brace.
	*/
	function isClosingCurlyBraceMustBeSpaced(spaced, penultimate) {
		const targetPenultimateType = arraysInObjectsException && isClosingBracketToken(penultimate) ? "TOMLArray" : objectsInObjectsException && isClosingBraceToken(penultimate) ? "TOMLInlineTable" : null;
		const node = sourceCode.getNodeByRangeIndex(penultimate.range[0]);
		return targetPenultimateType && node?.type === targetPenultimateType ? spaced === "never" : spaced === "always";
	}
	return {
		spaced,
		emptyObjects,
		isOpeningCurlyBraceMustBeSpaced,
		isClosingCurlyBraceMustBeSpaced
	};
}
var inline_table_curly_spacing_default = createRule("inline-table-curly-spacing", {
	meta: {
		docs: {
			description: "enforce consistent spacing inside braces",
			categories: ["standard"],
			extensionRule: "object-curly-spacing"
		},
		type: "layout",
		fixable: "whitespace",
		schema: [{
			type: "string",
			enum: ["always", "never"]
		}, {
			type: "object",
			properties: {
				arraysInObjects: { type: "boolean" },
				objectsInObjects: { type: "boolean" },
				emptyObjects: {
					type: "string",
					enum: [
						"ignore",
						"always",
						"never"
					]
				}
			},
			additionalProperties: false
		}],
		messages: {
			requireSpaceBefore: "A space is required before '{{token}}'.",
			requireSpaceAfter: "A space is required after '{{token}}'.",
			unexpectedSpaceBefore: "There should be no space before '{{token}}'.",
			unexpectedSpaceAfter: "There should be no space after '{{token}}'.",
			requiredSpaceInEmptyObject: "A space is required in empty inline table.",
			unexpectedSpaceInEmptyObject: "There should be no space in empty inline table."
		}
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		const options = parseOptions$1(context.options, sourceCode);
		/**
		* Reports that there shouldn't be a space after the first token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportNoBeginningSpace(node, token) {
			const nextToken = sourceCode.getTokenAfter(token, { includeComments: true });
			context.report({
				node,
				loc: {
					start: token.loc.end,
					end: nextToken.loc.start
				},
				messageId: "unexpectedSpaceAfter",
				data: { token: token.value },
				fix(fixer) {
					return fixer.removeRange([token.range[1], nextToken.range[0]]);
				}
			});
		}
		/**
		* Reports that there shouldn't be a space before the last token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportNoEndingSpace(node, token) {
			const previousToken = sourceCode.getTokenBefore(token, { includeComments: true });
			context.report({
				node,
				loc: {
					start: previousToken.loc.end,
					end: token.loc.start
				},
				messageId: "unexpectedSpaceBefore",
				data: { token: token.value },
				fix(fixer) {
					return fixer.removeRange([previousToken.range[1], token.range[0]]);
				}
			});
		}
		/**
		* Reports that there should be a space after the first token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportRequiredBeginningSpace(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "requireSpaceAfter",
				data: { token: token.value },
				fix(fixer) {
					return fixer.insertTextAfter(token, " ");
				}
			});
		}
		/**
		* Reports that there should be a space before the last token
		* @param node The node to report in the event of an error.
		* @param token The token to use for the report.
		*/
		function reportRequiredEndingSpace(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "requireSpaceBefore",
				data: { token: token.value },
				fix(fixer) {
					return fixer.insertTextBefore(token, " ");
				}
			});
		}
		/**
		* Determines if spacing in curly braces is valid.
		* @param node The AST node to check.
		* @param first The first token to check (should be the opening brace)
		* @param second The second token to check (should be first after the opening brace)
		* @param penultimate The penultimate token to check (should be last before closing brace)
		* @param last The last token to check (should be closing brace)
		*/
		function validateBraceSpacing(node, spaced, openingToken, second, penultimate, closingToken) {
			if (isTokenOnSameLine(openingToken, second)) {
				const firstSpaced = sourceCode.isSpaceBetween(openingToken, second);
				if (options.isOpeningCurlyBraceMustBeSpaced(spaced, second)) {
					if (!firstSpaced) reportRequiredBeginningSpace(node, openingToken);
				} else if (firstSpaced && second.type !== "Block") reportNoBeginningSpace(node, openingToken);
			}
			if (isTokenOnSameLine(penultimate, closingToken)) {
				const lastSpaced = sourceCode.isSpaceBetween(penultimate, closingToken);
				if (options.isClosingCurlyBraceMustBeSpaced(spaced, penultimate)) {
					if (!lastSpaced) reportRequiredEndingSpace(node, closingToken);
				} else if (lastSpaced) reportNoEndingSpace(node, closingToken);
			}
		}
		/**
		* Checks spacing in empty objects. Depending on the options, reports
		* if there is an unexpected space or if there is no space when there should be.
		* @param node The node to check.
		*/
		function checkSpaceInEmptyObject(node) {
			if (options.emptyObjects === "ignore") return;
			const openingToken = sourceCode.getFirstToken(node);
			const closingToken = sourceCode.getLastToken(node);
			const second = sourceCode.getTokenAfter(openingToken, { includeComments: true });
			if (second !== closingToken && isCommentToken(second)) {
				const penultimate = sourceCode.getTokenBefore(closingToken, { includeComments: true });
				validateBraceSpacing(node, options.emptyObjects, openingToken, second, penultimate, closingToken);
				return;
			}
			if (!isTokenOnSameLine(openingToken, closingToken)) return;
			const sourceBetween = sourceCode.text.slice(openingToken.range[1], closingToken.range[0]);
			if (sourceBetween.trim() !== "") return;
			if (options.emptyObjects === "always") {
				if (sourceBetween) return;
				context.report({
					node,
					loc: {
						start: openingToken.loc.end,
						end: closingToken.loc.start
					},
					messageId: "requiredSpaceInEmptyObject",
					fix(fixer) {
						return fixer.replaceTextRange([openingToken.range[1], closingToken.range[0]], " ");
					}
				});
			} else if (options.emptyObjects === "never") {
				if (!sourceBetween) return;
				context.report({
					node,
					loc: {
						start: openingToken.loc.end,
						end: closingToken.loc.start
					},
					messageId: "unexpectedSpaceInEmptyObject",
					fix(fixer) {
						return fixer.removeRange([openingToken.range[1], closingToken.range[0]]);
					}
				});
			}
		}
		/**
		* Gets '}' token of an object node.
		*
		* Because the last token of object patterns might be a type annotation,
		* this traverses tokens preceded by the last property, then returns the
		* first '}' token.
		* @param node The node to get. This node is an
		*      ObjectExpression or an ObjectPattern. And this node has one or
		*      more properties.
		* @returns '}' token.
		*/
		function getClosingBraceOfObject(node) {
			const lastProperty = node.body[node.body.length - 1];
			return sourceCode.getTokenAfter(lastProperty, isClosingBraceToken);
		}
		/**
		* Reports a given object node if spacing in curly braces is invalid.
		* @param node An ObjectExpression or ObjectPattern node to check.
		*/
		function checkForObject(node) {
			if (node.body.length === 0) {
				checkSpaceInEmptyObject(node);
				return;
			}
			const openingToken = sourceCode.getFirstToken(node);
			const closingToken = getClosingBraceOfObject(node);
			const second = sourceCode.getTokenAfter(openingToken, { includeComments: true });
			const penultimate = sourceCode.getTokenBefore(closingToken, { includeComments: true });
			validateBraceSpacing(node, options.spaced, openingToken, second, penultimate, closingToken);
		}
		return { TOMLInlineTable: checkForObject };
	}
});

//#endregion
//#region src/rules/inline-table-key-value-newline.ts
/**
* Parses the options for the rule and returns a normalized options object.
* @param options The raw options provided to the rule.
* @returns A normalized options object with default values applied.
*/
function parseOptions(options) {
	return { allowAllPropertiesOnSameLine: options?.allowAllPropertiesOnSameLine ?? true };
}
var inline_table_key_value_newline_default = createRule("inline-table-key-value-newline", {
	meta: {
		docs: {
			description: "enforce placing inline table key-value pairs on separate lines",
			categories: ["standard"],
			extensionRule: "object-property-newline"
		},
		fixable: "whitespace",
		hasSuggestions: false,
		schema: [{
			type: "object",
			properties: { allowAllPropertiesOnSameLine: { type: "boolean" } },
			additionalProperties: false
		}],
		messages: {
			propertiesOnNewlineAll: "Inline table key-value pairs must go on a new line if they aren't all on the same line.",
			propertiesOnNewline: "Inline table key-value pairs must go on a new line."
		},
		type: "layout"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		if (context.languageOptions.parserOptions?.tomlVersion) {
			const tomlVersion = context.languageOptions.parserOptions.tomlVersion.includes(".") && context.languageOptions.parserOptions.tomlVersion.split(".");
			if (tomlVersion && tomlVersion[0] === "1" && tomlVersion[1] === "0") return {};
		}
		const allowSameLine = parseOptions(context.options[0]).allowAllPropertiesOnSameLine;
		const messageId = allowSameLine ? "propertiesOnNewlineAll" : "propertiesOnNewline";
		/**
		* Checks whether the given tokens are on the same line.
		* @param token1 The first token to compare.
		* @param token2 The second token to compare.
		* @returns True if the tokens are on the same line, false otherwise.
		*/
		function check(node, children) {
			if (allowSameLine) {
				if (children.length > 1) {
					if (isTokenOnSameLine(sourceCode.getFirstToken(children[0]), sourceCode.getLastToken(children[children.length - 1]))) return;
				}
			}
			for (let i = 1; i < children.length; i++) {
				const lastTokenOfPreviousProperty = sourceCode.getLastToken(children[i - 1]);
				const firstTokenOfCurrentProperty = sourceCode.getFirstToken(children[i]);
				if (isTokenOnSameLine(lastTokenOfPreviousProperty, firstTokenOfCurrentProperty)) context.report({
					node,
					loc: firstTokenOfCurrentProperty.loc,
					messageId,
					fix(fixer) {
						const rangeAfterComma = [sourceCode.getTokenBefore(firstTokenOfCurrentProperty).range[1], firstTokenOfCurrentProperty.range[0]];
						if (sourceCode.text.slice(rangeAfterComma[0], rangeAfterComma[1]).trim()) return null;
						return fixer.replaceTextRange(rangeAfterComma, "\n");
					}
				});
			}
		}
		return { TOMLInlineTable(node) {
			check(node, node.body);
		} };
	}
});

//#endregion
//#region src/rules/key-spacing.ts
/**
* Checks whether a string contains a line terminator as defined in
* http://www.ecma-international.org/ecma-262/5.1/#sec-7.3
* @param {string} str String to test.
* @returns {boolean} True if str contains a line terminator.
*/
function containsLineTerminator(str) {
	return /[\n\r\u2028\u2029]/u.test(str);
}
/**
* Gets the last element of an array.
* @param {Array} arr An array.
* @returns {any} Last element of arr.
*/
function last(arr) {
	return arr[arr.length - 1];
}
/**
* Checks whether a node is contained on a single line.
* @param {ASTNode} node AST Node being evaluated.
* @returns {boolean} True if the node is a single line.
*/
function isSingleLine(node) {
	return node.loc.end.line === node.loc.start.line;
}
/**
* Checks whether the properties on a single line.
* @param {ASTNode[]} properties List of Property AST nodes.
* @returns {boolean} True if all properties is on a single line.
*/
function isSingleLineProperties(properties) {
	const [firstProp] = properties;
	const lastProp = last(properties);
	return firstProp.loc.start.line === lastProp.loc.end.line;
}
/**
* Initializes a single option property from the configuration with defaults for undefined values
* @param {Object} fromOptions Object to be initialized from
* @returns {Object} The object with correctly initialized options and values
*/
function initOptionProperty(fromOptions) {
	const mode = fromOptions.mode || "strict";
	let beforeEqual, afterEqual;
	if (typeof fromOptions.beforeEqual !== "undefined") beforeEqual = fromOptions.beforeEqual;
	else beforeEqual = true;
	if (typeof fromOptions.afterEqual !== "undefined") afterEqual = fromOptions.afterEqual;
	else afterEqual = true;
	let align = void 0;
	if (typeof fromOptions.align !== "undefined") if (typeof fromOptions.align === "object") align = fromOptions.align;
	else align = {
		on: fromOptions.align,
		mode,
		beforeEqual,
		afterEqual
	};
	return {
		mode,
		beforeEqual,
		afterEqual,
		align
	};
}
/**
* Initializes all the option values (singleLine, multiLine and align) from the configuration with defaults for undefined values
* @param {Object} fromOptions Object to be initialized from
* @returns {Object} The object with correctly initialized options and values
*/
function initOptions(fromOptions) {
	let align, multiLine, singleLine;
	if (typeof fromOptions.align === "object") {
		align = {
			...initOptionProperty(fromOptions.align),
			on: fromOptions.align.on || "equal",
			mode: fromOptions.align.mode || "strict"
		};
		multiLine = initOptionProperty(fromOptions.multiLine || fromOptions);
		singleLine = initOptionProperty(fromOptions.singleLine || fromOptions);
	} else {
		multiLine = initOptionProperty(fromOptions.multiLine || fromOptions);
		singleLine = initOptionProperty(fromOptions.singleLine || fromOptions);
		if (multiLine.align) align = {
			on: multiLine.align.on,
			mode: multiLine.align.mode || multiLine.mode,
			beforeEqual: multiLine.align.beforeEqual,
			afterEqual: multiLine.align.afterEqual
		};
	}
	return {
		align,
		multiLine,
		singleLine
	};
}
const ON_SCHEMA = { enum: ["equal", "value"] };
const OBJECT_WITHOUT_ON_SCHEMA = {
	type: "object",
	properties: {
		mode: { enum: ["strict", "minimum"] },
		beforeEqual: { type: "boolean" },
		afterEqual: { type: "boolean" }
	},
	additionalProperties: false
};
const ALIGN_OBJECT_SCHEMA = {
	type: "object",
	properties: {
		on: ON_SCHEMA,
		...OBJECT_WITHOUT_ON_SCHEMA.properties
	},
	additionalProperties: false
};
var key_spacing_default = createRule("key-spacing", {
	meta: {
		docs: {
			description: "enforce consistent spacing between keys and values in key/value pairs",
			categories: ["standard"],
			extensionRule: "key-spacing"
		},
		fixable: "whitespace",
		schema: [{ anyOf: [
			{
				type: "object",
				properties: {
					align: { anyOf: [ON_SCHEMA, ALIGN_OBJECT_SCHEMA] },
					...OBJECT_WITHOUT_ON_SCHEMA.properties
				},
				additionalProperties: false
			},
			{
				type: "object",
				properties: {
					singleLine: OBJECT_WITHOUT_ON_SCHEMA,
					multiLine: {
						type: "object",
						properties: {
							align: { anyOf: [ON_SCHEMA, ALIGN_OBJECT_SCHEMA] },
							...OBJECT_WITHOUT_ON_SCHEMA.properties
						},
						additionalProperties: false
					}
				},
				additionalProperties: false
			},
			{
				type: "object",
				properties: {
					singleLine: OBJECT_WITHOUT_ON_SCHEMA,
					multiLine: OBJECT_WITHOUT_ON_SCHEMA,
					align: ALIGN_OBJECT_SCHEMA
				},
				additionalProperties: false
			}
		] }],
		messages: {
			extraKey: "Extra space after key '{{key}}'.",
			extraValue: "Extra space before value for key '{{key}}'.",
			missingKey: "Missing space after key '{{key}}'.",
			missingValue: "Missing space before value for key '{{key}}'."
		},
		type: "layout"
	},
	create
});
/**
* Create rule visitor
*/
function create(context) {
	const sourceCode = context.sourceCode;
	if (!sourceCode.parserServices?.isTOML) return {};
	const { multiLine: multiLineOptions, singleLine: singleLineOptions, align: alignmentOptions } = initOptions(context.options[0] || {});
	/**
	* Determines if the given property is key/value property.
	* @param {ASTNode} property Property node to check.
	* @returns {boolean} Whether the property is a key/value property.
	*/
	function isKeyValueProperty(property) {
		return property.type === "TOMLKeyValue";
	}
	/**
	* Starting from the given a node (a property.key node here) looks forward
	* until it finds the last token before a equal punctuator and returns it.
	* @param {ASTNode} node The node to start looking from.
	* @returns {ASTNode} The last token before a equal punctuator.
	*/
	function getLastTokenBeforeEqual(node) {
		const equalToken = sourceCode.getTokenAfter(node, isEqualSign);
		return sourceCode.getTokenBefore(equalToken);
	}
	/**
	* Starting from the given a node (a property.key node here) looks forward
	* until it finds the equal punctuator and returns it.
	* @param {ASTNode} node The node to start looking from.
	* @returns {ASTNode} The equal punctuator.
	*/
	function getNextEqual(node) {
		return sourceCode.getTokenAfter(node, isEqualSign);
	}
	/**
	* Gets an object literal property's key as the identifier name or string value.
	* @param {ASTNode} property Property node whose key to retrieve.
	* @returns {string} The property's key.
	*/
	function getKey(property) {
		return property.key.keys.map((key) => sourceCode.getText().slice(key.range[0], key.range[1])).join(".");
	}
	/**
	* Reports an appropriately-formatted error if spacing is incorrect on one
	* side of the equal.
	* @param {ASTNode} property Key-value pair in an object literal.
	* @param {string} side Side being verified - either "key" or "value".
	* @param {string} whitespace Actual whitespace string.
	* @param {int} expected Expected whitespace length.
	* @param {string} mode Value of the mode as "strict" or "minimum"
	* @returns {void}
	*/
	function report(property, side, whitespace, expected, mode) {
		const diff = whitespace.length - expected;
		const nextEqual = getNextEqual(property.key);
		const tokenBeforeEqual = sourceCode.getTokenBefore(nextEqual, { includeComments: true });
		const tokenAfterEqual = sourceCode.getTokenAfter(nextEqual, { includeComments: true });
		if (!((mode === "strict" ? diff !== 0 : diff < 0 || diff > 0 && expected === 0) && !(expected && containsLineTerminator(whitespace)))) return;
		const { locStart, locEnd, missingLoc } = side === "key" ? {
			locStart: tokenBeforeEqual.loc.end,
			locEnd: nextEqual.loc.start,
			missingLoc: tokenBeforeEqual.loc
		} : {
			locStart: nextEqual.loc.start,
			locEnd: tokenAfterEqual.loc.start,
			missingLoc: tokenAfterEqual.loc
		};
		const { loc, messageId } = diff > 0 ? {
			loc: {
				start: locStart,
				end: locEnd
			},
			messageId: side === "key" ? "extraKey" : "extraValue"
		} : {
			loc: missingLoc,
			messageId: side === "key" ? "missingKey" : "missingValue"
		};
		context.report({
			node: property[side],
			loc,
			messageId,
			data: { key: getKey(property) },
			fix(fixer) {
				if (diff > 0) {
					if (side === "key") return fixer.removeRange([tokenBeforeEqual.range[1], tokenBeforeEqual.range[1] + diff]);
					return fixer.removeRange([tokenAfterEqual.range[0] - diff, tokenAfterEqual.range[0]]);
				}
				const spaces = " ".repeat(-diff);
				if (side === "key") return fixer.insertTextAfter(tokenBeforeEqual, spaces);
				return fixer.insertTextBefore(tokenAfterEqual, spaces);
			}
		});
	}
	/**
	* Gets the number of characters in a key, including quotes around string
	* keys and braces around computed property keys.
	* @param {ASTNode} property Property of on object literal.
	* @returns {int} Width of the key.
	*/
	function getKeyWidth(pair) {
		const startToken = sourceCode.getFirstToken(pair);
		return getLastTokenBeforeEqual(pair.key).range[1] - startToken.range[0];
	}
	/**
	* Gets the whitespace around the equal in an object literal property.
	* @param {ASTNode} property Property node from an object literal.
	* @returns {Object} Whitespace before and after the property's equal.
	*/
	function getPropertyWhitespace(pair) {
		const whitespace = /(\s*)=(\s*)/u.exec(sourceCode.getText().slice(pair.key.range[1], pair.value.range[0]));
		if (whitespace) return {
			beforeEqual: whitespace[1],
			afterEqual: whitespace[2]
		};
		return null;
	}
	/**
	* Verifies spacing of property conforms to specified options.
	* @param  {ASTNode} node Property node being evaluated.
	* @param {Object} lineOptions Configured singleLine or multiLine options
	* @returns {void}
	*/
	function verifySpacing(node, lineOptions) {
		const actual = getPropertyWhitespace(node);
		if (actual) {
			report(node, "key", actual.beforeEqual, lineOptions.beforeEqual ? 1 : 0, lineOptions.mode);
			report(node, "value", actual.afterEqual, lineOptions.afterEqual ? 1 : 0, lineOptions.mode);
		}
	}
	/**
	* Verifies spacing of each property in a list.
	* @param {ASTNode[]} properties List of Property AST nodes.
	* @param {Object} lineOptions Configured singleLine or multiLine options
	* @returns {void}
	*/
	function verifyListSpacing(properties, lineOptions) {
		const length = properties.length;
		for (let i = 0; i < length; i++) verifySpacing(properties[i], lineOptions);
	}
	if (alignmentOptions) return defineAlignmentVisitor(alignmentOptions);
	return defineSpacingVisitor();
	/**
	* Define alignment rule visitor
	*/
	function defineAlignmentVisitor(alignmentOptions) {
		return { "TOMLTopLevelTable, TOMLTable, TOMLInlineTable"(node) {
			if (isSingleLine(node)) {
				const body = node.body;
				verifyListSpacing(body.filter(isKeyValueProperty), singleLineOptions);
			} else verifyAlignment(node);
		} };
		/**
		* Verifies correct vertical alignment of a group of properties.
		* @param {ASTNode[]} properties List of Property AST nodes.
		* @returns {void}
		*/
		function verifyGroupAlignment(properties) {
			const length = properties.length;
			const widths = properties.map(getKeyWidth);
			const align = alignmentOptions.on;
			let targetWidth = Math.max(...widths);
			let beforeEqual, afterEqual, mode;
			if (alignmentOptions && length > 1) {
				beforeEqual = alignmentOptions.beforeEqual ? 1 : 0;
				afterEqual = alignmentOptions.afterEqual ? 1 : 0;
				mode = alignmentOptions.mode;
			} else {
				beforeEqual = multiLineOptions.beforeEqual ? 1 : 0;
				afterEqual = multiLineOptions.afterEqual ? 1 : 0;
				mode = alignmentOptions.mode;
			}
			targetWidth += align === "equal" ? beforeEqual : afterEqual;
			for (let i = 0; i < length; i++) {
				const property = properties[i];
				const whitespace = getPropertyWhitespace(property);
				if (whitespace) {
					const width = widths[i];
					if (align === "value") {
						report(property, "key", whitespace.beforeEqual, beforeEqual, mode);
						report(property, "value", whitespace.afterEqual, targetWidth - width, mode);
					} else {
						report(property, "key", whitespace.beforeEqual, targetWidth - width, mode);
						report(property, "value", whitespace.afterEqual, afterEqual, mode);
					}
				}
			}
		}
		/**
		* Checks whether a property is a member of the property group it follows.
		* @param {ASTNode} lastMember The last Property known to be in the group.
		* @param {ASTNode} candidate The next Property that might be in the group.
		* @returns {boolean} True if the candidate property is part of the group.
		*/
		function continuesPropertyGroup(lastMember, candidate) {
			const groupEndLine = lastMember.loc.start.line;
			const candidateStartLine = candidate.loc.start.line;
			if (candidateStartLine - groupEndLine <= 1) return true;
			const leadingComments = sourceCode.getCommentsBefore(candidate);
			if (leadingComments.length && leadingComments[0].loc.start.line - groupEndLine <= 1 && candidateStartLine - last(leadingComments).loc.end.line <= 1) {
				for (let i = 1; i < leadingComments.length; i++) if (leadingComments[i].loc.start.line - leadingComments[i - 1].loc.end.line > 1) return false;
				return true;
			}
			return false;
		}
		/**
		* Creates groups of properties.
		* @param  {ASTNode} node ObjectExpression node being evaluated.
		* @returns {Array.<ASTNode[]>} Groups of property AST node lists.
		*/
		function createGroups(node) {
			const pairs = node.body.filter(isKeyValueProperty);
			if (pairs.length === 1) return [pairs];
			return pairs.reduce((groups, property) => {
				const currentGroup = last(groups);
				const prev = last(currentGroup);
				if (!prev || continuesPropertyGroup(prev, property)) currentGroup.push(property);
				else groups.push([property]);
				return groups;
			}, [[]]);
		}
		/**
		* Verifies vertical alignment, taking into account groups of properties.
		* @param  {ASTNode} node ObjectExpression node being evaluated.
		* @returns {void}
		*/
		function verifyAlignment(node) {
			createGroups(node).forEach((group) => {
				const properties = group;
				if (properties.length > 0 && isSingleLineProperties(properties)) verifyListSpacing(properties, multiLineOptions);
				else verifyGroupAlignment(properties);
			});
		}
	}
	/**
	* Define spacing rule visitor
	*/
	function defineSpacingVisitor() {
		return { TOMLKeyValue(node) {
			if (!isKeyValueProperty(node)) return;
			verifySpacing(node, isSingleLine(node.parent) ? singleLineOptions : multiLineOptions);
		} };
	}
}

//#endregion
//#region src/rules/keys-order.ts
var keys_order_default = createRule("keys-order", {
	meta: {
		docs: {
			description: "disallow defining pair keys out-of-order",
			categories: ["standard"],
			extensionRule: false
		},
		fixable: "code",
		schema: [],
		messages: { outOfOrder: "'{{target}}' must be next to '{{before}}'." },
		type: "suggestion"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		/**
		* Apply new key
		*/
		function applyKey(tableKeys, node) {
			const keyNames = getStaticTOMLValue(node.key);
			let before = null;
			let keys = tableKeys;
			while (keyNames.length) {
				const key = keyNames.shift();
				let next = keys.find((e) => e.key === key);
				if (!next) {
					next = {
						key,
						node,
						keys: []
					};
					before = keys[keys.length - 1]?.node || null;
					keys.push(next);
				} else next.node = node;
				keys = next.keys;
			}
			return before;
		}
		/**
		* Verify table
		*/
		function verify(node) {
			const keys = [];
			let prev = null;
			for (const body of node.body) {
				if (body.type !== "TOMLKeyValue") continue;
				const before = applyKey(keys, body);
				if (before && before !== prev) context.report({
					node: body.key,
					messageId: "outOfOrder",
					data: {
						target: getStaticTOMLValue(body.key).join("."),
						before: getStaticTOMLValue(before.key).join(".")
					},
					fix(fixer) {
						const startToken = sourceCode.getTokenBefore(body);
						const start = node.type === "TOMLInlineTable" ? startToken.range[0] : startToken.range[1];
						const code = sourceCode.text.slice(start, body.range[1]);
						return [fixer.insertTextAfter(before, node.type === "TOMLInlineTable" ? code : `\n${code.trim()}`), fixer.removeRange([start, body.range[1]])];
					}
				});
				prev = body;
			}
		}
		return {
			TOMLTopLevelTable: verify,
			TOMLTable: verify,
			TOMLInlineTable: verify
		};
	}
});

//#endregion
//#region src/rules/no-mixed-type-in-array.ts
var no_mixed_type_in_array_default = createRule("no-mixed-type-in-array", {
	meta: {
		docs: {
			description: "disallow mixed data types in array",
			categories: null,
			extensionRule: false
		},
		schema: [{
			type: "object",
			properties: { typeMap: {
				type: "object",
				properties: {
					string: { type: "string" },
					boolean: { type: "string" },
					integer: { type: "string" },
					float: { type: "string" },
					offsetDateTime: { type: "string" },
					localDateTime: { type: "string" },
					localDate: { type: "string" },
					localTime: { type: "string" },
					array: { type: "string" },
					inlineTable: { type: "string" }
				},
				additionalProperties: false
			} },
			additionalProperties: false
		}],
		messages: { mixedDataType: "Data types may not be mixed in an array." },
		type: "suggestion"
	},
	create(context) {
		if (!context.sourceCode.parserServices?.isTOML) return {};
		const typeMap = {
			string: "String",
			integer: "Integer",
			float: "Float",
			boolean: "Boolean",
			offsetDateTime: "Datetime",
			localDateTime: "Datetime",
			localDate: "Datetime",
			localTime: "Datetime",
			array: "Array",
			inlineTable: "Inline Table",
			...context.options[0]?.typeMap || {}
		};
		/**
		* Get data type from given node
		*/
		function getDataType(node) {
			if (node.type === "TOMLArray") return "array";
			if (node.type === "TOMLInlineTable") return "inlineTable";
			if (node.type === "TOMLValue") {
				if (node.kind === "string" || node.kind === "integer" || node.kind === "float" || node.kind === "boolean") return node.kind;
				if (node.kind === "offset-date-time") return "offsetDateTime";
				if (node.kind === "local-date-time") return "localDateTime";
				if (node.kind === "local-date") return "localDate";
				if (node.kind === "local-time") return "localTime";
			}
			return null;
		}
		return { TOMLArray(node) {
			let typeName = null;
			for (const element of node.elements) {
				const type = getDataType(element);
				if (typeName == null) {
					if (type != null) typeName = typeMap[type];
				} else if (type == null || typeName !== typeMap[type]) context.report({
					node: element,
					messageId: "mixedDataType"
				});
			}
		} };
	}
});

//#endregion
//#region src/rules/no-non-decimal-integer.ts
/**
* Convert the given string to decimal string
*/
function toDecimalText(str, radix) {
	const digits = [0];
	for (const c of str) {
		let num = parseInt(c, radix);
		for (let place = 0; place < digits.length; place++) {
			num = digits[place] * radix + num;
			digits[place] = num % 10;
			num = Math.floor(num / 10);
		}
		while (num > 0) {
			digits.push(num % 10);
			num = Math.floor(num / 10);
		}
	}
	return digits.reverse().join("");
}
var no_non_decimal_integer_default = createRule("no-non-decimal-integer", {
	meta: {
		docs: {
			description: "disallow hexadecimal, octal and binary integer",
			categories: null,
			extensionRule: false
		},
		fixable: "code",
		schema: [{
			type: "object",
			properties: {
				allowHexadecimal: { type: "boolean" },
				allowOctal: { type: "boolean" },
				allowBinary: { type: "boolean" }
			},
			additionalProperties: false
		}],
		messages: {
			disallowHex: "Hexadecimal integers are forbidden.",
			disallowOctal: "Octal integers are forbidden.",
			disallowBinary: "Binary integers are forbidden."
		},
		type: "suggestion"
	},
	create(context) {
		if (!context.sourceCode.parserServices?.isTOML) return {};
		const allowHexadecimal = Boolean(context.options[0]?.allowHexadecimal);
		const allowOctal = Boolean(context.options[0]?.allowOctal);
		const allowBinary = Boolean(context.options[0]?.allowBinary);
		/**
		* Build fixer
		*/
		function buildFixer(node, text, mark) {
			if (allowHexadecimal || allowOctal || allowBinary) return;
			return (fixer) => {
				const d = mark === "x" ? 16 : mark === "o" ? 8 : 2;
				const decimalText = toDecimalText(text.slice(2), d);
				return fixer.replaceText(node, decimalText);
			};
		}
		/**
		* Verify number value node text
		*/
		function verifyText(node) {
			const text = node.number;
			if (text.startsWith("0")) {
				const maybeMark = text[1];
				if (maybeMark === "x" && !allowHexadecimal) context.report({
					node,
					messageId: "disallowHex",
					fix: buildFixer(node, text, maybeMark)
				});
				else if (maybeMark === "o" && !allowOctal) context.report({
					node,
					messageId: "disallowOctal",
					fix: buildFixer(node, text, maybeMark)
				});
				else if (maybeMark === "b" && !allowBinary) context.report({
					node,
					messageId: "disallowBinary",
					fix: buildFixer(node, text, maybeMark)
				});
			}
		}
		return { TOMLValue(node) {
			if (node.kind === "integer") verifyText(node);
		} };
	}
});

//#endregion
//#region src/rules/no-space-dots.ts
var no_space_dots_default = createRule("no-space-dots", {
	meta: {
		docs: {
			description: "disallow spacing around infix operators",
			categories: ["standard"],
			extensionRule: false
		},
		fixable: "whitespace",
		schema: [],
		messages: {
			unexpectedBefore: "Unexpected whitespace before dot.",
			unexpectedAfter: "Unexpected whitespace after dot."
		},
		type: "layout"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		return { TOMLKey(node) {
			for (let index = 0; index < node.keys.length - 1; index++) {
				const prev = node.keys[index];
				const next = node.keys[index + 1];
				const comma = sourceCode.getTokenAfter(prev);
				if (prev.range[1] < comma.range[0]) context.report({
					loc: {
						start: prev.loc.end,
						end: comma.loc.start
					},
					messageId: "unexpectedBefore",
					fix(fixer) {
						return fixer.removeRange([prev.range[1], comma.range[0]]);
					}
				});
				if (comma.range[1] < next.range[0]) context.report({
					loc: {
						start: comma.loc.end,
						end: next.loc.start
					},
					messageId: "unexpectedAfter",
					fix(fixer) {
						return fixer.removeRange([comma.range[1], next.range[0]]);
					}
				});
			}
		} };
	}
});

//#endregion
//#region src/rules/no-unreadable-number-separator.ts
var no_unreadable_number_separator_default = createRule("no-unreadable-number-separator", {
	meta: {
		docs: {
			description: "disallow number separators that to not enhance readability.",
			categories: ["recommended", "standard"],
			extensionRule: false
		},
		schema: [],
		messages: { unexpected: "Unexpected number separator that does not enhance readability." },
		type: "suggestion"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		/**
		* Parse character counts
		*/
		function parseCharCounts(text, startIndex, exitChars) {
			let start = startIndex;
			let count = 0;
			const charCounts = [];
			let index = startIndex;
			for (; index < text.length; index++) {
				const char = text[index];
				if (exitChars.includes(char)) break;
				if (char === "_") {
					charCounts.push({
						count,
						range: [start, index]
					});
					start = index + 1;
					count = 0;
				} else count++;
			}
			charCounts.push({
				count,
				range: [start, index]
			});
			return {
				charCounts,
				index
			};
		}
		/**
		* Verify parts of number
		*/
		function verifyParts(node, { integerCharCounts, fractionalCharCounts, exponentCharCounts }) {
			for (const { count, range } of [
				...integerCharCounts.slice(1),
				...fractionalCharCounts.slice(0, -1),
				...exponentCharCounts.slice(1)
			]) if (count === 1) context.report({
				loc: {
					start: sourceCode.getLocFromIndex(node.range[0] + range[0]),
					end: sourceCode.getLocFromIndex(node.range[0] + range[1])
				},
				messageId: "unexpected"
			});
		}
		/**
		* Verify number value node text
		*/
		function verifyText(node, text) {
			let index = 0;
			if (text[index] === "+" || text[index] === "-") index++;
			let integer = false;
			if (text[index] === "0") {
				const maybeMark = text[index + 1];
				if (maybeMark === "x" || maybeMark === "o" || maybeMark === "b") {
					index += 2;
					integer = true;
				}
			}
			const parsedInteger = parseCharCounts(text, index, integer ? [] : [
				".",
				"E",
				"e"
			]);
			const integerCharCounts = parsedInteger.charCounts;
			const fractionalCharCounts = [];
			const exponentCharCounts = [];
			if (!integer) {
				index = parsedInteger.index;
				if (text[index] === ".") {
					index++;
					const parsedFractional = parseCharCounts(text, index, ["E", "e"]);
					fractionalCharCounts.push(...parsedFractional.charCounts);
					index = parsedFractional.index;
				}
				if (text[index] === "e" || text[index] === "E") {
					index++;
					if (text[index] === "+" || text[index] === "-") index++;
					const parsedExponent = parseCharCounts(text, index, []);
					exponentCharCounts.push(...parsedExponent.charCounts);
				}
			}
			verifyParts(node, {
				integerCharCounts,
				fractionalCharCounts,
				exponentCharCounts
			});
		}
		return { TOMLValue(node) {
			if (node.kind === "integer" || node.kind === "float") {
				const text = sourceCode.getText(node);
				if (text.endsWith("nan") || text.endsWith("inf")) return;
				verifyText(node, text);
			}
		} };
	}
});

//#endregion
//#region src/rules/padding-line-between-pairs.ts
var padding_line_between_pairs_default = createRule("padding-line-between-pairs", {
	meta: {
		docs: {
			description: "require or disallow padding lines between pairs",
			categories: ["standard"],
			extensionRule: false
		},
		fixable: "whitespace",
		schema: [],
		messages: {
			unexpectedBlankLine: "Unexpected blank line before this pair.",
			expectedBlankLine: "Expected blank line before this pair."
		},
		type: "layout"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		/**
		* Verify pairs
		*/
		function verifyPairs(prevNode, prevKeys, nextNode, nextKeys) {
			const needPadding = nextKeys.length !== prevKeys.length || nextKeys.slice(0, -1).some((key, index) => key !== prevKeys[index]);
			const tokens = sourceCode.getTokensBetween(prevNode, nextNode, { includeComments: true });
			if (needPadding) {
				let prevTarget = prevNode;
				for (const token of [...tokens, nextNode]) {
					if (prevTarget.loc.end.line + 1 < token.loc.start.line) return;
					prevTarget = token;
				}
				context.report({
					node: nextNode.key,
					messageId: "expectedBlankLine",
					fix(fixer) {
						return fixer.insertTextAfter(prevNode, "\n");
					}
				});
			} else {
				const prevTarget = [prevNode, ...tokens].pop();
				if (prevTarget.loc.end.line + 1 >= nextNode.loc.start.line) return;
				context.report({
					node: nextNode.key,
					messageId: "unexpectedBlankLine",
					*fix(fixer) {
						yield fixer.replaceTextRange([prevTarget.range[1], nextNode.range[0]], "\n");
					}
				});
			}
		}
		/**
		* Verify table
		*/
		function verify(node) {
			let prev = null;
			for (const body of node.body) {
				if (body.type !== "TOMLKeyValue") continue;
				const keys = getStaticTOMLValue(body.key);
				if (prev) verifyPairs(prev.node, prev.keys, body, keys);
				prev = {
					node: body,
					keys
				};
			}
		}
		return {
			TOMLTopLevelTable: verify,
			TOMLTable: verify
		};
	}
});

//#endregion
//#region src/rules/padding-line-between-tables.ts
var padding_line_between_tables_default = createRule("padding-line-between-tables", {
	meta: {
		docs: {
			description: "require or disallow padding lines between tables",
			categories: ["standard"],
			extensionRule: false
		},
		fixable: "whitespace",
		schema: [],
		messages: {
			unexpectedBlankLine: "Unexpected blank line before this table.",
			expectedBlankLine: "Expected blank line before this table."
		},
		type: "layout"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		/**
		* Verify tables
		*/
		function verifyTables(prevNode, nextNode) {
			const tokens = sourceCode.getTokensBetween(prevNode, nextNode, { includeComments: true });
			let prevTarget = prevNode;
			for (const token of [...tokens, nextNode]) {
				if (prevTarget.loc.end.line + 1 < token.loc.start.line) return;
				prevTarget = token;
			}
			context.report({
				node: nextNode.key,
				messageId: "expectedBlankLine",
				fix(fixer) {
					return fixer.insertTextAfter(prevNode, "\n");
				}
			});
		}
		/**
		* Verify top level table
		*/
		function verify(node) {
			let prev = null;
			for (const body of node.body) {
				if (prev && body.type === "TOMLTable") verifyTables(prev, body);
				prev = body;
			}
		}
		return { TOMLTopLevelTable: verify };
	}
});

//#endregion
//#region src/rules/precision-of-fractional-seconds.ts
var precision_of_fractional_seconds_default = createRule("precision-of-fractional-seconds", {
	meta: {
		docs: {
			description: "disallow precision of fractional seconds greater than the specified value.",
			categories: ["recommended", "standard"],
			extensionRule: false
		},
		schema: [{
			type: "object",
			properties: { max: {
				type: "number",
				minimum: 0
			} },
			additionalProperties: false
		}],
		messages: { over: "Precision of fractional seconds greater than {{max}} are forbidden." },
		type: "problem"
	},
	create(context) {
		if (!context.sourceCode.parserServices?.isTOML) return {};
		const max = context.options[0]?.max ?? 3;
		/**
		* Verify date time value node text
		*/
		function verifyText(node) {
			const text = node.datetime;
			const fractional = /^\d{4}-\d{2}-\d{2}[ t]\d{2}:\d{2}:\d{2}.(\d+)/iu.exec(text)?.[1] || /^\d{2}:\d{2}:\d{2}.(\d+)/u.exec(text)?.[1];
			if (!fractional) return;
			if (fractional.length > max) context.report({
				node,
				messageId: "over",
				data: { max }
			});
		}
		return { TOMLValue(node) {
			if (node.kind === "offset-date-time" || node.kind === "local-date-time" || node.kind === "local-time") verifyText(node);
		} };
	}
});

//#endregion
//#region src/utils/bit.ts
/**
* Convert the given maxBit to max hex, octal, binary and decimal number strings
*
* *Export for testing.
*/
function maxBitToMaxValues(maxBit) {
	const binaryMax = [];
	const minusMax = [0];
	const plusMax = [0];
	const hexMax = [0];
	const octalMax = [0];
	for (let index = 0; index < maxBit; index++) {
		const binaryNum = index === 0 ? 1 : 0;
		binaryMax.push(binaryNum);
		processDigits(minusMax, binaryNum, 10);
		processDigits(hexMax, binaryNum, 16);
		processDigits(octalMax, binaryNum, 8);
		if (index > 0) processDigits(plusMax, 1, 10);
	}
	return {
		"+": plusMax.reverse().join(""),
		"-": minusMax.reverse().join(""),
		"0x": hexMax.map((i) => i.toString(16)).reverse().join("").toLowerCase(),
		"0o": octalMax.reverse().join(""),
		"0b": binaryMax.join("")
	};
	/** Process digits */
	function processDigits(digits, binaryNum, radix) {
		let num = binaryNum;
		for (let place = 0; place < digits.length; place++) {
			num = digits[place] * 2 + num;
			digits[place] = num % radix;
			num = Math.floor(num / radix);
		}
		while (num > 0) {
			digits.push(num % radix);
			num = Math.floor(num / radix);
		}
	}
}

//#endregion
//#region src/rules/precision-of-integer.ts
const cacheMaxValues = {};
/**
* Get max values
*/
function getMaxValues(bit) {
	if (cacheMaxValues[bit]) return cacheMaxValues[bit];
	return cacheMaxValues[bit] = maxBitToMaxValues(bit);
}
var precision_of_integer_default = createRule("precision-of-integer", {
	meta: {
		docs: {
			description: "disallow precision of integer greater than the specified value.",
			categories: ["recommended", "standard"],
			extensionRule: false
		},
		schema: [{
			type: "object",
			properties: { maxBit: {
				type: "number",
				minimum: 1
			} },
			additionalProperties: false
		}],
		messages: { over: "Integers with precision greater than {{maxBit}}-bit are forbidden." },
		type: "problem"
	},
	create(context) {
		if (!context.sourceCode.parserServices?.isTOML) return {};
		const maxBit = context.options[0]?.maxBit ?? 64;
		const maxValues = getMaxValues(maxBit);
		/**
		* Verify number text
		*/
		function verifyMaxValue(node, numText, max) {
			const num = numText.replace(/^0+/, "").toLowerCase();
			if (num.length < max.length) return;
			if (num.length === max.length && num <= max) return;
			context.report({
				node,
				messageId: "over",
				data: { maxBit }
			});
		}
		/**
		* Verify integer value node text
		*/
		function verifyText(node) {
			const text = node.number;
			if (text.startsWith("0")) {
				const maybeMark = text[1];
				if (maybeMark === "x") {
					verifyMaxValue(node, text.slice(2), maxValues["0x"]);
					return;
				} else if (maybeMark === "o") {
					verifyMaxValue(node, text.slice(2), maxValues["0o"]);
					return;
				} else if (maybeMark === "b") {
					verifyMaxValue(node, text.slice(2), maxValues["0b"]);
					return;
				}
			} else if (text.startsWith("-")) {
				verifyMaxValue(node, text.slice(1), maxValues["-"]);
				return;
			} else if (text.startsWith("+")) {
				verifyMaxValue(node, text.slice(1), maxValues["+"]);
				return;
			}
			verifyMaxValue(node, text, maxValues["+"]);
		}
		return { TOMLValue(node) {
			if (node.kind === "integer") verifyText(node);
		} };
	}
});

//#endregion
//#region src/rules/quoted-keys.ts
var quoted_keys_default = createRule("quoted-keys", {
	meta: {
		docs: {
			description: "require or disallow quotes around keys",
			categories: ["standard"],
			extensionRule: false
		},
		fixable: "code",
		schema: [{
			type: "object",
			properties: {
				prefer: { enum: ["as-needed", "always"] },
				numbers: { type: "boolean" }
			},
			additionalProperties: false
		}],
		messages: {
			unnecessarilyQuotedKey: "Unnecessarily quoted key '{{key}}' found.",
			unquotedNumericKey: "Unquoted number '{{key}}' used as key.",
			unquotedKeyFound: "Unquoted key '{{key}}' found."
		},
		type: "layout"
	},
	create(context) {
		if (!context.sourceCode.parserServices?.isTOML) return {};
		const prefer = context.options[0]?.prefer ?? "as-needed";
		const numbers = context.options[0]?.numbers !== false;
		return {
			TOMLBare(node) {
				if (prefer === "always") {
					context.report({
						node,
						messageId: "unquotedKeyFound",
						data: { key: node.name },
						fix(fixer) {
							return fixer.replaceText(node, `"${node.name}"`);
						}
					});
					return;
				}
				if (numbers && /^[\d-]+$/u.test(node.name)) context.report({
					node,
					messageId: "unquotedNumericKey",
					data: { key: node.name },
					fix(fixer) {
						return fixer.replaceText(node, `"${node.name}"`);
					}
				});
			},
			TOMLQuoted(node) {
				if (prefer === "always") return;
				if (/^[\w-]+$/u.test(node.value)) {
					if (numbers && /^[\d-]+$/u.test(node.value)) return;
					context.report({
						node,
						messageId: "unnecessarilyQuotedKey",
						data: { key: node.value },
						fix(fixer) {
							return fixer.replaceText(node, node.value);
						}
					});
				}
			}
		};
	}
});

//#endregion
//#region src/rules/space-eq-sign.ts
var space_eq_sign_default = createRule("space-eq-sign", {
	meta: {
		docs: {
			description: "require spacing around equals sign",
			categories: ["standard"],
			extensionRule: false
		},
		deprecated: true,
		replacedBy: ["key-spacing"],
		fixable: "whitespace",
		schema: [],
		messages: { missingSpace: "Equals sign '=' must be spaced." },
		type: "layout"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		/**
		* Reports an equal sign token as a rule violation
		*/
		function report(equalToken) {
			context.report({
				loc: equalToken.loc,
				messageId: "missingSpace",
				*fix(fixer) {
					const previousToken = sourceCode.getTokenBefore(equalToken);
					const afterToken = sourceCode.getTokenAfter(equalToken);
					if (previousToken.range[1] === equalToken.range[0]) yield fixer.insertTextBefore(equalToken, " ");
					if (equalToken.range[1] === afterToken.range[0]) yield fixer.insertTextAfter(equalToken, " ");
				}
			});
		}
		return { TOMLKeyValue(node) {
			const equalToken = sourceCode.getTokenBefore(node.value);
			if (node.key.range[1] === equalToken.range[0] || equalToken.range[1] === node.value.range[0]) report(equalToken);
		} };
	}
});

//#endregion
//#region src/rules/spaced-comment.ts
/**
* Escapes the control characters of a given string.
* @param {string} s A string to escape.
* @returns {string} An escaped string.
*/
function escapeText(s) {
	return `(?:${s.replace(/[$()*+.?[\\\]^{|}]/g, "\\$&")})`;
}
/**
* Escapes the control characters of a given string.
* And adds a repeat flag.
* @param {string} s A string to escape.
* @returns {string} An escaped string.
*/
function escapeAndRepeat(s) {
	return `${escapeText(s)}+`;
}
/**
* Creates string pattern for exceptions.
* Generated pattern:
*
* 1. A space or an exception pattern sequence.
* @param {string[]} exceptions An exception pattern list.
* @returns {string} A regular expression string for exceptions.
*/
function createExceptionsPattern(exceptions) {
	let pattern = "";
	if (exceptions.length === 0) pattern += "\\s";
	else {
		pattern += "(?:\\s|";
		if (exceptions.length === 1) pattern += escapeAndRepeat(exceptions[0]);
		else {
			pattern += "(?:";
			pattern += exceptions.map(escapeAndRepeat).join("|");
			pattern += ")";
		}
		pattern += "$)";
	}
	return pattern;
}
/**
* Creates RegExp object for `always` mode.
* Generated pattern for beginning of comment:
*
* 1. First, a marker or nothing.
* 2. Next, a space or an exception pattern sequence.
* @param {string[]} markers A marker list.
* @param {string[]} exceptions An exception pattern list.
* @returns {RegExp} A RegExp object for the beginning of a comment in `always` mode.
*/
function createAlwaysStylePattern(markers, exceptions) {
	let pattern = "^";
	if (markers.length === 1) pattern += escapeText(markers[0]);
	else {
		pattern += "(?:";
		pattern += markers.map(escapeText).join("|");
		pattern += ")";
	}
	pattern += "?";
	pattern += createExceptionsPattern(exceptions);
	return new RegExp(pattern, "u");
}
/**
* Creates RegExp object for `never` mode.
* Generated pattern for beginning of comment:
*
* 1. First, a marker or nothing (captured).
* 2. Next, a space or a tab.
* @param {string[]} markers A marker list.
* @returns {RegExp} A RegExp object for `never` mode.
*/
function createNeverStylePattern(markers) {
	const pattern = `^(${markers.map(escapeText).join("|")})?[ \t]+`;
	return new RegExp(pattern, "u");
}
var spaced_comment_default = createRule("spaced-comment", {
	meta: {
		docs: {
			description: "enforce consistent spacing after the `#` in a comment",
			categories: ["standard"],
			extensionRule: "spaced-comment"
		},
		fixable: "whitespace",
		schema: [{ enum: ["always", "never"] }, {
			type: "object",
			properties: {
				exceptions: {
					type: "array",
					items: { type: "string" }
				},
				markers: {
					type: "array",
					items: { type: "string" }
				}
			},
			additionalProperties: false
		}],
		messages: {
			unexpectedSpaceAfterMarker: "Unexpected space after marker ({{refChar}}) in comment.",
			expectedExceptionAfter: "Expected exception block, space after '{{refChar}}' in comment.",
			unexpectedSpaceAfter: "Unexpected space after '{{refChar}}' in comment.",
			expectedSpaceAfter: "Expected space after '{{refChar}}' in comment."
		},
		type: "suggestion"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		const requireSpace = context.options[0] !== "never";
		const config = context.options[1] || {};
		const markers = config.markers || [];
		const exceptions = config.exceptions || [];
		const styleRules = {
			beginRegex: requireSpace ? createAlwaysStylePattern(markers, exceptions) : createNeverStylePattern(markers),
			hasExceptions: exceptions.length > 0,
			captureMarker: new RegExp(`^(${markers.map(escapeText).join("|")})`, "u"),
			markers: new Set(markers)
		};
		/**
		* Reports a beginning spacing error with an appropriate message.
		* @param {ASTNode} node A comment node to check.
		* @param {string} messageId An error message to report.
		* @param {Array} match An array of match results for markers.
		* @param {string} refChar Character used for reference in the error message.
		* @returns {void}
		*/
		function reportBegin(node, messageId, match, refChar) {
			context.report({
				node,
				fix(fixer) {
					const start = node.range[0];
					let end = start + 1;
					if (requireSpace) {
						if (match) end += match[0].length;
						return fixer.insertTextAfterRange([start, end], " ");
					}
					end += match[0].length;
					return fixer.replaceTextRange([start, end], `#${match?.[1] ? match[1] : ""}`);
				},
				messageId,
				data: { refChar }
			});
		}
		/**
		* Reports a given comment if it's invalid.
		* @param {ASTNode} node a comment node to check.
		* @returns {void}
		*/
		function checkCommentForSpace(node) {
			if (node.value.length === 0 || styleRules.markers.has(node.value)) return;
			const beginMatch = styleRules.beginRegex.exec(node.value);
			if (requireSpace) {
				if (!beginMatch) {
					const hasMarker = styleRules.captureMarker.exec(node.value);
					const marker = hasMarker ? `#${hasMarker[0]}` : "#";
					if (styleRules.hasExceptions) reportBegin(node, "expectedExceptionAfter", hasMarker, marker);
					else reportBegin(node, "expectedSpaceAfter", hasMarker, marker);
				}
			} else if (beginMatch) if (!beginMatch[1]) reportBegin(node, "unexpectedSpaceAfter", beginMatch, "#");
			else reportBegin(node, "unexpectedSpaceAfterMarker", beginMatch, beginMatch[1]);
		}
		return { Program() {
			sourceCode.getAllComments().forEach(checkCommentForSpace);
		} };
	}
});

//#endregion
//#region src/rules/table-bracket-spacing.ts
var table_bracket_spacing_default = createRule("table-bracket-spacing", {
	meta: {
		docs: {
			description: "enforce consistent spacing inside table brackets",
			categories: ["standard"],
			extensionRule: "array-bracket-spacing"
		},
		fixable: "whitespace",
		schema: [{ enum: ["always", "never"] }],
		messages: {
			unexpectedSpaceAfter: "There should be no space after '{{tokenValue}}'.",
			unexpectedSpaceBefore: "There should be no space before '{{tokenValue}}'.",
			missingSpaceAfter: "A space is required after '{{tokenValue}}'.",
			missingSpaceBefore: "A space is required before '{{tokenValue}}'."
		},
		type: "layout"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		const prefer = context.options[0] || "never";
		/**
		* Reports that there should be a space after the first token
		* @param {ASTNode} node The node to report in the event of an error.
		* @param {Token} token The token to use for the report.
		* @returns {void}
		*/
		function reportRequiredBeginningSpace(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "missingSpaceAfter",
				data: { tokenValue: token.value },
				fix(fixer) {
					return fixer.insertTextAfter(token, " ");
				}
			});
		}
		/**
		* Reports that there shouldn't be a space after the first token
		* @param {ASTNode} node The node to report in the event of an error.
		* @param {Token} token The token to use for the report.
		* @returns {void}
		*/
		function reportNoBeginningSpace(node, token) {
			const nextToken = sourceCode.getTokenAfter(token);
			context.report({
				node,
				loc: {
					start: token.loc.end,
					end: nextToken.loc.start
				},
				messageId: "unexpectedSpaceAfter",
				data: { tokenValue: token.value },
				fix(fixer) {
					return fixer.removeRange([token.range[1], nextToken.range[0]]);
				}
			});
		}
		/**
		* Reports that there should be a space before the last token
		* @param {ASTNode} node The node to report in the event of an error.
		* @param {Token} token The token to use for the report.
		* @returns {void}
		*/
		function reportRequiredEndingSpace(node, token) {
			context.report({
				node,
				loc: token.loc,
				messageId: "missingSpaceBefore",
				data: { tokenValue: token.value },
				fix(fixer) {
					return fixer.insertTextBefore(token, " ");
				}
			});
		}
		/**
		* Reports that there shouldn't be a space before the last token
		* @param {ASTNode} node The node to report in the event of an error.
		* @param {Token} token The token to use for the report.
		* @returns {void}
		*/
		function reportNoEndingSpace(node, token) {
			const previousToken = sourceCode.getTokenBefore(token);
			context.report({
				node,
				loc: {
					start: previousToken.loc.end,
					end: token.loc.start
				},
				messageId: "unexpectedSpaceBefore",
				data: { tokenValue: token.value },
				fix(fixer) {
					return fixer.removeRange([previousToken.range[1], token.range[0]]);
				}
			});
		}
		/**
		* Validates the spacing around table brackets
		* @param {ASTNode} node The node we're checking for spacing
		* @returns {void}
		*/
		function validateArraySpacing(node) {
			const key = node.key;
			const first = sourceCode.getTokenBefore(key);
			const last = sourceCode.getTokenAfter(key);
			if (prefer === "always" && first.range[1] === key.range[0]) reportRequiredBeginningSpace(node, first);
			if (prefer === "never" && first.range[1] < key.range[0]) reportNoBeginningSpace(node, first);
			if (prefer === "always" && key.range[1] === last.range[0]) reportRequiredEndingSpace(node, last);
			if (prefer === "never" && key.range[1] < last.range[0]) reportNoEndingSpace(node, last);
		}
		return { TOMLTable: validateArraySpacing };
	}
});

//#endregion
//#region src/rules/tables-order.ts
/**
* Get first key node
*/
function getFirst(keys) {
	const first = keys[0];
	if (first) return getFirst(first.keys) || first.node;
	return null;
}
/**
* Get last key node
*/
function getLast(keys) {
	const last = keys[keys.length - 1];
	if (last) return getLast(last.keys) || last.node;
	return null;
}
var tables_order_default = createRule("tables-order", {
	meta: {
		docs: {
			description: "disallow defining tables out-of-order",
			categories: ["standard"],
			extensionRule: false
		},
		fixable: "code",
		schema: [],
		messages: {
			outOfOrder: "'{{target}}' must be next to '{{before}}'.",
			outOfOrderToBefore: "'{{target}}' must be previous to '{{after}}'."
		},
		type: "suggestion"
	},
	create(context) {
		const sourceCode = context.sourceCode;
		if (!sourceCode.parserServices?.isTOML) return {};
		/**
		* Apply new key
		*/
		function applyKey(rootKeys, node) {
			const keyNames = [...node.resolvedKey];
			let before = null;
			let keys = rootKeys;
			while (keyNames.length) {
				const key = keyNames.shift();
				const isLast = !keyNames.length;
				let next = keys.find((e) => e.key === key);
				if (!next) {
					next = {
						key,
						node,
						keys: []
					};
					if (isLast) before = getLast(keys);
					keys.push(next);
				} else if (isLast) {
					if (next.keys.length > 0) return {
						before: null,
						after: getFirst(next.keys)
					};
					/* istanbul ignore next */
					before = getLast(keys);
				}
				keys = next.keys;
			}
			return {
				before,
				after: null
			};
		}
		/**
		* Verify tables
		*/
		function verify(node) {
			const keys = [];
			let prev = null;
			for (const body of node.body) {
				if (body.type !== "TOMLTable") continue;
				const { before, after } = applyKey(keys, body);
				if (after) context.report({
					node: body.key,
					messageId: "outOfOrderToBefore",
					data: {
						target: getStaticTOMLValue(body.key).join("."),
						after: getStaticTOMLValue(after.key).join(".")
					},
					fix(fixer) {
						const startToken = sourceCode.getTokenBefore(body);
						const code = sourceCode.text.slice(startToken.range[1], body.range[1]);
						return [fixer.insertTextBefore(after, `${code.trim()}\n`), fixer.removeRange([startToken.range[1], body.range[1]])];
					}
				});
				else if (before && before !== prev) context.report({
					node: body.key,
					messageId: "outOfOrder",
					data: {
						target: getStaticTOMLValue(body.key).join("."),
						before: getStaticTOMLValue(before.key).join(".")
					},
					fix(fixer) {
						const startToken = sourceCode.getTokenBefore(body);
						const code = sourceCode.text.slice(startToken.range[1], body.range[1]);
						return [fixer.insertTextAfter(before, `\n${code.trim()}`), fixer.removeRange([startToken.range[1], body.range[1]])];
					}
				});
				prev = body;
			}
		}
		return { TOMLTopLevelTable: verify };
	}
});

//#endregion
//#region src/rules/vue-custom-block/no-parsing-error.ts
var no_parsing_error_default = createRule("vue-custom-block/no-parsing-error", {
	meta: {
		docs: {
			description: "disallow parsing errors in Vue custom blocks",
			categories: ["recommended", "standard"],
			extensionRule: false
		},
		schema: [],
		messages: {},
		type: "problem"
	},
	create(context, { customBlock }) {
		if (!customBlock) return {};
		const parseError = context.sourceCode.parserServices?.parseError;
		if (parseError && typeof parseError === "object") {
			let loc = void 0;
			if ("column" in parseError && "lineNumber" in parseError) loc = {
				line: parseError.lineNumber,
				column: parseError.column
			};
			return { Program(node) {
				context.report({
					node,
					loc,
					message: parseError.message
				});
			} };
		}
		return {};
	}
});

//#endregion
//#region src/utils/rules.ts
const rules$1 = [
	array_bracket_newline_default,
	array_bracket_spacing_default,
	array_element_newline_default,
	comma_style_default,
	indent_default,
	inline_table_curly_newline_default,
	inline_table_curly_spacing_default,
	inline_table_key_value_newline_default,
	key_spacing_default,
	keys_order_default,
	no_mixed_type_in_array_default,
	no_non_decimal_integer_default,
	no_space_dots_default,
	no_unreadable_number_separator_default,
	padding_line_between_pairs_default,
	padding_line_between_tables_default,
	precision_of_fractional_seconds_default,
	precision_of_integer_default,
	quoted_keys_default,
	space_eq_sign_default,
	spaced_comment_default,
	table_bracket_spacing_default,
	tables_order_default,
	no_parsing_error_default
];

//#endregion
//#region src/configs/flat/base.ts
var base_default = [{ plugins: { get toml() {
	return src_default;
} } }, {
	files: ["*.toml", "**/*.toml"],
	language: "toml/toml",
	rules: {
		"no-irregular-whitespace": "off",
		"spaced-comment": "off"
	}
}];

//#endregion
//#region src/configs/flat/recommended.ts
var recommended_default = [...base_default, { rules: {
	"toml/no-unreadable-number-separator": "error",
	"toml/precision-of-fractional-seconds": "error",
	"toml/precision-of-integer": "error",
	"toml/vue-custom-block/no-parsing-error": "error"
} }];

//#endregion
//#region src/configs/flat/standard.ts
var standard_default = [...base_default, { rules: {
	"toml/array-bracket-newline": "error",
	"toml/array-bracket-spacing": "error",
	"toml/array-element-newline": "error",
	"toml/comma-style": "error",
	"toml/indent": "error",
	"toml/inline-table-curly-newline": "error",
	"toml/inline-table-curly-spacing": "error",
	"toml/inline-table-key-value-newline": "error",
	"toml/key-spacing": "error",
	"toml/keys-order": "error",
	"toml/no-space-dots": "error",
	"toml/no-unreadable-number-separator": "error",
	"toml/padding-line-between-pairs": "error",
	"toml/padding-line-between-tables": "error",
	"toml/precision-of-fractional-seconds": "error",
	"toml/precision-of-integer": "error",
	"toml/quoted-keys": "error",
	"toml/spaced-comment": "error",
	"toml/table-bracket-spacing": "error",
	"toml/tables-order": "error",
	"toml/vue-custom-block/no-parsing-error": "error"
} }];

//#endregion
//#region package.json
var name$1 = "eslint-plugin-toml";
var version$1 = "1.3.1";

//#endregion
//#region src/meta.ts
var meta_exports = /* @__PURE__ */ __exportAll({
	name: () => name,
	version: () => version
});
const name = name$1;
const version = version$1;

//#endregion
//#region src/language/toml-source-code.ts
/**
* @fileoverview The TOMLSourceCode class.
*/
const commentParser = new ConfigCommentParser();
/**
* Pattern to match ESLint inline configuration comments in TOML.
* Matches: eslint, eslint-disable, eslint-enable, eslint-disable-line, eslint-disable-next-line
*/
const INLINE_CONFIG = /^\s*eslint(?:-enable|-disable(?:(?:-next)?-line)?)?(?:\s|$)/u;
/**
* TOML Source Code Object
*/
var TOMLSourceCode = class extends TextSourceCodeBase {
	#steps = null;
	#cacheTokensAndComments = null;
	#inlineConfigComments = null;
	/**
	* Creates a new instance.
	*/
	constructor(config) {
		super({
			ast: config.ast,
			text: config.text
		});
		this.hasBOM = Boolean(config.hasBOM);
		this.parserServices = config.parserServices;
		this.visitorKeys = config.visitorKeys || {};
		this.tokenStore = new TokenStore({
			tokens: [...config.ast.tokens, ...config.ast.comments],
			isComment: (token) => token.type === "Block"
		});
	}
	traverse() {
		if (this.#steps != null) return this.#steps;
		const steps = [];
		this.#steps = steps;
		const root = this.ast;
		steps.push(new CallMethodStep({
			target: "onCodePathStart",
			args: [{}, root]
		}));
		traverseNodes(root, {
			enterNode(n) {
				steps.push(new VisitNodeStep({
					target: n,
					phase: 1,
					args: [n]
				}));
			},
			leaveNode(n) {
				steps.push(new VisitNodeStep({
					target: n,
					phase: 2,
					args: [n]
				}));
			}
		});
		steps.push(new CallMethodStep({
			target: "onCodePathEnd",
			args: [{}, root]
		}));
		return steps;
	}
	/**
	* Gets all tokens and comments.
	*/
	get tokensAndComments() {
		return this.#cacheTokensAndComments ??= [...this.ast.tokens, ...this.ast.comments].sort((a, b) => a.range[0] - b.range[0]);
	}
	getLines() {
		return this.lines;
	}
	getAllComments() {
		return this.tokenStore.getAllComments();
	}
	/**
	* Returns an array of all inline configuration nodes found in the source code.
	* This includes eslint-disable, eslint-enable, eslint-disable-line,
	* eslint-disable-next-line, and eslint (for inline config) comments.
	*/
	getInlineConfigNodes() {
		if (!this.#inlineConfigComments) this.#inlineConfigComments = this.ast.comments.filter((comment) => INLINE_CONFIG.test(comment.value));
		return this.#inlineConfigComments;
	}
	/**
	* Returns directives that enable or disable rules along with any problems
	* encountered while parsing the directives.
	*/
	getDisableDirectives() {
		const problems = [];
		const directives = [];
		this.getInlineConfigNodes().forEach((comment) => {
			const directive = commentParser.parseDirective(comment.value);
			if (!directive) return;
			const { label, value, justification } = directive;
			if (label === "eslint-disable-line" && comment.loc.start.line !== comment.loc.end.line) {
				const message = `${label} comment should not span multiple lines.`;
				problems.push({
					ruleId: null,
					message,
					loc: comment.loc
				});
				return;
			}
			switch (label) {
				case "eslint-disable":
				case "eslint-enable":
				case "eslint-disable-next-line":
				case "eslint-disable-line": {
					const directiveType = label.slice(7);
					directives.push(new Directive({
						type: directiveType,
						node: comment,
						value,
						justification
					}));
					break;
				}
			}
		});
		return {
			problems,
			directives
		};
	}
	/**
	* Returns inline rule configurations along with any problems
	* encountered while parsing the configurations.
	*/
	applyInlineConfig() {
		const problems = [];
		const configs = [];
		this.getInlineConfigNodes().forEach((comment) => {
			const directive = commentParser.parseDirective(comment.value);
			if (!directive) return;
			const { label, value } = directive;
			if (label === "eslint") {
				const parseResult = commentParser.parseJSONLikeConfig(value);
				if (parseResult.ok) configs.push({
					config: { rules: parseResult.config },
					loc: comment.loc
				});
				else problems.push({
					ruleId: null,
					message: parseResult.error.message,
					loc: comment.loc
				});
			}
		});
		return {
			configs,
			problems
		};
	}
	getNodeByRangeIndex(index) {
		let node = find([this.ast]);
		if (!node) return null;
		while (true) {
			const child = find(this._getChildren(node));
			if (!child) return node;
			node = child;
		}
		/**
		* Finds a node that contains the given index.
		*/
		function find(nodes) {
			for (const node of nodes) if (node.range[0] <= index && index < node.range[1]) return node;
			return null;
		}
	}
	getFirstToken(node, options) {
		return this.tokenStore.getFirstToken(node, options);
	}
	getFirstTokens(node, options) {
		return this.tokenStore.getFirstTokens(node, options);
	}
	getLastToken(node, options) {
		return this.tokenStore.getLastToken(node, options);
	}
	getLastTokens(node, options) {
		return this.tokenStore.getLastTokens(node, options);
	}
	getTokenBefore(node, options) {
		return this.tokenStore.getTokenBefore(node, options);
	}
	getTokensBefore(node, options) {
		return this.tokenStore.getTokensBefore(node, options);
	}
	getTokenAfter(node, options) {
		return this.tokenStore.getTokenAfter(node, options);
	}
	getTokensAfter(node, options) {
		return this.tokenStore.getTokensAfter(node, options);
	}
	getFirstTokenBetween(left, right, options) {
		return this.tokenStore.getFirstTokenBetween(left, right, options);
	}
	getFirstTokensBetween(left, right, options) {
		return this.tokenStore.getFirstTokensBetween(left, right, options);
	}
	getLastTokenBetween(left, right, options) {
		return this.tokenStore.getLastTokenBetween(left, right, options);
	}
	getLastTokensBetween(left, right, options) {
		return this.tokenStore.getLastTokensBetween(left, right, options);
	}
	getTokens(node, options) {
		return this.tokenStore.getTokens(node, options);
	}
	getTokensBetween(left, right, options) {
		return this.tokenStore.getTokensBetween(left, right, options);
	}
	getCommentsInside(nodeOrToken) {
		return this.tokenStore.getCommentsInside(nodeOrToken);
	}
	getCommentsBefore(nodeOrToken) {
		return this.tokenStore.getCommentsBefore(nodeOrToken);
	}
	getCommentsAfter(nodeOrToken) {
		return this.tokenStore.getCommentsAfter(nodeOrToken);
	}
	commentsExistBetween(first, second) {
		return this.tokenStore.commentsExistBetween(first, second);
	}
	isSpaceBetween(first, second) {
		const [left, right] = first.range[1] <= second.range[0] ? [first, second] : [second, first];
		return this.tokenStore.isSpaceBetween(left, right);
	}
	/**
	* Compatibility for ESLint's SourceCode API
	* @deprecated TOML does not have scopes
	*/
	getScope(node) {
		if (node?.type !== "Program") return null;
		return createFakeGlobalScope(this.ast);
	}
	/**
	* Compatibility for ESLint's SourceCode API
	* @deprecated TOML does not have scopes
	*/
	get scopeManager() {
		return {
			scopes: [],
			globalScope: createFakeGlobalScope(this.ast),
			acquire: (node) => {
				if (node.type === "Program") return createFakeGlobalScope(this.ast);
				return null;
			},
			getDeclaredVariables: () => [],
			addGlobals: () => {}
		};
	}
	/**
	* Compatibility for ESLint's SourceCode API
	* @deprecated
	*/
	isSpaceBetweenTokens(first, second) {
		return this.isSpaceBetween(first, second);
	}
	_getChildren(node) {
		const keys = this.visitorKeys[node.type] || [];
		const children = [];
		for (const key of keys) {
			const value = node[key];
			if (Array.isArray(value)) {
				for (const element of value) if (isNode(element)) children.push(element);
			} else if (isNode(value)) children.push(value);
		}
		return children;
	}
};
/**
* Determines whether the given value is a TOML AST node.
*/
function isNode(value) {
	return typeof value === "object" && value !== null && typeof value.type === "string" && Array.isArray(value.range) && Boolean(value.loc) && typeof value.loc === "object";
}
/**
* Creates a fake global scope for TOML files.
* @deprecated TOML does not have scopes
*/
function createFakeGlobalScope(node) {
	const fakeGlobalScope = {
		type: "global",
		block: node,
		set: /* @__PURE__ */ new Map(),
		through: [],
		childScopes: [],
		variableScope: null,
		variables: [],
		references: [],
		functionExpressionScope: false,
		isStrict: false,
		upper: null,
		implicit: {
			variables: [],
			set: /* @__PURE__ */ new Map()
		}
	};
	fakeGlobalScope.variableScope = fakeGlobalScope;
	return fakeGlobalScope;
}

//#endregion
//#region src/language/toml-language.ts
/**
* The TOML language implementation for ESLint.
*/
var TOMLLanguage = class {
	constructor() {
		this.fileType = "text";
		this.lineStart = 1;
		this.columnStart = 0;
		this.nodeTypeKey = "type";
	}
	/**
	* Validates the language options.
	*/
	validateLanguageOptions(_languageOptions) {}
	normalizeLanguageOptions(languageOptions) {
		return {
			ecmaVersion: "latest",
			...languageOptions,
			parserOptions: { ...languageOptions.parserOptions }
		};
	}
	/**
	* Parses the given file into an AST.
	*/
	parse(file, context) {
		const text = file.body;
		try {
			return {
				ok: true,
				ast: parseTOML(text, {
					filePath: file.path,
					tomlVersion: context.languageOptions?.parserOptions?.tomlVersion
				})
			};
		} catch (error) {
			const message = error instanceof Error ? error.message : String(error);
			const parseError = error;
			return {
				ok: false,
				errors: [{
					message,
					line: parseError.lineNumber ?? 1,
					column: parseError.column ?? 1
				}]
			};
		}
	}
	/**
	* Creates a new SourceCode object for the given file and parse result.
	*/
	createSourceCode(file, parseResult) {
		return new TOMLSourceCode({
			text: file.body,
			ast: parseResult.ast,
			hasBOM: file.bom,
			parserServices: { isTOML: true },
			visitorKeys: VisitorKeys
		});
	}
};

//#endregion
//#region src/index.ts
const configs = {
	base: base_default,
	recommended: recommended_default,
	standard: standard_default,
	"flat/base": base_default,
	"flat/recommended": recommended_default,
	"flat/standard": standard_default
};
const rules = rules$1.reduce((obj, r) => {
	obj[r.meta.docs.ruleName] = r;
	return obj;
}, {});
const languages = { toml: new TOMLLanguage() };
var src_default = {
	meta: meta_exports,
	configs,
	rules,
	languages
};

//#endregion
export { configs, src_default as default, languages, meta_exports as meta, rules };