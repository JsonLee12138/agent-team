import type { ConfigurationDiagnostic, FormatRequest, Formatter, GlobalConfiguration } from "./common.js";
export type { ConfigurationDiagnostic, FileMatchingInfo, FormatRequest, Formatter, GlobalConfiguration, Host, PluginInfo, } from "./common.js";
/** A formatter returned from adding a plugin to a context. */
export interface ContextFormatter {
    /** Formats the specified file text using this plugin. */
    formatText(request: FormatRequest): string;
    /** Gets the resolved configuration for this plugin. */
    getResolvedConfig(): Record<string, unknown>;
    /** Gets the configuration diagnostics for this plugin. */
    getConfigDiagnostics(): ConfigurationDiagnostic[];
}
/** A context for managing multiple formatters with shared configuration. */
export interface FormatterContext {
    /**
     * Adds a plugin to the context.
     * @param source - The buffer or Wasm module of the plugin (e.g., from `@dprint/json` getBuffer()).
     * @param param - Plugin config.
     * @returns A formatter for directly formatting with this plugin.
     */
    addPlugin(source: BufferSource | WebAssembly.Module, pluginConfig?: Record<string, unknown>): ContextFormatter;
    /**
     * Adds a plugin to the context.
     * @param source - Source response object.
     * @param pluginConfig - Plugin config.
     * @returns A formatter for directly formatting with this plugin.
     */
    addPluginStreaming(source: ResponseLike, pluginConfig?: Record<string, unknown>): Promise<ContextFormatter>;
    /**
     * Formats the specified file text, automatically selecting the appropriate plugin.
     * @param request - Data to format.
     * @returns The formatted text.
     * @throws If no plugin matches the file or there is an error formatting.
     */
    formatText(request: FormatRequest): string;
    /**
     * Gets all configuration diagnostics from all plugins.
     */
    getConfigDiagnostics(): ConfigurationDiagnostic[];
}
/**
 * Creates a formatter context for managing multiple plugins with shared configuration.
 * @param globalConfig - Global configuration shared across all plugins.
 */
export declare function createContext(globalConfig?: GlobalConfiguration): FormatterContext;
export interface ResponseLike {
    status: number;
    arrayBuffer(): Promise<BufferSource>;
    text(): Promise<string>;
    headers: {
        get(name: string): string | null;
    };
}
/**
 * Creates a formatter from the specified streaming source.
 * @remarks This is the most efficient way to create a formatter.
 * @param response - The streaming source to create the formatter from.
 */
export declare function createStreaming(responsePromise: Promise<ResponseLike> | ResponseLike): Promise<Formatter>;
/**
 * Creates a formatter from the specified wasm module bytes.
 * @param wasmModuleBuffer - The buffer of the wasm module.
 */
export declare function createFromBuffer(wasmModuleBuffer: BufferSource): Formatter;
export declare function createFromWasmModule(wasmModule: WebAssembly.Module): Formatter;
//# sourceMappingURL=mod.d.ts.map